package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRelativeTo(t *testing.T) {
	require.Equal(t, "pkg/marker.txt", relativeTo("/repo", "/repo/pkg/marker.txt"))
	require.Equal(t, "../other/marker.txt", relativeTo("/repo/pkg", "/repo/other/marker.txt"))
}

func TestCachePkg_HasReason(t *testing.T) {
	require.False(t, cachePkg{Package: "pkg"}.hasReason(), "no trace captured at all")
	require.False(t, cachePkg{Package: "pkg", Reasons: []string{noReasonLogged}}.hasReason(), "trace captured but silent on this package")
	require.True(t, cachePkg{Package: "pkg", Reasons: []string{"miss: no prior cached result found"}}.hasReason())
}

func TestCategorizeTestcachePayload(t *testing.T) {
	tests := []struct {
		name       string
		payload    string
		wantReason string
		wantDetail string
	}{
		{
			name:       "local directory mode",
			payload:    "caching disabled in local directory mode",
			wantReason: "disabled: local directory mode",
		},
		{
			name:       "outside module root",
			payload:    "caching disabled for package outside of module root, GOPATH, or GOROOT",
			wantReason: "disabled: outside module root, GOPATH, or GOROOT",
		},
		{
			name:       "non-cacheable test argument",
			payload:    "caching disabled for test argument: -count=2",
			wantReason: "disabled: non-cacheable test argument",
			wantDetail: "-count=2",
		},
		{
			name:       "input list not found",
			payload:    "input list not found: cache entry not found: open /tmp/gocache/5e/5ec4...-a: no such file or directory",
			wantReason: "miss: no prior cached result found",
		},
		{
			name:       "test output not found maps to the same reason as input list not found",
			payload:    "test output not found: cache entry not found: open /tmp/gocache/5e/5ec4...-d: no such file or directory",
			wantReason: "miss: no prior cached result found",
		},
		{
			name:       "input list malformed",
			payload:    `input list malformed ("garbage")`,
			wantReason: "miss: input list malformed",
		},
		{
			name:       "test output malformed",
			payload:    "test output malformed",
			wantReason: "miss: test output malformed",
		},
		{
			name:       "test output expired",
			payload:    "test output expired due to go clean -testcache",
			wantReason: "miss: cache entry expired (go clean -testcache)",
		},
		{
			name:       "cover profile missing",
			payload:    "cached cover profile missing: cache entry not found: open /tmp/gocache/5e/5ec4...-a: no such file or directory",
			wantReason: "miss: cached cover profile missing",
		},
		{
			name:    "informational test ID line is not a finding",
			payload: "test ID abc123 => def456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason, detail, matched := categorizeTestcachePayload(tt.payload)
			if tt.wantReason == "" {
				require.False(t, matched)

				return
			}

			require.True(t, matched)
			require.Equal(t, tt.wantReason, reason)
			require.Equal(t, tt.wantDetail, detail)
		})
	}
}

// TestCategorizeTestcachePayload_FilePaths covers the two categories whose Detail is a file
// path: unlike the fixed-string cases above, the path go reports is always absolute and gets
// rewritten relative to the current working directory (see inputFilePath/relativeTo), so the
// expected value here is computed the same way rather than hardcoded.
func TestCategorizeTestcachePayload_FilePaths(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	abs := filepath.Join(cwd, "pkg", "marker.txt")

	reason, detail, matched := categorizeTestcachePayload("input file " + abs + ": file used as input is too new")
	require.True(t, matched)
	require.Equal(t, "miss: input file too new", reason)
	require.Equal(t, filepath.Join("pkg", "marker.txt"), detail)

	reason, detail, matched = categorizeTestcachePayload("input file " + abs + ": permission denied")
	require.True(t, matched)
	require.Equal(t, "miss: input file error", reason)
	require.Equal(t, filepath.Join("pkg", "marker.txt"), detail)
}

