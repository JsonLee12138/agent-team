package internal

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

type apiTestEnvelope struct {
	Data  json.RawMessage `json:"data"`
	Error *apiTestError   `json:"error"`
}

type apiTestError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type apiTestRole struct {
	Name         string     `json:"name"`
	Source       string     `json:"source"`
	Status       string     `json:"status"`
	InstallCount int        `json:"install_count"`
	VerifiedAt   *time.Time `json:"verified_at"`
}

type apiTestList struct {
	Items  []apiTestRole `json:"items"`
	Total  int           `json:"total"`
	Status string        `json:"status"`
}

func TestCatalogAPIListDefaultVerified(t *testing.T) {
	catalogPath := writeTestCatalog(t)
	srv := httptest.NewServer(NewCatalogAPIHandler(catalogPath))
	defer srv.Close()

	resp := mustRequest(t, srv.URL+"/api/roles")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	var envelope apiTestEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if envelope.Error != nil {
		t.Fatalf("unexpected error: %+v", envelope.Error)
	}
	var payload apiTestList
	if err := json.Unmarshal(envelope.Data, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.Total != 1 || len(payload.Items) != 1 {
		t.Fatalf("expected 1 verified role, got %d", payload.Total)
	}
	if payload.Items[0].InstallCount != 3 {
		t.Fatalf("expected install_count=3, got %d", payload.Items[0].InstallCount)
	}
	if payload.Status != "verified" {
		t.Fatalf("expected status=verified, got %q", payload.Status)
	}
}

func TestCatalogAPISearchAllStatus(t *testing.T) {
	catalogPath := writeTestCatalog(t)
	srv := httptest.NewServer(NewCatalogAPIHandler(catalogPath))
	defer srv.Close()

	resp := mustRequest(t, srv.URL+"/api/roles/search?q=roles&status=all")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	var envelope apiTestEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	var payload apiTestList
	if err := json.Unmarshal(envelope.Data, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.Total != 2 {
		t.Fatalf("expected 2 roles, got %d", payload.Total)
	}
	if payload.Status != "all" {
		t.Fatalf("expected status=all, got %q", payload.Status)
	}
}

func TestCatalogAPIRoleDetail(t *testing.T) {
	catalogPath := writeTestCatalog(t)
	srv := httptest.NewServer(NewCatalogAPIHandler(catalogPath))
	defer srv.Close()

	resp := mustRequest(t, srv.URL+"/api/roles/frontend?source=acme/roles")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	var envelope apiTestEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	var payload struct {
		Item apiTestRole `json:"item"`
	}
	if err := json.Unmarshal(envelope.Data, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.Item.Name != "frontend" {
		t.Fatalf("expected frontend role, got %q", payload.Item.Name)
	}
	if payload.Item.VerifiedAt == nil {
		t.Fatal("expected verified_at to be present")
	}
}

func TestCatalogAPIRepoDetail(t *testing.T) {
	catalogPath := writeTestCatalog(t)
	srv := httptest.NewServer(NewCatalogAPIHandler(catalogPath))
	defer srv.Close()

	resp := mustRequest(t, srv.URL+"/api/repos/acme/roles")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	var envelope apiTestEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	var payload struct {
		Repo  string        `json:"repo"`
		Items []apiTestRole `json:"items"`
		Total int           `json:"total"`
	}
	if err := json.Unmarshal(envelope.Data, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.Repo != "acme/roles" {
		t.Fatalf("expected repo acme/roles, got %q", payload.Repo)
	}
	if payload.Total != 1 || len(payload.Items) != 1 {
		t.Fatalf("expected 1 verified role, got %d", payload.Total)
	}
}

func TestCatalogAPIStats(t *testing.T) {
	catalogPath := writeTestCatalog(t)
	srv := httptest.NewServer(NewCatalogAPIHandler(catalogPath))
	defer srv.Close()

	resp := mustRequest(t, srv.URL+"/api/stats")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	var envelope apiTestEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	var payload struct {
		Total        int            `json:"total"`
		Repositories int            `json:"repositories"`
		ByStatus     map[string]int `json:"by_status"`
	}
	if err := json.Unmarshal(envelope.Data, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.Total != 2 {
		t.Fatalf("expected total=2, got %d", payload.Total)
	}
	if payload.ByStatus["verified"] != 1 {
		t.Fatalf("expected verified=1, got %d", payload.ByStatus["verified"])
	}
}

func TestCatalogAPIInvalidStatus(t *testing.T) {
	catalogPath := writeTestCatalog(t)
	srv := httptest.NewServer(NewCatalogAPIHandler(catalogPath))
	defer srv.Close()

	resp := mustRequest(t, srv.URL+"/api/roles?status=bad")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	var envelope apiTestEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if envelope.Error == nil || envelope.Error.Code != "invalid_status" {
		t.Fatalf("expected invalid_status error, got %+v", envelope.Error)
	}
}

func writeTestCatalog(t *testing.T) string {
	t.Helper()
	now := time.Date(2026, 3, 5, 3, 4, 5, 0, time.UTC)
	verifiedAt := now.Add(-2 * time.Hour)
	catalog := RoleRepoCatalog{
		Version: RoleRepoCatalogVersion,
		Entries: []RoleRepoCatalogEntry{
			{
				Name:         "frontend",
				Source:       "acme/roles",
				SourceType:   "github",
				SourceURL:    "https://github.com/acme/roles",
				RolePath:     "skills/frontend",
				FolderHash:   "hash-frontend",
				Status:       CatalogStatusVerified,
				DiscoveredAt: now.Add(-24 * time.Hour),
				VerifiedAt:   &verifiedAt,
				UpdatedAt:    now,
				InstallCount: 3,
			},
			{
				Name:         "backend-api",
				Source:       "acme/roles",
				SourceType:   "github",
				SourceURL:    "https://github.com/acme/roles",
				RolePath:     "skills/backend-api",
				FolderHash:   "hash-backend",
				Status:       CatalogStatusDiscovered,
				DiscoveredAt: now.Add(-12 * time.Hour),
				UpdatedAt:    now,
				InstallCount: 1,
			},
		},
	}

	root := t.TempDir()
	catalogPath := filepath.Join(root, ".agents", "catalog.json")
	if err := WriteRoleRepoCatalog(catalogPath, catalog); err != nil {
		t.Fatalf("write catalog: %v", err)
	}
	return catalogPath
}

func mustRequest(t *testing.T, url string) *http.Response {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	return resp
}

func readBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return body
}
