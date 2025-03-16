package sqlite

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/bool64/sqluct"
	"github.com/vearutop/teststat/app/model"
)

type Repository struct {
	RecentlyFailedRuns       time.Duration // Stores only failed or successful runs that have failed within duration.
	OnlyRecentlyFailedTotals bool
	SkipOutputSamples        bool

	st *sqluct.Storage

	packages       map[Hash]bool
	tests          map[Hash]bool
	revisions      map[Hash]bool
	recentlyFailed map[Hash]bool
	totals         map[Hash]Total

	tx     *sql.Tx
	rowsTx int
}

func NewRepository(fn string, options ...func(r *Repository)) (*Repository, error) {
	st, err := newStorage(fn)
	if err != nil {
		return nil, err
	}

	r := &Repository{
		st:        st,
		packages:  make(map[Hash]bool),
		tests:     make(map[Hash]bool),
		revisions: make(map[Hash]bool),
		totals:    make(map[Hash]Total),
	}

	for _, option := range options {
		option(r)
	}

	if r.RecentlyFailedRuns > 0 {
		if err := r.populateRecentlyFailed(); err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (r *Repository) populateRecentlyFailed() error {
	ctx := context.Background()
	ref := r.st.MakeReferencer()
	row := &Total{}

	ref.AddTableAlias(row, "")

	var rows []Total
	qb := r.st.SelectStmt(totalsTable, nil).
		Column(ref.Col(&row.Hash)).
		Where(ref.Fmt("%s > ?", &row.LastFailed), time.Now().Add(-r.RecentlyFailedRuns).UnixMilli())

	if err := r.st.Select(ctx, qb, &rows); err != nil {
		return err
	}

	r.recentlyFailed = make(map[Hash]bool)

	for _, row := range rows {
		r.recentlyFailed[row.Hash] = true
	}

	return nil
}

func (r *Repository) AddRun(run Run) (err error) {
	if r.tx == nil {
		r.tx, err = r.st.DB().Begin()
	}

	t := Test{
		Hash:        run.TestHash,
		PackageHash: StringHash(run.Package),
		Test:        run.Fn,
	}

	if r.RecentlyFailedRuns > 0 && r.OnlyRecentlyFailedTotals {
		if !r.recentlyFailed[t.Hash] {
			if run.Result == model.Passed || run.Result == model.Skipped {
				return nil
			}
		}
	}

	tot := r.totals[t.Hash]
	tot.Hash = t.Hash
	switch run.Result {
	case model.Passed:
		tot.Passed++
	case model.Failed:
		tot.Failed++
	case model.Skipped:
		tot.Skipped++
	case model.DataRace:
		tot.DataRaces++
	case model.Unfinished:
		tot.Unfinished++
	}
	if run.Cached {
		tot.Cached++
	}
	tot.Pauses += run.Pauses
	tot.Elapsed += run.Elapsed
	tot.OutputLines += run.OutputLines
	if tot.FirstRevision == 0 {
		tot.FirstRevision = run.RevisionHash
	}
	tot.LastRevision = run.RevisionHash
	if tot.First == 0 {
		tot.First = run.Started
	}
	tot.Last = run.Started
	if run.Result != model.Passed && run.Result != model.Skipped {
		tot.LastFailed = run.Started

		if len(run.Output) > 0 && !r.SkipOutputSamples {
			o := Output{
				TestHash: run.TestHash,
				RunHash:  run.Hash,
				Output:   strings.Join(run.Output, ""),
			}
			_, err = r.st.InsertStmt(outputsTable, o).
				Suffix("ON CONFLICT(test_hash) DO UPDATE SET output = excluded.output, run_hash = excluded.run_hash").
				RunWith(r.tx).Exec()
			if err != nil {
				println(err.Error())
				return err
			}

			r.rowsTx++
		}
	}
	tot.Runs++
	r.totals[t.Hash] = tot

	if !r.packages[t.PackageHash] {
		if _, err := r.st.InsertStmt(packagesTable, Package{
			Hash:    t.PackageHash,
			Package: run.Package,
		}, sqluct.InsertIgnore).RunWith(r.tx).Exec(); err != nil {
			return err
		}

		r.rowsTx++

		r.packages[t.PackageHash] = true
	}

	if !r.tests[run.TestHash] {
		if _, err := r.st.InsertStmt(testsTable, t, sqluct.InsertIgnore).RunWith(r.tx).Exec(); err != nil {
			return err
		}

		r.rowsTx++

		r.tests[run.TestHash] = true
	}

	if !r.revisions[run.RevisionHash] {
		if _, err := r.st.InsertStmt(revisionsTable, Revision{
			Hash:     run.RevisionHash,
			Revision: run.Revision,
		}, sqluct.InsertIgnore).RunWith(r.tx).Exec(); err != nil {
			return err
		}

		r.rowsTx++

		r.revisions[run.RevisionHash] = true
	}

	if r.RecentlyFailedRuns > 0 {
		if !r.recentlyFailed[t.Hash] {
			if run.Result == model.Passed || run.Result == model.Skipped {
				return nil
			} else {
				r.recentlyFailed[t.Hash] = true
			}
		}
	}

	_, err = r.st.InsertStmt(runsTable, run, sqluct.InsertIgnore).RunWith(r.tx).Exec()
	if err != nil {
		return err
	}

	r.rowsTx++

	if r.rowsTx > 500 {
		err = r.tx.Commit()
		r.tx = nil
		r.rowsTx = 0
		return err
	}

	return nil
}

func (r *Repository) SyncTotals() error {
	if r.tx != nil {
		if err := r.tx.Commit(); err != nil {
			return err
		}
	}

	suffix := "ON CONFLICT(test_hash) DO UPDATE SET " +
		"failed = failed + excluded.failed," +
		"passed = passed + excluded.passed," +
		"unfinished = unfinished + excluded.unfinished," +
		"skipped = skipped + excluded.skipped," +
		"output_lines = output_lines + excluded.output_lines," +
		"data_races = data_races + excluded.data_races," +
		"pauses = pauses + excluded.pauses," +
		"runs = runs + excluded.runs," +
		"cached = cached + excluded.cached," +
		"elapsed = elapsed + excluded.elapsed," +
		"first_ums = min(first_ums, excluded.first_ums)," +
		"last_ums = max(last_ums, excluded.last_ums)," +
		"last_failed_ums = max(last_failed_ums, excluded.last_failed_ums)," +
		"last_rev = excluded.last_rev;"

	var totals []Total
	for _, t := range r.totals {
		totals = append(totals, t)

		if len(totals) >= 1000 {
			_, err := r.st.InsertStmt(totalsTable, totals).Suffix(suffix).Exec()
			if err != nil {
				return err
			}

			totals = totals[:0]
		}
	}

	if len(totals) > 0 {
		_, err := r.st.InsertStmt(totalsTable, totals).Suffix(suffix).Exec()
		if err != nil {
			return err
		}
	}

	return nil
}
