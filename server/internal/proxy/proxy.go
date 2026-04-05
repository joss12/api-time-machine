package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/joss12/api-time-machine/internal/db"
	"github.com/joss12/api-time-machine/internal/models"
)

type Proxy struct {
	SessionID string
	Target    string
}

func New(sessionID, target string) *Proxy {
	return &Proxy{
		SessionID: sessionID,
		Target:    target,
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request, path string) {
	start := time.Now()

	// Read body
	var bodyBytes []byte
	if r.Body != nil {
		bodyBytes, _ = io.ReadAll(r.Body)
	}

	// Capture headers
	headers := make(map[string]string)
	for k, v := range r.Header {
		headers[k] = v[0]
	}

	// Build full target URL
	targetURL := p.Target + path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	// Build outgoing request
	outReq, err := http.NewRequest(r.Method, targetURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		http.Error(w, "failed to build request", 500)
		return
	}

	// Copy headers
	for k, v := range r.Header {
		outReq.Header.Set(k, v[0])
	}

	// Fire request
	client := &http.Client{}
	resp, err := client.Do(outReq)
	if err != nil {
		http.Error(w, "failed to forward request", 502)
		return
	}
	defer resp.Body.Close()

	latency := int(time.Since(start).Milliseconds())

	// Read response
	respBody, _ := io.ReadAll(resp.Body)

	for k, v := range resp.Header {
		w.Header().Set(k, v[0])
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)

	// Save to DB
	headersJSON, _ := json.Marshal(headers)
	req := models.Request{
		SessionID:  p.SessionID,
		Method:     r.Method,
		URL:        targetURL,
		Headers:    headers,
		Body:       string(bodyBytes),
		StatusCode: resp.StatusCode,
		Response:   string(respBody),
		LatencyMs:  latency,
	}

	if err := db.SaveRequest(req, headersJSON); err != nil {
		log.Println("Failed to save request:", err)
	} else {
		log.Printf(" [%s] %s → %d (%dms)\n", req.Method, targetURL, req.StatusCode, req.LatencyMs)
	}
}
