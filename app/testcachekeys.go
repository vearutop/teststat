package app

import (
	"encoding/json"
	"os"
	"sort"
)

// testcachePkgKeys is one package's captured GOCACHE object keys: Hits from packages go's own
// cache reported `(cached)` for, Misses from cacheStats' per-package findings. Both use the same
// "xx/hash-a" object key format, so gocacheprog can check either set against the remote/manifest
// without translation. A package with neither is omitted entirely rather than written empty.
type testcachePkgKeys struct {
	Hits   []string `json:"hits,omitempty"`
	Misses []string `json:"misses,omitempty"`
}

// buildTestcacheKeys collects, per package, the distinct GOCACHE object keys go's own
// GODEBUG=gocachetest=1 trace named -- both the key a cache *hit* actually matched (from the
// hit/miss-agnostic "test ID => input ID => key" line, see parseTestcacheLookupKey) and the
// key(s) a cache *miss* went looking for and didn't find (see extractCacheKey). Hits capture
// today's known-good, worth-prioritizing keys; misses feed the existing gocacheprog forensics
// (was it in the manifest, on the remote, too old). A package with no extractable key in either
// direction (e.g. GODEBUG=gocachetest=1 wasn't set) is omitted rather than included empty.
//
// A miss's key isn't always recoverable from its own payload: go's cache.GetBytes first looks up
// a small ActionID index entry, then separately reads the (potentially large) output blob it
// points to; if the index entry is found but the blob isn't, the error is a path-less "bad
// checksum" rather than "open <path>: no such file or directory", so extractCacheKey finds
// nothing even though this is arguably the more interesting case (something was restored, the
// content wasn't). The lookup-key line is printed unconditionally once go gets far enough to
// compute it -- before checking whether the output blob is actually there -- so it still has the
// exact key for exactly this case, and is folded in as a fallback below.
func (p *processor) buildTestcacheKeys() map[string]testcachePkgKeys {
	result := map[string]testcachePkgKeys{}

	for pkg, ps := range p.packageStats {
		if !ps.Cached {
			continue
		}

		if keys := dedupSortedKeys(p.testcacheLookupKeys[pkg]); len(keys) > 0 {
			e := result[pkg]
			e.Hits = keys
			result[pkg] = e
		}
	}

	for _, m := range p.cacheStats().Misses {
		var keys []string

		seen := map[string]bool{}

		for _, r := range p.testcacheFindings[m.Package] {
			if r.Key == "" || seen[r.Key] {
				continue
			}

			seen[r.Key] = true

			keys = append(keys, r.Key)
		}

		for _, k := range p.testcacheLookupKeys[m.Package] {
			if k == "" || seen[k] {
				continue
			}

			seen[k] = true

			keys = append(keys, k)
		}

		if len(keys) == 0 {
			continue
		}

		sort.Strings(keys)

		e := result[m.Package]
		e.Misses = keys
		result[m.Package] = e
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

// dedupSortedKeys returns keys with duplicates removed and sorted, or nil if empty.
func dedupSortedKeys(keys []string) []string {
	if len(keys) == 0 {
		return nil
	}

	seen := map[string]bool{}

	var out []string

	for _, k := range keys {
		if seen[k] {
			continue
		}

		seen[k] = true

		out = append(out, k)
	}

	sort.Strings(out)

	return out
}

// storeTestcacheKeysJSON writes -testcache-keys' output file: {package: {hits, misses}} for
// every package with at least one captured key. Kept as one file/flag (rather than a separate
// one for hits) since both are the same per-package key data, just split by outcome -- a second
// flag later would just be this same data cut a different way.
func (p *processor) storeTestcacheKeysJSON() {
	if p.fl.TestcacheKeys == "" {
		return
	}

	keys := p.buildTestcacheKeys()

	data, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		p.println("failed to marshal testcache keys json: " + err.Error())

		return
	}

	if err := os.WriteFile(p.fl.TestcacheKeys, data, 0o600); err != nil {
		p.println("failed to store testcache keys json: " + err.Error())
	}
}
