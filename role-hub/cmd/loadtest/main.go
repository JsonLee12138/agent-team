package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/JsonLee12138/agent-team/role-hub/internal/ingest"
)

func main() {
	target := getenvDefault("ROLE_HUB_LOADTEST_TARGET", "http://localhost:8080/api/v1/ingest")
	requests := getenvDefaultInt("ROLE_HUB_LOADTEST_REQUESTS", 500)
	concurrency := getenvDefaultInt("ROLE_HUB_LOADTEST_CONCURRENCY", 50)
	timeout := getenvDefaultDuration("ROLE_HUB_LOADTEST_TIMEOUT", 5*time.Second)
	idPrefix := getenvDefault("ROLE_HUB_LOADTEST_ID_PREFIX", "load")

	if requests <= 0 || concurrency <= 0 {
		fmt.Println("requests and concurrency must be > 0")
		os.Exit(2)
	}

	client := &http.Client{Timeout: timeout}
	jobs := make(chan int)
	var success uint64
	var failure uint64
	var statusCounts sync.Map

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for idx := range jobs {
				payload := ingest.IngestRequest{
					IdempotencyKey: fmt.Sprintf("%s-%d", idPrefix, idx),
					TraceID:        fmt.Sprintf("trace-%d", idx),
					Timestamp:      time.Now().UTC().Format(time.RFC3339),
					Query:          "loadtest",
					ResultCount:    1,
					Results: []ingest.IngestResult{{
						Repo:     "acme/roles",
						RolePath: "skills/backend",
						Name:     "backend",
					}},
				}
				body, _ := json.Marshal(payload)
				req, _ := http.NewRequest(http.MethodPost, target, bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				resp, err := client.Do(req)
				if err != nil {
					atomic.AddUint64(&failure, 1)
					incrementStatus(&statusCounts, "error")
					continue
				}
				_ = resp.Body.Close()
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					atomic.AddUint64(&success, 1)
				} else {
					atomic.AddUint64(&failure, 1)
				}
				incrementStatus(&statusCounts, strconv.Itoa(resp.StatusCode))
			}
		}(i)
	}

	start := time.Now()
	for i := 0; i < requests; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	dur := time.Since(start)

	total := success + failure
	rate := float64(success) / float64(total) * 100
	fmt.Printf("Requests: %d\n", total)
	fmt.Printf("Success:  %d\n", success)
	fmt.Printf("Failure:  %d\n", failure)
	fmt.Printf("Success Rate: %.2f%%\n", rate)
	fmt.Printf("Duration: %s\n", dur)
	fmt.Println("Status Counts:")
	statusCounts.Range(func(key, value any) bool {
		fmt.Printf("  %s: %d\n", key, value.(uint64))
		return true
	})
}

func incrementStatus(m *sync.Map, key string) {
	value, _ := m.LoadOrStore(key, uint64(0))
	m.Store(key, value.(uint64)+1)
}

func getenvDefault(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

func getenvDefaultInt(key string, def int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return def
}

func getenvDefaultDuration(key string, def time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if parsed, err := time.ParseDuration(val); err == nil {
			return parsed
		}
	}
	return def
}
