package sqlite

import (
	"strconv"
	"time"

	"github.com/vearutop/teststat/app/model"
)

const (
	testsTable     = "tests"
	totalsTable    = "totals"
	revisionsTable = "revisions"
	runsTable      = "runs"
)

type Test struct {
	Hash    Hash   `db:"hash"`
	Package string `db:"package"`
	Test    string `db:"test"`
}

type Total struct {
	Hash          Hash      `db:"hash"`
	First         time.Time `db:"first"`
	Last          time.Time `db:"last"`
	Failed        int       `db:"failed"`
	Passed        int       `db:"passed"`
	Unfinished    int       `db:"unfinished"`
	Skipped       int       `db:"skipped"`
	OutputLines   int       `db:"output_lines"`
	DataRaces     int       `db:"data_races"`
	Pauses        int       `db:"pauses"`
	Runs          int       `db:"runs"`
	Cached        int       `db:"cached"`
	Elapsed       float64   `db:"elapsed"`
	FirstRevision Hash      `db:"first_rev"`
	LastRevision  Hash      `db:"last_rev"`
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
		Hash:         StringHash(r.Package + "/" + r.Fn + "/" + r.Revision + "/" + strconv.Itoa(int(r.Started.UnixNano()))),
		TestHash:     TestHash(r.Package, r.Fn),
		RevisionHash: StringHash(r.Revision),
		TestRun:      r,
	}
}
