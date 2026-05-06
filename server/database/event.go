package database

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	Timestamp time.Time       `json:"timestamp" db:"timestamp"`
	EventType string          `json:"event-type" db:"event_type"`
	Data      json.RawMessage `json:"data" db:"data"`
}

func (store *DBStore) StoreEvent(eventType string, data []byte) (*Event, error) {
	query := `
        INSERT INTO events (event_type, data) 
        VALUES ($1, $2) 
        RETURNING id, timestamp, event_type, data
    `
	var event Event
	err := store.DB.QueryRow(
		query,
		eventType,
		data,
	).Scan(
		&event.ID,
		&event.Timestamp,
		&event.EventType,
		&event.Data,
	)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (store *DBStore) GetEventByID(id uuid.UUID) (*Event, error) {
	query := `SELECT id, timestamp, event_type, data FROM events WHERE id = $1`

	var e Event
	err := store.DB.QueryRow(query, id).Scan(
		&e.ID,
		&e.Timestamp,
		&e.EventType,
		&e.Data,
	)
	if err != nil {
		return &Event{}, err
	}
	return &e, nil
}

func (store *DBStore) GetEventsByType(eventType string) ([]Event, error) {
	query := `SELECT id, timestamp, event_type, data FROM events where event_type = $1`

	rows, err := store.DB.Query(query, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event

	for rows.Next() {
		var e Event

		err := rows.Scan(&e.ID, &e.Timestamp, &e.EventType, &e.Data)
		if err != nil {
			return nil, err
		}

		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (store *DBStore) GetAllEvents() ([]Event, error) {
	query := `SELECT id, timestamp, event_type, data FROM events`

	rows, err := store.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		err := rows.Scan(&event.ID, &event.Timestamp, &event.EventType, &event.Data)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if events == nil {
		events = []Event{}
	}

	return events, nil
}

func (store *DBStore) DeleteEvent(id uuid.UUID) error {
	query := `DELETE FROM events WHERE id = $1`

	result, err := store.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("no event found to delete")
	}
	return nil
}
