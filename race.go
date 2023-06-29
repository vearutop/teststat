package main

import (
	"strings"
)

func (p *processor) checkRace(t test, output []string) bool {
	raceFound := false
	raceFailed := false

	for _, l := range output {
		if l == "WARNING: DATA RACE\n" {
			raceFound = true

			break
		}

		if strings.Contains(l, "race detected during execution of test") {
			raceFailed = true

			break
		}
	}

	if raceFailed {
		return true
	}

	if !raceFound {
		return false
	}

	p.counts.DataRace++
	p.dataRaces[t] = strings.Join(output, "")

	sk := strippedKey(stripDataRace(output), p.fl.RaceDepth)
	p.strippedDataRaces[sk] = strings.Join(output, "")
	p.strippedTests[sk] = append(p.strippedTests[sk], t.String())

	return true
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
