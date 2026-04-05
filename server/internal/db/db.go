package db

import (
	"database/sql"
	"log"
	"os"

	"github.com/joss12/api-time-machine/internal/models"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to open DB connection: ", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("Failed to ping DB: ", err)
	}

	log.Println("->Connected to db...")
}

func SaveRequest(req models.Request, headersJSON []byte) error {
	query := `
	INSERT INTO requests (session_id, method, url, headers, body, status_code, response, latency_ms)
	VALUES ($1, $2, $3, $4::jsonb, $5, $6, $7, $8)
	`
	_, err := DB.Exec(query,
		req.SessionID,
		req.Method,
		req.URL,
		string(headersJSON),
		req.Body,
		req.StatusCode,
		req.Response,
		req.LatencyMs,
	)
	return err
}

func DeleteSession(id string) error {
	_, err := DB.Exec(`DELETE FROM sessions WHERE id = $1`, id)
	return err
}
