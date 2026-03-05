package ingest

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	repoPattern     = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)
	rolePathPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+(/[A-Za-z0-9_.-]+)*$`)
)

func ValidateRequest(req IngestRequest, maxResults int) []FieldDetail {
	var issues []FieldDetail
	if strings.TrimSpace(req.IdempotencyKey) == "" {
		issues = append(issues, FieldDetail{Field: "idempotency_key", Message: "required"})
	}
	if len(req.IdempotencyKey) > 128 {
		issues = append(issues, FieldDetail{Field: "idempotency_key", Message: "too long"})
	}
	if strings.TrimSpace(req.TraceID) == "" {
		issues = append(issues, FieldDetail{Field: "trace_id", Message: "required"})
	}
	if strings.TrimSpace(req.Timestamp) == "" {
		issues = append(issues, FieldDetail{Field: "timestamp", Message: "required"})
	} else if _, err := time.Parse(time.RFC3339, req.Timestamp); err != nil {
		issues = append(issues, FieldDetail{Field: "timestamp", Message: "must be RFC3339"})
	}
	if strings.TrimSpace(req.Query) == "" {
		issues = append(issues, FieldDetail{Field: "query", Message: "required"})
	}
	if req.ResultCount < 0 {
		issues = append(issues, FieldDetail{Field: "result_count", Message: "must be >= 0"})
	}
	if req.ResultCount != len(req.Results) {
		issues = append(issues, FieldDetail{Field: "result_count", Message: fmt.Sprintf("expected %d results", req.ResultCount)})
	}
	if len(req.Results) > maxResults {
		issues = append(issues, FieldDetail{Field: "results", Message: fmt.Sprintf("max %d results", maxResults)})
	}
	for i, result := range req.Results {
		issues = append(issues, ValidateResult(result, i)...)
	}
	return issues
}

func ValidateResult(result IngestResult, index int) []FieldDetail {
	var issues []FieldDetail
	fieldPrefix := fmt.Sprintf("results[%d]", index)
	if strings.TrimSpace(result.Repo) == "" {
		issues = append(issues, FieldDetail{Field: fieldPrefix + ".repo", Message: "required"})
	} else if !repoPattern.MatchString(result.Repo) {
		issues = append(issues, FieldDetail{Field: fieldPrefix + ".repo", Message: "must be owner/repo"})
	}
	if strings.TrimSpace(result.RolePath) == "" {
		issues = append(issues, FieldDetail{Field: fieldPrefix + ".role_path", Message: "required"})
	} else if !rolePathPattern.MatchString(result.RolePath) {
		issues = append(issues, FieldDetail{Field: fieldPrefix + ".role_path", Message: "invalid role path"})
	}
	if result.SourceURL != "" {
		if err := validateGitHubURL(result.SourceURL); err != nil {
			issues = append(issues, FieldDetail{Field: fieldPrefix + ".source_url", Message: err.Error()})
		}
	}
	if len(result.Tags) > 64 {
		issues = append(issues, FieldDetail{Field: fieldPrefix + ".tags", Message: "too many tags"})
	}
	return issues
}

func validateGitHubURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid url")
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return fmt.Errorf("must be http or https")
	}
	if !strings.EqualFold(parsed.Host, "github.com") {
		return fmt.Errorf("must be a github.com URL")
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 {
		return fmt.Errorf("must include owner/repo")
	}
	return nil
}