func TestParseTestcacheLine(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	abs := filepath.Join(cwd, "marker.txt")

	r, ok := parseTestcacheLine("testcache: example.com/pkg: input file " + abs + ": file used as input is too new")
	require.True(t, ok)
	require.Equal(t, testcacheReason{
		Package: "example.com/pkg",
		Reason:  "miss: input file too new",
		Detail:  "marker.txt",
	}, r)

	_, ok = parseTestcacheLine("not a testcache line")
	require.False(t, ok)

	_, ok = parseTestcacheLine("testcache: example.com/pkg: test ID abc => def")
	require.False(t, ok, "informational lines never produce a finding")
}

func TestParseTestcacheLine_ExtractsKeyForNotFoundCategories(t *testing.T) {
	r, ok := parseTestcacheLine("testcache: example.com/pkg: input list not found: cache entry not found: open /tmp/gocache-123/5e/5ec4abc-a: no such file or directory")
	require.True(t, ok)
	require.Equal(t, "5e/5ec4abc-a", r.Key)

	r, ok = parseTestcacheLine("testcache: example.com/pkg: test output not found: cache entry not found: open /tmp/gocache-123/5e/5ec4abc-d: no such file or directory")
	require.True(t, ok)
	require.Equal(t, "5e/5ec4abc-d", r.Key)

	r, ok = parseTestcacheLine("testcache: example.com/pkg: cached cover profile missing: cache entry not found: open /tmp/gocache-123/93/93d888-a: no such file or directory")
	require.True(t, ok)
	require.Equal(t, "93/93d888-a", r.Key)
}

func TestExtractCacheKey(t *testing.T) {
	key, ok := extractCacheKey("input list not found: cache entry not found: open /tmp/gocache-123/5e/5ec4abc-a: no such file or directory")
	require.True(t, ok)
	require.Equal(t, "5e/5ec4abc-a", key)

	key, ok = extractCacheKey("input list not found: cache entry not found: open C:\\gocache\\5e\\5ec4abc-a: no such file or directory")
	require.True(t, ok)
	require.Equal(t, "5e/5ec4abc-a", key, "backslash-separated (Windows) paths are normalized too")

	_, ok = extractCacheKey("input list malformed (\"garbage\")")
	require.False(t, ok, "no \": open \" marker at all")

	_, ok = extractCacheKey("caching disabled for test argument: -count=2")
	require.False(t, ok, "has a colon-space but no \"open\" path to extract")
}

func TestExtractLookupKey(t *testing.T) {
	key, ok := extractLookupKey("test ID abc123 => input ID def456 => 5ec4abc")
	require.True(t, ok)
	require.Equal(t, "5e/5ec4abc-a", key)

	_, ok = extractLookupKey("test ID abc123 => def456")
	require.False(t, ok, "no \"input ID\" segment -- the shorter, first informational line")

	_, ok = extractLookupKey("input list malformed")
	require.False(t, ok, "not a lookup-key line at all")
}

func TestParseTestcacheLookupKey(t *testing.T) {
	pkg, key, ok := parseTestcacheLookupKey("testcache: example.com/pkg: test ID abc123 => input ID def456 => 5ec4abc")
	require.True(t, ok)
	require.Equal(t, "example.com/pkg", pkg)
	require.Equal(t, "5e/5ec4abc-a", key)

	_, _, ok = parseTestcacheLookupKey("testcache: example.com/pkg: input list not found: cache entry not found: open /tmp/gocache/5e/5ec4abc-a: no such file or directory")
	require.False(t, ok, "a miss line, not a lookup-key line")

	_, _, ok = parseTestcacheLookupKey("not a testcache line")
	require.False(t, ok)
}

