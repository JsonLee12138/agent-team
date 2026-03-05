package ingest

import "time"

type IngestRequest struct {
	IdempotencyKey string         `json:"idempotency_key"`
	TraceID        string         `json:"trace_id"`
	Timestamp      string         `json:"timestamp"`
	Query          string         `json:"query"`
	ResultCount    int            `json:"result_count"`
	Results        []IngestResult `json:"results"`
}

type IngestResult struct {
	Repo        string   `json:"repo"`
	RolePath    string   `json:"role_path"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	SourceURL   string   `json:"source_url"`
	Score       *float64 `json:"score"`
	Tags        []string `json:"tags"`
}

type IngestResponse struct {
	Status         string          `json:"status"`
	IdempotencyKey string          `json:"idempotency_key"`
	Accepted       int             `json:"accepted"`
	Rejected       int             `json:"rejected"`
	Errors         []IngestFailure `json:"errors,omitempty"`
}

type IngestFailure struct {
	Index   int    `json:"index"`
	Repo    string `json:"repo"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code          string        `json:"code"`
	Message       string        `json:"message"`
	Details       []FieldDetail `json:"details,omitempty"`
	MissingFields []string      `json:"missing_fields,omitempty"`
}

type FieldDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (r IngestRequest) ParsedTimestamp() (time.Time, error) {
	return time.Parse(time.RFC3339, r.Timestamp)
}
