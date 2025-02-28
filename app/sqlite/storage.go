package sqlite

import (
	"database/sql"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/bool64/sqluct"
	"github.com/jmoiron/sqlx"
	"github.com/vearutop/gooselite"
	"github.com/vearutop/gooselite/iofs"
	_ "modernc.org/sqlite" // SQLite3 driver.
	"time"
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

	if err := gooselite.SetDialect(dialect); err != nil {
		return nil, fmt.Errorf("set migrations dialect: %w", err)
	}

	// Apply migrations.
	if err := iofs.Up(db, migrations, "migrations"); err != nil {
		return nil, fmt.Errorf("run up migrations: %w", err)
	}

	return st, nil
}
