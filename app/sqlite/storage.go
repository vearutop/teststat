package sqlite

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/bool64/sqluct"
	"github.com/jmoiron/sqlx"
	"github.com/vearutop/gooselite"
	"github.com/vearutop/gooselite/iofs"
	_ "modernc.org/sqlite" // SQLite3 driver.
)

const driverName = "sqlite"

func newStorage(fn string) (*sqluct.Storage, error) {
	db, err := sql.Open(driverName, fn+"?_time_format=sqlite")
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(time.Hour)

	st := sqluct.NewStorage(sqlx.NewDb(db, driverName))
	st.Mapper = &sqluct.Mapper{}

	st.Format = squirrel.Question
	st.IdentifierQuoter = sqluct.QuoteBackticks
	st.Mapper.Dialect = sqluct.DialectSQLite3
	dialect := "sqlite3"

	gooselite.SetLogger(log.New(io.Discard, "", 0))

	if err := gooselite.SetDialect(dialect); err != nil {
		return nil, fmt.Errorf("set migrations dialect: %w", err)
	}

	// Apply migrations.
	if err := iofs.Up(db, migrations, "migrations"); err != nil {
		return nil, fmt.Errorf("run up migrations: %w", err)
	}

	_, err = db.Exec("pragma journal_mode=off;")
	if err != nil {
		return nil, err
	}

	return st, nil
}
