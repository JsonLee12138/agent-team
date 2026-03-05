package ingest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/JsonLee12138/agent-team/role-hub/internal/storage"
)

type Store interface {
	GetIngestEvent(ctx context.Context, key string) (*storage.IngestEvent, error)
	InsertIngestEvent(ctx context.Context, key string, responseCode int, responseBody []byte) error
	UpsertRoleRecord(ctx context.Context, record storage.RoleRecord) error
}

type Handler struct {
	store         Store
	maxBodyBytes  int64
	maxResults    int
	rateLimiter   *RateLimiter
	cleanupTicker *time.Ticker
}

func NewHandler(store Store, maxBodyBytes int64, maxResults int, limiter *RateLimiter) *Handler {
	h := &Handler{
		store:        store,
		maxBodyBytes: maxBodyBytes,
		maxResults:   maxResults,
		rateLimiter:  limiter,
	}
	if limiter != nil {
		h.cleanupTicker = time.NewTicker(10 * time.Minute)
		go func() {
			for range h.cleanupTicker.C {
				limiter.Cleanup(30 * time.Minute)
			}
		}()
	}
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only POST is supported", nil)
		return
	}

	if h.maxBodyBytes > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, h.maxBodyBytes)
	}
	defer r.Body.Close()

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "unable to read request body", nil)
		return
	}

	if isLegacyPayload(payload) {
		writeError(w, http.StatusBadRequest, "UNSUPPORTED_PAYLOAD_VERSION", "legacy roles[] payload is not supported", nil)
		return
	}

	var req IngestRequest
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		if errors.Is(err, io.EOF) {
			writeError(w, http.StatusBadRequest, "INVALID_BODY", "empty request body", nil)
			return
		}
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "invalid JSON payload", nil)
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "invalid JSON payload", nil)
		return
	}

	issues := ValidateRequest(req, h.maxResults)
	if len(issues) > 0 {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "payload validation failed", issues)
		return
	}

	if event, err := h.store.GetIngestEvent(r.Context(), req.IdempotencyKey); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to check idempotency", nil)
		return
	} else if event != nil {
		writeCached(w, event)
		return
	}

	clientKey := clientKey(r)
	if h.rateLimiter != nil && !h.rateLimiter.Allow(clientKey) {
		writeError(w, http.StatusTooManyRequests, "RATE_LIMITED", "rate limit exceeded", nil)
		return
	}

	resp := h.process(r.Context(), req)
	body, err := json.Marshal(resp)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to encode response", nil)
		return
	}
	if err := h.store.InsertIngestEvent(r.Context(), req.IdempotencyKey, http.StatusOK, body); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to persist idempotency", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Idempotency-Key", req.IdempotencyKey)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func (h *Handler) process(ctx context.Context, req IngestRequest) IngestResponse {
	resp := IngestResponse{
		Status:         "ok",
		IdempotencyKey: req.IdempotencyKey,
	}

	for i, result := range req.Results {
		record := mapToRecord(result)
		if err := h.store.UpsertRoleRecord(ctx, record); err != nil {
			resp.Rejected++
			resp.Errors = append(resp.Errors, IngestFailure{Index: i, Repo: result.Repo, Message: "upsert failed"})
			continue
		}
		resp.Accepted++
	}
	return resp
}

func mapToRecord(result IngestResult) storage.RoleRecord {
	parts := strings.SplitN(result.Repo, "/", 2)
	owner, repo := "", ""
	if len(parts) == 2 {
		owner = parts[0]
		repo = parts[1]
	}
	return storage.RoleRecord{
		RepoOwner:   owner,
		RepoName:    repo,
		RolePath:    result.RolePath,
		Name:        result.Name,
		Description: result.Description,
		SourceURL:   result.SourceURL,
		Score:       result.Score,
		Tags:        result.Tags,
	}
}

func isLegacyPayload(payload []byte) bool {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		return false
	}
	_, ok := raw["roles"]
	return ok
}

func writeCached(w http.ResponseWriter, event *storage.IngestEvent) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(event.ResponseCode)
	_, _ = w.Write(event.ResponseBody)
}

func writeError(w http.ResponseWriter, status int, code, message string, details []FieldDetail) {
	resp := ErrorResponse{Error: ErrorBody{Code: code, Message: message, Details: details}}
	body, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func clientKey(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}
