package app

import (
	"encoding/json"
	"os"
)

// metrics is teststat's total-run summary for this invocation, meant to be attached to a CI
// cache session's own report rather than read directly -- see gocacheprogd's
// `report_<name>=<path>` DSN param, which reads a file back at -github-actions-done time and
// inlines it into that session's sessions.jsonl line under "<name>" if it parses as JSON (which
// this always does).
type metrics struct {
	Pass       int `json:"pass"`
	Fail       int `json:"fail"`
	Unfinished int `json:"unfinished"`
	Flaky      int `json:"flaky"`
	Skip       int `json:"skip"`
	DataRaces  int `json:"data_races"`
	Slow       int `json:"slow"`

	ElapsedS     float64 `json:"elapsed_s"`
	ElapsedSlowS float64 `json:"elapsed_slow_s"`

	// Pkg* mirrors cacheStats: PkgTotal includes PkgNoTests, and PkgFailing is excluded from the
	// cache-miss judgment entirely (go test never caches a failing result).
	PkgTotal   int `json:"pkg_total"`
	PkgCached  int `json:"pkg_cached"`
	PkgMiss    int `json:"pkg_cache_miss"`
	PkgFailing int `json:"pkg_failing"`
	PkgNoTests int `json:"pkg_no_tests"`
}

func (p *processor) buildMetrics() metrics {
	total, cached, miss, failing, noTests, _ := p.cacheStats()

	return metrics{
		Pass:       p.counts.Pass,
		Fail:       p.counts.Fail,
		Unfinished: len(p.unfinished),
		Flaky:      p.counts.Flaky,
		Skip:       p.counts.Skip,
		DataRaces:  p.counts.DataRace,
		Slow:       p.counts.Slow,

		ElapsedS:     p.elapsed.Seconds(),
		ElapsedSlowS: p.elapsedSlow.Seconds(),

		PkgTotal:   total,
		PkgCached:  cached,
		PkgMiss:    miss,
		PkgFailing: failing,
		PkgNoTests: noTests,
	}
}

func (p *processor) storeMetricsJSON() {
	if p.fl.MetricsJSON == "" {
		return
	}

	data, err := json.MarshalIndent(p.buildMetrics(), "", "  ")
	if err != nil {
		p.println("failed to marshal metrics json: " + err.Error())

		return
	}

	if err := os.WriteFile(p.fl.MetricsJSON, data, 0o600); err != nil {
		p.println("failed to store metrics json: " + err.Error())
	}
}
