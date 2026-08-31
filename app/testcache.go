package app

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// cachePkg is one package's cache-miss verdict for the report.
type cachePkg struct {
	Package string
	Reasons []string // empty if the testcache trace wasn't captured, or captured but silent on why.
}

// noReasonLogged is the placeholder used when a testcache trace was captured (see
// processor.testcacheSeen) but produced no specific finding for this package -- as opposed to
// Reasons being empty outright, which means no trace was captured at all.
const noReasonLogged = "no testcache reason logged"

// hasReason reports whether this package's miss has an actual, actionable explanation attached --
// false for both an empty Reasons (no GODEBUG=gocachetest=1 trace captured) and the sole
// noReasonLogged placeholder (a trace was captured but didn't explain this particular miss).
func (c cachePkg) hasReason() bool {
	if len(c.Reasons) == 0 {
		return false
	}

	return len(c.Reasons) != 1 || c.Reasons[0] != noReasonLogged
}

// cacheStatsResult is cacheStats' breakdown, as named fields rather than a long positional
// return - some callers only need Misses, and a 6-value positional return forces those into a
// wall of blank identifiers to get there.
type cacheStatsResult struct {
	Total, Cached, Miss, Failing, NoTests int
	Misses                                []cachePkg
}

// cacheStats assesses this single run's own test-result cache health: how many packages showed
// `(cached)`, how many didn't (and why, if the GODEBUG=gocachetest=1 trace was captured), and how
// many were never a caching candidate at all (no test files, or -run/-bench/-fuzz matched
// nothing). Total counts every package go test touched, including those, so it always accounts
// for the whole run. A package that failed is excluded from the miss judgment: go test never
// caches a failing result, so it not being cached is expected, not a finding.
func (p *processor) cacheStats() cacheStatsResult {
	var stats cacheStatsResult

	for pkg, ps := range p.packageStats {
		stats.Total++

		switch {
		case !ps.HadTests:
			stats.NoTests++
		case ps.Failed:
			stats.Failing++
		case ps.Cached:
			stats.Cached++
		default:
			stats.Miss++

			seen := map[string]bool{}

			var reasons []string

			for _, r := range p.testcacheFindings[pkg] {
				line := r.Reason
				if r.Detail != "" {
					line += ": " + r.Detail
				}

				if seen[line] {
					continue // Same reason logged again in a later round (e.g. a retry) -- noise, not a second finding.
				}

				seen[line] = true

				reasons = append(reasons, line)
			}

			if len(reasons) == 0 && p.testcacheSeen {
				reasons = []string{noReasonLogged}
			}

			stats.Misses = append(stats.Misses, cachePkg{Package: pkg, Reasons: reasons})
		}
	}

	sort.Slice(stats.Misses, func(i, j int) bool { return stats.Misses[i].Package < stats.Misses[j].Package })

	return stats
}

// testcachePrefix marks a GODEBUG=gocachetest=1 trace line. These land on stderr, so they only
// reach teststat when the caller combines stdout+stderr into one stream (e.g. `go test ... |&
// teststat -`), the same way data race reports already do.
const testcachePrefix = "testcache: "

func isTestcacheLine(line string) bool {
	return strings.HasPrefix(line, testcachePrefix)
}

// testcacheReason is one categorized, actionable finding parsed from a single `testcache: pkg:
// ...` line. Informational lines (e.g. "test ID %x => %x", which just narrate a successful
// lookup, not a miss) are not findings and never produce one.
type testcacheReason struct {
	Package string
	Reason  string // short, stable category label, safe to group/count by
	Detail  string // extra actionable context beyond Reason (e.g. a file path); empty when Reason
	// already says everything worth saying -- go's own cache-implementation detail (a temp file
	// path under GOCACHE, an internal "cache entry not found" phrasing) is deliberately never
	// included here, since it's not something a reader could act on.
	Key string // GOCACHE-relative object key (e.g. "xx/hash-a") this lookup was for, when the
	// payload names one (the three "not found" categories below); empty otherwise. Machine-only:
	// unlike Detail, this is never shown in the human report (see categorizeTestcachePayload's
	// doc comment) -- it exists for -testcache-keys to hand to gocacheprog for remote/manifest
	// forensics on exactly the object Go went looking for.
}

// splitTestcacheLine splits a `testcache: pkg: payload` line into its package and payload,
// shared by parseTestcacheLine (findings) and parseTestcacheLookupKey (the hit/miss-agnostic
// lookup key).
func splitTestcacheLine(line string) (pkg, payload string, ok bool) {
	rest := strings.TrimPrefix(line, testcachePrefix)

	pkg, payload, ok = strings.Cut(rest, ": ")
	if !ok || pkg == "" || payload == "" {
		return "", "", false
	}

	return pkg, payload, true
}

// parseTestcacheLine parses one line of GODEBUG=gocachetest=1 output. See test.go in
// cmd/go/internal/test for the exact strings this matches against -- verified against Go's own
// source rather than assumed, since these messages aren't a documented, stable API and could in
// principle change between Go versions.
func parseTestcacheLine(line string) (testcacheReason, bool) {
	pkg, payload, ok := splitTestcacheLine(line)
	if !ok {
		return testcacheReason{}, false
	}

	reason, detail, matched := categorizeTestcachePayload(payload)
	if !matched {
		return testcacheReason{}, false
	}

	key, _ := extractCacheKey(payload)

	return testcacheReason{Package: pkg, Reason: reason, Detail: detail, Key: key}, true
}

