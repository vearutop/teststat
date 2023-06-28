package main

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
}

type flags struct {
	Slow        time.Duration
	HistBuckets int
	RaceDepth   int
	Slowest     int
	Store       string
	FailedTests string
	Progress    bool
	Markdown    bool
	SkipReport  bool
	Allure      string
	Version     bool
}

func main() {
	fl := flags{}

	flag.DurationVar(&fl.Slow, "slow", time.Second, "minimal duration of slow test")
	flag.IntVar(&fl.HistBuckets, "buckets", 10, "number of buckets for histogram")
	flag.IntVar(&fl.RaceDepth, "race-depth", 5, "stacktrace depth to group similar data races")
	flag.IntVar(&fl.Slowest, "slowest", 30, "limit number of slowest tests to list")
	flag.StringVar(&fl.Store, "store", "", "store received json lines to file, useful for STDIN")
	flag.StringVar(&fl.FailedTests, "failed-tests", "", "store regexp of failed tests to a file, useful for a retry run")
	flag.BoolVar(&fl.Progress, "progress", false, "show progress")
	flag.BoolVar(&fl.Markdown, "markdown", false, "render output as markdown")
	flag.BoolVar(&fl.SkipReport, "skip-report", false, "skip reporting, useful for multiple retries")
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
