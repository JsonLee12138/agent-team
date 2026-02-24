# Brainstorming + OpenSpec Integration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the file-based task system with OpenSpec and add brainstorming flow to the SKILL.md controller layer.

**Architecture:** OpenSpec CLI invoked via `exec.Command` from Go. New `internal/openspec.go` wraps install-check, init, and status-parsing. `assign` creates OpenSpec changes with proposals. SKILL.md guides the controller AI through brainstorming before assign.

**Tech Stack:** Go 1.24, Cobra CLI, OpenSpec CLI (`@fission-ai/openspec`), YAML

---

### Task 1: Create `internal/openspec.go` — OpenSpec CLI helpers

**Files:**
- Create: `internal/openspec.go`
- Test: `internal/openspec_test.go`

**Step 1: Write the failing tests**

```go
// internal/openspec_test.go
package internal

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestEnsureOpenSpec_AlreadyInstalled(t *testing.T) {
	// If openspec is on PATH, EnsureOpenSpec should succeed without installing
	if _, err := exec.LookPath("openspec"); err != nil {
		t.Skip("openspec not installed, skipping")
	}
	err := EnsureOpenSpec()
	if err != nil {
		t.Fatalf("EnsureOpenSpec: %v", err)
	}
}

func TestOpenSpecInit(t *testing.T) {
	if _, err := exec.LookPath("openspec"); err != nil {
		t.Skip("openspec not installed, skipping")
	}

	dir := t.TempDir()
	// openspec init requires a git repo
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s (%v)", args, out, err)
		}
	}
	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")
	run("commit", "--allow-empty", "-m", "init")

	err := OpenSpecInit(dir)
	if err != nil {
		t.Fatalf("OpenSpecInit: %v", err)
	}

	// Verify openspec directory was created
	if _, err := os.Stat(filepath.Join(dir, "openspec")); os.IsNotExist(err) {
		t.Error("openspec/ directory not created")
	}
	if _, err := os.Stat(filepath.Join(dir, "openspec", "changes")); os.IsNotExist(err) {
		t.Error("openspec/changes/ directory not created")
	}
}

func TestCreateChange(t *testing.T) {
	dir := t.TempDir()
	// Create minimal openspec structure
	os.MkdirAll(filepath.Join(dir, "openspec", "changes"), 0755)

	changeName := "2026-02-24-fix-login"
	proposal := "# Proposal\n\nFix the login flow by adding JWT validation."

	changePath, err := CreateChange(dir, changeName, proposal)
	if err != nil {
		t.Fatalf("CreateChange: %v", err)
	}

	// Verify change directory
	if _, err := os.Stat(changePath); os.IsNotExist(err) {
		t.Error("change directory not created")
	}

	// Verify .openspec.yaml
	metaPath := filepath.Join(changePath, ".openspec.yaml")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Error(".openspec.yaml not created")
	}

	// Verify proposal.md
	proposalPath := filepath.Join(changePath, "proposal.md")
	data, err := os.ReadFile(proposalPath)
	if err != nil {
		t.Fatalf("read proposal.md: %v", err)
	}
	if string(data) != proposal {
		t.Errorf("proposal content = %q, want %q", string(data), proposal)
	}
}

func TestCreateChangeEmptyProposal(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "openspec", "changes"), 0755)

	changePath, err := CreateChange(dir, "2026-02-24-test", "")
	if err != nil {
		t.Fatalf("CreateChange: %v", err)
	}

	// proposal.md should exist but be empty
	data, _ := os.ReadFile(filepath.Join(changePath, "proposal.md"))
	if len(data) != 0 {
		t.Errorf("expected empty proposal, got %q", string(data))
	}
}

func TestParseOpenSpecStatus(t *testing.T) {
	// Test parsing the JSON output from openspec status
	jsonData := `{"changes":[{"name":"fix-login","artifacts":{"proposal":{"status":"done"},"specs":{"status":"ready"},"design":{"status":"blocked"},"tasks":{"status":"blocked"}}},{"name":"add-auth","artifacts":{"proposal":{"status":"done"},"specs":{"status":"done"},"design":{"status":"done"},"tasks":{"status":"done"}}}]}`

	result, err := ParseOpenSpecStatus([]byte(jsonData))
	if err != nil {
		t.Fatalf("ParseOpenSpecStatus: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(result))
	}
	if result[0].Name != "fix-login" {
		t.Errorf("change[0].Name = %q, want fix-login", result[0].Name)
	}
	if result[0].Phase != "planning" {
		t.Errorf("change[0].Phase = %q, want planning", result[0].Phase)
	}
	if result[1].Phase != "ready" {
		t.Errorf("change[1].Phase = %q, want ready", result[1].Phase)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/jsonlee/Projects/agent-team && go test ./internal/ -run "TestEnsureOpenSpec|TestOpenSpecInit|TestCreateChange|TestParseOpenSpecStatus" -v`
