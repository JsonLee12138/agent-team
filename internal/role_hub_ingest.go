package internal

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	defaultRoleHubIngestURL = "https://role-hub.agent-team.dev/api/v1/ingest"
	roleHubIngestTimeout    = 5 * time.Second
	roleHubMaxRetries       = 2
	roleHubBaseBackoff      = 500 * time.Millisecond
)

// IngestPayload is the data sent to role-hub after a find operation.
type IngestPayload struct {
	IdempotencyKey string                  `json:"idempotency_key"`
	TraceID        string                  `json:"trace_id"`
	Timestamp      string                  `json:"timestamp"`
	Query          string                  `json:"query"`
	ResultCount    int                     `json:"result_count"`
	Results        []IngestSearchResultRef `json:"results"`
}

// IngestSearchResultRef is a lightweight reference to a search result.
type IngestSearchResultRef struct {
	Name      string `json:"name"`
	Repo      string `json:"repo"`
	RolePath  string `json:"role_path"`
	SourceURL string `json:"source_url"`
}

// IngestClient reports role-repo find results to the role-hub service.
type IngestClient struct {
	httpClient *http.Client
	baseURL    string
	logger     *log.Logger
	enabled    bool
}

// NewIngestClient creates an IngestClient.
// It is disabled when AGENT_TEAM_ROLE_HUB_URL is set to "off" or empty.
func NewIngestClient() *IngestClient {
	url := strings.TrimSpace(os.Getenv("AGENT_TEAM_ROLE_HUB_URL"))
	enabled := true
	if url == "off" {
		enabled = false
		url = ""
	}
	if url == "" && enabled {
		url = defaultRoleHubIngestURL
	}
	return &IngestClient{
		httpClient: &http.Client{Timeout: roleHubIngestTimeout},
		baseURL:    url,
		logger:     log.New(os.Stderr, "[role-hub-ingest] ", log.LstdFlags),
		enabled:    enabled,
	}
}

// NewIngestClientForTest creates an IngestClient pointing at a test server.
func NewIngestClientForTest(baseURL string, httpClient *http.Client) *IngestClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: roleHubIngestTimeout}
	}
	return &IngestClient{
		httpClient: httpClient,
		baseURL:    baseURL,
		logger:     log.New(io.Discard, "", 0),
		enabled:    true,
	}
}

// GenerateTraceID creates a unique trace identifier for the find session.
func GenerateTraceID() string {
	now := time.Now()
	b := make([]byte, 8)
	for i := range b {
		b[i] = byte(rand.IntN(256))
	}
	return fmt.Sprintf("find-%s-%s",
		now.Format("20060102T150405"),
		hex.EncodeToString(b),
	)
}

// GenerateIdempotencyKey creates a deterministic key from query + results
// so duplicate reports for the same search are safely deduplicated server-side.
func GenerateIdempotencyKey(query string, results []RoleRepoSearchResult) string {
	h := sha256.New()
	h.Write([]byte(query))
	h.Write([]byte{0})
	for _, r := range results {
		h.Write([]byte(r.Repo))
		h.Write([]byte{0})
		h.Write([]byte(r.RolePath))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))[:32]
}

// buildPayload constructs the ingest payload from search inputs/outputs.
func buildPayload(query string, results []RoleRepoSearchResult, traceID string) IngestPayload {
	refs := make([]IngestSearchResultRef, len(results))
	for i, r := range results {
		refs[i] = IngestSearchResultRef{
			Name:      r.Name,
			Repo:      r.Repo,
			RolePath:  r.RolePath,
			SourceURL: r.SourceURL,
		}
	}
	return IngestPayload{
		IdempotencyKey: GenerateIdempotencyKey(query, results),
		TraceID:        traceID,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		Query:          query,
		ResultCount:    len(results),
		Results:        refs,
	}
}

// ReportAsync fires off a non-blocking goroutine to report find results.
// It never blocks the caller and silently degrades on failure after retries.
// Returns a WaitGroup that tests can use to synchronize; callers may ignore it.
func (c *IngestClient) ReportAsync(query string, results []RoleRepoSearchResult, traceID string) *sync.WaitGroup {
	var wg sync.WaitGroup
	if !c.enabled || len(results) == 0 {
		return &wg
	}

	payload := buildPayload(query, results, traceID)

	wg.Add(1)
	go func() {
		defer wg.Done()
		c.reportWithRetry(payload)
	}()
	return &wg
}

// reportWithRetry attempts to POST the payload with exponential backoff.
func (c *IngestClient) reportWithRetry(payload IngestPayload) {
	body, err := json.Marshal(payload)
	if err != nil {
		c.logger.Printf("marshal error (trace=%s): %v", payload.TraceID, err)
		return
	}

	var lastErr error
	for attempt := 0; attempt <= roleHubMaxRetries; attempt++ {
		if attempt > 0 {
			backoff := roleHubBaseBackoff * time.Duration(1<<(attempt-1))
			jitter := time.Duration(rand.Int64N(int64(backoff) / 2))
			time.Sleep(backoff + jitter)
		}
		lastErr = c.doPost(payload.TraceID, body)
		if lastErr == nil {
			return
		}
		c.logger.Printf("attempt %d/%d failed (trace=%s): %v",
			attempt+1, roleHubMaxRetries+1, payload.TraceID, lastErr)
	}
	// Graceful degradation: log and move on; never surface to user.
	c.logger.Printf("all retries exhausted (trace=%s): %v", payload.TraceID, lastErr)
}

// doPost sends a single HTTP POST to the ingest endpoint.
func (c *IngestClient) doPost(traceID string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), roleHubIngestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "agent-team-cli")
	req.Header.Set("X-Trace-ID", traceID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("post ingest: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))

	if resp.StatusCode >= 500 {
		return fmt.Errorf("server error %d", resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		// Client errors (4xx) are not retryable; treat as permanent failure.
		c.logger.Printf("non-retryable error %d (trace=%s)", resp.StatusCode, traceID)
		return nil
	}
	return nil
}
