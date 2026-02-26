// internal/openspec.go
package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// EnsureOpenSpec checks if openspec CLI is available, installs if missing.
func EnsureOpenSpec() error {
	if _, err := exec.LookPath("openspec"); err == nil {
		return nil
	}
	fmt.Println("OpenSpec not found, installing...")
	cmd := exec.Command("npm", "install", "-g", "@fission-ai/openspec@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// OpenSpecInit runs openspec init in the given directory.
func OpenSpecInit(dir string) error {
	cmd := exec.Command("openspec", "init", "--tools", "claude,codex,opencode")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// CreateChange creates an OpenSpec change directory with proposal and optional design files.
func CreateChange(wtPath, changeName, proposal, design string) (string, error) {
	changePath := filepath.Join(wtPath, "openspec", "changes", changeName)
	if err := os.MkdirAll(changePath, 0755); err != nil {
		return "", fmt.Errorf("create change directory: %w", err)
	}

	// Write .openspec.yaml metadata
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	meta := fmt.Sprintf("schema: default\ncreated_at: %s\n", now)
	metaPath := filepath.Join(changePath, ".openspec.yaml")
	if err := os.WriteFile(metaPath, []byte(meta), 0644); err != nil {
		return "", fmt.Errorf("write .openspec.yaml: %w", err)
	}

	// Write proposal.md
	proposalPath := filepath.Join(changePath, "proposal.md")
	if err := os.WriteFile(proposalPath, []byte(proposal), 0644); err != nil {
		return "", fmt.Errorf("write proposal.md: %w", err)
	}

	// Write design.md (if provided)
	if design != "" {
		designPath := filepath.Join(changePath, "design.md")
		if err := os.WriteFile(designPath, []byte(design), 0644); err != nil {
			return "", fmt.Errorf("write design.md: %w", err)
		}
	}

	return changePath, nil
}

// ChangeStatus represents a parsed OpenSpec change with its current phase.
type ChangeStatus struct {
	Name  string
	Phase string // "planning", "ready", "implementing", "completed"
}

type openspecStatusJSON struct {
	Changes []openspecChangeJSON `json:"changes"`
}

type openspecChangeJSON struct {
	Name      string                     `json:"name"`
	Artifacts map[string]openspecArtifact `json:"artifacts"`
}

type openspecArtifact struct {
	Status string `json:"status"`
}

// ParseOpenSpecStatus parses the JSON output of `openspec status --json`.
func ParseOpenSpecStatus(data []byte) ([]ChangeStatus, error) {
	var raw openspecStatusJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse openspec status: %w", err)
	}

	var result []ChangeStatus
	for _, c := range raw.Changes {
		phase := classifyPhase(c.Artifacts)
		result = append(result, ChangeStatus{Name: c.Name, Phase: phase})
	}
	return result, nil
}

func classifyPhase(artifacts map[string]openspecArtifact) string {
	tasksStatus := artifacts["tasks"].Status
	if tasksStatus == "done" {
		// "completed" requires verify artifact to also be done
		if artifacts["verify"].Status == "done" {
			return "completed"
		}
		return "ready"
	}
	// If proposal is done but tasks not → still planning
	if artifacts["proposal"].Status == "done" {
		// Check if currently applying
		for _, a := range artifacts {
			if a.Status == "in_progress" {
				return "implementing"
			}
		}
		return "planning"
	}
	return "planning"
}

// GetOpenSpecStatus runs `openspec status --json` in the given directory and parses results.
func GetOpenSpecStatus(wtPath string) ([]ChangeStatus, error) {
	cmd := exec.Command("openspec", "status", "--json")
	cmd.Dir = wtPath
	out, err := cmd.Output()
	if err != nil {
		return nil, nil // No changes or openspec not initialized
	}
	return ParseOpenSpecStatus(out)
}
