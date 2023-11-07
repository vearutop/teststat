package app

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

func (p *processor) reportFlaky() {
	var flaky []flakyTest

	for t, count := range p.failed {
		if p.passed[t] != 0 {
			flaky = append(flaky, flakyTest{
				test:   t,
				passed: p.passed[t],
				failed: count,
			})
		}
	}

	sort.Slice(flaky, func(i, j int) bool {
		return flaky[i].test.String() > flaky[j].test.String()
	})

	if len(flaky) > 0 {
		p.counts.Flaky = len(flaky)

		if p.fl.Markdown {
			p.println("### Flaky tests")
			p.println("<details>")
			p.printf("<summary>Tests: %d</summary>\n\n", len(flaky))

			p.println("| Pass | Fail | Test |")
			p.println("| - | - | - |")

			for _, ft := range flaky {
				p.printf("| %d | %d | %s |\n", ft.passed, ft.failed, ft.test)
			}

			p.println("</details>")
		} else {
			p.println("Flaky tests:")

			for _, ft := range flaky {
				p.printf("%s: %d passed, %d failed\n", ft.test, ft.passed, ft.failed)
			}
		}

		p.println()
	}
}

func (p *processor) reportSlowest() {
	sort.Slice(p.slowest, func(i, j int) bool {
		return *p.slowest[i].Elapsed > *p.slowest[j].Elapsed
	})

	if len(p.slowest) > 0 {
		if p.fl.Markdown {
			p.println("### Slow tests")
			p.println("<details>")
			p.printf("<summary>Total slow runs: %d</summary>\n\n", len(p.slowest))

			p.println("| Result | Duration | Package | Test |")
			p.println("| - | - | - | - |")

			for i, l := range p.slowest {
				if i >= p.fl.Slowest {
					break
				}

				dur := time.Duration(*l.Elapsed * float64(time.Second))
				p.printf("| %s | %s | %s | %s |\n", l.Action, dur.String(), l.Package, l.Test)
			}

			p.println("</details>")
		} else {
			p.println("Slowest tests:")

			for i, l := range p.slowest {
				if i >= p.fl.Slowest {
					break
				}

				dur := time.Duration(*l.Elapsed * float64(time.Second))
				p.printf("%s %s %s %s\n", l.Action, l.Package, l.Test, dur.String())
			}
		}

		p.println()
	}
}

