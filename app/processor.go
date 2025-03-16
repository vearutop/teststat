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
	"github.com/vearutop/teststat/app/model"
	"github.com/vearutop/teststat/app/sqlite"
)

type counts struct {
	DataRace   int
	Fail       int
	Unfinished int
	Flaky      int
	Output     int
	Pause      int
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
	run    = "run" // Test starts.
	pause  = "pause"
	start  = "start" // Package starts.
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
	case pause:
		c.Pause++
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

	tests    map[test]model.TestRun
	testRuns []model.TestRun

	unfinished     map[test]bool
	passed, failed map[test]int
	failures       map[test][]string
	outputs        map[test][]string

	// buildFailures contain lines with malformed JSON, typically build errors from STDERR.
	buildFailures []string

	allureFormatter *report.Formatter
	repo            *sqlite.Repository

	progressStatus string
	progressLast   time.Time

	rep         io.Writer
	repLimitHit bool
}

type packageStat struct {
	Package string
	Started time.Time
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

func newProcessor(fl flags) (*processor, error) {
	p := &processor{
		dataRaces:         map[test]string{},
		strippedDataRaces: map[string]string{},
		strippedTests:     map[string][]string{},
		tests:             make(map[test]model.TestRun),
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
		progressLast: time.Now(),
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

	if fl.Sqlite != "" {
		repo, err := sqlite.NewRepository(fl.Sqlite, func(r *sqlite.Repository) {
			if fl.RecentlyFailed > 0 {
				r.RecentlyFailedRuns = time.Duration(fl.RecentlyFailed) * 24 * time.Hour
				r.OnlyRecentlyFailedTotals = true
			}
		})
		if err != nil {
			return nil, err
		}

		p.repo = repo
	}

	if fl.LimitReport > 0 {
		p.rep = &limitingWriter{
			w:        os.Stdout,
			lim:      fl.LimitReport,
			limitHit: &p.repLimitHit,
		}
	}

	return p, nil
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

	if force || time.Since(p.progressLast) > 5*time.Second {
		st := p.status()

		if p.progressStatus != st {
			p.progressLast = time.Now()
			p.progressStatus = st
			println(st)
		}
	}
}

func (p *processor) pkgLine(l Line) {
	t := test{pkg: l.Package}
	tr := p.tests[t]
	tr.Package = l.Package

	if l.Elapsed != nil {
		ps := p.packageStats[l.Package]
		ps.Elapsed = *l.Elapsed
		ps.Package = l.Package
		p.packageStats[l.Package] = ps
		tr.Elapsed = *l.Elapsed
	}

	switch l.Action {
	case output:
		if strings.Contains(l.Output, "(cached)") {
			p.counts.PkgCached++

			ps := p.packageStats[l.Package]
			ps.Cached = true
			ps.Package = l.Package
			p.packageStats[l.Package] = ps

			tr.Cached = true
			tr.Started = int(l.Time.UnixMilli())
			p.tests[t] = tr
		}
	case fail:
		ps := p.packageStats[l.Package]
		ps.Failed = true
		ps.Package = l.Package
		p.packageStats[l.Package] = ps

		tr.Result = model.Failed
		p.testRuns = append(p.testRuns, tr)
		delete(p.tests, t)
	case pass:
		tr.Result = model.Passed
		p.testRuns = append(p.testRuns, tr)
		delete(p.tests, t)
	}

	p.tests[t] = tr
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
	case start:
		ps := p.packageStats[l.Package]
		ps.Started = l.Time
		p.packageStats[l.Package] = ps

		p.tests[t] = model.TestRun{
			Package: l.Package,
			Fn:      t.fn,
			Result:  model.Unfinished,
			Started: int(l.Time.UnixMilli()),
		}
	case run:
		p.tests[t] = model.TestRun{
			Package: l.Package,
			Fn:      t.fn,
			Result:  model.Unfinished,
			Started: int(l.Time.UnixMilli()),
		}
		p.unfinished[t] = true
	case output:
		p.outputs[t] = append(p.outputs[t], l.Output)

		return nil, true
	case pass:
		p.progress(false)
		tr := p.tests[t]
		tr.Result = model.Passed
		if l.Elapsed != nil {
			tr.Elapsed = *l.Elapsed
		}
		tr.OutputLines = len(p.outputs[t])
		p.testRuns = append(p.testRuns, tr)

		p.passed[t]++

		delete(p.tests, t)
		delete(p.unfinished, t)
		delete(p.outputs, t)
	case fail:
		p.progress(false)

		p.failed[t]++
		out = p.outputs[t]

		tr := p.tests[t]
		tr.Result = model.Failed
		if l.Elapsed != nil {
			tr.Elapsed = *l.Elapsed
		}
		tr.Output = out
		tr.OutputLines = len(out)
		p.tests[t] = tr

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

		p.testRuns = append(p.testRuns, p.tests[t])
		delete(p.tests, t)
	case skip:
		tr := p.tests[t]
		tr.OutputLines = len(p.outputs[t])
		tr.Result = model.Skipped
		p.testRuns = append(p.testRuns, tr)

		delete(p.tests, t)
		delete(p.unfinished, t)
		delete(p.outputs, t)
	case pause:
		tr := p.tests[t]
		tr.Pauses++
		p.tests[t] = tr
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
