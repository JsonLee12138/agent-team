package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type Store struct {
	db              *sql.DB
	dialect         string
	stmtUpsertRole  string
	stmtGetEvent    string
	stmtInsertEvent string
}

func NewStore(db *sql.DB, dialect string) (*Store, error) {
	var upsert, getEvent, insertEvent string
	switch dialect {
	case "postgres":
		upsert = `
INSERT INTO role_records (repo_owner, repo_name, role_path, name, description, source_url, score, tags, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)
ON CONFLICT (repo_owner, repo_name, role_path)
DO UPDATE SET name=EXCLUDED.name,
              description=EXCLUDED.description,
              source_url=EXCLUDED.source_url,
              score=EXCLUDED.score,
              tags=EXCLUDED.tags,
              updated_at=CURRENT_TIMESTAMP;
`
		getEvent = "SELECT response_code, response_body FROM ingest_events WHERE idempotency_key = $1"
		insertEvent = "INSERT INTO ingest_events (idempotency_key, response_code, response_body) VALUES ($1,$2,$3)"
	case "sqlite":
		upsert = `
INSERT INTO role_records (repo_owner, repo_name, role_path, name, description, source_url, score, tags, created_at, updated_at)
VALUES (?,?,?,?,?,?,?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (repo_owner, repo_name, role_path)
DO UPDATE SET name=excluded.name,
              description=excluded.description,
              source_url=excluded.source_url,
              score=excluded.score,
              tags=excluded.tags,
              updated_at=CURRENT_TIMESTAMP;
`
		getEvent = "SELECT response_code, response_body FROM ingest_events WHERE idempotency_key = ?"
		insertEvent = "INSERT INTO ingest_events (idempotency_key, response_code, response_body) VALUES (?,?,?)"
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dialect)
	}

	return &Store{db: db, dialect: dialect, stmtUpsertRole: upsert, stmtGetEvent: getEvent, stmtInsertEvent: insertEvent}, nil
}

type IngestEvent struct {
	ResponseCode int
	ResponseBody []byte
}

func (s *Store) GetIngestEvent(ctx context.Context, key string) (*IngestEvent, error) {
	row := s.db.QueryRowContext(ctx, s.stmtGetEvent, key)
	var code int
	var body string
	if err := row.Scan(&code, &body); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &IngestEvent{ResponseCode: code, ResponseBody: []byte(body)}, nil
}

func (s *Store) InsertIngestEvent(ctx context.Context, key string, responseCode int, responseBody []byte) error {
	_, err := s.db.ExecContext(ctx, s.stmtInsertEvent, key, responseCode, string(responseBody))
	return err
}

func (s *Store) UpsertRoleRecord(ctx context.Context, record RoleRecord) error {
	var tagsJSON any
	if record.Tags != nil {
		encoded, err := json.Marshal(record.Tags)
		if err != nil {
			return err
		}
		tagsJSON = string(encoded)
	} else {
		tagsJSON = nil
	}
	_, err := s.db.ExecContext(ctx, s.stmtUpsertRole,
		record.RepoOwner,
		record.RepoName,
		record.RolePath,
		record.Name,
		record.Description,
		record.SourceURL,
		record.Score,
		tagsJSON,
	)
	return err
}

// RoleRecord is the normalized storage shape.
type RoleRecord struct {
	RepoOwner   string
	RepoName    string
	RolePath    string
	Name        string
	Description string
	SourceURL   string
	Score       *float64
	Tags        []string
}