Expected: compilation errors — functions not defined

**Step 3: Write minimal implementation**

```go
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

// CreateChange creates an OpenSpec change directory with a proposal file.
func CreateChange(wtPath, changeName, proposal string) (string, error) {
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
	Name      string                       `json:"name"`
	Artifacts map[string]openspecArtifact   `json:"artifacts"`
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
		// Check if all artifacts are done (verify included)
		allDone := true
		for _, a := range artifacts {
			if a.Status != "done" {
				allDone = false
				break
			}
		}
		if allDone {
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
```

**Step 4: Run tests to verify they pass**

Run: `cd /Users/jsonlee/Projects/agent-team && go test ./internal/ -run "TestCreateChange|TestParseOpenSpecStatus" -v`
Expected: PASS (skip tests requiring openspec binary unless installed)

**Step 5: Commit**

```bash
git add internal/openspec.go internal/openspec_test.go
git commit -m "feat: add OpenSpec CLI helper functions"
```

---

### Task 2: Update `internal/role.go` — Remove task references, update templates

**Files:**
- Modify: `internal/role.go:94-144` (GenerateClaudeMD + PromptMDContent)
- Modify: `internal/role_test.go`

**Step 1: Write the failing tests**

Update `internal/role_test.go` — change `TestGenerateClaudeMD` to expect OpenSpec instructions instead of task instructions, and change `TestPromptMDTemplate` to expect Workflow section:

```go
// In TestGenerateClaudeMD — add new assertions:
func TestGenerateClaudeMD(t *testing.T) {
	// ... existing setup ...

	err := GenerateClaudeMD(wtPath, "dev", dir)
	if err != nil {
		t.Fatalf("GenerateClaudeMD: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(wtPath, "CLAUDE.md"))
	content := string(data)

	if !strings.HasPrefix(content, "# Role: dev") {
		t.Error("CLAUDE.md should start with prompt.md content")
	}
	if !strings.Contains(content, "Development Environment") {
		t.Error("CLAUDE.md should contain worktree context")
	}
	if !strings.Contains(content, "team/dev") {
		t.Error("CLAUDE.md should contain branch name")
	}
	// NEW: should NOT contain old task references
	if strings.Contains(content, "tasks/pending") {
		t.Error("CLAUDE.md should not reference tasks/pending")
	}
	if strings.Contains(content, "tasks/done") {
		t.Error("CLAUDE.md should not reference tasks/done")
	}
}

// Update TestPromptMDTemplate:
func TestPromptMDTemplate(t *testing.T) {
	content := PromptMDContent("backend")
	if !strings.Contains(content, "# Role: backend") {
		t.Error("prompt template should contain role name")
	}
	if !strings.Contains(content, "Communication Protocol") {
		t.Error("prompt template should contain communication protocol")
	}
	// NEW: should contain Workflow section with OpenSpec
	if !strings.Contains(content, "## Workflow") {
		t.Error("prompt template should contain Workflow section")
	}
	if !strings.Contains(content, "/opsx:continue") {
		t.Error("prompt template should reference /opsx:continue")
	}
	if !strings.Contains(content, "/opsx:apply") {
		t.Error("prompt template should reference /opsx:apply")
	}
	// Should NOT contain old task references
	if strings.Contains(content, "tasks/pending") {
		t.Error("prompt template should not reference tasks/pending")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/jsonlee/Projects/agent-team && go test ./internal/ -run "TestGenerateClaudeMD|TestPromptMDTemplate" -v`
