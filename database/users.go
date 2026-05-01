package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reimagined-pancake/global"

	"github.com/mattn/go-sqlite3"
)

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

func GetUserByUsername(db *sql.DB, ctx context.Context, username string) (*global.User, error) {
	query := `SELECT id, username, email, password FROM users WHERE username = ?`
	row := db.QueryRowContext(ctx, query, username)

	var u global.User
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: username %s", ErrUserNotFound, username)
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
