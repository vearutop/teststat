package sqlite

import "github.com/bool64/sqluct"

type Repository struct {
	st *sqluct.Storage

	tests     map[Hash]bool
	revisions map[Hash]bool
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

func (r *Repository) AddRun(run Run) error {
	_, err := r.st.InsertStmt(runsTable, run, sqluct.InsertIgnore).Exec()
	if err != nil {
		return err
	}

	if !r.tests[run.TestHash] {
		if _, err := r.st.InsertStmt(testsTable, Test{
			Hash:    run.TestHash,
			Package: run.Package,
			Test:    run.Fn,
		}, sqluct.InsertIgnore).Exec(); err != nil {
			return err
		}

		r.tests[run.TestHash] = true
	}

	if !r.revisions[run.RevisionHash] {
		if _, err := r.st.InsertStmt(revisionsTable, Revision{
			Hash:     run.RevisionHash,
			Revision: run.Revision,
		}, sqluct.InsertIgnore).Exec(); err != nil {
			return err
		}

		r.revisions[run.RevisionHash] = true
	}

	return nil
}

func (r *Repository) SyncTotals() error {
	//r.st.
	return nil
}
