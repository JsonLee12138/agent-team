package internal

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestGenerateIdempotencyKeyDeterministic(t *testing.T) {
	results := []RoleRepoSearchResult{
		{Name: "frontend", Repo: "acme/roles", RolePath: "skills/frontend"},
		{Name: "backend", Repo: "acme/roles", RolePath: "skills/backend"},
	}
	key1 := GenerateIdempotencyKey("test-query", results, "trace-a")
	key2 := GenerateIdempotencyKey("test-query", results, "trace-a")
	if key1 != key2 {
		t.Fatalf("idempotency keys should be deterministic: %s != %s", key1, key2)
	}
	if len(key1) != 32 {
		t.Fatalf("idempotency key length = %d, want 32", len(key1))
	}
}

func TestGenerateIdempotencyKeyDiffersOnQuery(t *testing.T) {
	results := []RoleRepoSearchResult{
		{Name: "frontend", Repo: "acme/roles", RolePath: "skills/frontend"},
	}
	key1 := GenerateIdempotencyKey("query-a", results, "trace-a")
	key2 := GenerateIdempotencyKey("query-b", results, "trace-a")
	if key1 == key2 {
		t.Fatal("different queries should produce different keys")
	}
}

func TestGenerateIdempotencyKeyDiffersOnResults(t *testing.T) {
	resultsA := []RoleRepoSearchResult{
		{Name: "frontend", Repo: "acme/roles", RolePath: "skills/frontend"},
	}
	resultsB := []RoleRepoSearchResult{
		{Name: "backend", Repo: "acme/roles", RolePath: "skills/backend"},
	}
	key1 := GenerateIdempotencyKey("same-query", resultsA, "trace-a")
	key2 := GenerateIdempotencyKey("same-query", resultsB, "trace-a")
	if key1 == key2 {
		t.Fatal("different results should produce different keys")
	}
}

func TestGenerateTraceIDFormat(t *testing.T) {
	id := GenerateTraceID()
	if !strings.HasPrefix(id, "find-") {
		t.Fatalf("trace ID should start with 'find-': %s", id)
	}
	// Format: find-20060102T150405-<16 hex chars>
	parts := strings.SplitN(id, "-", 3)
	if len(parts) != 3 {
		t.Fatalf("trace ID should have 3 parts: %s", id)
	}
	if len(parts[2]) != 16 {
		t.Fatalf("random suffix should be 16 hex chars, got %d: %s", len(parts[2]), parts[2])
	}
}

func TestGenerateTraceIDUnique(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := GenerateTraceID()
		if ids[id] {
			t.Fatalf("duplicate trace ID: %s", id)
		}
		ids[id] = true
	}
}

func TestIngestClientReportAsyncSuccess(t *testing.T) {
	var received atomic.Int32
	var lastPayload IngestPayload

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content-type")
		}
		if r.Header.Get("X-Trace-ID") == "" {
			t.Error("missing X-Trace-ID header")
		}

		var p IngestPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		lastPayload = p
		received.Add(1)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	client := NewIngestClientForTest(srv.URL, srv.Client())
	results := []RoleRepoSearchResult{
		{Name: "frontend", Repo: "acme/roles", RolePath: "skills/frontend", SourceURL: "https://github.com/acme/roles"},
	}

	wg := client.ReportAsync("test-query", results, "trace-test-123")
	wg.Wait()

	if received.Load() != 1 {
		t.Fatalf("expected 1 request, got %d", received.Load())
	}
	if lastPayload.Query != "test-query" {
		t.Fatalf("payload query = %q, want %q", lastPayload.Query, "test-query")
	}
	if lastPayload.TraceID != "trace-test-123" {
		t.Fatalf("payload trace_id = %q, want %q", lastPayload.TraceID, "trace-test-123")
	}
	if lastPayload.ResultCount != 1 {
		t.Fatalf("payload result_count = %d, want 1", lastPayload.ResultCount)
	}
	if lastPayload.IdempotencyKey == "" {
		t.Fatal("payload idempotency_key should not be empty")
	}
	if lastPayload.Timestamp == "" {
		t.Fatal("payload timestamp should not be empty")
	}
	if _, err := time.Parse(time.RFC3339, lastPayload.Timestamp); err != nil {
		t.Fatalf("timestamp not RFC3339: %v", err)
	}
}

func TestIngestClientReportAsyncRetriesOnServerError(t *testing.T) {
	var attempts atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := attempts.Add(1)
		if count <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	client := NewIngestClientForTest(srv.URL, srv.Client())
	results := []RoleRepoSearchResult{
		{Name: "frontend", Repo: "acme/roles", RolePath: "skills/frontend"},
	}

	wg := client.ReportAsync("retry-query", results, "trace-retry")
	wg.Wait()

	got := attempts.Load()
	if got != 3 {
		t.Fatalf("expected 3 attempts (1 initial + 2 retries), got %d", got)
	}
}

