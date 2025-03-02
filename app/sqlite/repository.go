package sqlite

import (
	"database/sql"

	"github.com/bool64/sqluct"
)

type Repository struct {
	st *sqluct.Storage

	tests     map[Hash]bool
	revisions map[Hash]bool

	tx     *sql.Tx
	rowsTx int
}

func NewRepository(fn string) (*Repository, error) {
	st, err := newStorage(fn)
	if err != nil {
		return nil, err
	}

	return &Repository{
		st:        st,
		tests:     make(map[Hash]bool),
		revisions: make(map[Hash]bool),
	}, nil
}

func (r *Repository) AddRun(run Run) (err error) {
	if r.tx == nil {
		r.tx, err = r.st.DB().Begin()
	}

	_, err = r.st.InsertStmt(runsTable, run, sqluct.InsertIgnore).RunWith(r.tx).Exec()
	if err != nil {
		return err
	}

	r.rowsTx++

	if !r.tests[run.TestHash] {
		if _, err := r.st.InsertStmt(testsTable, Test{
			Hash:    run.TestHash,
			Package: run.Package,
			Test:    run.Fn,
		}, sqluct.InsertIgnore).RunWith(r.tx).Exec(); err != nil {
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

	// r.st.
	return nil
}
