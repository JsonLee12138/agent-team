// internal/hook_test.go
package internal

import (
	"os"
	"strings"
	"testing"
)

func TestDetectProvider(t *testing.T) {
	// Save and clear relevant env vars
	envVars := []string{"CLAUDE_PLUGIN_ROOT", "GEMINI_PROJECT_DIR", "OPENCODE_SESSION"}
	saved := make(map[string]string)
	for _, v := range envVars {
		saved[v] = os.Getenv(v)
		os.Unsetenv(v)
	}
	defer func() {
		for k, v := range saved {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
	}()

	// Unknown when no env vars set
	if got := DetectProvider(); got != ProviderUnknown {
		t.Errorf("empty env: got %s, want %s", got, ProviderUnknown)
	}

	// Claude
	os.Setenv("CLAUDE_PLUGIN_ROOT", "/tmp/plugin")
	if got := DetectProvider(); got != ProviderClaude {
		t.Errorf("CLAUDE_PLUGIN_ROOT set: got %s, want %s", got, ProviderClaude)
	}
	os.Unsetenv("CLAUDE_PLUGIN_ROOT")

	// Gemini
	os.Setenv("GEMINI_PROJECT_DIR", "/tmp/project")
	if got := DetectProvider(); got != ProviderGemini {
		t.Errorf("GEMINI_PROJECT_DIR set: got %s, want %s", got, ProviderGemini)
	}
	os.Unsetenv("GEMINI_PROJECT_DIR")

	// OpenCode
	os.Setenv("OPENCODE_SESSION", "abc123")
	if got := DetectProvider(); got != ProviderOpenCode {
		t.Errorf("OPENCODE_SESSION set: got %s, want %s", got, ProviderOpenCode)
	}
	os.Unsetenv("OPENCODE_SESSION")

	// Priority: Claude > Gemini > OpenCode
	os.Setenv("CLAUDE_PLUGIN_ROOT", "/tmp/plugin")
	os.Setenv("GEMINI_PROJECT_DIR", "/tmp/project")
	if got := DetectProvider(); got != ProviderClaude {
		t.Errorf("both set: got %s, want %s", got, ProviderClaude)
	}
	os.Unsetenv("CLAUDE_PLUGIN_ROOT")
	os.Unsetenv("GEMINI_PROJECT_DIR")
}

func TestParseProvider(t *testing.T) {
	tests := []struct {
		input string
		want  Provider
	}{
		{"claude", ProviderClaude},
		{"gemini", ProviderGemini},
		{"opencode", ProviderOpenCode},
		{"unknown", ProviderUnknown},
		{"", ProviderUnknown},
		{"invalid", ProviderUnknown},
	}
	for _, tt := range tests {
		if got := ParseProvider(tt.input); got != tt.want {
			t.Errorf("ParseProvider(%q) = %s, want %s", tt.input, got, tt.want)
		}
	}
}

func TestParseHookInput_ClaudeFormat(t *testing.T) {
	json := `{"cwd": "/path/to/worktree", "session_id": "sess-123"}`
	input, err := ParseHookInput(strings.NewReader(json))
	if err != nil {
		t.Fatalf("ParseHookInput: %v", err)
	}
	if input.CWD != "/path/to/worktree" {
		t.Errorf("CWD = %q, want /path/to/worktree", input.CWD)
	}
	if input.SessionID != "sess-123" {
		t.Errorf("SessionID = %q, want sess-123", input.SessionID)
	}
}

func TestParseHookInput_GeminiFormat(t *testing.T) {
	json := `{"cwd": "/path", "session_id": "abc", "hook_event_name": "SessionStart", "source": "startup"}`
	input, err := ParseHookInput(strings.NewReader(json))
	if err != nil {
		t.Fatalf("ParseHookInput: %v", err)
	}
	if input.CWD != "/path" {
		t.Errorf("CWD = %q, want /path", input.CWD)
	}
	if input.SessionID != "abc" {
		t.Errorf("SessionID = %q, want abc", input.SessionID)
	}
	if input.Raw == nil {
		t.Error("Raw should not be nil")
	}
}

func TestParseHookInput_PreToolUse(t *testing.T) {
	json := `{"cwd": "/path", "tool_name": "Write", "tool_input": {"file_path": "/foo.txt"}}`
	input, err := ParseHookInput(strings.NewReader(json))
	if err != nil {
		t.Fatalf("ParseHookInput: %v", err)
	}
	if input.ToolName != "Write" {
		t.Errorf("ToolName = %q, want Write", input.ToolName)
	}
	if input.ToolInput == nil {
		t.Error("ToolInput should not be nil")
	}
}

func TestParseHookInput_EmptyStdin(t *testing.T) {
	input, err := ParseHookInput(strings.NewReader(""))
	if err != nil {
		t.Fatalf("ParseHookInput empty: %v", err)
	}
	if input.CWD == "" {
		t.Error("CWD should fallback to os.Getwd()")
	}
}

func TestParseHookInput_WhitespaceStdin(t *testing.T) {
	input, err := ParseHookInput(strings.NewReader("  \n  "))
	if err != nil {
		t.Fatalf("ParseHookInput whitespace: %v", err)
	}
	if input.CWD == "" {
		t.Error("CWD should fallback to os.Getwd()")
	}
}

func TestParseHookInput_InvalidJSON(t *testing.T) {
	_, err := ParseHookInput(strings.NewReader("{invalid"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestExtractAgentTeamSection(t *testing.T) {
	content := `# Project

Some text before.

<!-- AGENT_TEAM:START -->
## Role: frontend-dev

You are a frontend developer.
<!-- AGENT_TEAM:END -->

More text after.
`
	section := ExtractAgentTeamSection(content)
	if section == "" {
		t.Fatal("ExtractAgentTeamSection returned empty")
	}
	if !strings.Contains(section, "AGENT_TEAM:START") {
		t.Error("section should contain start marker")
	}
	if !strings.Contains(section, "AGENT_TEAM:END") {
		t.Error("section should contain end marker")
	}
	if !strings.Contains(section, "frontend-dev") {
		t.Error("section should contain role content")
	}
}

func TestExtractAgentTeamSection_NotFound(t *testing.T) {
	content := "# Just a regular file\nNo markers here."
	section := ExtractAgentTeamSection(content)
	if section != "" {
		t.Errorf("expected empty, got %q", section)
	}
}

func TestExtractAgentTeamSection_MissingEnd(t *testing.T) {
	content := "<!-- AGENT_TEAM:START -->\nContent without end marker"
	section := ExtractAgentTeamSection(content)
	if section != "" {
		t.Errorf("expected empty for missing end marker, got %q", section)
	}
}

func TestReadRoleQualityChecks(t *testing.T) {
	tmpDir := t.TempDir()
	refsDir := tmpDir + "/references"
	os.MkdirAll(refsDir, 0755)

	// Write role.yaml with quality_checks
	yamlContent := `name: test-role
quality_checks:
  - "go vet ./..."
  - "golangci-lint run"
`
	os.WriteFile(refsDir+"/role.yaml", []byte(yamlContent), 0644)

	checks, err := ReadRoleQualityChecks(tmpDir)
	if err != nil {
		t.Fatalf("ReadRoleQualityChecks: %v", err)
	}
	if len(checks) != 2 {
		t.Fatalf("expected 2 checks, got %d", len(checks))
	}
	if checks[0] != "go vet ./..." {
		t.Errorf("checks[0] = %q, want 'go vet ./...'", checks[0])
	}
}

func TestReadRoleQualityChecks_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	checks, err := ReadRoleQualityChecks(tmpDir)
	if err != nil {
		t.Fatalf("ReadRoleQualityChecks: %v", err)
	}
	if checks != nil {
		t.Errorf("expected nil, got %v", checks)
	}
}

func TestReadRoleQualityChecks_NoChecks(t *testing.T) {
	tmpDir := t.TempDir()
	refsDir := tmpDir + "/references"
	os.MkdirAll(refsDir, 0755)

	yamlContent := `name: test-role
description: A test role
`
	os.WriteFile(refsDir+"/role.yaml", []byte(yamlContent), 0644)

	checks, err := ReadRoleQualityChecks(tmpDir)
	if err != nil {
		t.Fatalf("ReadRoleQualityChecks: %v", err)
	}
	if len(checks) != 0 {
		t.Errorf("expected 0 checks, got %d", len(checks))
	}
}
