package main

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
			fmt.Println("### Flaky tests")
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
		return *p.slowest[i].Elapsed > *p.slowest[j].Elapsed
	})

	if len(p.slowest) > 0 {
		if p.fl.Markdown {
			fmt.Println("### Slow tests")
			fmt.Println("<details>")
			fmt.Printf("<summary>Total slow runs: %d</summary>\n\n", len(p.slowest))

			fmt.Println("| Result | Duration | Package | Test |")
			fmt.Println("| - | - | - | - |")

			for i, l := range p.slowest {
				if i >= p.fl.Slowest {
					break
				}

				dur := time.Duration(*l.Elapsed * float64(time.Second))
				fmt.Printf("| %s | %s | %s | %s |\n", l.Action, dur.String(), l.Package, l.Test)
			}

			fmt.Println("</details>")
		} else {
			fmt.Println("Slowest tests:")

			for i, l := range p.slowest {
				if i >= p.fl.Slowest {
					break
				}

				dur := time.Duration(*l.Elapsed * float64(time.Second))
				fmt.Printf("%s %s %s %s\n", l.Action, l.Package, l.Test, dur.String())
			}
		}

		fmt.Println()
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
		fmt.Println("### Data races")
		fmt.Println("<details>")
		fmt.Printf("<summary>Total data races: %d, unique: %d</summary>\n\n",
			len(p.dataRaces), len(p.strippedDataRaces))

		for _, k := range keys {
			r := shortedDataRace(p.strippedDataRaces[k])
			t := p.strippedTests[k]

			t = uniq(t)

			fmt.Println("<details>")
			fmt.Printf("<summary><code>%s</code></summary>\n\n", t[0])

			if len(t) > 1 {
				fmt.Println("Other affected tests:")
				fmt.Println("```")

				for _, tt := range t[1:] {
					fmt.Println(tt)
				}

				fmt.Println("```")
			}

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
			fmt.Println(shortedDataRace(p.strippedDataRaces[k]))
		}

		fmt.Println()
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
			fmt.Println("### Slowest test packages")
			fmt.Println("<details>")
			fmt.Printf("<summary>Total packages with tests: %d</summary>\n\n", len(p.packageStats))

			fmt.Println("| Duration | Package |")
			fmt.Println("| - | - |")

			for i, ps := range pstats {
				dur := time.Duration(ps.Elapsed * float64(time.Second)).String()
				if ps.Cached {
					dur += " (cached)"
				}

				fmt.Printf("| %s | %s |\n", dur, ps.Package)

				if i > p.fl.Slowest {
					break
				}
			}

			fmt.Println("</details>")
			fmt.Println()
		}
	}
}

func (p *processor) reportFailed() {
	if len(p.failures) > 0 {
		if p.fl.Markdown {
			fmt.Println("### Failed tests")
			fmt.Println("<details>")
			fmt.Printf("<summary>Failed: %d</summary>\n\n", len(p.failures))

			for test, output := range p.failures {
				fmt.Println("<details>")
				fmt.Printf("<summary><code>%s</code></summary>\n\n", test)

				fmt.Println("```")
				fmt.Println(strings.Join(output, ""))
				fmt.Println("```")

				fmt.Println("</details>")
			}

			fmt.Println("</details>")
			fmt.Println()
		} else {
			fmt.Println("Failed tests:")
			for test, output := range p.failures {
				fmt.Println(test)
				fmt.Println(strings.Join(output, ""))
			}
		}
	}
}

func (p *processor) storeFailed() {
	if p.fl.FailedTests == "" || len(p.failed) == 0 {
		return
	}

	failedRegex := ""

	for t := range p.failed {
		failedRegex += "^" + t.fn + "$|"
	}

	failedRegex = failedRegex[0 : len(failedRegex)-1]

	if err := os.WriteFile(p.fl.FailedTests, []byte(failedRegex), 0o600); err != nil {
		fmt.Println("failed to store failed tests regexp: " + err.Error())
	}
}

func (p *processor) report() {
	if p.prStatus != "" {
		fmt.Println()
	}

	p.storeFailed()

	if p.fl.SkipReport {
		return
	}

	p.reportFlaky()
	p.reportSlowest()
	p.reportRaces()
	p.reportPackages()
	p.reportFailed()

	if p.fl.Markdown {
		fmt.Println("### Metrics")
		fmt.Println()

		fmt.Printf("```\n%s\n```\n\n", p.status())
		fmt.Println("Elapsed:", p.elapsed.String())
		fmt.Println("Slow:", p.elapsedSlow.String())

		fmt.Println()

		fmt.Println("### Test time distribution (seconds)")
		fmt.Println("```")
		fmt.Println(p.hist.String())
		fmt.Println("```")
	} else {
		fmt.Println("Total", p.status())
		fmt.Println("Elapsed:", p.elapsed.String())
		fmt.Println("Slow:", p.elapsedSlow.String())

		fmt.Println()

		fmt.Println("Test time distribution (seconds):")
		fmt.Println(p.hist.String())
	}
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