func TestIngestClientReportAsyncDegracesAfterMaxRetries(t *testing.T) {
	var attempts atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewIngestClientForTest(srv.URL, srv.Client())
	results := []RoleRepoSearchResult{
		{Name: "frontend", Repo: "acme/roles", RolePath: "skills/frontend"},
	}

	wg := client.ReportAsync("degrade-query", results, "trace-degrade")
	wg.Wait()

	// 1 initial + 2 retries = 3 total
	got := attempts.Load()
	if got != 3 {
		t.Fatalf("expected 3 total attempts, got %d", got)
	}
	// No panic, no error returned — graceful degradation.
}

func TestIngestClientReportAsyncNoRetryOn4xx(t *testing.T) {
	var attempts atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	client := NewIngestClientForTest(srv.URL, srv.Client())
	results := []RoleRepoSearchResult{
		{Name: "frontend", Repo: "acme/roles", RolePath: "skills/frontend"},
	}

	wg := client.ReportAsync("bad-query", results, "trace-4xx")
	wg.Wait()

	got := attempts.Load()
	if got != 1 {
		t.Fatalf("4xx should not retry, expected 1 attempt, got %d", got)
	}
}

func TestIngestClientReportAsyncSkipsWhenDisabled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not call server when disabled")
	}))
	defer srv.Close()

	client := NewIngestClientForTest(srv.URL, srv.Client())
	client.enabled = false

	results := []RoleRepoSearchResult{
		{Name: "frontend", Repo: "acme/roles", RolePath: "skills/frontend"},
	}
	wg := client.ReportAsync("disabled-query", results, "trace-disabled")
	wg.Wait()
}

func TestIngestClientReportAsyncSkipsEmptyResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not call server for empty results")
	}))
	defer srv.Close()

	client := NewIngestClientForTest(srv.URL, srv.Client())
	wg := client.ReportAsync("empty-query", nil, "trace-empty")
	wg.Wait()
}

func TestBuildPayloadFields(t *testing.T) {
	results := []RoleRepoSearchResult{
		{Name: "a", Repo: "acme/r", RolePath: "skills/a", SourceURL: "https://github.com/acme/r"},
		{Name: "b", Repo: "other/r2", RolePath: "skills/b", SourceURL: "https://github.com/other/r2"},
	}

	payload := buildPayload("test", results, "trace-build")
	if payload.Query != "test" {
		t.Fatalf("query = %q, want %q", payload.Query, "test")
	}
	if payload.TraceID != "trace-build" {
		t.Fatalf("traceID = %q, want %q", payload.TraceID, "trace-build")
	}
	if payload.ResultCount != 2 {
		t.Fatalf("result_count = %d, want 2", payload.ResultCount)
	}
	if len(payload.Results) != 2 {
		t.Fatalf("results len = %d, want 2", len(payload.Results))
	}
	if payload.Results[0].Name != "a" || payload.Results[1].Repo != "other/r2" {
		t.Fatalf("unexpected results: %+v", payload.Results)
	}
	if payload.IdempotencyKey == "" {
		t.Fatal("idempotency key should not be empty")
	}
}

func TestBuildPayloadIdempotencyKeyDiffersAcrossInstallEvents(t *testing.T) {
	results := []RoleRepoSearchResult{
		{Name: "frontend", Repo: "acme/r", RolePath: "skills/frontend", SourceURL: "https://github.com/acme/r"},
	}

	first := buildPayload("install:acme/r", results, "trace-install-a")
	second := buildPayload("install:acme/r", results, "trace-install-b")

	if first.IdempotencyKey == second.IdempotencyKey {
		t.Fatalf("install events with different trace IDs should have different idempotency keys: %s", first.IdempotencyKey)
	}
}

func TestGenerateIdempotencyKeyDiffersOnTraceID(t *testing.T) {
	results := []RoleRepoSearchResult{
		{Name: "frontend", Repo: "acme/roles", RolePath: "skills/frontend"},
	}

	keyA := GenerateIdempotencyKey("install:acme/roles", results, "trace-a")
	keyB := GenerateIdempotencyKey("install:acme/roles", results, "trace-b")
	if keyA == keyB {
		t.Fatal("different trace IDs should produce different keys")
	}
}

func TestNewIngestClientDefaultsToVercelIngestURL(t *testing.T) {
	orig := os.Getenv("AGENT_TEAM_ROLE_HUB_URL")
	t.Cleanup(func() {
		if orig == "" {
			os.Unsetenv("AGENT_TEAM_ROLE_HUB_URL")
			return
		}
		os.Setenv("AGENT_TEAM_ROLE_HUB_URL", orig)
	})
	os.Unsetenv("AGENT_TEAM_ROLE_HUB_URL")

	client := NewIngestClient()
	if client.baseURL != "https://role-hub.vercel.app/api/v1/ingest" {
		t.Fatalf("default baseURL=%q, want %q", client.baseURL, "https://role-hub.vercel.app/api/v1/ingest")
	}
}

