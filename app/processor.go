package app

import (
	"bufio"
	"bytes"
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
	DataRace   int
	Fail       int
	Unfinished int
	Flaky      int
	Output     int
	Pass       int
	PkgTotal   int
	PkgCached  int
	Run        int
	Skip       int
	Slow       int
}

const (
	fail   = "fail"
	pass   = "pass"
	skip   = "skip"
	output = "output"
	run    = "run"

	// Added in go1.24.
	buildOutput = "build-output"
	buildFail   = "build-fail"
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

	unfinished     map[test]bool
	passed, failed map[test]int
	failures       map[test][]string
	outputs        map[test][]string

	// buildFailures contain lines with malformed JSON, typically build errors from STDERR.
	buildFailures []string

	allureFormatter *report.Formatter

	prStatus string
	prLast   time.Time

	rep         io.Writer
	repLimitHit bool
}

type packageStat struct {
	Package string
	Elapsed float64
	Cached  bool
	Failed  bool
}

type limitingWriter struct {
	w        io.Writer
	lim      int
	written  int
	limitHit *bool
}

func (l *limitingWriter) Write(p []byte) (n int, err error) {
	if *l.limitHit {
		return len(p), nil
	}

	l.written += len(p)

	if l.written > l.lim {
		*l.limitHit = true

		_, err := l.w.Write([]byte("...truncated"))
		if err != nil {
			return 0, err
		}
	}

	return l.w.Write(p)
}

func newProcessor(fl flags) *processor {
	p := &processor{
		dataRaces:         map[test]string{},
		strippedDataRaces: map[string]string{},
		strippedTests:     map[string][]string{},
		unfinished:        map[test]bool{},
		passed:            map[test]int{},
		failed:            map[test]int{},
		failures:          map[test][]string{},
		outputs:           map[test][]string{},
		fl:                fl,
		hist: &dynhist.Collector{
			BucketsLimit: fl.HistBuckets,
			WeightFunc:   dynhist.ExpWidth(1.2, 0.9),
			PrintSum:     true,
		},
		packageStats: map[string]packageStat{},
		prLast:       time.Now(),
		rep:          os.Stdout,
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

	if fl.LimitReport > 0 {
		p.rep = &limitingWriter{
			w:        os.Stdout,
			lim:      fl.LimitReport,
			limitHit: &p.repLimitHit,
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

	if c.Unfinished != 0 {
		res += fmt.Sprintf(", unfinished: %d", c.Unfinished)
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

	switch l.Action {
	case output:
		if strings.Contains(l.Output, "(cached)") {
			p.counts.PkgCached++

			ps := p.packageStats[l.Package]
			ps.Cached = true
			ps.Package = l.Package
			p.packageStats[l.Package] = ps
		}
	case fail:
		ps := p.packageStats[l.Package]
		ps.Failed = true
		ps.Package = l.Package
		p.packageStats[l.Package] = ps
	}
}

type test struct {
	pkg, fn string
}

func (t test) String() string {
	return t.pkg + "." + t.fn
}

func (p *processor) iterate(scanner *bufio.Scanner) error {
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		b := scanner.Bytes()
		if len(b) == 0 || b[0] != '{' {
			if bytes.HasPrefix(b, []byte("go: downloading")) || bytes.HasPrefix(b, []byte("go test")) ||
				bytes.HasPrefix(b, []byte("make:")) {
				continue
			}

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

		if l.ImportPath != "" && l.Action == buildOutput {
			p.buildFailures = append(p.buildFailures, strings.TrimSuffix(l.Output, "\n"))
		}

		// Skipping package-level stats.
		if l.Test == "" {
			p.pkgLine(l)

			continue
		}

		p.counts.add(l.Action)

		t := test{pkg: l.Package, fn: l.Test}

		out, skipLine := p.action(l, t)
		if skipLine {
			continue
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

func (p *processor) action(l Line, t test) (out []string, skipLine bool) {
	switch l.Action {
	case run:
		p.unfinished[t] = true
	case output:
		p.outputs[t] = append(p.outputs[t], l.Output)

		return nil, true
	case pass:
		p.progress(false)

		p.passed[t]++

		delete(p.unfinished, t)
		delete(p.outputs, t)
	case fail:
		p.progress(false)

		p.failed[t]++
		out = p.outputs[t]

		delete(p.unfinished, t)
		delete(p.outputs, t)

		if p.fl.Verbosity > 0 {
			println("FAIL:", t.String())
		}

		if p.fl.Verbosity > 1 {
			println(strings.Join(out, ""))
		}

		if !p.checkRace(t, out) {
			p.failures[t] = out
		}
	case skip:
		delete(p.unfinished, t)
		delete(p.outputs, t)
	}

	return out, false
}

func (p *processor) countElapsed(l Line) {
	if l.Elapsed == nil {
		return
	}

	dur := time.Duration(*l.Elapsed * float64(time.Second))
	p.elapsed += dur

	p.hist.Add(*l.Elapsed)

	if *l.Elapsed >= p.fl.Slow.Seconds() {
		p.elapsedSlow += dur
		p.counts.Slow++

		p.slowest = append(p.slowest, l)
	}
}

type flakyTest struct {
	test           test
	passed, failed int
}
