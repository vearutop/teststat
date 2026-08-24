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

// cacheStats assesses this single run's own test-result cache health: how many packages showed
// `(cached)`, how many didn't (and why, if the GODEBUG=gocachetest=1 trace was captured), and how
// many were never a caching candidate at all (no test files, or -run/-bench/-fuzz matched
// nothing). total counts every package go test touched, including those, so it always accounts
// for the whole run. A package that failed is excluded from the miss judgment: go test never
// caches a failing result, so it not being cached is expected, not a finding.
func (p *processor) cacheStats() (total, cached, miss, failing, noTests int, misses []cachePkg) {
	for pkg, ps := range p.packageStats {
		total++

		switch {
		case !ps.HadTests:
			noTests++
		case ps.Failed:
			failing++
		case ps.Cached:
			cached++
		default:
			miss++

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

			misses = append(misses, cachePkg{Package: pkg, Reasons: reasons})
		}
	}

	sort.Slice(misses, func(i, j int) bool { return misses[i].Package < misses[j].Package })

	return total, cached, miss, failing, noTests, misses
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
}

// parseTestcacheLine parses one line of GODEBUG=gocachetest=1 output. See test.go in
// cmd/go/internal/test for the exact strings this matches against -- verified against Go's own
// source rather than assumed, since these messages aren't a documented, stable API and could in
// principle change between Go versions.
func parseTestcacheLine(line string) (testcacheReason, bool) {
	rest := strings.TrimPrefix(line, testcachePrefix)

	pkg, payload, ok := strings.Cut(rest, ": ")
	if !ok || pkg == "" || payload == "" {
		return testcacheReason{}, false
	}

	reason, detail, matched := categorizeTestcachePayload(payload)
	if !matched {
		return testcacheReason{}, false
	}

	return testcacheReason{Package: pkg, Reason: reason, Detail: detail}, true
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
