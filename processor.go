package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
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
	packageElapsed               map[string]float64

	passed, failed map[string]int

	isDataRace bool
	dataRace   []string
}

func newProcessor(fl flags) *processor {
	return &processor{
		counts:            map[string]int{},
		dataRaces:         map[string]string{},
		strippedDataRaces: map[string]string{},
		strippedTests:     map[string][]string{},
		passed:            map[string]int{},
		failed:            map[string]int{},
		packageElapsed:    map[string]float64{},
		fl:                fl,
		hist: &dynhist.Collector{
			BucketsLimit: fl.HistBuckets,
			WeightFunc:   dynhist.ExpWidth(1.2, 0.9),
		},
	}
}

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
				p.packageElapsed[l.Package] = l.Elapsed
			}

			continue
		}

		p.counts[l.Action]++

		if l.Action == "pass" {
			p.passed[l.Package+"."+l.Test]++
		} else if l.Action == "fail" {
			p.failed[l.Package+"."+l.Test]++
		}

		p.checkRace(l)

		if l.Action == "output" || l.Action == "skip" {
			continue
		}

		p.countElapsed(l)
	}

	return nil
}

func (p *processor) checkRace(l Line) {
	if l.Output == "WARNING: DATA RACE\n" {
		p.counts["data_race"]++

		p.isDataRace = true
	}

	if p.isDataRace && l.Output == "==================\n" {
		p.isDataRace = false

		if len(p.dataRace) != 0 {
			p.dataRaces[l.Package+"."+l.Test] = strings.Join(p.dataRace, "")

			sk := strippedKey(stripDataRace(p.dataRace), p.fl.RaceDepth)
			p.strippedDataRaces[sk] = strings.Join(p.dataRace, "")
			p.strippedTests[sk] = append(p.strippedTests[sk], l.Test)

			p.dataRace = p.dataRace[:0]
		}
	}

	if p.isDataRace && l.Action == "output" {
		p.dataRace = append(p.dataRace, l.Output)
	}
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

func (p *processor) reportFlaky() {
	var flaky []flakyTest

	for test, count := range p.failed {
		if p.passed[test] != 0 {
			flaky = append(flaky, flakyTest{
				test:   test,
				passed: p.passed[test],
				failed: count,
			})
		}
	}

	sort.Slice(flaky, func(i, j int) bool {
		return flaky[i].test > flaky[j].test
	})

	if len(flaky) > 0 {
		p.counts["flaky"] = len(flaky)

		if p.fl.Markdown {
			fmt.Println("## Flaky tests")
			fmt.Println("<details>")
			fmt.Printf("<summary>Tests: %d</summary>\n\n", len(flaky))

			fmt.Println("| Pass | Fail | Test |")
			fmt.Println("| - | - | - |")

			for _, ft := range flaky {
				fmt.Printf("| %d | %d | %s |\n", ft.passed, ft.failed, ft.test)
			}

			fmt.Println("</details>")
		} else {
			fmt.Println("Flaky tests:")

			for _, ft := range flaky {
				fmt.Printf("%s: %d passed, %d failed\n", ft.test, ft.passed, ft.failed)
			}
		}

		fmt.Println()
	}
}

func (p *processor) reportSlowest() {
	sort.Slice(p.slowest, func(i, j int) bool {
		return p.slowest[i].Elapsed > p.slowest[j].Elapsed
	})

	if len(p.slowest) > 0 {
		if p.fl.Markdown {
			fmt.Println("## Slow tests")
			fmt.Println("<details>")
			fmt.Printf("<summary>Total slow runs: %d</summary>\n\n", len(p.slowest))

			fmt.Println("| Result | Duration | Package | Test |")
			fmt.Println("| - | - | - | - |")

			for i, l := range p.slowest {
				if i >= p.fl.Slowest {
					break
				}

				dur := time.Duration(l.Elapsed * float64(time.Second))
				fmt.Printf("| %s | %s | %s | %s |\n", l.Action, dur.String(), l.Package, l.Test)
			}

			fmt.Println("</details>")
		} else {
			fmt.Println("Slowest tests:")

			for i, l := range p.slowest {
				if i >= p.fl.Slowest {
					break
				}

				dur := time.Duration(l.Elapsed * float64(time.Second))
				fmt.Printf("%s %s %s %s\n", l.Action, l.Package, l.Test, dur.String())
			}
		}

		fmt.Println()
	}
}

func (p *processor) reportRaces() {
	if len(p.strippedDataRaces) > 0 {
		var keys []string

		for k := range p.strippedDataRaces {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		if p.fl.Markdown {
			fmt.Println("## Data races")
			fmt.Println("<details>")
			fmt.Printf("<summary>Total data races: %d, unique: %d</summary>\n\n",
				len(p.dataRaces), len(p.strippedDataRaces))

			for _, k := range keys {
				r := p.strippedDataRaces[k]
				t := p.strippedTests[k]

				if len(t) > 3 {
					t = append(t[0:3], "...")
				}

				fmt.Println("<details>")
				fmt.Printf("<summary>%s</summary>\n\n", strings.Join(t, ", "))
				fmt.Println("\n```")
				fmt.Println(r)
				fmt.Println("```")
				fmt.Println("</details>")
				fmt.Println()
			}

			fmt.Println("</details>")
			fmt.Println()
		} else {
			fmt.Println("Data races:")

			for _, k := range keys {
				t := p.strippedTests[k]

				if len(t) > 3 {
					t = append(t[0:3], "...")
				}

				fmt.Println(strings.Join(t, ", "))
				fmt.Println(p.strippedDataRaces[k])
			}

			fmt.Println()
		}
	}
}

func (p *processor) report() {
	p.reportFlaky()
	p.reportSlowest()
	p.reportRaces()

	if p.fl.Markdown {
		fmt.Println("## Metrics")
		fmt.Println()

		fmt.Printf("```\n%v\n```\n\n", p.counts)
		fmt.Println("Elapsed:", p.elapsed.String())
		fmt.Println("Slow:", p.elapsedSlow.String())

		fmt.Println()

		fmt.Println("## Elapsed distribution")
		fmt.Println("```")
		fmt.Println(p.hist.String())
		fmt.Println("```")
	} else {
		fmt.Printf("Metrics: %v\n", p.counts)
		fmt.Println("Elapsed:", p.elapsed.String())
		fmt.Println("Slow:", p.elapsedSlow.String())

		fmt.Println()

		fmt.Println("Elapsed distribution:")
		fmt.Println(p.hist.String())
	}
}

func strippedKey(stripped [][]string, limit int) string {
	var res string

	if len(stripped) < 2 {
		return ""
	}

	// Taking only first 2 traces (Read/Write and Previous Write).
	for i := 0; i <= 1; i++ {
		if len(stripped[i]) >= limit {
			res += strings.Join(stripped[i][0:limit], "\n")
		} else {
			res += strings.Join(stripped[i], "\n")
		}

		res += "\n=====\n"
	}

	return res
}

func stripDataRace(data []string) [][]string {
	var (
		traces [][]string
		trace  []string
	)

	for _, l := range data {
		if strings.Trim(l, " \n") == "" {
			if len(trace) > 0 {
				traces = append(traces, trace)

				trace = nil
			}

			continue
		}

		if strings.HasPrefix(l, "      ") {
			l = strings.Trim(l, " ")
			l = strings.Split(l, " ")[0]
			trace = append(trace, l)
		}
	}

	if len(trace) > 0 {
		traces = append(traces, trace)
	}

	return traces
}
