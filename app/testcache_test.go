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

	total, cached, miss, failing, noTests, misses := p.cacheStats()
	require.Equal(t, 1, total)
	require.Equal(t, 0, cached)
	require.Equal(t, 1, miss)
	require.Equal(t, 0, failing)
	require.Equal(t, 0, noTests)
	require.Len(t, misses, 1)
	require.Equal(t, []string{
		"miss: no prior cached result found",
		"miss: input file too new: /repo/marker.txt",
	}, misses[0].Reasons)
}
