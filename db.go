package main

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite"
)

func initDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", "./dice.db")
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS rolls (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			player    TEXT    NOT NULL DEFAULT 'Anonymous',
			result    INTEGER NOT NULL,
			rolled_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}
