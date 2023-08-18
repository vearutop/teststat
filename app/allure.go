// Package main provides teststat utility.
package app

import (
	"log"
	"strings"

	"github.com/godogx/allure/report"
)

func (p *processor) updateAllure(l Line, output []string) {
	if p.allureFormatter == nil {
		return
	}

	if p.allureFormatter.LastTime == 0 {
		p.allureFormatter.Container.Start = report.TimeMs(l.Time)
	}

	// Skipping package level stats.
	if l.Test == "" {
		return
	}

	f := p.allureFormatter

	if l.Elapsed != nil {
		stop := report.TimeMs(l.Time)
		start := stop - report.TimestampMs(int(1000**l.Elapsed))

		f.StartNewResult(report.Result{
			Name:      l.Test,
			FullName:  l.Package + "." + l.Test,
			HistoryID: l.Package + "." + l.Test,
			Start:     start,
			Stop:      stop,
			Labels: []report.Label{
				{Name: "feature", Value: l.Package},
				{Name: "suite", Value: f.Container.Name},
				{Name: "framework", Value: "go test"},
				{Name: "language", Value: "Go"},
			},
		})

		switch l.Action {
		case pass:
			f.StepFinished("test", report.Passed, nil, func(s *report.Step) {
				s.Stop = stop
				s.Start = start
				f.Res.Stop = stop
				f.LastTime = stop
			})
		case fail:
			f.StepFinished("test", report.Failed, nil, func(s *report.Step) {
				if len(output) > 0 {
					att, err := f.BytesAttachment([]byte(strings.Join(output, "\n")), "")
					if err != nil {
						log.Println(err)
					} else {
						s.Attachments = append(s.Attachments, *att)
					}
				}
			})
		case skip:
			f.StepFinished("test", report.Skipped, nil, nil)
		}
	}
}
