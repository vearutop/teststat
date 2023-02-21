package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/godogx/allure/report"
	"github.com/google/uuid"
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
	packageStats                 map[string]packageStat

	passed, failed map[string]int
	failures       map[string][]string

	allureFormatter *report.Formatter

	done int
}

type packageStat struct {
	Package string
	Elapsed float64
	Cached  bool
}

func newProcessor(fl flags) *processor {
	p := &processor{
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
			PrintSum:     true,
		},
		packageStats: map[string]packageStat{},
	}

	if fl.Allure != "" {
		name := os.Getenv("ALLURE_SUITE_NAME")
		if name == "" {
			name = "Go Test"
		}

		p.allureFormatter = &report.Formatter{
			ResultsPath: strings.TrimSuffix(fl.Allure, "/"),
			Container: &report.Container{
				UUID:  uuid.New().String(),
				Start: report.GetTimestampMs(),
				Name:  name,
			},
		}
	}

	return p
}

func (p *processor) process(fn string) (err error) {
	var r io.Reader

	// Read file.
	if fn == "-" || fn == "/dev/stdin" || fn == "" {
		r = os.Stdin
	} else {
		f, oErr := os.Open(fn) //nolint:gosec
		if oErr != nil {
			return oErr
		}

		defer func() {
			if clErr := f.Close(); clErr != nil && err == nil {
				err = clErr
			}
		}()

		r = f
	}

	if p.fl.Store != "" {
		w, err := os.Create(p.fl.Store)
		if err != nil {
			return err
		}

		r = io.TeeReader(r, w)
	}

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1e7), 1e7)

	return p.iterate(scanner)
}

func (p *processor) progress(status string) {
	if p.fl.Progress {
		print(status)

		p.done++
		if p.done >= 80 {
			println()

			p.done = 0
		}
	}
}

func (p *processor) pkgLine(l Line) {
	if l.Elapsed != nil {
		ps := p.packageStats[l.Package]
		ps.Elapsed = *l.Elapsed
		ps.Package = l.Package
		p.packageStats[l.Package] = ps
	}

	if l.Action == "output" {
		if strings.Contains(l.Output, "(cached)") {
			p.counts["pkg_cached"]++

			ps := p.packageStats[l.Package]
			ps.Cached = true
			ps.Package = l.Package
			p.packageStats[l.Package] = ps
		}
	}
}

func (p *processor) iterate(scanner *bufio.Scanner) error {
	outputs := map[string][]string{}

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		b := scanner.Bytes()
		if len(b) == 0 || b[0] != '{' {
			continue
		}

		var l Line
		if err := json.Unmarshal(b, &l); err != nil {
			if !errors.Is(err, io.EOF) {
				log.Println(err)

				continue
			}

			break
		}

		// Skipping package-level stats.
		if l.Test == "" {
			p.pkgLine(l)

			continue
		}

		p.counts[l.Action]++

		test := l.Package + "." + l.Test

		var output []string

		switch l.Action {
		case "output":
			outputs[test] = append(outputs[test], l.Output)

			continue
		case "pass":
			p.progress(".")
			p.passed[test]++
			delete(outputs, test)
		case "fail":
			p.progress("F")
			p.failed[test]++
			output = outputs[test]
			delete(outputs, test)

			if !p.checkRace(test, output) {
				p.failures[test] = output
			}
		case "skip":
			delete(outputs, test)
		}

		p.countElapsed(l)
		p.updateAllure(l, output)
	}

	if p.allureFormatter != nil {
		p.allureFormatter.Container.Stop = p.allureFormatter.Res.Stop
		p.allureFormatter.Finish(report.Executor{})
	}

	p.counts["pkg_total"] = len(p.packageStats)

	return nil
}

func (p *processor) countElapsed(l Line) {
	if l.Elapsed == nil {
		return
	}

	dur := time.Duration(*l.Elapsed * float64(time.Second))
	p.elapsed += dur

	p.hist.Add(*l.Elapsed)

	if *l.Elapsed > p.fl.Slow.Seconds() {
		p.elapsedSlow += dur
		p.counts["slow"]++

		p.slowest = append(p.slowest, l)
	}
}

type flakyTest struct {
	test           string
	passed, failed int
}
