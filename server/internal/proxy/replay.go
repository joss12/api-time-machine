package proxy

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/joss12/api-time-machine/internal/db"
)

type ReplayResult struct {
	RequestID  string `json:"request_id"`
	Method     string `json:"method"`
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	LatencyMs  int    `json:"latency_ms"`
	Response   string `json:"response"`
}

func ReplaySession(sessionID string, speed float64) ([]ReplayResult, error) {
	log.Printf("Fetching requests for session %s", sessionID)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	requests, err := db.GetSessionRequests(ctx, sessionID)
	if err != nil {
		log.Printf("DB error: %v", err)
		return nil, err
	}

	log.Printf("Got %d requests to replay", len(requests))

	var results []ReplayResult

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for i, req := range requests {
		log.Printf("Replaying [%d] %s %s", i, req.Method, req.URL)

		start := time.Now()

		outReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bytes.NewBufferString(req.Body))
		if err != nil {
			log.Printf("Failed to build request: %v", err)
			continue
		}

		for k, v := range req.Headers {
			outReq.Header.Set(k, v)
		}

		log.Printf("Firing...")
		resp, err := client.Do(outReq)
		if err != nil {
			log.Printf("Request failed: %v", err)
			results = append(results, ReplayResult{
				RequestID:  req.ID,
				Method:     req.Method,
				URL:        req.URL,
				StatusCode: 0,
				LatencyMs:  int(time.Since(start).Milliseconds()),
				Response:   "request failed: " + err.Error(),
			})
			continue
		}

		log.Printf("Response: %d", resp.StatusCode)
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		results = append(results, ReplayResult{
			RequestID:  req.ID,
			Method:     req.Method,
			URL:        req.URL,
			StatusCode: resp.StatusCode,
			LatencyMs:  int(time.Since(start).Milliseconds()),
			Response:   string(respBody),
		})
	}

	log.Printf("Replay done. %d results", len(results))
	return results, nil
}