// parseTestcacheLookupKey extracts the package and exact GOCACHE object key from a "test ID %x
// => input ID %x => %x" line. Unlike parseTestcacheLine's categories, which only fire on a miss,
// go prints this line on every lookup attempt that gets far enough to compute the key -- so it's
// the only place a cache *hit*'s key is ever observable, and it's deliberately kept out of
// categorizeTestcachePayload/testcacheReason: mixing it into the findings-per-package list would
// make every hit package look like it has a "reason" too, breaking hasReason's miss-only meaning.
func parseTestcacheLookupKey(line string) (pkg, key string, ok bool) {
	pkg, payload, ok := splitTestcacheLine(line)
	if !ok {
		return "", "", false
	}

	key, ok = extractLookupKey(payload)
	if !ok {
		return "", "", false
	}

	return pkg, key, true
}

// categorizeTestcachePayload maps a testcache line's payload (everything after "pkg: ") to a
// short, stable category plus (when there's something genuinely actionable beyond the category
// itself, such as which file) a short detail string. Order matters: more specific
// prefixes/suffixes are checked before more generic ones that would otherwise also match.
func categorizeTestcachePayload(payload string) (reason, detail string, matched bool) {
	switch {
	case strings.HasPrefix(payload, "caching disabled in local directory mode"):
		return "disabled: local directory mode", "", true
	case strings.HasPrefix(payload, "caching disabled for package outside of module root"):
		return "disabled: outside module root, GOPATH, or GOROOT", "", true
	case strings.HasPrefix(payload, "caching disabled for test argument:"):
		arg := strings.TrimSpace(strings.TrimPrefix(payload, "caching disabled for test argument:"))

		return "disabled: non-cacheable test argument", arg, true

	case strings.HasPrefix(payload, "input list not found:"), strings.HasPrefix(payload, "test output not found:"):
		// "No prior cache entry for this exact input/output key" -- the expected, uninteresting
		// state on a cold run. Only worth a second look if it keeps showing up on runs that are
		// supposed to be warm.
		return "miss: no prior cached result found", "", true
	case strings.HasPrefix(payload, "input list malformed"):
		return "miss: input list malformed", "", true
	case strings.HasPrefix(payload, "input file ") && strings.HasSuffix(payload, "file used as input is too new"):
		return "miss: input file too new", inputFilePath(payload), true
	case strings.HasPrefix(payload, "input file "):
		return "miss: input file error", inputFilePath(payload), true

	case strings.HasPrefix(payload, "test output malformed"):
		return "miss: test output malformed", "", true
	case strings.HasPrefix(payload, "test output expired"):
		return "miss: cache entry expired (go clean -testcache)", "", true

	case strings.HasPrefix(payload, "cached cover profile missing:"):
		return "miss: cached cover profile missing", "", true

	default:
		// Includes the informational "test ID %x => %x", "test ID %x => input ID %x => %x", and
		// "save test ID ..." lines -- these narrate a successful lookup, not a problem.
		return "", "", false
	}
}

// inputFilePath extracts the file path from an "input file <path>: <reason>" payload, dropping
// go's own trailing explanation text -- the category (see categorizeTestcachePayload's caller)
// already states the reason; the path is the only part worth adding. Go always reports this as
// an absolute path; relativeTo trims it back down to whatever's actually useful to read.
func inputFilePath(payload string) string {
	rest := strings.TrimPrefix(payload, "input file ")
	path, _, _ := strings.Cut(rest, ": ")

	cwd, err := os.Getwd()
	if err != nil {
		return path
	}

	return relativeTo(cwd, path)
}

// relativeTo rewrites path relative to base (typically the current working directory, i.e.
// wherever the `go test` that produced it was run from), falling back to path unchanged if
// that's not possible (e.g. different volumes on Windows).
func relativeTo(base, path string) string {
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return path
	}

	return rel
}

// extractCacheKey pulls the GOCACHE-relative object key out of a "... cache entry not found:
// open <gocache-root>/xx/hash-a: no such file or directory" payload (the three "not found"
// categories in categorizeTestcachePayload all share this exact go-internal wording). The key
// is the object's last two path segments (2-char shard directory + hash-suffixed filename),
// which is gocacheprog's own on-disk relative object path convention (see
// internal/gocache/store.go's objectPath in the gocacheprog repo) -- so it's directly usable to
// ask gocacheprog "do you have this" without any further translation.
func extractCacheKey(payload string) (string, bool) {
	_, rest, found := strings.Cut(payload, ": open ")
	if !found {
		return "", false
	}

	path, _, ok := strings.Cut(rest, ": ")
	if !ok {
		return "", false
	}

	// Split on either separator explicitly, rather than filepath.ToSlash (which only
	// normalizes the host's own separator -- go's own trace was produced by whatever OS
	// actually ran the test, not necessarily this one, e.g. when re-parsing a captured log).
	parts := strings.FieldsFunc(path, func(r rune) bool { return r == '/' || r == '\\' })
	if len(parts) < 2 {
		return "", false
	}

	return parts[len(parts)-2] + "/" + parts[len(parts)-1], true
}

// extractLookupKey pulls the GOCACHE object key out of a "test ID %x => input ID %x => %x"
// payload -- the third %x is exactly testAndInputKey(testID, testInputsID), the ActionID go
// checks the cache for, whether or not it's actually there. Reassembled into the same
// "xx/hash-a" shard convention as extractCacheKey so both feed the same downstream format.
func extractLookupKey(payload string) (string, bool) {
	parts := strings.Split(payload, " => ")
	if len(parts) != 3 || !strings.HasPrefix(parts[0], "test ID ") || !strings.HasPrefix(parts[1], "input ID ") {
		return "", false
	}

	hexHash := parts[2]
	if len(hexHash) < 2 {
		return "", false
	}

	return hexHash[:2] + "/" + hexHash + "-a", true
}