func (p *processor) reportRaces() {
	if len(p.strippedDataRaces) == 0 {
		return
	}

	keys := make([]string, 0, len(p.strippedDataRaces))

	for k := range p.strippedDataRaces {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	if p.fl.Markdown {
		p.println("### Data races")
		p.println("<details>")
		p.printf("<summary>Total data races: %d, unique: %d</summary>\n\n",
			len(p.dataRaces), len(p.strippedDataRaces))

		for _, k := range keys {
			r := shortedDataRace(p.strippedDataRaces[k])
			t := p.strippedTests[k]

			t = uniq(t)

			p.println("<details>")
			p.printf("<summary><code>%s</code></summary>\n\n", t[0])

			if len(t) > 1 {
				p.println("Other affected tests:")
				p.println("```")

				for _, tt := range t[1:] {
					p.println(tt)
				}

				p.println("```")
			}

			p.println("\n```")
			p.println(r)
			p.println("```")
			p.println("</details>")
			p.println()
		}

		p.println("</details>")
		p.println()
	} else {
		p.println("Data races:")

		for _, k := range keys {
			t := p.strippedTests[k]

			if len(t) > 3 {
				t = append(t[0:3], "...")
			}

			p.println(strings.Join(t, ", "))
			p.println(shortedDataRace(p.strippedDataRaces[k]))
		}

		p.println()
	}
}

func (p *processor) reportPackages() {
	if len(p.packageStats) > 0 {
		pstats := make([]packageStat, 0, len(p.packageStats))
		cached := 0

		for _, v := range p.packageStats {
			if v.Cached {
				cached++
			}

			pstats = append(pstats, v)
		}

		sort.Slice(pstats, func(i, j int) bool {
			return pstats[i].Elapsed > pstats[j].Elapsed
		})

		if p.fl.Markdown {
			p.println("### Slowest test packages")
			p.println("<details>")
			p.printf("<summary>Total packages with tests: %d</summary>\n\n", len(p.packageStats))

			p.println("| Duration | Package |")
			p.println("| - | - |")

			for i, ps := range pstats {
				dur := time.Duration(ps.Elapsed * float64(time.Second)).String()
				if ps.Cached {
					dur += " (cached)"
				}

				p.printf("| %s | %s |\n", dur, ps.Package)

				if i > p.fl.Slowest {
					break
				}
			}

			p.println("</details>")
			p.println()
		}
	}
}

func (p *processor) reportFailed() {
	if len(p.failures) == 0 && len(p.buildFailures) == 0 {
		return
	}

	if p.fl.Markdown {
		p.println("### Failures")

		if len(p.buildFailures) > 0 {
			p.println("<details>")
			p.printf("<summary>Failed builds</summary>\n\n")
			p.println("```")

			for _, output := range p.buildFailures {
				p.println(output)
			}

			p.println("```\n\n</details>")
			p.println()
		}

		if len(p.failures) > 0 {
			p.println("<details>")
			p.printf("<summary>Failed tests (including flaky): %d</summary>\n\n", len(p.failures))

			var failures []test

			for t := range p.failures {
				failures = append(failures, t)
			}

			sort.Slice(failures, func(i, j int) bool {
				return failures[i].String() < failures[j].String()
			})

			for _, t := range failures {
				output := p.failures[t]

				p.println("<details>")
				p.printf("<summary><code>%s</code></summary>\n\n", t)

				p.println("```")
				p.println(strings.Join(output, ""))
				p.println("```")

				p.println("</details>")
			}

			p.println("</details>")
			p.println()
		}
	} else {
		if len(p.buildFailures) > 0 {
			p.println("Failed builds:")
			for _, output := range p.buildFailures {
				p.println(output)
			}
		}

		if len(p.failures) > 0 {
			p.println("Failed tests (including flaky):")
			for test, output := range p.failures {
				p.println(test)
				p.println(strings.Join(output, ""))
			}
		}
	}
}

func (p *processor) storeFailureStats() {
	if p.fl.FailureStats == "" {
		return
	}

	rep := ""

	if len(p.buildFailures) > 0 {
		failed := 0

		for _, l := range p.buildFailures {
			if strings.Contains(l, "[build failed]") {
				failed++
			}
		}

		rep += fmt.Sprintf(", %d package(s) failed build", failed)
	}

	if len(p.failures) > 0 {
		flaky := 0
		failed := 0

		for t := range p.failures {
			if p.passed[t] > 0 {
				flaky++
			} else {
				failed++
			}
		}

		if failed > 0 {
			rep += fmt.Sprintf(", %d failed test(s)", failed)
		}

		if flaky > 0 {
			rep += fmt.Sprintf(", %d flaky test(s)", flaky)
		}
	}

	if len(p.dataRaces) > 0 {
		rep += fmt.Sprintf(", %d data race(s)", len(p.dataRaces))
	}

	if rep == "" {
		rep = "no failures"
	} else {
		rep = rep[2:]
	}

	if err := os.WriteFile(p.fl.FailureStats, []byte(rep+"\n"), 0o600); err != nil {
		p.println("failed to store failure stats: " + err.Error())
	}
}

func (p *processor) storeFailed() {
	if p.fl.FailedTests == "" || len(p.failed) == 0 {
		return
	}

	failedRegex := map[string]bool{}

	for t := range p.failed {
		failedRegex["^"+t.fn+"$"] = true
	}

	if p.fl.SkipParent {
		for k := range failedRegex {
			for {
				p := strings.LastIndex(k, "/")
				if p == -1 {
					break
				}

				k = k[0:p] + "$"
				delete(failedRegex, k)
			}
		}
	}

	fr := make([]string, 0, len(failedRegex))

	for k := range failedRegex {
		fr = append(fr, k)
	}

	sort.Strings(fr)

	if err := os.WriteFile(p.fl.FailedTests, []byte(strings.Join(fr, "|")), 0o600); err != nil {
		p.println("failed to store failed tests regexp: " + err.Error())
	}
}

func (p *processor) storeBuildFailures() {
	if p.fl.FailedBuilds == "" || len(p.buildFailures) == 0 {
		return
	}

	if err := os.WriteFile(p.fl.FailedBuilds, []byte(strings.Join(p.buildFailures, "")), 0o600); err != nil {
		p.println("failed to store build failed: " + err.Error())
	}
}

func (p *processor) println(a ...interface{}) {
	if p.repLimitHit {
		return
	}

	if _, err := fmt.Fprintln(p.rep, a...); err != nil {
		panic(err.Error())
	}
}

func (p *processor) printf(format string, a ...interface{}) {
	if p.repLimitHit {
		return
	}

	if _, err := fmt.Fprintf(p.rep, format, a...); err != nil {
		panic(err.Error())
	}
}

func (p *processor) filterUniqBuildFailures() {
	if len(p.buildFailures) == 0 {
		return
	}

	u := map[string]bool{}

	var res []string

	for _, l := range p.buildFailures {
		if !u[l] {
			res = append(res, l)
			u[l] = true
		}
	}

	p.buildFailures = res
}

func (p *processor) report() {
	if p.prStatus != "" {
		p.println()
	}

	p.filterUniqBuildFailures()
	p.storeFailed()
	p.storeFailureStats()
	p.storeBuildFailures()

	if p.fl.SkipReport {
		return
	}

	p.reportFailed()

	if p.fl.Markdown {
		p.println("### Metrics")
		p.println()

		p.printf("```\n%s\n```\n\n", p.status())
		p.println("Elapsed:", p.elapsed.String())
		p.println("Slow:", p.elapsedSlow.String())

		p.println()

		p.println("### Test time distribution (seconds)")
		p.println("```")
		p.println(p.hist.String())
		p.println("```")
	} else {
		p.println("Total", p.status())
		p.println("Elapsed:", p.elapsed.String())
		p.println("Slow:", p.elapsedSlow.String())

		p.println()

		p.println("Test time distribution (seconds):")
		p.println(p.hist.String())
	}

	p.reportFlaky()
	p.reportSlowest()
	p.reportRaces()
	p.reportPackages()
}

func uniq(a []string) []string {
	if len(a) <= 1 {
		return a
	}

	idx := map[string]bool{}
	res := make([]string, 0, len(a))

	for _, s := range a {
		if idx[s] {
			continue
		}

		res = append(res, s)
		idx[s] = true
	}

	return res
}

func shortedDataRace(r string) string {
	maxLen := 5000

	if len(r) < maxLen {
		return r
	}

	p := strings.Index(r, "WARNING: DATA RACE")
	p2 := strings.Index(r[p+1:], "WARNING: DATA RACE")

	if p2 > 0 {
		return r[0:p+p2] + "\n...... other data race(s) truncated\n"
	}

	return r
}