Expected: FAIL — old templates still contain task references

**Step 3: Update `GenerateClaudeMD` in `internal/role.go:94-123`**

Replace the Git Rules section (line 114-120):

```go
func GenerateClaudeMD(wtPath, name, root string) error {
	teamsDir := filepath.Join(wtPath, "agents", "teams", name)
	promptPath := filepath.Join(teamsDir, "prompt.md")
	claudePath := filepath.Join(wtPath, "CLAUDE.md")

	prompt, err := os.ReadFile(promptPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var b strings.Builder
	b.Write(prompt)
	b.WriteString("\n## Development Environment\n\n")
	b.WriteString("You are working in an **isolated git worktree**. All development MUST happen here:\n\n")
	fmt.Fprintf(&b, "- **Working directory**: `%s`\n", wtPath)
	fmt.Fprintf(&b, "- **Git branch**: `team/%s` (your dedicated branch)\n", name)
	fmt.Fprintf(&b, "- **Main project root**: `%s`\n\n", root)
	b.WriteString("### Git Rules\n\n")
	fmt.Fprintf(&b, "- All changes and commits go to the `team/%s` branch — this is already checked out\n", name)
	b.WriteString("- **Never** run `git checkout`, `git switch`, or change branches\n")
	b.WriteString("- **Never** merge or rebase from within this worktree\n")
	b.WriteString("- Commit regularly with clear messages as you complete work\n\n")
	b.WriteString("The main controller will merge your branch back to main when ready.\n")

	return os.WriteFile(claudePath, []byte(b.String()), 0644)
}
```

**Step 4: Update `PromptMDContent` in `internal/role.go:125-144`**

```go
func PromptMDContent(name string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Role: %s\n\n", name)
	b.WriteString("## Description\n")
	b.WriteString("Describe this role's responsibilities here.\n\n")
	b.WriteString("## Expertise\n")
	b.WriteString("- List key areas of expertise\n\n")
	b.WriteString("## Behavior\n")
	b.WriteString("- How this role approaches tasks\n")
	b.WriteString("- Communication style and boundaries\n\n")
	b.WriteString("## Workflow\n\n")
	b.WriteString("When you receive a `[New Change Assigned]` message:\n")
	b.WriteString("1. Read the proposal at the specified change path\n")
	b.WriteString("2. Run `/opsx:continue` to create remaining artifacts (specs, design, tasks)\n")
	b.WriteString("3. Run `/opsx:apply` to implement tasks\n")
	b.WriteString("4. Run `/opsx:verify` to validate implementation\n")
	b.WriteString("5. Commit your work regularly\n\n")
	b.WriteString("## Communication Protocol\n\n")
	b.WriteString("When you need clarification or have a question for the main controller, use:\n\n")
	b.WriteString("```bash\n")
	fmt.Fprintf(&b, "ask claude \"%s: <your question here>\"\n", name)
	b.WriteString("```\n\n")
	b.WriteString("Wait for the main controller to reply. Replies will appear as:\n")
	b.WriteString("`[Main Controller Reply]`\n\n")
	b.WriteString("Do NOT proceed on blocked tasks until you receive a reply.\n")
	return b.String()
}
```

**Step 4: Run tests to verify they pass**

Run: `cd /Users/jsonlee/Projects/agent-team && go test ./internal/ -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/role.go internal/role_test.go
git commit -m "refactor: update role templates to use OpenSpec workflow"
```

---

### Task 3: Update `cmd/create.go` — Replace tasks/ with OpenSpec init

**Files:**
- Modify: `cmd/create.go:39-44`
- Modify: `cmd/create_test.go:86-91`

**Step 1: Update test in `cmd/create_test.go`**

Replace the task directory assertion (lines 86-91) with OpenSpec directory check:

```go
// In TestRunCreate, replace:
//   // Verify task directories
//   pendingDir := ...
//   if _, err := os.Stat(pendingDir); ...
// With:

	// Verify openspec directory (created by openspec init)
	// Note: In test environment openspec may not be installed,
	// so we check that tasks/pending is NOT created (old behavior removed)
	pendingDir := filepath.Join(wtPath, "agents", "teams", "backend", "tasks", "pending")
	if _, err := os.Stat(pendingDir); err == nil {
		t.Error("tasks/pending should no longer be created")
	}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jsonlee/Projects/agent-team && go test ./cmd/ -run TestRunCreate -v`
Expected: FAIL — create still makes tasks/ directories

**Step 3: Update `cmd/create.go` — replace tasks/ creation with OpenSpec init**

Replace lines 39-44:

```go
// OLD:
//	for _, sub := range []string{"tasks/pending", "tasks/done"} {
//		if err := os.MkdirAll(filepath.Join(teamsDir, sub), 0755); err != nil {
//			return err
//		}
//	}

