package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"time"

	"github.com/vearutop/dynhist-go"
)

type processor struct {
	counts                       map[string]int
	elapsed, elapsedSlow         time.Duration
	hist                         *dynhist.Collector
	slowest                      []Line
	dataRaces, strippedDataRaces map[string]string
	strippedTests                map[string][]string
	fl                           flags
	packageElapsed               []packageStat

	passed, failed map[string]int
	failures       map[string][]string
}

type packageStat struct {
	Package string
	Elapsed float64
}

func newProcessor(fl flags) *processor {
	return &processor{
		counts:            map[string]int{},
		dataRaces:         map[string]string{},
		strippedDataRaces: map[string]string{},
		strippedTests:     map[string][]string{},
		passed:            map[string]int{},
		failed:            map[string]int{},
		failures:          map[string][]string{},
		fl:                fl,
		hist: &dynhist.Collector{
			BucketsLimit: fl.HistBuckets,
			WeightFunc:   dynhist.ExpWidth(1.2, 0.9),
		},
	}
}

// nolint:funlen
func (p *processor) process(fn string) (err error) {
	var f *os.File

	// Read file.
	if fn == "-" {
		f = os.Stdin
	} else {
		f, err = os.Open(fn) // nolint:gosec
		if err != nil {
			return err
		}

		defer func() {
			if clErr := f.Close(); clErr != nil && err == nil {
				err = clErr
			}
		}()
	}

	dec := json.NewDecoder(f)
	outputs := map[string][]string{}

	for {
		var l Line
		if err := dec.Decode(&l); err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}

			break
		}

		// Skipping package-level stats.
		if l.Test == "" {
			if l.Elapsed > 0 {
				p.packageElapsed = append(p.packageElapsed, packageStat{Package: l.Package, Elapsed: l.Elapsed})
			}

			continue
		}

		p.counts[l.Action]++

		test := l.Package + "." + l.Test

		switch l.Action {
		case "output":
			outputs[test] = append(outputs[test], l.Output)

			continue
		case "pass":
			p.passed[test]++
			delete(outputs, test)
		case "fail":
			p.failed[test]++
			output := outputs[test]
			delete(outputs, test)

			if !p.checkRace(test, output) {
				p.failures[test] = output
			}
		case "skip":
			delete(outputs, test)
		}

		p.countElapsed(l)
	}

	return nil
}

func (p *processor) countElapsed(l Line) {
	if l.Elapsed <= 0 {
		return
	}

	dur := time.Duration(l.Elapsed * float64(time.Second))
	p.elapsed += dur

	p.hist.Add(l.Elapsed)

	if l.Elapsed > 1 {
		p.elapsedSlow += dur
		p.counts["slow"]++

		p.slowest = append(p.slowest, l)
	}
}

type flakyTest struct {
	test           string
	passed, failed int
}
