package main

import (
	"database/sql"
	"fmt"
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
	hashed_password TEXT NOT NULL
);
`

type DB struct {
	*sql.DB
}

type Payload struct {
	ID      string
	Event   string
	RawData string
}

type User struct {
	ID             string
	Username       string
	Email          string
	HashedPassword string
}

func InitDB(filepath string) (*DB, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return &DB{db}, nil
}

func (db *DB) SavePayload(p Payload) error {
	query := `INSERT INTO payloads (id, event, raw_data) VALUES (?, ?, ?)`
	_, err := db.Exec(query, p.ID, p.Event, p.RawData)
	return err
}

func (db *DB) GetPayloadByID(id string) (*Payload, error) {
	query := `SELECT id, event, raw_data FROM payloads WHERE id = ?`
	row := db.QueryRow(query, id)

	var p Payload
	if err := row.Scan(&p.ID, &p.Event, &p.RawData); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payload with id %s not found", id)
		}
		return nil, err
	}
	return &p, nil
}

func (db *DB) GetAllPayloads() ([]Payload, error) {
	query := `SELECT id, event, raw_data FROM payloads`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payloads []Payload
	for rows.Next() {
		var p Payload
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

func (db *DB) DeletePayloadByID(id string) error {
	query := `DELETE FROM payloads WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}

func (db *DB) SaveUser(u User) error {
	query := `INSERT INTO users (id, username, email, hashed_password) VALUES (?, ?, ?, ?)`
	_, err := db.Exec(query, u.ID, u.Username, u.Email, u.HashedPassword)
	return err
}

func (db *DB) GetUserByID(id string) (*User, error) {
	query := `SELECT id, username, email, hashed_password FROM users WHERE id = ?`
	row := db.QueryRow(query, id)

	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.HashedPassword); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with id %s not found", id)
		}
		return nil, err
	}
	return &u, nil
}

func (db *DB) GetAllUsers() ([]User, error) {
	query := `SELECT id, username, email, hashed_password FROM users`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.HashedPassword); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (db *DB) UpdateUserByID(id string, u User) error {
	query := `UPDATE users SET username = ?, email = ?, hashed_password = ? WHERE id = ?`
	res, err := db.Exec(query, u.Username, u.Email, u.HashedPassword, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no user found with id %s", id)
	}

	return nil
}

func (db *DB) DeleteUserByID(id string) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}
