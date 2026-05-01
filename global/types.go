package global

import "database/sql"

type DB struct {
	*sql.DB
}

type Payload struct {
	ID      string
	Event   string
	RawData string
}

type User struct {
	ID       string
	Username string
	Email    string
	Password string
}
