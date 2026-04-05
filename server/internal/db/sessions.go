package db

import (
	"context"
	"encoding/json"

	"github.com/joss12/api-time-machine/internal/models"
)

func CreateSession(name string) (models.Session, error) {
	var s models.Session
	query := `
	INSERT INTO sessions (name)
	VALUES ($1)
	RETURNING id, name, created_at, ended_at
	`
	err := DB.QueryRow(query, name).Scan(&s.ID, &s.Name, &s.CreatedAt, &s.EndedAt)
	return s, err
}

func EndSession(id string) error {
	_, err := DB.Exec(`UPDATE sessions SET ended_at = NOW() WHERE id = $1`, id)
	return err
}

func GetAllSessions() ([]models.Session, error) {
	rows, err := DB.Query(`SELECT id, name, created_at, ended_at FROM sessions ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		var s models.Session
		err := rows.Scan(&s.ID, &s.Name, &s.CreatedAt, &s.EndedAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func GetSessionRequests(ctx context.Context, sessionID string) ([]models.Request, error) {
	rows, err := DB.QueryContext(ctx, `
	SELECT id, session_id, method, url, headers, body, status_code, response, latency_ms, timestamp
	FROM requests
	WHERE session_id = $1
	ORDER BY timestamp ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []models.Request
	for rows.Next() {
		var r models.Request
		var headersJSON []byte
		err := rows.Scan(
			&r.ID,
			&r.SessionID,
			&r.Method,
			&r.URL,
			&headersJSON,
			&r.Body,
			&r.StatusCode,
			&r.Response,
			&r.LatencyMs,
			&r.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		if headersJSON != nil {
			json.Unmarshal(headersJSON, &r.Headers)
		}
		requests = append(requests, r)
	}
	return requests, nil
}

func ClosestaleSession() error {

	_, err := DB.Exec(`
	UPDATE sessions
	SET ended_at = NOW()
	WHERE ended_at IS NULL
	AND id NOT IN (
		SELECT DISTINCT session_id FROM requests
		WHERE timestamp > NOW() - INTERVAL '1 hour'
	)
	AND created_at < NOW() - INTERVAL '1 hour'
	`)
	return err
}