func TestNewIngestClientNormalizesRootURLToIngestPath(t *testing.T) {
	orig := os.Getenv("AGENT_TEAM_ROLE_HUB_URL")
	t.Cleanup(func() {
		if orig == "" {
			os.Unsetenv("AGENT_TEAM_ROLE_HUB_URL")
			return
		}
		os.Setenv("AGENT_TEAM_ROLE_HUB_URL", orig)
	})
	os.Setenv("AGENT_TEAM_ROLE_HUB_URL", "https://role-hub.vercel.app/")

	client := NewIngestClient()
	if client.baseURL != "https://role-hub.vercel.app/api/v1/ingest" {
		t.Fatalf("normalized baseURL=%q, want %q", client.baseURL, "https://role-hub.vercel.app/api/v1/ingest")
	}
}

func TestNewIngestClientUsesDefaultTimeout(t *testing.T) {
	orig := os.Getenv("AGENT_TEAM_ROLE_HUB_TIMEOUT")
	t.Cleanup(func() {
		if orig == "" {
			os.Unsetenv("AGENT_TEAM_ROLE_HUB_TIMEOUT")
			return
		}
		os.Setenv("AGENT_TEAM_ROLE_HUB_TIMEOUT", orig)
	})
	os.Unsetenv("AGENT_TEAM_ROLE_HUB_TIMEOUT")

	client := NewIngestClient()
	if client.httpClient.Timeout != 15*time.Second {
		t.Fatalf("default timeout=%v, want %v", client.httpClient.Timeout, 15*time.Second)
	}
}

func TestNewIngestClientUsesTimeoutFromEnv(t *testing.T) {
	orig := os.Getenv("AGENT_TEAM_ROLE_HUB_TIMEOUT")
	t.Cleanup(func() {
		if orig == "" {
			os.Unsetenv("AGENT_TEAM_ROLE_HUB_TIMEOUT")
			return
		}
		os.Setenv("AGENT_TEAM_ROLE_HUB_TIMEOUT", orig)
	})
	os.Setenv("AGENT_TEAM_ROLE_HUB_TIMEOUT", "30s")

	client := NewIngestClient()
	if client.httpClient.Timeout != 30*time.Second {
		t.Fatalf("configured timeout=%v, want %v", client.httpClient.Timeout, 30*time.Second)
	}
}

func TestParseProxyURLAddsScheme(t *testing.T) {
	parsed := parseProxyURL("127.0.0.1:7890")
	if parsed == nil {
		t.Fatal("expected proxy URL")
	}
	if parsed.Scheme != "http" || parsed.Host != "127.0.0.1:7890" {
		t.Fatalf("parsed=%s, want http://127.0.0.1:7890", parsed.String())
	}
}

func TestParseDarwinSystemProxySettings(t *testing.T) {
	raw := `
HTTPEnable : 1
HTTPProxy : 127.0.0.1
HTTPPort : 8080
HTTPSEnable : 1
HTTPSProxy : 127.0.0.1
HTTPSPort : 7890
SOCKSEnable : 1
SOCKSProxy : 127.0.0.1
SOCKSPort : 1080
`
	settings := parseDarwinSystemProxySettings(raw)
	if settings.HTTP == nil || settings.HTTPS == nil || settings.SOCKS == nil {
		t.Fatalf("unexpected settings: %+v", settings)
	}
	if settings.HTTP.Host != "127.0.0.1:8080" {
		t.Fatalf("http host=%q, want 127.0.0.1:8080", settings.HTTP.Host)
	}
	if settings.HTTPS.Host != "127.0.0.1:7890" {
		t.Fatalf("https host=%q, want 127.0.0.1:7890", settings.HTTPS.Host)
	}
	if settings.SOCKS.Scheme != "socks5" || settings.SOCKS.Host != "127.0.0.1:1080" {
		t.Fatalf("socks proxy=%v, want socks5://127.0.0.1:1080", settings.SOCKS)
	}
}