// NEW:
	// Initialize OpenSpec in worktree
	if err := internal.EnsureOpenSpec(); err != nil {
		return fmt.Errorf("install openspec: %w", err)
	}
	if err := internal.OpenSpecInit(wtPath); err != nil {
		return fmt.Errorf("openspec init: %w", err)
	}
```

Update the import in `cmd/create.go` — remove `"path/filepath"` if no longer needed (it's still used for `teamsDir`). Actually `teamsDir` uses `internal.TeamsDir()` which returns a full path, so `path/filepath` can stay since it's used by the `internal` package call indirectly — actually check: no, `filepath` is not directly used in create.go after removal. Remove it.

Wait — `filepath.Join(teamsDir, sub)` is removed, but `teamsDir` is still used for config.yaml and prompt.md paths. Check: `filepath.Join(teamsDir, "config.yaml")` — yes, `filepath` is still needed. Keep the import.

Actually, looking at the code again: `cfg.Save(filepath.Join(teamsDir, "config.yaml"))` and `filepath.Join(teamsDir, "prompt.md")` — yes, `filepath` is still used. Keep it.

**Step 4: Run tests**

Run: `cd /Users/jsonlee/Projects/agent-team && go test ./cmd/ -run TestRunCreate -v`
Expected: PASS (note: openspec init may fail in test env if openspec not installed — see Task 3b)

**Step 3b: Handle test environment where openspec is not installed**

Add a test helper and make `create` gracefully handle OpenSpec not being available in tests. Update `App` struct to allow injecting an OpenSpec runner:

Actually, simpler approach: extract the OpenSpec calls into a function on App that tests can skip. Or: in `cmd/create.go`, add a package-level variable for testability:

```go
// cmd/create.go — add near top:
// openSpecSetup can be overridden in tests to skip OpenSpec initialization
var openSpecSetup = func(wtPath string) error {
	if err := internal.EnsureOpenSpec(); err != nil {
		return fmt.Errorf("install openspec: %w", err)
	}
	return internal.OpenSpecInit(wtPath)
}
```

Then in `RunCreate`, replace the direct calls with:
```go
	if err := openSpecSetup(wtPath); err != nil {
		return fmt.Errorf("openspec setup: %w", err)
	}
```

In `cmd/create_test.go`, override it:
```go
func TestRunCreate(t *testing.T) {
	// Skip OpenSpec in test environment
	openSpecSetup = func(wtPath string) error { return nil }
	defer func() { openSpecSetup = defaultOpenSpecSetup }()
	// ... rest of test ...
```

Actually, even cleaner — keep the var as the default and just set it in tests. Let me restructure:

```go
// cmd/create.go
var defaultOpenSpecSetup = func(wtPath string) error {
	if err := internal.EnsureOpenSpec(); err != nil {
		return fmt.Errorf("install openspec: %w", err)
	}
	return internal.OpenSpecInit(wtPath)
}

var openSpecSetup = defaultOpenSpecSetup
```

**Step 5: Run all tests**

Run: `cd /Users/jsonlee/Projects/agent-team && go test ./... -v`
Expected: ALL PASS

**Step 6: Commit**

```bash
git add cmd/create.go cmd/create_test.go
git commit -m "feat: replace tasks/ creation with OpenSpec init in create command"
```

---

### Task 4: Rewrite `cmd/assign.go` — OpenSpec change creation

**Files:**
- Modify: `cmd/assign.go` (full rewrite)
- Modify: `cmd/assign_test.go` (full rewrite)

**Step 1: Write the failing tests**

```go
// cmd/assign_test.go
package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JsonLee12138/agent-team/internal"
)

