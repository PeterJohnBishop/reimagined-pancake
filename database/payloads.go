package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reimagined-pancake/global"
)

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
