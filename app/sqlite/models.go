package sqlite

import (
	"strconv"

	"github.com/vearutop/teststat/app/model"
)

const (
	packagesTable  = "packages"
	testsTable     = "tests"
	totalsTable    = "totals"
	revisionsTable = "revisions"
	runsTable      = "runs"
	outputsTable   = "outputs"
)

type Package struct {
	Hash    Hash   `db:"hash"`
	Package string `db:"package"`
}

type Test struct {
	Hash        Hash   `db:"hash"`
	PackageHash Hash   `db:"package_hash"`
	Test        string `db:"test"`
}

type Output struct {
	TestHash Hash   `db:"test_hash"`
	RunHash  Hash   `db:"run_hash"`
	Output   string `db:"output"`
}

type Total struct {
	Hash          Hash    `db:"test_hash"`
	First         int     `db:"first_ums" description:"Unix timestamp in milliseconds"`
	Last          int     `db:"last_ums" description:"Unix timestamp in milliseconds"`
	LastFailed    int     `db:"last_failed_ums" description:"Unix timestamp in milliseconds"`
	Failed        int     `db:"failed"`
	Passed        int     `db:"passed"`
	Unfinished    int     `db:"unfinished"`
	Skipped       int     `db:"skipped"`
	OutputLines   int     `db:"output_lines"`
	DataRaces     int     `db:"data_races"`
	Pauses        int     `db:"pauses"`
	Runs          int     `db:"runs"`
	Cached        int     `db:"cached"`
	Elapsed       float64 `db:"elapsed"`
	FirstRevision Hash    `db:"first_rev"`
	LastRevision  Hash    `db:"last_rev"`
}

type Revision struct {
	Hash     Hash   `db:"hash"`
	Revision string `db:"revision"`
}

type Run struct {
	Hash         Hash `db:"hash"` // test_hash, revision_hash, started
	TestHash     Hash `db:"test_hash"`
	RevisionHash Hash `db:"rev_hash"`
	model.TestRun
}

func TestHash(pkg, fn string) Hash {
	return StringHash(pkg + "." + fn)
}

func NewRun(r model.TestRun) Run {
	return Run{
		Hash:         StringHash(r.Package + "/" + r.Fn + "/" + r.Revision + "/" + strconv.Itoa(r.Started)),
		TestHash:     TestHash(r.Package, r.Fn),
		RevisionHash: StringHash(r.Revision),
		TestRun:      r,
	}
}
