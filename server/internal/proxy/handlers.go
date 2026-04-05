// Package proxy
package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/joss12/api-time-machine/internal/db"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	respond(w, 200, map[string]string{"status": "ok"})
}

func respond(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func StartSessionHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name   string `json:"name"`
		Target string `json:"target"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respond(w, 400, map[string]string{"error": "invalid body"})
		return
	}

	session, err := db.CreateSession(body.Name)
	if err != nil {
		respond(w, 500, map[string]string{"error": err.Error()})
		return
	}
	respond(w, 201, session)
}

func EndSessionHanlder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := db.EndSession(id); err != nil {
		respond(w, 500, map[string]string{"error": err.Error()})
		return
	}
	respond(w, 200, map[string]string{"status": "session ended"})
}

func ListSessionsHandler(w http.ResponseWriter, r *http.Request) {
	session, err := db.GetAllSessions()
	if err != nil {
		respond(w, 500, map[string]string{"error": err.Error()})
		return
	}
	respond(w, 200, session)
}

func GetSessionRequestsHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	requests, err := db.GetSessionRequests(r.Context(), id)
	if err != nil {
		respond(w, 500, map[string]string{"error": err.Error()})
		return
	}
	respond(w, 200, requests)
}

func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	target := r.URL.Query().Get("target")

	if sessionID == "" || target == "" {
		respond(w, 400, map[string]string{"error": "session_id and target are required"})
		return
	}

	q := r.URL.Query()
	q.Del("session_id")
	q.Del("target")
	r.URL.RawQuery = q.Encode()

	path := strings.TrimPrefix(r.URL.Path, "/proxy")

	p := New(sessionID, target)
	p.ServeHTTP(w, r, path)
}

func ReplaySessionHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	speed := 1.0
	if s := r.URL.Query().Get("speed"); s != "" {
		fmt.Sscanf(s, "%f", &speed)
	}

	// Run replay in background
	go func() {
		ctx := context.Background()
		requests, err := db.GetSessionRequests(ctx, id)
		if err != nil {
			log.Printf(" Replay failed: %v", err)
			return
		}
		ReplaySession(id, speed)
		log.Printf(" Replay done for session %s — %d requests", id, len(requests))
	}()

	respond(w, 200, map[string]string{"status": "replay started"})
}

func DeleteSessionHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := db.DeleteSession(id); err != nil {
		respond(w, 500, map[string]string{"error": err.Error()})
		return
	}
	respond(w, 200, map[string]string{"status": "deleted"})
}
