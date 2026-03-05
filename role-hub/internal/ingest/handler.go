package ingest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/JsonLee12138/agent-team/role-hub/internal/storage"
)

type Store interface {
	GetIngestEvent(ctx context.Context, key string) (*storage.IngestEvent, error)
	InsertIngestEvent(ctx context.Context, key string, responseCode int, responseBody []byte) error
	UpsertRoleRecords(ctx context.Context, records []storage.RoleRecord) []error
}

type Handler struct {
	store         Store
	maxBodyBytes  int64
	maxResults    int
	rateLimiter   *RateLimiter
	cleanupTicker *time.Ticker
	inflight      chan struct{}
	dbTimeout     time.Duration
}

func NewHandler(store Store, maxBodyBytes int64, maxResults int, limiter *RateLimiter, maxInflight int, dbTimeout time.Duration) *Handler {
	h := &Handler{
		store:        store,
		maxBodyBytes: maxBodyBytes,
		maxResults:   maxResults,
		rateLimiter:  limiter,
		dbTimeout:    dbTimeout,
	}
	if maxInflight > 0 {
		h.inflight = make(chan struct{}, maxInflight)
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

	rawMap, rawErr := decodeRaw(payload)
	if rawErr == nil && isLegacyPayload(rawMap) {
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

	missingFields := []string(nil)
	if rawErr == nil {
		missingFields = findMissingFields(rawMap)
	}
	issues := ValidateRequest(req, h.maxResults)
	if len(issues) > 0 || len(missingFields) > 0 {
		writeErrorWithMissing(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "payload validation failed", issues, missingFields)
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

	if !h.acquireInflight() {
		writeError(w, http.StatusTooManyRequests, "CONCURRENCY_LIMIT", "too many concurrent requests", nil)
		return
	}
	defer h.releaseInflight()

	ctx := r.Context()
	if h.dbTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, h.dbTimeout)
		defer cancel()
	}

	resp := h.process(ctx, req)
	body, err := json.Marshal(resp)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to encode response", nil)
		return
	}
	if err := h.store.InsertIngestEvent(ctx, req.IdempotencyKey, http.StatusOK, body); err != nil {
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

	records := make([]storage.RoleRecord, 0, len(req.Results))
	for _, result := range req.Results {
		records = append(records, mapToRecord(result))
	}
	errs := h.store.UpsertRoleRecords(ctx, records)
	for i, result := range req.Results {
		var err error
		if i < len(errs) {
			err = errs[i]
		}
		if err != nil {
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

func decodeRaw(payload []byte) (map[string]json.RawMessage, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func isLegacyPayload(raw map[string]json.RawMessage) bool {
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

func writeErrorWithMissing(w http.ResponseWriter, status int, code, message string, details []FieldDetail, missing []string) {
	if len(missing) == 0 {
		writeError(w, status, code, message, details)
		return
	}
	resp := ErrorResponse{Error: ErrorBody{Code: code, Message: message, Details: details, MissingFields: missing}}
	body, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func (h *Handler) acquireInflight() bool {
	if h.inflight == nil {
		return true
	}
	select {
	case h.inflight <- struct{}{}:
		return true
	default:
		return false
	}
}

func (h *Handler) releaseInflight() {
	if h.inflight == nil {
		return
	}
	select {
	case <-h.inflight:
	default:
	}
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

func findMissingFields(raw map[string]json.RawMessage) []string {
	missing := make(map[string]struct{})
	checkRequired(raw, missing, "idempotency_key", "trace_id", "timestamp", "query", "result_count", "results")

	if rawResults, ok := raw["results"]; ok && !isMissing(rawResults) {
		var items []map[string]json.RawMessage
		if err := json.Unmarshal(rawResults, &items); err == nil {
			for i, item := range items {
				prefix := "results[" + strconv.Itoa(i) + "]"
				checkRequired(item, missing, prefix+".repo", prefix+".role_path")
			}
		}
	}

	if len(missing) == 0 {
		return nil
	}
	ordered := make([]string, 0, len(missing))
	for field := range missing {
		ordered = append(ordered, field)
	}
	sort.Strings(ordered)
	return ordered
}

func checkRequired(raw map[string]json.RawMessage, missing map[string]struct{}, fields ...string) {
	for _, field := range fields {
		key := field
		if dot := strings.LastIndex(field, "."); dot != -1 {
			key = field[dot+1:]
		}
		if val, ok := raw[key]; !ok || isMissing(val) {
			missing[field] = struct{}{}
		}
	}
}

func isMissing(raw json.RawMessage) bool {
	if len(bytes.TrimSpace(raw)) == 0 {
		return true
	}
	trimmed := bytes.TrimSpace(raw)
	return bytes.Equal(trimmed, []byte("null"))
}