func TestRunAssign(t *testing.T) {
	openSpecSetup = func(wtPath string) error { return nil }
	defer func() { openSpecSetup = defaultOpenSpecSetup }()

	app, dir := initTestApp(t)
	mock := &MockBackend{
		SpawnedID:  "99",
		AlivePanes: map[string]bool{"99": true},
	}
	app.Session = mock
	app.RunCreate("worker")

	// Manually create openspec/changes/ directory (normally done by openspec init)
	wtPath := filepath.Join(dir, ".worktrees", "worker")
	os.MkdirAll(filepath.Join(wtPath, "openspec", "changes"), 0755)

	// Set pane as running
	configPath := filepath.Join(wtPath, "agents", "teams", "worker", "config.yaml")
	cfg, _ := internal.LoadRoleConfig(configPath)
	cfg.PaneID = "99"
	cfg.Save(configPath)

	err := app.RunAssign("worker", "Fix the login bug", "", "", "")
	if err != nil {
		t.Fatalf("RunAssign: %v", err)
	}

	// Change directory should exist under openspec/changes/
	changesDir := filepath.Join(wtPath, "openspec", "changes")
	entries, _ := os.ReadDir(changesDir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 change directory, got %d", len(entries))
	}

	// proposal.md should exist (empty since no --proposal)
	proposalPath := filepath.Join(changesDir, entries[0].Name(), "proposal.md")
	if _, err := os.Stat(proposalPath); os.IsNotExist(err) {
		t.Error("proposal.md not created")
	}

	// .openspec.yaml should exist
	metaPath := filepath.Join(changesDir, entries[0].Name(), ".openspec.yaml")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Error(".openspec.yaml not created")
	}

	// Notification should be sent to pane
	if len(mock.SentTexts) == 0 {
		t.Error("no notification sent to pane")
	}
	if !strings.Contains(mock.SentTexts[0], "[New Change Assigned]") {
		t.Errorf("notification = %q, want [New Change Assigned] prefix", mock.SentTexts[0])
	}
	if !strings.Contains(mock.SentTexts[0], "/opsx:continue") {
		t.Errorf("notification should mention /opsx:continue")
	}
}

func TestRunAssignWithProposal(t *testing.T) {
	openSpecSetup = func(wtPath string) error { return nil }
	defer func() { openSpecSetup = defaultOpenSpecSetup }()

	app, dir := initTestApp(t)
	mock := &MockBackend{
		SpawnedID:  "99",
		AlivePanes: map[string]bool{"99": true},
	}
	app.Session = mock
	app.RunCreate("worker")

	wtPath := filepath.Join(dir, ".worktrees", "worker")
	os.MkdirAll(filepath.Join(wtPath, "openspec", "changes"), 0755)

	configPath := filepath.Join(wtPath, "agents", "teams", "worker", "config.yaml")
	cfg, _ := internal.LoadRoleConfig(configPath)
	cfg.PaneID = "99"
	cfg.Save(configPath)

	// Create a proposal file
	proposalFile := filepath.Join(dir, "proposal.md")
	proposalContent := "# Proposal\n\nFix the login flow with JWT."
	os.WriteFile(proposalFile, []byte(proposalContent), 0644)

	err := app.RunAssign("worker", "Fix the login bug", "", "", proposalFile)
	if err != nil {
		t.Fatalf("RunAssign: %v", err)
	}

	// Verify proposal content was written
	changesDir := filepath.Join(wtPath, "openspec", "changes")
	entries, _ := os.ReadDir(changesDir)
	proposalPath := filepath.Join(changesDir, entries[0].Name(), "proposal.md")
	data, _ := os.ReadFile(proposalPath)
	if string(data) != proposalContent {
		t.Errorf("proposal content = %q, want %q", string(data), proposalContent)
	}
}

