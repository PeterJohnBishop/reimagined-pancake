package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reimagined-pancake/global"

	"github.com/mattn/go-sqlite3"
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

	return db, nil
}

func SavePayload(db *sql.DB, ctx context.Context, p global.Payload) error {
	query := `INSERT INTO payloads (id, event, raw_data) VALUES (?, ?, ?)`
	_, err := db.ExecContext(ctx, query, p.ID, p.Event, p.RawData)
	return err
}

func GetPayloadByID(db *sql.DB, ctx context.Context, id string) (*global.Payload, error) {
	query := `SELECT id, event, raw_data FROM payloads WHERE id = ?`
	row := db.QueryRowContext(ctx, query, id)

	var p global.Payload
	if err := row.Scan(&p.ID, &p.Event, &p.RawData); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: id %s", ErrPayloadNotFound, id)
		}
		return nil, err
	}
	return &p, nil
}

func GetAllPayloads(db *sql.DB, ctx context.Context) ([]global.Payload, error) {
	query := `SELECT id, event, raw_data FROM payloads`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payloads []global.Payload
	for rows.Next() {
		var p global.Payload
		if err := rows.Scan(&p.ID, &p.Event, &p.RawData); err != nil {
			return nil, err
		}
		payloads = append(payloads, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return payloads, nil
}

func DeletePayloadByID(db *sql.DB, ctx context.Context, id string) error {
	query := `DELETE FROM payloads WHERE id = ?`
	_, err := db.ExecContext(ctx, query, id)
	return err
}

func SaveUser(db *sql.DB, ctx context.Context, u global.User) error {
	query := `INSERT INTO users (id, username, email, password) VALUES (?, ?, ?, ?)`
	_, err := db.ExecContext(ctx, query, u.ID, u.Username, u.Email, u.Password)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				// Note: This could trigger for duplicate email OR duplicate username
				return ErrEmailAlreadyExists
			}
		}
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

func GetUserByID(db *sql.DB, ctx context.Context, id string) (*global.User, error) {
	query := `SELECT id, username, email, password FROM users WHERE id = ?`
	row := db.QueryRowContext(ctx, query, id)

	var u global.User
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: id %s", ErrUserNotFound, id)
		}
		return nil, err
	}
	return &u, nil
}

func GetUserByEmail(db *sql.DB, ctx context.Context, email string) (*global.User, error) {
	query := `SELECT id, username, email, password FROM users WHERE email = ?`
	row := db.QueryRowContext(ctx, query, email)

	var u global.User
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: email %s", ErrUserNotFound, email)
		}
		return nil, err
	}
	return &u, nil
}

func GetAllUser(db *sql.DB, ctx context.Context) ([]global.User, error) {
	query := `SELECT id, username, email, password FROM users`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []global.User
	for rows.Next() {
		var u global.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Password); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func UpdateUserByID(db *sql.DB, ctx context.Context, id string, u global.User) error {
	query := `UPDATE users SET username = ?, email = ?, password = ? WHERE id = ?`
	res, err := db.ExecContext(ctx, query, u.Username, u.Email, u.Password, id)
	if err != nil {
		// Also catch unique constraint here, in case they update to an existing email/username
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				return ErrEmailAlreadyExists
			}
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w: id %s", ErrUserNotFound, id)
	}

	return nil
}

func DeleteUserByID(db *sql.DB, ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := db.ExecContext(ctx, query, id)
	return err
}
