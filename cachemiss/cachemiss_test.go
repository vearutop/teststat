//go:build cachemiss

package cachemiss_test

import (
	"os"
	"testing"
	"time"
)

func TestThatMissesCache(t *testing.T) {
	// Writing then reading a repo-local file makes this test's own tracked inputs include a
	// file whose mtime is always "now" -- go test's test-result cache never trusts a tracked
	// input read within 2 seconds of the cache decision (modTimeCutoff in
	// cmd/go/internal/test/test.go), so this test can never show (cached), on any run.
	//
	// The file is deliberately NOT removed afterward: go's cache bookkeeping (computeTestInputsID
	// in cmd/go/internal/test/test.go) re-stats every tracked input only after the test function
	// returns, not live as it's read. Cleaning up (e.g. a deferred os.Remove) before then makes
	// that re-stat see "file doesn't exist" instead of "file too new" -- a different, and
	// perfectly cacheable (since it's the same deterministic error both times), code path that
	// silently defeats the whole point of this test.
	//
	// Kept in its own scenario, run directly rather than through run-imperfect.sh's retry loop:
	// a package whose only test never fails still gets invoked on every retry round with a -run
	// pattern that matches nothing in it, and those "no tests to run" invocations can collide
	// with each other's own (unrelated, input-free, trivially cacheable) cache entries, printing
	// a spurious "(cached)" that permanently masks this test's real, always-reproducible miss in
	// teststat's aggregated report (see cacheStats/pkgLine: a package's Cached flag is an OR
	// across every fed-in round, never reset).
	const marker = "marker.txt"

	if err := os.WriteFile(marker, []byte(time.Now().String()), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := os.ReadFile(marker); err != nil {
		t.Fatal(err)
	}
}