func TestRunAssignNotFound(t *testing.T) {
	app, _ := initTestApp(t)
	err := app.RunAssign("ghost", "task", "", "", "")
	if err == nil {
		t.Error("expected error for nonexistent role")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/jsonlee/Projects/agent-team && go test ./cmd/ -run "TestRunAssign" -v`
Expected: FAIL — RunAssign signature changed

**Step 3: Rewrite `cmd/assign.go`**

```go
// cmd/assign.go
package cmd

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newAssignCmd() *cobra.Command {
	var model string
	var proposal string
	cmd := &cobra.Command{
		Use:   `assign <name> "<description>" [provider]`,
		Short: "Create an OpenSpec change and notify the role session",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := ""
			if len(args) > 2 {
				provider = args[2]
			}
			return GetApp(cmd).RunAssign(args[0], args[1], provider, model, proposal)
		},
	}
	cmd.Flags().StringVarP(&model, "model", "m", "", "AI model identifier")
	cmd.Flags().StringVarP(&proposal, "proposal", "p", "", "Path to proposal file (use - for stdin)")
	return cmd
}

func (a *App) RunAssign(name, desc, provider, model, proposalPath string) error {
	root := a.Git.Root()
	teamsDir := internal.TeamsDir(root, a.WtBase, name)
	configPath := internal.ConfigPath(root, a.WtBase, name)
	wtPath := internal.WtPath(root, a.WtBase, name)

	if _, err := os.Stat(teamsDir); os.IsNotExist(err) {
		return fmt.Errorf("role '%s' not found", name)
	}

	// Read proposal content
	var proposalContent string
	if proposalPath != "" {
		var data []byte
		var err error
		if proposalPath == "-" {
			data, err = io.ReadAll(os.Stdin)
		} else {
			data, err = os.ReadFile(proposalPath)
		}
		if err != nil {
			return fmt.Errorf("read proposal: %w", err)
		}
		proposalContent = string(data)
	}

	// Create OpenSpec change
	ts := time.Now().Format("2006-01-02-15-04-05")
	slug := internal.Slugify(desc, 50)
	changeName := fmt.Sprintf("%s-%s", ts, slug)

	changePath, err := internal.CreateChange(wtPath, changeName, proposalContent)
	if err != nil {
		return err
	}
	fmt.Printf("✓ Change created: %s\n", changePath)

	// Ensure session is running
	cfg, err := internal.LoadRoleConfig(configPath)
	if err != nil {
		return err
	}

	if !a.Session.PaneAlive(cfg.PaneID) {
		fmt.Printf("Role '%s' is not running, opening session first...\n", name)
		if err := a.RunOpen(name, provider, model); err != nil {
			return err
		}
		cfg, err = internal.LoadRoleConfig(configPath)
		if err != nil {
			return err
		}
		fmt.Println("  Waiting for AI to initialize...")
		time.Sleep(3 * time.Second)
	}

	// Notify role
	changeRel := fmt.Sprintf("openspec/changes/%s/", changeName)
	msg := fmt.Sprintf("[New Change Assigned] %s\nChange: %s\nProposal ready. Run /opsx:continue to proceed.",
		desc, changeRel)
	a.Session.PaneSend(cfg.PaneID, msg)

	fmt.Printf("✓ Assigned to '%s': %s\n", name, desc)
	return nil
}
```

**Step 4: Run tests**

Run: `cd /Users/jsonlee/Projects/agent-team && go test ./cmd/ -run "TestRunAssign" -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add cmd/assign.go cmd/assign_test.go
git commit -m "feat: rewrite assign command to create OpenSpec changes"
```

---

### Task 5: Update `cmd/status.go` — Read OpenSpec status

**Files:**
- Modify: `cmd/status.go` (full rewrite)
- Modify: `cmd/commands_test.go:52-64` (TestRunStatus)

**Step 1: Update test**

```go
// In cmd/commands_test.go, update TestRunStatus:
func TestRunStatus(t *testing.T) {
	openSpecSetup = func(wtPath string) error { return nil }
	defer func() { openSpecSetup = defaultOpenSpecSetup }()

	app, dir := initTestApp(t)
	app.RunCreate("alpha")
	app.RunCreate("beta")

	// Create openspec/changes/ directories for testing
	for _, role := range []string{"alpha", "beta"} {
		os.MkdirAll(filepath.Join(dir, ".worktrees", role, "openspec", "changes"), 0755)
	}

	err := app.RunStatus()
	if err != nil {
		t.Fatalf("RunStatus: %v", err)
	}
}
```

**Step 2: Run test to verify it still passes (baseline)**

Run: `cd /Users/jsonlee/Projects/agent-team && go test ./cmd/ -run TestRunStatus -v`

**Step 3: Rewrite `cmd/status.go`**

```go
// cmd/status.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show all roles, running state, and OpenSpec change status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunStatus()
		},
	}
}

