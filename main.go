package main

import (
	"flag"
	"fmt"
	"log"
	"time"
)

// Line structure describes single event of `go test -json` report.
type Line struct {
	Time    string  `json:"Time,omitempty"`
	Action  string  `json:"Action,omitempty"`
	Package string  `json:"Package,omitempty"`
	Test    string  `json:"Test,omitempty"`
	Output  string  `json:"Output,omitempty"`
	Elapsed float64 `json:"Elapsed,omitempty"`
}

type flags struct {
	Slow        time.Duration
	HistBuckets int
	RaceDepth   int
	Slowest     int
	Markdown    bool
}

func main() {
	fl := flags{}

	flag.DurationVar(&fl.Slow, "slow", time.Second, "minimal duration of slow test")
	flag.IntVar(&fl.HistBuckets, "buckets", 10, "number of buckets for histogram")
	flag.IntVar(&fl.RaceDepth, "race-depth", 5, "stacktrace depth to group similar data races")
	flag.IntVar(&fl.Slowest, "slowest", 30, "limit number of slowest tests to list")
	flag.BoolVar(&fl.Markdown, "markdown", false, "render output as markdown")

	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: teststat [options] report.jsonl ...")
		fmt.Println("	Use `-` as file name to read from STDIN.")
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
