// internal/hook.go
package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Provider represents an AI coding tool provider.
type Provider string

const (
	ProviderClaude   Provider = "claude"
	ProviderGemini   Provider = "gemini"
	ProviderOpenCode Provider = "opencode"
	ProviderUnknown  Provider = "unknown"
)

// HookInput is the unified input structure for all hook events.
type HookInput struct {
	CWD       string          `json:"cwd"`
	SessionID string          `json:"session_id,omitempty"`
	ToolName  string          `json:"tool_name,omitempty"`
	ToolInput json.RawMessage `json:"tool_input,omitempty"`
	ParentCWD string          `json:"parent_cwd,omitempty"`
	Provider  Provider        `json:"-"`
	Raw       json.RawMessage `json:"-"`
}

// WorktreeInfo holds information about a detected agent-team worktree.
type WorktreeInfo struct {
	WtPath   string // worktree absolute path
	WorkerID string // basename of WtPath
	MainRoot string // main repository root
}

// DetectProvider determines the current provider from environment variables.
func DetectProvider() Provider {
	if os.Getenv("CLAUDE_PLUGIN_ROOT") != "" {
		return ProviderClaude
	}
	if os.Getenv("GEMINI_PROJECT_DIR") != "" {
		return ProviderGemini
	}
	if os.Getenv("OPENCODE_SESSION") != "" {
		return ProviderOpenCode
	}
	return ProviderUnknown
}

// ParseProvider parses a provider string, returning ProviderUnknown for unrecognized values.
func ParseProvider(s string) Provider {
	switch Provider(s) {
	case ProviderClaude, ProviderGemini, ProviderOpenCode:
		return Provider(s)
	default:
		return ProviderUnknown
	}
}

// ParseHookInput reads and parses hook JSON from the given reader.
// Returns a valid HookInput even for empty stdin (OpenCode scenario).
func ParseHookInput(r io.Reader) (*HookInput, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read stdin: %w", err)
	}

	input := &HookInput{}

	// Handle empty stdin (e.g. OpenCode calls with --provider flag)
	if len(bytes.TrimSpace(data)) == 0 {
		cwd, _ := os.Getwd()
		input.CWD = cwd
		return input, nil
	}

	input.Raw = json.RawMessage(data)
	if err := json.Unmarshal(data, input); err != nil {
		return nil, fmt.Errorf("parse hook JSON: %w", err)
	}

	// Fallback CWD to current working directory
	if input.CWD == "" {
		cwd, _ := os.Getwd()
		input.CWD = cwd
	}

	return input, nil
}

// ResolveWorktree detects if cwd is inside an agent-team managed worktree.
// Returns nil (not an error) if not in a managed worktree.
func ResolveWorktree(cwd string) (*WorktreeInfo, error) {
	topCmd := exec.Command("git", "rev-parse", "--show-toplevel")
	topCmd.Dir = cwd
	topOut, err := topCmd.Output()
	if err != nil {
		return nil, nil // not in a git repo
	}
	wtPath := strings.TrimSpace(string(topOut))

	// Check if the path contains /.worktrees/ or /worktrees/ (agent-team convention)
	for _, marker := range []string{"/.worktrees/", "/worktrees/"} {
		if idx := strings.Index(wtPath, marker); idx >= 0 {
			mainRoot := wtPath[:idx]
			return &WorktreeInfo{
				WtPath:   wtPath,
				WorkerID: filepath.Base(wtPath),
				MainRoot: mainRoot,
			}, nil
		}
	}

	return nil, nil // not a managed worktree
}

// ResolveMainRootFromCWD derives the main repository root using git-common-dir.
func ResolveMainRootFromCWD(cwd string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-common-dir")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}
	commonDir := strings.TrimSpace(string(out))

	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(cwd, commonDir)
	}
	commonDir = filepath.Clean(commonDir)

	if strings.HasSuffix(commonDir, string(filepath.Separator)+".git") {
		return strings.TrimSuffix(commonDir, string(filepath.Separator)+".git"), nil
	}

	return filepath.Dir(commonDir), nil
}

// LoadWorkerFromWorktree loads the WorkerConfig from a worktree's worker.yaml.
func LoadWorkerFromWorktree(wtPath string) (*WorkerConfig, error) {
	return LoadWorkerConfig(WorkerYAMLPath(wtPath))
}

// roleYAMLWithChecks is used to read quality_checks from role.yaml.
type roleYAMLWithChecks struct {
	QualityChecks []string `yaml:"quality_checks"`
}

// ReadRoleQualityChecks reads the quality_checks list from a role's role.yaml.
func ReadRoleQualityChecks(rolePath string) ([]string, error) {
	yamlPath := filepath.Join(rolePath, "references", "role.yaml")
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var ry roleYAMLWithChecks
	if err := yaml.Unmarshal(data, &ry); err != nil {
		return nil, err
	}
	return ry.QualityChecks, nil
}

// ExtractAgentTeamSection extracts the <!-- AGENT_TEAM:START -->...<!-- AGENT_TEAM:END --> block
// from the given content. Returns empty string if not found.
func ExtractAgentTeamSection(content string) string {
	const startMarker = "<!-- AGENT_TEAM:START -->"
	const endMarker = "<!-- AGENT_TEAM:END -->"

	startIdx := strings.Index(content, startMarker)
	if startIdx < 0 {
		return ""
	}
	endIdx := strings.Index(content[startIdx:], endMarker)
	if endIdx < 0 {
		return ""
	}
	return content[startIdx : startIdx+endIdx+len(endMarker)]
}