func (a *App) RunStatus() error {
	root := a.Git.Root()
	roles := internal.ListRoles(root, a.WtBase)
	if len(roles) == 0 {
		fmt.Println("No roles found. Create one with: agent-team create <name>")
		return nil
	}

	fmt.Printf("%-16s %-24s %s\n", "Role", "Status", "Changes")
	fmt.Printf("%-16s %-24s %s\n", "────────────────", "────────────────────────", "──────────────────────────")

	for _, role := range roles {
		configPath := internal.ConfigPath(root, a.WtBase, role)
		wtPath := internal.WtPath(root, a.WtBase, role)

		cfg, _ := internal.LoadRoleConfig(configPath)
		status := "✗ offline"
		if cfg != nil && a.Session.PaneAlive(cfg.PaneID) {
			status = fmt.Sprintf("✓ running [p:%s]", cfg.PaneID)
		}

		// Count changes by reading openspec/changes/ directory
		changesDir := filepath.Join(wtPath, "openspec", "changes")
		changesSummary := "0"
		if entries, err := os.ReadDir(changesDir); err == nil {
			active := 0
			for _, e := range entries {
				if e.IsDir() && e.Name() != "archive" {
					active++
				}
			}
			if active > 0 {
				changesSummary = fmt.Sprintf("%d active", active)
			} else {
				changesSummary = "0"
			}
		}

		fmt.Printf("%-16s %-24s %s\n", role, status, changesSummary)
	}
	return nil
}
```

**Step 4: Run tests**

Run: `cd /Users/jsonlee/Projects/agent-team && go test ./cmd/ -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add cmd/status.go cmd/commands_test.go
git commit -m "feat: update status command to show OpenSpec changes"
```

---

### Task 6: Update `skills/agent-team/SKILL.md` — Add brainstorming section

**Files:**
- Modify: `skills/agent-team/SKILL.md`
- Modify: `skills/agent-team/references/details.md`

**Step 1: Update SKILL.md**

Add brainstorming section and update assign documentation. Full new content:

```markdown
---
name: agent-team
description: >
  AI team role manager for multi-agent development workflows.
  Use when the user wants to create/delete team roles, open role sessions in terminal tabs,
  assign tasks to roles, check team status, or merge role branches.
  Triggers on /agent-team commands, "create a team role", "open role session",
  "assign task to role", "show team status", "merge role branch".
---

# agent-team

Manages AI team roles using git worktrees + terminal multiplexer tabs. Each role runs in its own isolated worktree (branch `team/<name>`) and opens as a full-permission AI session in a new tab.

For directory layout and bidirectional communication details, see [references/details.md](references/details.md).

## Install

```bash
brew tap JsonLee12138/agent-team && brew install agent-team
```

## Usage

Run from within a project git repository:

```bash
agent-team <command>
```

Use tmux backend (default is WezTerm):

```bash
AGENT_TEAM_BACKEND=tmux agent-team <command>
```

## Brainstorming (Required Before Assign)

When the user intends to assign new work to a role, you MUST brainstorm first:

1. **Explore context** — check the role's `prompt.md`, existing `openspec/specs/`, and project state
2. **Ask clarifying questions** — one at a time, prefer multiple choice when possible
3. **Propose 2-3 approaches** — with trade-offs and your recommendation
4. **User confirms design** — get explicit approval
5. **Write proposal** — save the confirmed design to a temp file
6. **Execute assign** — run `agent-team assign <name> "<desc>" --proposal <file>`

**Rules:**
- Brainstorming is **mandatory** for new work assignments
- Can be skipped when user explicitly says "just assign" or provides a complete design
- One question at a time. YAGNI. Explore alternatives before settling.

## Commands

