package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/joss12/api-time-machine/internal/db"
	"github.com/joss12/api-time-machine/internal/models"
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

		// Copy stored headers — skip Accept-Encoding
		for k, v := range req.Headers {
			if strings.ToLower(k) == "accept-encoding" {
				continue
			}
			outReq.Header.Set(k, v)
		}
		outReq.Header.Set("Accept-Encoding", "identity")

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

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		latency := int(time.Since(start).Milliseconds())

		log.Printf("Got response: %d", resp.StatusCode)

		// Save replayed request to DB
		headers := make(map[string]string)
		for k, v := range outReq.Header {
			headers[k] = v[0]
		}
		headersJSON, _ := json.Marshal(headers)
		newReq := models.Request{
			SessionID:  sessionID,
			Method:     req.Method,
			URL:        req.URL,
			Headers:    headers,
			Body:       req.Body,
			StatusCode: resp.StatusCode,
			Response:   sanitizeString(string(respBody)),
			LatencyMs:  latency,
		}
		if err := db.SaveRequest(newReq, headersJSON); err != nil {
			log.Printf("Failed to save replayed request: %v", err)
		}

		results = append(results, ReplayResult{
			RequestID:  req.ID,
			Method:     req.Method,
			URL:        req.URL,
			StatusCode: resp.StatusCode,
			LatencyMs:  latency,
			Response:   sanitizeString(string(respBody)),
		})
	}

	log.Printf("Replay done. %d results", len(results))
	return results, nil
}
