package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
)

const schema = `
CREATE TABLE IF NOT EXISTS payloads (
	id TEXT PRIMARY KEY,
	event TEXT NOT NULL,
	raw_data TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL
);
`

type DB struct {
	*sql.DB
}

var (
	ErrEmailAlreadyExists = errors.New("email address (or username) is already in use")
	ErrUserNotFound       = errors.New("user not found")
	ErrPayloadNotFound    = errors.New("payload not found")
)

func InitDB(ctx context.Context, filepath string) (*sql.DB, error) {
	// Note: mattn/go-sqlite3 registers as "sqlite3", not "sqlite"
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	log.Println("Sqlite database connection established.")
	return db, nil
}
