// Package models
package models

import "time"

type Session struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	EndedAt   *time.Time `json:"ended_at"`
}

type Request struct {
	ID         string            `json:"id"`
	SessionID  string            `json:"session_id"`
	Method     string            `json:"method"`
	URL        string            `json:"url"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	StatusCode int               `json:"status_code"`
	Response   string            `json:"response"`
	LatencyMs  int               `json:"latency_ms"`
	Timestamp  time.Time         `json:"timestamp"`
}
