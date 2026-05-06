package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

type DBStore struct {
	DB *sql.DB
}

func ConnectPGDB() (*DBStore, error) {
	host := os.Getenv("PG_HOST")
	port := os.Getenv("PG_PORT")
	user := os.Getenv("PG_USER")
	password := os.Getenv("PG_PASSWORD")
	dbname := os.Getenv("PG_DB")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("Successfully connected to the PostgreSQL Docker container!")

	return &DBStore{DB: db}, nil
}

func (store *DBStore) CreateTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS events (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMPTZ DEFAULT NOW(),    
    event_type VARCHAR(255) NOT NULL,       
    data JSONB NOT NULL                     
	);

	-- Fast lookups by event type
	CREATE INDEX idx_events_type ON events (event_type);

	-- Fast sorting by time
	CREATE INDEX idx_events_time ON events (timestamp DESC);

	-- Fast searching inside the JSON payload
	CREATE INDEX idx_events_data ON events USING GIN (data);
	`

	_, err := store.DB.Exec(query)
	if err != nil {
		return err
	}

	return nil
}
