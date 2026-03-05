package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const catalogAPIBasePath = "/api"

// NewCatalogAPIHandler builds the HTTP handler for catalog queries.
func NewCatalogAPIHandler(catalogPath string) http.Handler {
	api := &catalogAPI{catalogPath: catalogPath}
	mux := http.NewServeMux()
	mux.HandleFunc(catalogAPIBasePath+"/roles/search", api.handleRoleSearch)
	mux.HandleFunc(catalogAPIBasePath+"/roles/", api.handleRoleDetail)
	mux.HandleFunc(catalogAPIBasePath+"/roles", api.handleRoleList)
	mux.HandleFunc(catalogAPIBasePath+"/repos/", api.handleRepoDetail)
	mux.HandleFunc(catalogAPIBasePath+"/stats", api.handleStats)
	return mux
}

type catalogAPI struct {
	catalogPath string
}

type apiEnvelope struct {
	Data  any       `json:"data,omitempty"`
	Error *apiError `json:"error,omitempty"`
}

type apiError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

type roleDTO struct {
	Name         string     `json:"name"`
	Source       string     `json:"source"`
	SourceType   string     `json:"source_type"`
	SourceURL    string     `json:"source_url"`
	RolePath     string     `json:"role_path"`
	FolderHash   string     `json:"folder_hash"`
	Status       string     `json:"status"`
	StatusReason string     `json:"status_reason,omitempty"`
	DiscoveredAt time.Time  `json:"discovered_at"`
	VerifiedAt   *time.Time `json:"verified_at,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
	InstallCount int        `json:"install_count"`
}

type listResponse struct {
	Items  []roleDTO `json:"items"`
	Total  int       `json:"total"`
	Status string    `json:"status"`
}

type detailResponse struct {
	Item roleDTO `json:"item"`
}

type repoResponse struct {
	Repo      string    `json:"repo"`
	SourceURL string    `json:"source_url"`
	Items     []roleDTO `json:"items"`
	Total     int       `json:"total"`
	Status    string    `json:"status"`
}

type statsResponse struct {
	Total        int            `json:"total"`
	Repositories int            `json:"repositories"`
	ByStatus     map[string]int `json:"by_status"`
}

func (api *catalogAPI) handleRoleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	status, statusLabel, err := parseStatusParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_status", err.Error())
		return
	}
	catalog, err := ReadRoleRepoCatalog(api.catalogPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "catalog_read_failed", err.Error())
		return
	}

	entries := FilterCatalogByStatus(catalog, status)
	items := make([]roleDTO, 0, len(entries))
	for _, entry := range entries {
		items = append(items, roleToDTO(entry))
	}
	writeData(w, http.StatusOK, listResponse{Items: items, Total: len(items), Status: statusLabel})
}

func (api *catalogAPI) handleRoleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		query = strings.TrimSpace(r.URL.Query().Get("query"))
	}
	if query == "" {
		writeError(w, http.StatusBadRequest, "missing_query", "query parameter 'q' is required")
		return
	}
	status, statusLabel, err := parseStatusParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_status", err.Error())
		return
	}
	catalog, err := ReadRoleRepoCatalog(api.catalogPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "catalog_read_failed", err.Error())
		return
	}

	entries := SearchCatalog(catalog, query, status)
	items := make([]roleDTO, 0, len(entries))
	for _, entry := range entries {
		items = append(items, roleToDTO(entry))
	}
	writeData(w, http.StatusOK, listResponse{Items: items, Total: len(items), Status: statusLabel})
}

func (api *catalogAPI) handleRoleDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, catalogAPIBasePath+"/roles/")
	name = strings.Trim(name, "/")
	if name == "" || strings.Contains(name, "/") {
		writeError(w, http.StatusBadRequest, "invalid_role", "role name must be provided in path")
		return
	}
	status, _, err := parseStatusParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_status", err.Error())
		return
	}
	catalog, err := ReadRoleRepoCatalog(api.catalogPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "catalog_read_failed", err.Error())
		return
	}
	entries := FilterCatalogByStatus(catalog, status)
	entry, found := findEntry(entries, strings.TrimSpace(r.URL.Query().Get("source")), name)
	if !found {
		writeError(w, http.StatusNotFound, "role_not_found", "role not found")
		return
	}
	writeData(w, http.StatusOK, detailResponse{Item: roleToDTO(entry)})
}

func (api *catalogAPI) handleRepoDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	repo, ok := parseRepoPath(r.URL.Path)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_repo", "repo path must be /api/repos/{owner}/{repo}")
		return
	}
	status, statusLabel, err := parseStatusParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_status", err.Error())
		return
	}
	catalog, err := ReadRoleRepoCatalog(api.catalogPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "catalog_read_failed", err.Error())
		return
	}
	entries := FilterCatalogByStatus(catalog, status)
	items := make([]roleDTO, 0)
	sourceURL := ""
	for _, entry := range entries {
		if strings.EqualFold(entry.Source, repo) {
			if sourceURL == "" && entry.SourceURL != "" {
				sourceURL = entry.SourceURL
			}
			items = append(items, roleToDTO(entry))
		}
	}
	if sourceURL == "" {
		sourceURL = "https://github.com/" + repo
	}
	writeData(w, http.StatusOK, repoResponse{
		Repo:      repo,
		SourceURL: sourceURL,
		Items:     items,
		Total:     len(items),
		Status:    statusLabel,
	})
}

func (api *catalogAPI) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	catalog, err := ReadRoleRepoCatalog(api.catalogPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "catalog_read_failed", err.Error())
		return
	}
	stats := CatalogStats(catalog)
	byStatus := map[string]int{}
	for status, count := range stats {
		byStatus[string(status)] = count
	}
	repos := GroupCatalogByRepo(catalog.Entries)
	writeData(w, http.StatusOK, statsResponse{
		Total:        len(catalog.Entries),
		Repositories: len(repos),
		ByStatus:     byStatus,
	})
}

func parseStatusParam(r *http.Request) (RoleRepoCatalogStatus, string, error) {
	raw := strings.TrimSpace(r.URL.Query().Get("status"))
	if raw == "" {
		return CatalogStatusVerified, string(CatalogStatusVerified), nil
	}
	if strings.EqualFold(raw, "all") {
		return "", "all", nil
	}
	status := RoleRepoCatalogStatus(raw)
	if !ValidCatalogStatuses[status] {
		return "", "", fmt.Errorf("invalid status %q; use: discovered, verified, invalid, unreachable, all", raw)
	}
	return status, raw, nil
}

func roleToDTO(entry RoleRepoCatalogEntry) roleDTO {
	var verifiedAt *time.Time
	if entry.VerifiedAt != nil {
		t := *entry.VerifiedAt
		verifiedAt = &t
	}
	var updatedAt *time.Time
	if !entry.UpdatedAt.IsZero() {
		t := entry.UpdatedAt
		updatedAt = &t
	}
	return roleDTO{
		Name:         entry.Name,
		Source:       entry.Source,
		SourceType:   entry.SourceType,
		SourceURL:    entry.SourceURL,
		RolePath:     entry.RolePath,
		FolderHash:   entry.FolderHash,
		Status:       string(entry.Status),
		StatusReason: entry.StatusReason,
		DiscoveredAt: entry.DiscoveredAt,
		VerifiedAt:   verifiedAt,
		UpdatedAt:    updatedAt,
		InstallCount: entry.InstallCount,
	}
}

func parseRepoPath(path string) (string, bool) {
	trimmed := strings.TrimPrefix(path, catalogAPIBasePath+"/repos/")
	trimmed = strings.Trim(trimmed, "/")
	if trimmed == "" {
		return "", false
	}
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", false
	}
	return parts[0] + "/" + parts[1], true
}

func findEntry(entries []RoleRepoCatalogEntry, source, name string) (RoleRepoCatalogEntry, bool) {
	if source != "" {
		for _, entry := range entries {
			if strings.EqualFold(entry.Source, source) && entry.Name == name {
				return entry, true
			}
		}
		return RoleRepoCatalogEntry{}, false
	}
	for _, entry := range entries {
		if entry.Name == name {
			return entry, true
		}
	}
	return RoleRepoCatalogEntry{}, false
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
}

func writeData(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, apiEnvelope{Data: data})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, apiEnvelope{Error: &apiError{Code: code, Message: message}})
}

func writeJSON(w http.ResponseWriter, status int, payload apiEnvelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
