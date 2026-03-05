package ingest

import "testing"

func TestValidateRequest(t *testing.T) {
	req := IngestRequest{
		IdempotencyKey: "",
		TraceID:        "",
		Timestamp:      "not-a-time",
		Query:          "",
		ResultCount:    1,
		Results: []IngestResult{{
			Repo:      "bad repo",
			RolePath:  "../oops",
			SourceURL: "https://example.com/acme/roles",
		}},
	}

	issues := ValidateRequest(req, 10)
	if len(issues) == 0 {
		t.Fatalf("expected validation errors")
	}
}
