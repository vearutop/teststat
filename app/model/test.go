package model

import "time"

type Test struct {
	Pkg, Fn string
}

type Result string

const (
	Passed     = Result("P")
	Failed     = Result("F")
	Skipped    = Result("S")
	DataRace   = Result("D")
	Unfinished = Result("U")
)

type TestRun struct {
	Package  string `db:"-"`
	Fn       string `db:"-"`
	Revision string `db:"-"`

	Started     time.Time `db:"started"`
	Result      Result    `db:"result,omitempty"`
	OutputLines int       `db:"output_lines"`
	Pauses      int       `db:"pauses"`
	Cached      bool      `db:"cached"`
	Elapsed     float64   `db:"elapsed"`
}
