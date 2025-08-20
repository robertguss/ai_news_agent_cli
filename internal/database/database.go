package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	"github.com/robertguss/ai-news-agent-cli/pkg/errs"
	"github.com/robertguss/ai-news-agent-cli/pkg/logging"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

func Open(dataSource string) (*sql.DB, *Queries, error) {
	dsn := dataSource
	if dsn != ":memory:" && dsn != "" {
		dsn = fmt.Sprintf("%s?_busy_timeout=3000&_journal_mode=WAL&_synchronous=NORMAL", dataSource)
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		wrappedErr := errs.Wrap("open sqlite database", err)
		logging.Error("database_open", wrappedErr)
		return nil, nil, wrappedErr
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		wrappedErr := errs.Wrap("ping database", err)
		logging.Error("database_ping", wrappedErr)
		return nil, nil, wrappedErr
	}

	logging.Info("database_open", fmt.Sprintf("Successfully opened database: %s", dataSource))
	return db, New(db), nil
}

func InitSchema(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, schemaSQL)
	if err != nil {
		wrappedErr := errs.Wrap("initialize database schema", err)
		logging.Error("database_init_schema", wrappedErr)
		return wrappedErr
	}

	logging.Info("database_init_schema", "Database schema initialized successfully")
	return nil
}