func TestCacheStats_DedupesSameReasonAcrossRounds(t *testing.T) {
	p := newProcessor(flags{})
	p.packageStats["example.com/pkg"] = packageStat{Package: "example.com/pkg", HadTests: true}
	// Same (Reason, Detail) recorded twice, as it would be from two rounds fed into one report.
	p.testcacheFindings["example.com/pkg"] = []testcacheReason{
		{Package: "example.com/pkg", Reason: "miss: no prior cached result found"},
		{Package: "example.com/pkg", Reason: "miss: input file too new", Detail: "/repo/marker.txt"},
		{Package: "example.com/pkg", Reason: "miss: no prior cached result found"},
		{Package: "example.com/pkg", Reason: "miss: input file too new", Detail: "/repo/marker.txt"},
	}

	stats := p.cacheStats()
	require.Equal(t, 1, stats.Total)
	require.Equal(t, 0, stats.Cached)
	require.Equal(t, 1, stats.Miss)
	require.Equal(t, 0, stats.Failing)
	require.Equal(t, 0, stats.NoTests)
	require.Len(t, stats.Misses, 1)
	require.Equal(t, []string{
		"miss: no prior cached result found",
		"miss: input file too new: /repo/marker.txt",
	}, stats.Misses[0].Reasons)
}

func TestBuildTestcacheKeys(t *testing.T) {
	p := newProcessor(flags{})

	// A miss with keys, deduped across repeated rounds.
	p.packageStats["example.com/withkey"] = packageStat{Package: "example.com/withkey", HadTests: true}
	p.testcacheFindings["example.com/withkey"] = []testcacheReason{
		{Package: "example.com/withkey", Reason: "miss: no prior cached result found", Key: "5e/5ec4abc-a"},
		{Package: "example.com/withkey", Reason: "miss: no prior cached result found", Key: "5e/5ec4abc-a"},
		{Package: "example.com/withkey", Reason: "miss: no prior cached result found", Key: "31/317aa-a"},
	}

	// A miss whose only reason carries no key at all (e.g. "input file too new" has its path
	// in Detail, not Key) -- expected to be omitted entirely, not included with an empty list.
	p.packageStats["example.com/nokey"] = packageStat{Package: "example.com/nokey", HadTests: true}
	p.testcacheFindings["example.com/nokey"] = []testcacheReason{
		{Package: "example.com/nokey", Reason: "miss: input file too new", Detail: "/repo/marker.txt"},
	}

	// A cached package with a captured lookup key -- its hit key, deduped across rounds.
	p.packageStats["example.com/cached"] = packageStat{Package: "example.com/cached", HadTests: true, Cached: true}
	p.testcacheLookupKeys["example.com/cached"] = []string{"ab/cachedkey-a", "ab/cachedkey-a"}

	// A cached package with no captured lookup key (e.g. GODEBUG=gocachetest=1 wasn't set) --
	// never included regardless of stray testcacheFindings.
	p.packageStats["example.com/cachednokey"] = packageStat{Package: "example.com/cachednokey", HadTests: true, Cached: true}

	// A miss whose payload carries no recoverable key at all (go's "bad checksum" case: the
	// small ActionID index entry was found, but its referenced output blob wasn't) -- falls back
	// to the lookup-key line's key instead of being omitted.
	p.packageStats["example.com/badchecksum"] = packageStat{Package: "example.com/badchecksum", HadTests: true}
	p.testcacheFindings["example.com/badchecksum"] = []testcacheReason{
		{Package: "example.com/badchecksum", Reason: "miss: no prior cached result found"},
	}
	p.testcacheLookupKeys["example.com/badchecksum"] = []string{"cd/lookupkey-a"}

	keys := p.buildTestcacheKeys()
	require.Equal(t, map[string]testcachePkgKeys{
		"example.com/withkey":     {Misses: []string{"31/317aa-a", "5e/5ec4abc-a"}},
		"example.com/cached":      {Hits: []string{"ab/cachedkey-a"}},
		"example.com/badchecksum": {Misses: []string{"cd/lookupkey-a"}},
	}, keys)
}

func TestBuildTestcacheKeys_NoneCapturedReturnsNil(t *testing.T) {
	p := newProcessor(flags{})
	p.packageStats["example.com/pkg"] = packageStat{Package: "example.com/pkg", HadTests: true, Cached: true}

	require.Nil(t, p.buildTestcacheKeys())
}
