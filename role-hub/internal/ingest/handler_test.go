package ingest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JsonLee12138/agent-team/role-hub/internal/storage"
)

func setupTestStore(t *testing.T) *storage.Store {
	t.Helper()
	ctx := context.Background()
	db, err := storage.Open(ctx, "sqlite", "file:rolehub_test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := storage.ApplyMigrations(ctx, db, "sqlite"); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	store, err := storage.NewStore(db, "sqlite")
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	return store
}

func TestIngestHandler_AcceptsNewPayload(t *testing.T) {
	store := setupTestStore(t)
	limiter := NewRateLimiter(10, 10)
	handler := NewHandler(store, 1<<20, 100, limiter)

	payload := IngestRequest{
		IdempotencyKey: "idem-1",
		TraceID:        "trace-1",
		Timestamp:      "2026-03-05T00:00:00Z",
		Query:          "search",
		ResultCount:    1,
		Results: []IngestResult{{
			Repo:     "acme/roles",
			RolePath: "skills/backend",
			Name:     "backend",
		}},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ingest", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var resp IngestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Accepted != 1 || resp.Rejected != 0 {
		t.Fatalf("unexpected counts: %+v", resp)
	}
}

func TestIngestHandler_RejectsLegacyPayload(t *testing.T) {
	store := setupTestStore(t)
	handler := NewHandler(store, 1<<20, 100, nil)

	legacy := []byte(`{"roles":[{"name":"legacy"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ingest", bytes.NewReader(legacy))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("UNSUPPORTED_PAYLOAD_VERSION")) {
		t.Fatalf("expected unsupported payload error, got %s", w.Body.String())
	}
}

func TestIngestHandler_Idempotency(t *testing.T) {
	store := setupTestStore(t)
	handler := NewHandler(store, 1<<20, 100, nil)

	payload := IngestRequest{
		IdempotencyKey: "idem-2",
		TraceID:        "trace-2",
		Timestamp:      "2026-03-05T00:00:00Z",
		Query:          "search",
		ResultCount:    1,
		Results: []IngestResult{{
			Repo:     "acme/roles",
			RolePath: "skills/backend",
			Name:     "backend",
		}},
	}
	body, _ := json.Marshal(payload)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ingest", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("attempt %d status = %d", i+1, w.Code)
		}
	}
}

func TestIngestHandler_RateLimit(t *testing.T) {
	store := setupTestStore(t)
	limiter := NewRateLimiter(0.1, 1)
	handler := NewHandler(store, 1<<20, 100, limiter)

	payload := IngestRequest{
		IdempotencyKey: "idem-3",
		TraceID:        "trace-3",
		Timestamp:      "2026-03-05T00:00:00Z",
		Query:          "search",
		ResultCount:    1,
		Results: []IngestResult{{
			Repo:     "acme/roles",
			RolePath: "skills/backend",
			Name:     "backend",
		}},
	}
	body, _ := json.Marshal(payload)

	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/ingest", bytes.NewReader(body))
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first status = %d", w1.Code)
	}

	payload.IdempotencyKey = "idem-4"
	body2, _ := json.Marshal(payload)
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/ingest", bytes.NewReader(body2))
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("second status = %d", w2.Code)
	}
}
