package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/godogx/allure/report"
	"github.com/google/uuid"
	"github.com/vearutop/dynhist-go"
)

type counts struct {
	DataRace  int
	Fail      int
	Flaky     int
	Output    int
	Pass      int
	PkgTotal  int
	PkgCached int
	Run       int
	Skip      int
	Slow      int
}

const (
	fail   = "fail"
	pass   = "pass"
	skip   = "skip"
	output = "output"
	run    = "run"
)

func (c *counts) add(key string) {
	switch key {
	case fail:
		c.Fail++
	case pass:
		c.Pass++
	case output:
		c.Output++
	case run:
		c.Run++
	case skip:
		c.Skip++
	}
}

type processor struct {
	counts               counts
	elapsed, elapsedSlow time.Duration
	hist                 *dynhist.Collector
	slowest              []Line
	dataRaces            map[test]string
	strippedDataRaces    map[string]string
	strippedTests        map[string][]string
	fl                   flags
	packageStats         map[string]packageStat

	passed, failed map[test]int
	failures       map[test][]string

	// buildFailures contain lines with malformed JSON, typically build errors from STDERR.
	buildFailures []string

	allureFormatter *report.Formatter

	prStatus string
	prLast   time.Time
}

type packageStat struct {
	Package string
	Elapsed float64
	Cached  bool
}

func newProcessor(fl flags) *processor {
	p := &processor{
		dataRaces:         map[test]string{},
		strippedDataRaces: map[string]string{},
		strippedTests:     map[string][]string{},
		passed:            map[test]int{},
		failed:            map[test]int{},
		failures:          map[test][]string{},
		fl:                fl,
		hist: &dynhist.Collector{
			BucketsLimit: fl.HistBuckets,
			WeightFunc:   dynhist.ExpWidth(1.2, 0.9),
			PrintSum:     true,
		},
		packageStats: map[string]packageStat{},
		prLast:       time.Now(),
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

func (p *processor) status() string {
	c := p.counts

	res := fmt.Sprintf("pass: %d", c.Pass)

	if c.Fail != 0 {
		res += fmt.Sprintf(", fail: %d", c.Fail)
	}

	if c.Skip != 0 {
		res += fmt.Sprintf(", skip: %d", c.Skip)
	}

	if c.DataRace != 0 {
		res += fmt.Sprintf(", data races: %d", c.DataRace)
	}

	if c.Flaky != 0 {
		res += fmt.Sprintf(", flaky tests: %d", c.Flaky)
	}

	if c.Slow != 0 {
		res += fmt.Sprintf(", slow: %d", c.Slow)
	}

	if c.PkgCached != 0 {
		res += fmt.Sprintf(", cached pkg runs: %d", c.PkgCached)
	}

	if c.PkgTotal != 0 {
		res += fmt.Sprintf(", total pkg: %d", c.PkgTotal)
	}

	return res
}

func (p *processor) progress(force bool) {
	if !p.fl.Progress {
		return
	}

	if force || time.Since(p.prLast) > 5*time.Second {
		st := p.status()

		if p.prStatus != st {
			p.prLast = time.Now()
			p.prStatus = st
			println(st)
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
			p.counts.PkgCached++

			ps := p.packageStats[l.Package]
			ps.Cached = true
			ps.Package = l.Package
			p.packageStats[l.Package] = ps
		}
	}
}

type test struct {
	pkg, fn string
}

func (t test) String() string {
	return t.pkg + "." + t.fn
}

func (p *processor) iterate(scanner *bufio.Scanner) error {
	outputs := map[test][]string{}

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		b := scanner.Bytes()
		if len(b) == 0 || b[0] != '{' {
			p.buildFailures = append(p.buildFailures, scanner.Text())

			continue
		}

		var l Line
		if err := json.Unmarshal(b, &l); err != nil {
			if !errors.Is(err, io.EOF) {
				p.buildFailures = append(p.buildFailures, scanner.Text())

				continue
			}

			break
		}

		// Skipping package-level stats.
		if l.Test == "" {
			p.pkgLine(l)

			continue
		}

		p.counts.add(l.Action)

		t := test{pkg: l.Package, fn: l.Test}

		var out []string

		switch l.Action {
		case output:
			outputs[t] = append(outputs[t], l.Output)

			continue
		case pass:
			p.progress(false)
			p.passed[t]++
			delete(outputs, t)
		case fail:
			p.progress(false)
			p.failed[t]++
			out = outputs[t]
			delete(outputs, t)

			if !p.checkRace(t, out) {
				p.failures[t] = out
			}
		case skip:
			delete(outputs, t)
		}

		p.countElapsed(l)
		p.updateAllure(l, out)
	}

	// Print final progress.
	p.progress(true)

	if p.allureFormatter != nil {
		p.allureFormatter.Container.Stop = p.allureFormatter.Res.Stop
		p.allureFormatter.Finish(report.Executor{})
	}

	p.counts.PkgTotal = len(p.packageStats)

	return scanner.Err()
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
		p.counts.Slow++

		p.slowest = append(p.slowest, l)
	}
}

type flakyTest struct {
	test           test
	passed, failed int
}
