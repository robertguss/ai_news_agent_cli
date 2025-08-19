package database

import (
	"context"
	"database/sql"
	_ "embed"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

func Open(dataSource string) (*sql.DB, *Queries, error) {
	db, err := sql.Open("sqlite", dataSource)
	if err != nil {
		return nil, nil, err
	}

	if err := db.PingContext(context.Background()); err != nil {
		return nil, nil, err
	}

	return db, New(db), nil
}

func InitSchema(db *sql.DB) error {
	_, err := db.Exec(schemaSQL)
	return err
}
