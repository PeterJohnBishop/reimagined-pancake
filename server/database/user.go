package database

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Password string    `json:"-"`
}

// CREATE
func (store *DBStore) CreateUser(name, email, hashedPassword string) (*User, error) {
	var user User

	query := `
		INSERT INTO users (name, email, password) 
		VALUES ($1, $2, $3) 
		RETURNING id, name, email, password`

	err := store.DB.QueryRow(query, name, email, hashedPassword).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GET USER BY ID
func (store *DBStore) GetUserByID(id uuid.UUID) (*User, error) {
	var user User
	query := `SELECT id, name, email, password FROM users WHERE id = $1`

	err := store.DB.QueryRow(query, id).Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// GET USER BY EMAIL
func (store *DBStore) GetUserByEmail(email string) (*User, error) {
	var user User
	query := `SELECT id, name, email, password FROM users WHERE email = $1`

	err := store.DB.QueryRow(query, email).Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GET ALL USERS
func (store *DBStore) GetAllUsers() ([]User, error) {
	query := `SELECT id, name, email, password FROM users`

	rows, err := store.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Password)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if users == nil {
		users = []User{}
	}

	return users, nil
}

// UPDATE
func (store *DBStore) UpdateUser(id uuid.UUID, name, email string) (*User, error) {
	var user User

	query := `
		UPDATE users 
		SET name = $1, email = $2 
		WHERE id = $3 
		RETURNING id, name, email, password`

	err := store.DB.QueryRow(query, name, email, id).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no user found to update")
		}
		return nil, err
	}

	return &user, nil
}

// UPDATE PASSWORD
func (store *DBStore) UpdateUserPassword(id uuid.UUID, newHashedPassword string) error {
	query := `UPDATE users SET password = $1 WHERE id = $2`

	result, err := store.DB.Exec(query, newHashedPassword, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("no user found to update")
	}
	return nil
}

// DELETE
func (store *DBStore) DeleteUser(id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := store.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("no user found to delete")
	}
	return nil
}
