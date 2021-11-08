package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/vearutop/dynhist-go"
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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: teststat report.jsonl")

		return
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	counts := map[string]int{}
	elapsed := time.Duration(0)
	elapsedSlow := time.Duration(0)
	hist := dynhist.Collector{
		BucketsLimit: 10,
		WeightFunc:   dynhist.ExpWidth(1.2, 0.9),
	}

	var (
		slowest []Line
		dec     = json.NewDecoder(f)
	)

	for {
		var l Line
		if err := dec.Decode(&l); err != nil {
			break
		}

		counts[l.Action]++
		dur := time.Duration(l.Elapsed * float64(time.Second))
		elapsed += dur

		if l.Action == "output" || l.Action == "skip" {
			continue
		}

		hist.Add(l.Elapsed)

		if l.Elapsed > 1 {
			elapsedSlow += dur

			slowest = append(slowest, l)
		}
	}

	sort.Slice(slowest, func(i, j int) bool {
		return slowest[i].Elapsed > slowest[j].Elapsed
	})

	for _, l := range slowest {
		dur := time.Duration(l.Elapsed * float64(time.Second))
		fmt.Printf("%s %s %s %s %s\n", l.Time, l.Action, l.Package, l.Test, dur.String())
	}

	fmt.Printf("Events: %+v\n", counts)
	fmt.Println("Elapsed:", elapsed.String(), "Slow:", elapsedSlow.String())
	fmt.Println("Latency distribution:", hist.String())
}
