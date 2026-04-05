// Package db
package db

import "log"

func CreateTable() {
	query := `
	CREATE TABLE IF NOT EXISTS sessions (
		id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name        TEXT NOT NULL,
		created_at  TIMESTAMP DEFAULT NOW(),
		ended_at    TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS requests (
		id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		session_id  UUID REFERENCES sessions(id) ON DELETE CASCADE,
		method      TEXT NOT NULL,
		url         TEXT NOT NULL,
		headers     JSONB,
		body        TEXT,
		status_code INT,
		response    TEXT,
		latency_ms  INT,
		timestamp   TIMESTAMP DEFAULT NOW()
	);
	`
	_, err := DB.Exec(query)
	if err != nil {
		log.Fatal("Failed to create tables: ", err)
	}

	log.Println("Table ready")
}
