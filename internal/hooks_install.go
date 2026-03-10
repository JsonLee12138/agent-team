// internal/hooks_install.go
package internal

import (
	_ "embed"
	"encoding/json"
	"os"
	"path/filepath"
)

//go:embed hooks.json
var embeddedHooksJSON []byte

// Claude Code supported hook events (2026-03).
var claudeHookEvents = map[string]bool{
	"PreToolUse":   true,
	"PostToolUse":  true,
	"Stop":         true,
	"SubagentStop": true,
}

// InstallHooksToSettings writes Claude-compatible hook events from the embedded
// hooks.json into the worktree's .claude/settings.json.
// Existing settings.json content is preserved; hooks are merged by event name.
func InstallHooksToSettings(wtPath string) error {
	claudeDir := filepath.Join(wtPath, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return err
	}
	settingsPath := filepath.Join(claudeDir, "settings.json")

	// Parse embedded hooks.json
	var raw struct {
		Hooks map[string]json.RawMessage `json:"hooks"`
	}
	if err := json.Unmarshal(embeddedHooksJSON, &raw); err != nil {
		return err
	}

	// Filter to Claude-compatible events only
	filtered := make(map[string]json.RawMessage)
	for event, groups := range raw.Hooks {
		if claudeHookEvents[event] {
			filtered[event] = groups
		}
	}

	// Read existing settings.json if present, merge rather than overwrite
	existing := make(map[string]json.RawMessage)
	if data, err := os.ReadFile(settingsPath); err == nil {
		_ = json.Unmarshal(data, &existing) // parse failure → start fresh
	}

	// Merge hooks: agent-team events override same-name events, others preserved
	existingHooks := make(map[string]json.RawMessage)
	if raw, ok := existing["hooks"]; ok {
		_ = json.Unmarshal(raw, &existingHooks)
	}
	for event, groups := range filtered {
		existingHooks[event] = groups
	}
	hooksBytes, err := json.Marshal(existingHooks)
	if err != nil {
		return err
	}
	existing["hooks"] = json.RawMessage(hooksBytes)

	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, data, 0644)
}