### Create a role
```bash
agent-team create <name>
```
Creates `team/<name>` git branch + worktree at `.worktrees/<name>/`. Generates:
- `agents/teams/<name>/config.yaml` — provider, description, pane tracking
- `agents/teams/<name>/prompt.md` — role system prompt (edit this to define the role)
- `openspec/` — OpenSpec project structure for change management

After creating, guide the user to edit `prompt.md` to define the role's expertise and behavior.

### Open a role session
```bash
agent-team open <name> [claude|codex|opencode] [--model <model>]
```
- Generates `CLAUDE.md` in worktree root from `prompt.md` (auto-injected as system context)
- Spawns a new terminal tab titled `<name>` running the chosen AI provider
- Provider priority: CLI argument > `config.yaml default_provider` > claude
- Model priority: `--model` flag > `config.yaml default_model` > provider default

### Open all sessions
```bash
agent-team open-all [claude|codex|opencode] [--model <model>]
```
Opens every role that has a config.yaml.

### Assign a change
```bash
agent-team assign <name> "<description>" [claude|codex|opencode] [--model <model>] [--proposal <file>]
```
1. Creates an OpenSpec change at `openspec/changes/<timestamp>-<slug>/`
2. Writes the proposal file from `--proposal` flag (or empty if not provided)
3. Auto-opens the role session if not running
4. Sends a `[New Change Assigned]` notification to the running session

The role will then use `/opsx:continue` to proceed through specs → design → tasks → apply.

### Reply to a role
```bash
agent-team reply <name> "<answer>"
```
Sends a reply to a role's running session, prefixed with `[Main Controller Reply]`.

### Check status
```bash
agent-team status
```
Shows all roles, session status (running/stopped), and active OpenSpec changes.

### Merge completed work
```bash
agent-team merge <name>
```
Merges `team/<name>` into the current branch with `--no-ff`. Run `delete` afterward to clean up.

### Delete a role
```bash
agent-team delete <name>
```
Closes the running session (if any), removes the worktree, and deletes the `team/<name>` branch.
```

**Step 2: Update `references/details.md`**

```markdown
# agent-team Reference Details

## Role directory layout

```
.worktrees/<name>/
  CLAUDE.md                          <- auto-generated from prompt.md on open
  agents/teams/<name>/
    config.yaml                      <- name, default_provider, default_model, pane_id
    prompt.md                        <- role system prompt (edit manually)
  openspec/
    specs/                           <- project specifications
    changes/                         <- active changes (managed by OpenSpec)
      <change-name>/
        .openspec.yaml               <- change metadata
        proposal.md                  <- brainstorming output from controller
        specs/                       <- delta specs (created by role)
        design.md                    <- design artifact (created by role)
        tasks.md                     <- task breakdown (created by role)
    config.yaml                      <- OpenSpec configuration
```

## Change workflow

Changes are managed by OpenSpec. The controller creates a change with a proposal via `agent-team assign`. The role then proceeds through the OpenSpec workflow:

1. `/opsx:continue` — create remaining artifacts (specs, design, tasks)
2. `/opsx:apply` — implement the tasks
3. `/opsx:verify` — validate implementation matches the design

## Bidirectional communication

Role asks a question:
```bash
ask claude "<rolename>: <question>"
```

Main controller replies:
```bash
agent-team reply <rolename> "<answer>"
```

Reply appears in the role's terminal tab as `[Main Controller Reply]`. The role AI must NOT proceed on blocked tasks until it receives a reply.

The `prompt.md` template includes this communication protocol automatically.
```

**Step 3: Commit**

```bash
git add skills/agent-team/SKILL.md skills/agent-team/references/details.md
git commit -m "docs: update SKILL.md with brainstorming flow and OpenSpec commands"
```

---

### Task 7: Run full test suite and verify

**Step 1: Run all tests**

Run: `cd /Users/jsonlee/Projects/agent-team && go test ./... -v`
Expected: ALL PASS

**Step 2: Build binary**

Run: `cd /Users/jsonlee/Projects/agent-team && make build`
Expected: Binary builds successfully

**Step 3: Run linter**

Run: `cd /Users/jsonlee/Projects/agent-team && make lint`
Expected: No issues

**Step 4: Commit if any fixes needed, then final commit**

```bash
git add -A
git commit -m "chore: final cleanup for brainstorming + OpenSpec integration"
```
