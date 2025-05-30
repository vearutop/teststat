package app

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/bool64/dev/version"
)

// Line structure describes single event of `go test -json` report.
type Line struct {
	Time    time.Time `json:"Time,omitempty"`
	Action  string    `json:"Action,omitempty"`
	Package string    `json:"Package,omitempty"`
	Test    string    `json:"Test,omitempty"`
	Output  string    `json:"Output,omitempty"`
	Elapsed *float64  `json:"Elapsed,omitempty"`

	// Added in go1.24
	ImportPath  string `json:"ImportPath,omitempty"`
	FailedBuild string `json:"FailedBuild,omitempty"`
}

type flags struct {
	Slow         time.Duration
	HistBuckets  int
	RaceDepth    int
	Slowest      int
	Store        string
	FailedTests  string
	FailureStats string
	SkipParent   bool
	FailedBuilds string
	Progress     bool
	Verbosity    int
	Markdown     bool
	SkipReport   bool
	LimitReport  int
	Allure       string
	Version      bool
}

// Main is an app entry point.
func Main() {
	fl := flags{}

	flag.DurationVar(&fl.Slow, "slow", time.Second, "minimal duration of slow test")
	flag.IntVar(&fl.HistBuckets, "buckets", 10, "number of buckets for histogram")
	flag.IntVar(&fl.RaceDepth, "race-depth", 5, "stacktrace depth to group similar data races")
	flag.IntVar(&fl.Slowest, "slowest", 30, "limit number of slowest tests to list")
	flag.StringVar(&fl.Store, "store", "", "store received json lines to file, useful for STDIN")
	flag.StringVar(&fl.FailedTests, "failed-tests", "", "store regexp of failed tests to a file, useful for a retry run")
	flag.BoolVar(&fl.SkipParent, "skip-parent", false, "exclude parent tests of subtests in regexp of failed tests, this may help to avoid running full suite on single failure")
	flag.StringVar(&fl.FailedBuilds, "failed-builds", "", "store build failures to a file")
	flag.StringVar(&fl.FailureStats, "failure-stats", "", "store failure stats (total) to a file")
	flag.BoolVar(&fl.Progress, "progress", false, "show progress")
	flag.IntVar(&fl.Verbosity, "verbosity", 0, "output verbosity, 0 for no output, 1 for failed test names, 2 for failure message")
	flag.BoolVar(&fl.Markdown, "markdown", false, "render output as markdown")
	flag.BoolVar(&fl.SkipReport, "skip-report", false, "skip reporting, useful for multiple retries")
	flag.IntVar(&fl.LimitReport, "limit-report", 60000, "maximum report length, exceeding part is truncated")
	flag.StringVar(&fl.Allure, "allure", "", "path to write allure report")

	flag.BoolVar(&fl.Version, "version", false, "show version and exit")

	flag.Parse()

	if fl.Version {
		fmt.Println(version.Info().Version)

		return
	}

	if flag.NArg() < 1 {
		fmt.Println("Usage: teststat [options] report.jsonl ...")
		fmt.Println("	Use `-` or `/dev/stdin` as file name to read from STDIN.")
		flag.PrintDefaults()

		return
	}

	p := newProcessor(fl)

	for _, f := range flag.Args() {
		if err := p.process(f); err != nil {
			log.Fatalf("%s: %s", f, err)
		}
	}

	p.report()
}