func TestNewRoleHubProxyFuncPrefersExplicitProxy(t *testing.T) {
	origProxyEnv := os.Getenv(roleHubProxyEnv)
	origProxyFromEnv := proxyFromEnvironment
	origLoadSystem := loadRoleHubSystemProxySettings
	t.Cleanup(func() {
		if origProxyEnv == "" {
			os.Unsetenv(roleHubProxyEnv)
		} else {
			os.Setenv(roleHubProxyEnv, origProxyEnv)
		}
		proxyFromEnvironment = origProxyFromEnv
		loadRoleHubSystemProxySettings = origLoadSystem
	})

	os.Setenv(roleHubProxyEnv, "127.0.0.1:7890")
	proxyFromEnvironment = func(*http.Request) (*neturl.URL, error) {
		return parseProxyURL("http://127.0.0.1:8080"), nil
	}
	loadRoleHubSystemProxySettings = func() roleHubSystemProxySettings {
		return roleHubSystemProxySettings{HTTPS: parseProxyURL("http://127.0.0.1:9090")}
	}

	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	got, err := newRoleHubProxyFunc()(req)
	if err != nil {
		t.Fatalf("proxy error: %v", err)
	}
	if got == nil || got.Host != "127.0.0.1:7890" {
		t.Fatalf("proxy host=%v, want 127.0.0.1:7890", got)
	}
}

func TestNewRoleHubProxyFuncFallsBackToSystemProxy(t *testing.T) {
	origProxyEnv := os.Getenv(roleHubProxyEnv)
	origProxyFromEnv := proxyFromEnvironment
	origLoadSystem := loadRoleHubSystemProxySettings
	t.Cleanup(func() {
		if origProxyEnv == "" {
			os.Unsetenv(roleHubProxyEnv)
		} else {
			os.Setenv(roleHubProxyEnv, origProxyEnv)
		}
		proxyFromEnvironment = origProxyFromEnv
		loadRoleHubSystemProxySettings = origLoadSystem
	})

	os.Unsetenv(roleHubProxyEnv)
	proxyFromEnvironment = func(*http.Request) (*neturl.URL, error) { return nil, nil }
	loadRoleHubSystemProxySettings = func() roleHubSystemProxySettings {
		return roleHubSystemProxySettings{HTTPS: parseProxyURL("http://127.0.0.1:7891")}
	}

	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	got, err := newRoleHubProxyFunc()(req)
	if err != nil {
		t.Fatalf("proxy error: %v", err)
	}
	if got == nil || got.Host != "127.0.0.1:7891" {
		t.Fatalf("proxy host=%v, want 127.0.0.1:7891", got)
	}
}

func TestNewRoleHubProxyFuncFallsBackToSystemSocksProxy(t *testing.T) {
	origProxyEnv := os.Getenv(roleHubProxyEnv)
	origProxyFromEnv := proxyFromEnvironment
	origLoadSystem := loadRoleHubSystemProxySettings
	t.Cleanup(func() {
		if origProxyEnv == "" {
			os.Unsetenv(roleHubProxyEnv)
		} else {
			os.Setenv(roleHubProxyEnv, origProxyEnv)
		}
		proxyFromEnvironment = origProxyFromEnv
		loadRoleHubSystemProxySettings = origLoadSystem
	})

	os.Unsetenv(roleHubProxyEnv)
	proxyFromEnvironment = func(*http.Request) (*neturl.URL, error) { return nil, nil }
	loadRoleHubSystemProxySettings = func() roleHubSystemProxySettings {
		return roleHubSystemProxySettings{SOCKS: parseProxyURL("socks5://127.0.0.1:1080")}
	}

	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	got, err := newRoleHubProxyFunc()(req)
	if err != nil {
		t.Fatalf("proxy error: %v", err)
	}
	if got == nil || got.Scheme != "socks5" || got.Host != "127.0.0.1:1080" {
		t.Fatalf("proxy=%v, want socks5://127.0.0.1:1080", got)
	}
}

func TestNewRoleHubProxyFuncPrefersEnvProxyOverSystem(t *testing.T) {
	origProxyEnv := os.Getenv(roleHubProxyEnv)
	origProxyFromEnv := proxyFromEnvironment
	origLoadSystem := loadRoleHubSystemProxySettings
	t.Cleanup(func() {
		if origProxyEnv == "" {
			os.Unsetenv(roleHubProxyEnv)
		} else {
			os.Setenv(roleHubProxyEnv, origProxyEnv)
		}
		proxyFromEnvironment = origProxyFromEnv
		loadRoleHubSystemProxySettings = origLoadSystem
	})

	os.Unsetenv(roleHubProxyEnv)
	proxyFromEnvironment = func(*http.Request) (*neturl.URL, error) {
		return parseProxyURL("http://127.0.0.1:8111"), nil
	}
	loadRoleHubSystemProxySettings = func() roleHubSystemProxySettings {
		return roleHubSystemProxySettings{HTTPS: parseProxyURL("http://127.0.0.1:7892")}
	}

	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	got, err := newRoleHubProxyFunc()(req)
	if err != nil {
		t.Fatalf("proxy error: %v", err)
	}
	if got == nil || got.Host != "127.0.0.1:8111" {
		t.Fatalf("proxy host=%v, want 127.0.0.1:8111", got)
	}
}
