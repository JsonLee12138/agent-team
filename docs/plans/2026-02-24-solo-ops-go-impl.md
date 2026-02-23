# solo-ops Go Rewrite Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Rewrite solo-ops from Python to Go, producing a single binary distributable via Homebrew.

**Architecture:** Strategy pattern for terminal backends (SessionBackend interface), Facade for git operations (GitClient), dependency injection via App struct. All internal logic in `internal/`, cobra subcommands in `cmd/`, wired in `main.go`.

**Tech Stack:** Go 1.23+, cobra (CLI), yaml.v3 (config), goreleaser (distribution)

---

### Task 1: Project Scaffolding

**Files:**
- Create: `go.mod`
- Create: `Makefile`
- Create: `internal/.gitkeep` (placeholder)
- Create: `cmd/.gitkeep` (placeholder)

**Step 1: Initialize Go module and directories**

```bash
cd /Users/jsonlee/Projects/agent-team
go mod init github.com/leeforge/agent-team
mkdir -p cmd internal
```

**Step 2: Install dependencies**

```bash
go get github.com/spf13/cobra@latest
go get gopkg.in/yaml.v3@latest
```

**Step 3: Create Makefile**

```makefile
BINARY := solo-ops
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build test lint clean

build:
	go build -ldflags "-X github.com/leeforge/agent-team/cmd.Version=$(VERSION)" -o $(BINARY) .

test:
	go test ./... -v

lint:
	go vet ./...

clean:
	rm -f $(BINARY)
```

**Step 4: Commit**

```bash
git add go.mod go.sum Makefile cmd/ internal/
git commit -m "chore: scaffold Go project with cobra and yaml.v3"
```

---

### Task 2: internal/config — RoleConfig

**Files:**
- Create: `internal/config.go`
- Test: `internal/config_test.go`

**Step 1: Write the failing test**

```go
// internal/config_test.go
package internal

import (
	"path/filepath"
	"testing"
)

func TestRoleConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	original := &RoleConfig{
		Name:            "backend",
		Description:     "Backend developer",
		DefaultProvider: "claude",
		DefaultModel:    "claude-sonnet-4-6",
		CreatedAt:       "2026-02-24T10:00:00Z",
		PaneID:          "42",
	}

	if err := original.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := LoadRoleConfig(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Name != original.Name {
		t.Errorf("Name = %q, want %q", loaded.Name, original.Name)
	}
	if loaded.DefaultProvider != original.DefaultProvider {
		t.Errorf("DefaultProvider = %q, want %q", loaded.DefaultProvider, original.DefaultProvider)
	}
	if loaded.PaneID != original.PaneID {
		t.Errorf("PaneID = %q, want %q", loaded.PaneID, original.PaneID)
	}
}

func TestLoadRoleConfigNotFound(t *testing.T) {
	_, err := LoadRoleConfig("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestRoleConfigSaveUpdate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &RoleConfig{Name: "test", PaneID: ""}
	cfg.Save(path)

	cfg.PaneID = "99"
	cfg.Save(path)

	reloaded, _ := LoadRoleConfig(path)
	if reloaded.PaneID != "99" {
		t.Errorf("PaneID after update = %q, want %q", reloaded.PaneID, "99")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/ -run TestRoleConfig -v`
Expected: FAIL — `RoleConfig` not defined

**Step 3: Write minimal implementation**

```go
// internal/config.go
package internal

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type RoleConfig struct {
	Name            string `yaml:"name"`
	Description     string `yaml:"description"`
	DefaultProvider string `yaml:"default_provider"`
	DefaultModel    string `yaml:"default_model"`
	CreatedAt       string `yaml:"created_at"`
	PaneID          string `yaml:"pane_id"`
}

func LoadRoleConfig(path string) (*RoleConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	var cfg RoleConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	return &cfg, nil
}

func (c *RoleConfig) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/ -run TestRoleConfig -v`
Expected: PASS (3 tests)

**Step 5: Commit**

```bash
git add internal/config.go internal/config_test.go
git commit -m "feat: add RoleConfig YAML load/save"
```

---

### Task 3: internal/session — SessionBackend Interface + Implementations

**Files:**
- Create: `internal/session.go`
- Test: `internal/session_test.go`

**Step 1: Write the failing test**

```go
// internal/session_test.go
package internal

import (
	"testing"
)

func TestNewSessionBackendDefault(t *testing.T) {
	t.Setenv("SOLO_OPS_BACKEND", "")
	b := NewSessionBackend()
	if _, ok := b.(*WeztermBackend); !ok {
		t.Errorf("expected WeztermBackend, got %T", b)
	}
}

func TestNewSessionBackendTmux(t *testing.T) {
	t.Setenv("SOLO_OPS_BACKEND", "tmux")
	b := NewSessionBackend()
	if _, ok := b.(*TmuxBackend); !ok {
		t.Errorf("expected TmuxBackend, got %T", b)
	}
}

func TestNewSessionBackendCaseInsensitive(t *testing.T) {
	t.Setenv("SOLO_OPS_BACKEND", "  TMUX  ")
	b := NewSessionBackend()
	if _, ok := b.(*TmuxBackend); !ok {
		t.Errorf("expected TmuxBackend for 'TMUX', got %T", b)
	}
}

func TestPaneAliveEmptyID(t *testing.T) {
	t.Setenv("SOLO_OPS_BACKEND", "")
	b := NewSessionBackend()
	if b.PaneAlive("") {
		t.Error("empty pane ID should not be alive")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/ -run TestNewSessionBackend -v`
Expected: FAIL — `SessionBackend` not defined

**Step 3: Write implementation**

```go
// internal/session.go
package internal

import (
	"os"
	"os/exec"
	"strings"
	"time"
)

// SessionBackend abstracts terminal multiplexer operations (Strategy pattern).
type SessionBackend interface {
	PaneAlive(paneID string) bool
	PaneSend(paneID string, text string) error
	SpawnPane(cwd string) (paneID string, err error)
	KillPane(paneID string) error
	SetTitle(paneID string, title string) error
	ActivatePane(paneID string) error
}

func NewSessionBackend() SessionBackend {
	backend := strings.TrimSpace(strings.ToLower(os.Getenv("SOLO_OPS_BACKEND")))
	if backend == "tmux" {
		return &TmuxBackend{}
	}
	return &WeztermBackend{}
}

// --- WeztermBackend ---

type WeztermBackend struct{}

func (w *WeztermBackend) PaneAlive(paneID string) bool {
	if paneID == "" {
		return false
	}
	out, err := exec.Command("wezterm", "cli", "list").Output()
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(out), "\n")[1:] {
		parts := strings.Fields(line)
		if len(parts) >= 3 && parts[2] == paneID {
			return true
		}
	}
	return false
}

func (w *WeztermBackend) PaneSend(paneID string, text string) error {
	cmd := exec.Command("wezterm", "cli", "send-text", "--pane-id", paneID, "--no-paste")
	cmd.Stdin = strings.NewReader(text)
	if err := cmd.Run(); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	cmd2 := exec.Command("wezterm", "cli", "send-text", "--pane-id", paneID, "--no-paste")
	cmd2.Stdin = strings.NewReader("\r")
	return cmd2.Run()
}

func (w *WeztermBackend) SpawnPane(cwd string) (string, error) {
	out, err := exec.Command("wezterm", "cli", "spawn", "--cwd", cwd).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (w *WeztermBackend) KillPane(paneID string) error {
	return exec.Command("wezterm", "cli", "kill-pane", "--pane-id", paneID).Run()
}

func (w *WeztermBackend) SetTitle(paneID string, title string) error {
	return exec.Command("wezterm", "cli", "set-tab-title", "--pane-id", paneID, title).Run()
}

func (w *WeztermBackend) ActivatePane(paneID string) error {
	return exec.Command("wezterm", "cli", "activate-pane", "--pane-id", paneID).Run()
}

// --- TmuxBackend ---

type TmuxBackend struct{}

func (t *TmuxBackend) PaneAlive(paneID string) bool {
	if paneID == "" {
		return false
	}
	out, err := exec.Command("tmux", "list-panes", "-a", "-F", "#{pane_id}").Output()
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == paneID {
			return true
		}
	}
	return false
}

func (t *TmuxBackend) PaneSend(paneID string, text string) error {
	if err := exec.Command("tmux", "send-keys", "-t", paneID, "-l", text).Run(); err != nil {
		return err
	}
	return exec.Command("tmux", "send-keys", "-t", paneID, "Enter").Run()
}

func (t *TmuxBackend) SpawnPane(cwd string) (string, error) {
	out, err := exec.Command("tmux", "new-session", "-d", "-P", "-F", "#{pane_id}", "-c", cwd).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (t *TmuxBackend) KillPane(paneID string) error {
	return exec.Command("tmux", "kill-pane", "-t", paneID).Run()
}

func (t *TmuxBackend) SetTitle(paneID string, title string) error {
	return exec.Command("tmux", "rename-window", "-t", paneID, title).Run()
}

func (t *TmuxBackend) ActivatePane(_ string) error {
	return nil // tmux does not steal focus
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/ -run "TestNewSessionBackend|TestPaneAlive" -v`
Expected: PASS (4 tests)

**Step 5: Commit**

```bash
git add internal/session.go internal/session_test.go
git commit -m "feat: add SessionBackend interface with Wezterm and Tmux implementations"
```

---

### Task 4: internal/git — GitClient

**Files:**
- Create: `internal/git.go`
- Test: `internal/git_test.go`

**Step 1: Write the failing test**

```go
// internal/git_test.go
package internal

import (
	"os/exec"
	"testing"
)

func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
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
	return dir
}

func TestNewGitClient(t *testing.T) {
	dir := initTestRepo(t)
	gc, err := NewGitClient(dir)
	if err != nil {
		t.Fatalf("NewGitClient: %v", err)
	}
	if gc.Root() != dir {
		t.Errorf("Root() = %q, want %q", gc.Root(), dir)
	}
}

func TestGitClientCurrentBranch(t *testing.T) {
	dir := initTestRepo(t)
	gc, _ := NewGitClient(dir)
	branch, err := gc.CurrentBranch()
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	// git init creates "main" or "master" depending on config
	if branch == "" {
		t.Error("CurrentBranch returned empty string")
	}
}

func TestGitClientWorktreeAddRemove(t *testing.T) {
	dir := initTestRepo(t)
	gc, _ := NewGitClient(dir)

	wtPath := dir + "/.worktrees/test-role"
	if err := gc.WorktreeAdd(wtPath, "team/test-role"); err != nil {
		t.Fatalf("WorktreeAdd: %v", err)
	}

	if err := gc.WorktreeRemove(wtPath); err != nil {
		t.Fatalf("WorktreeRemove: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/ -run TestGitClient -v`
Expected: FAIL — `NewGitClient` not defined

**Step 3: Write implementation**

```go
// internal/git.go
package internal

import (
	"fmt"
	"os/exec"
	"strings"
)

type GitClient struct {
	root string
}

func NewGitClient(dir string) (*GitClient, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("not a git repository: %w", err)
	}
	return &GitClient{root: strings.TrimSpace(string(out))}, nil
}

func (g *GitClient) Root() string {
	return g.root
}

func (g *GitClient) CurrentBranch() (string, error) {
	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	cmd.Dir = g.root
	out, err := cmd.Output()
	if err != nil {
		return "main", nil
	}
	return strings.TrimSpace(string(out)), nil
}

func (g *GitClient) WorktreeAdd(path, branch string) error {
	cmd := exec.Command("git", "worktree", "add", path, "-b", branch)
	cmd.Dir = g.root
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("worktree add: %s (%w)", out, err)
	}
	return nil
}

func (g *GitClient) WorktreeRemove(path string) error {
	cmd := exec.Command("git", "worktree", "remove", path, "--force")
	cmd.Dir = g.root
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("worktree remove: %s (%w)", out, err)
	}
	return nil
}

func (g *GitClient) Merge(branch, message string) error {
	cmd := exec.Command("git", "merge", branch, "--no-ff", "-m", message)
	cmd.Dir = g.root
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("merge: %s (%w)", out, err)
	}
	return nil
}

func (g *GitClient) DeleteBranch(branch string) error {
	cmd := exec.Command("git", "branch", "-D", branch)
	cmd.Dir = g.root
	cmd.CombinedOutput() // ignore error — branch may not exist
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/ -run TestGitClient -v`
Expected: PASS (3 tests)

**Step 5: Commit**

```bash
git add internal/git.go internal/git_test.go
git commit -m "feat: add GitClient facade for git operations"
```

---

### Task 5: internal/role — Role Discovery, Helpers, Templates

**Files:**
- Create: `internal/role.go`
- Test: `internal/role_test.go`

**Step 1: Write the failing test**

```go
// internal/role_test.go
package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindWtBase(t *testing.T) {
	dir := t.TempDir()

	// no worktrees dir exists → default ".worktrees"
	if got := FindWtBase(dir); got != ".worktrees" {
		t.Errorf("FindWtBase(empty) = %q, want .worktrees", got)
	}

	// create .worktrees
	os.Mkdir(filepath.Join(dir, ".worktrees"), 0755)
	if got := FindWtBase(dir); got != ".worktrees" {
		t.Errorf("FindWtBase(.worktrees) = %q, want .worktrees", got)
	}
}

func TestListRoles(t *testing.T) {
	dir := t.TempDir()
	wtBase := ".worktrees"

	// empty → no roles
	roles := ListRoles(dir, wtBase)
	if len(roles) != 0 {
		t.Errorf("ListRoles(empty) = %v, want empty", roles)
	}

	// create a role
	teamsDir := filepath.Join(dir, wtBase, "backend", "agents", "teams", "backend")
	os.MkdirAll(teamsDir, 0755)
	os.WriteFile(filepath.Join(teamsDir, "config.yaml"), []byte("name: backend\n"), 0644)

	roles = ListRoles(dir, wtBase)
	if len(roles) != 1 || roles[0] != "backend" {
		t.Errorf("ListRoles = %v, want [backend]", roles)
	}
}

func TestBuildLaunchCmd(t *testing.T) {
	tests := []struct {
		provider, model, want string
	}{
		{"claude", "", "claude --dangerously-skip-permissions"},
		{"codex", "gpt-5", "codex --dangerously-bypass-approvals-and-sandbox --model gpt-5"},
		{"opencode", "", "opencode"},
		{"", "", "claude --dangerously-skip-permissions"},
	}
	for _, tt := range tests {
		got := BuildLaunchCmd(tt.provider, tt.model)
		if got != tt.want {
			t.Errorf("BuildLaunchCmd(%q, %q) = %q, want %q", tt.provider, tt.model, got, tt.want)
		}
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"Fix the login bug", "fix-the-login-bug"},
		{"Hello, World!!!", "hello-world"},
		{"", "task"},
	}
	for _, tt := range tests {
		got := Slugify(tt.input, 50)
		if got != tt.want {
			t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestGenerateClaudeMD(t *testing.T) {
	dir := t.TempDir()
	wtPath := filepath.Join(dir, ".worktrees", "dev")
	teamsDir := filepath.Join(wtPath, "agents", "teams", "dev")
	os.MkdirAll(teamsDir, 0755)

	promptContent := "# Role: dev\n\nA developer role.\n"
	os.WriteFile(filepath.Join(teamsDir, "prompt.md"), []byte(promptContent), 0644)

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
}

func TestPromptMDTemplate(t *testing.T) {
	content := PromptMDContent("backend")
	if !strings.Contains(content, "# Role: backend") {
		t.Error("prompt template should contain role name")
	}
	if !strings.Contains(content, "Communication Protocol") {
		t.Error("prompt template should contain communication protocol")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/ -run "TestFindWtBase|TestListRoles|TestBuildLaunchCmd|TestSlugify|TestGenerateClaudeMD|TestPromptMD" -v`
Expected: FAIL — functions not defined

**Step 3: Write implementation**

```go
// internal/role.go
package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var SupportedProviders = map[string]bool{
	"claude":   true,
	"codex":    true,
	"opencode": true,
}

var launchCommands = map[string]string{
	"claude":   "claude --dangerously-skip-permissions",
	"codex":    "codex --dangerously-bypass-approvals-and-sandbox",
	"opencode": "opencode",
}

func FindWtBase(root string) string {
	if info, err := os.Stat(filepath.Join(root, ".worktrees")); err == nil && info.IsDir() {
		return ".worktrees"
	}
	if info, err := os.Stat(filepath.Join(root, "worktrees")); err == nil && info.IsDir() {
		return "worktrees"
	}
	return ".worktrees"
}

func ListRoles(root, wtBase string) []string {
	base := filepath.Join(root, wtBase)
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil
	}
	var roles []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		configPath := filepath.Join(base, name, "agents", "teams", name, "config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			roles = append(roles, name)
		}
	}
	return roles
}

func TeamsDir(root, wtBase, name string) string {
	return filepath.Join(root, wtBase, name, "agents", "teams", name)
}

func WtPath(root, wtBase, name string) string {
	return filepath.Join(root, wtBase, name)
}

func ConfigPath(root, wtBase, name string) string {
	return filepath.Join(TeamsDir(root, wtBase, name), "config.yaml")
}

func BuildLaunchCmd(provider, model string) string {
	if provider == "" {
		provider = "claude"
	}
	base, ok := launchCommands[provider]
	if !ok {
		base = launchCommands["claude"]
	}
	if model != "" {
		return base + " --model " + model
	}
	return base
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func Slugify(text string, maxLen int) string {
	s := slugRe.ReplaceAllString(strings.ToLower(text), "-")
	s = strings.Trim(s, "-")
	if len(s) > maxLen {
		s = s[:maxLen]
	}
	if s == "" {
		return "task"
	}
	return s
}

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
	b.WriteString("- Commit regularly with clear messages as you complete work\n")
	b.WriteString("- When your task is fully done, move its file from `tasks/pending/` to `tasks/done/`\n\n")
	b.WriteString("The main controller will merge your branch back to main when ready.\n")

	return os.WriteFile(claudePath, []byte(b.String()), 0644)
}

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

**Step 4: Run test to verify it passes**

Run: `go test ./internal/ -run "TestFindWtBase|TestListRoles|TestBuildLaunchCmd|TestSlugify|TestGenerateClaudeMD|TestPromptMD" -v`
Expected: PASS (6 tests)

**Step 5: Commit**

```bash
git add internal/role.go internal/role_test.go
git commit -m "feat: add role discovery, path helpers, and template generation"
```

---

### Task 6: cmd/root — App Struct + Cobra Root Command

**Files:**
- Create: `cmd/root.go`
- Create: `main.go`

**Step 1: Write root command and App struct**

```go
// cmd/root.go
package cmd

import (
	"fmt"
	"os"

	"github.com/leeforge/agent-team/internal"
	"github.com/spf13/cobra"
)

var Version = "dev"

type App struct {
	Git     *internal.GitClient
	Session internal.SessionBackend
	WtBase  string
}

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "solo-ops",
		Short:   "AI team role manager",
		Long:    "Manages AI team roles using git worktrees and terminal multiplexer tabs.",
		Version: Version,
	}

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Skip init for help/version/completion
		if cmd.Name() == "help" || cmd.Name() == "version" || cmd.Name() == "completion" {
			return nil
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}

		gc, err := internal.NewGitClient(cwd)
		if err != nil {
			return fmt.Errorf("not in a git repository")
		}

		app := &App{
			Git:     gc,
			Session: internal.NewSessionBackend(),
			WtBase:  internal.FindWtBase(gc.Root()),
		}
		cmd.SetContext(WithApp(cmd.Context(), app))
		return nil
	}

	return rootCmd
}

// Context key for App
type appKey struct{}

func WithApp(ctx context.Context, app *App) context.Context {
	return context.WithValue(ctx, appKey{}, app)
}

func GetApp(cmd *cobra.Command) *App {
	return cmd.Context().Value(appKey{}).(*App)
}
```

Note: add `"context"` to imports.

**Step 2: Write main.go**

```go
// main.go
package main

import (
	"os"

	"github.com/leeforge/agent-team/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	cmd.RegisterCommands(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

**Step 3: Add RegisterCommands placeholder**

Add to bottom of `cmd/root.go`:

```go
func RegisterCommands(rootCmd *cobra.Command) {
	// Commands will be added in subsequent tasks
}
```

**Step 4: Build and verify**

Run: `go build -o solo-ops . && ./solo-ops --version`
Expected: `solo-ops version dev`

Run: `./solo-ops --help`
Expected: Help text with "AI team role manager"

**Step 5: Commit**

```bash
git add cmd/root.go main.go
git commit -m "feat: add cobra root command with App dependency injection"
```

---

### Task 7: cmd/create

**Files:**
- Create: `cmd/create.go`
- Test: `cmd/create_test.go`

**Step 1: Write the failing test**

```go
// cmd/create_test.go
package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/leeforge/agent-team/internal"
)

func initTestApp(t *testing.T) (*App, string) {
	t.Helper()
	dir := t.TempDir()
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

	gc, err := internal.NewGitClient(dir)
	if err != nil {
		t.Fatal(err)
	}

	return &App{
		Git:     gc,
		Session: &MockBackend{},
		WtBase:  ".worktrees",
	}, dir
}

// MockBackend implements SessionBackend for testing
type MockBackend struct {
	AlivePanes map[string]bool
	SentTexts  []string
	SpawnedID  string
}

func (m *MockBackend) PaneAlive(id string) bool     { return m.AlivePanes[id] }
func (m *MockBackend) PaneSend(_ string, t string) error { m.SentTexts = append(m.SentTexts, t); return nil }
func (m *MockBackend) SpawnPane(_ string) (string, error) { return m.SpawnedID, nil }
func (m *MockBackend) KillPane(_ string) error       { return nil }
func (m *MockBackend) SetTitle(_, _ string) error     { return nil }
func (m *MockBackend) ActivatePane(_ string) error    { return nil }

func TestRunCreate(t *testing.T) {
	app, dir := initTestApp(t)

	if err := app.RunCreate("backend"); err != nil {
		t.Fatalf("RunCreate: %v", err)
	}

	// Verify worktree directory
	wtPath := filepath.Join(dir, ".worktrees", "backend")
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Error("worktree directory not created")
	}

	// Verify config.yaml
	configPath := filepath.Join(wtPath, "agents", "teams", "backend", "config.yaml")
	cfg, err := internal.LoadRoleConfig(configPath)
	if err != nil {
		t.Fatalf("LoadRoleConfig: %v", err)
	}
	if cfg.Name != "backend" {
		t.Errorf("config.Name = %q, want backend", cfg.Name)
	}
	if cfg.DefaultProvider != "claude" {
		t.Errorf("config.DefaultProvider = %q, want claude", cfg.DefaultProvider)
	}

	// Verify prompt.md
	promptPath := filepath.Join(wtPath, "agents", "teams", "backend", "prompt.md")
	if _, err := os.Stat(promptPath); os.IsNotExist(err) {
		t.Error("prompt.md not created")
	}

	// Verify task directories
	pendingDir := filepath.Join(wtPath, "agents", "teams", "backend", "tasks", "pending")
	if _, err := os.Stat(pendingDir); os.IsNotExist(err) {
		t.Error("tasks/pending not created")
	}
}

func TestRunCreateDuplicate(t *testing.T) {
	app, _ := initTestApp(t)
	app.RunCreate("dup")
	err := app.RunCreate("dup")
	if err == nil {
		t.Error("expected error for duplicate role")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/ -run TestRunCreate -v`
Expected: FAIL — `RunCreate` not defined

**Step 3: Write implementation**

```go
// cmd/create.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/leeforge/agent-team/internal"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new role with git worktree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunCreate(args[0])
		},
	}
}

func (a *App) RunCreate(name string) error {
	root := a.Git.Root()
	wtPath := internal.WtPath(root, a.WtBase, name)
	branch := "team/" + name

	if _, err := os.Stat(wtPath); err == nil {
		return fmt.Errorf("role '%s' already exists at %s", name, wtPath)
	}

	fmt.Printf("Creating role '%s'...\n", name)
	if err := a.Git.WorktreeAdd(wtPath, branch); err != nil {
		return err
	}

	teamsDir := internal.TeamsDir(root, a.WtBase, name)
	for _, sub := range []string{"tasks/pending", "tasks/done"} {
		if err := os.MkdirAll(filepath.Join(teamsDir, sub), 0755); err != nil {
			return err
		}
	}

	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	cfg := &internal.RoleConfig{
		Name:            name,
		Description:     "",
		DefaultProvider: "claude",
		DefaultModel:    "",
		CreatedAt:       now,
		PaneID:          "",
	}
	if err := cfg.Save(filepath.Join(teamsDir, "config.yaml")); err != nil {
		return err
	}

	promptPath := filepath.Join(teamsDir, "prompt.md")
	if err := os.WriteFile(promptPath, []byte(internal.PromptMDContent(name)), 0644); err != nil {
		return err
	}

	fmt.Printf("✓ Created role '%s' at %s\n", name, wtPath)
	fmt.Printf("  → Edit %s/prompt.md to define the role\n", teamsDir)
	fmt.Printf("  → Edit %s/config.yaml to set default_provider\n", teamsDir)
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/ -run TestRunCreate -v`
Expected: PASS (2 tests)

**Step 5: Register command and commit**

Update `RegisterCommands` in `cmd/root.go`:

```go
func RegisterCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(newCreateCmd())
}
```

```bash
git add cmd/create.go cmd/create_test.go cmd/root.go
git commit -m "feat: add create command"
```

---

### Task 8: cmd/delete

**Files:**
- Create: `cmd/delete.go`
- Test: `cmd/delete_test.go`

**Step 1: Write the failing test**

```go
// cmd/delete_test.go
package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunDelete(t *testing.T) {
	app, dir := initTestApp(t)
	app.RunCreate("ephemeral")

	// Verify it exists
	wtPath := filepath.Join(dir, ".worktrees", "ephemeral")
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Fatal("setup: role not created")
	}

	if err := app.RunDelete("ephemeral"); err != nil {
		t.Fatalf("RunDelete: %v", err)
	}

	// Worktree should be gone
	if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
		t.Error("worktree directory still exists after delete")
	}
}

func TestRunDeleteNotFound(t *testing.T) {
	app, _ := initTestApp(t)
	err := app.RunDelete("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent role")
	}
}

func TestRunDeleteKillsRunningPane(t *testing.T) {
	app, dir := initTestApp(t)
	app.RunCreate("running")

	// Simulate a running pane
	mock := &MockBackend{AlivePanes: map[string]bool{"123": true}}
	app.Session = mock

	configPath := filepath.Join(dir, ".worktrees", "running", "agents", "teams", "running", "config.yaml")
	cfg, _ := internal.LoadRoleConfig(configPath)
	cfg.PaneID = "123"
	cfg.Save(configPath)

	app.RunDelete("running")
	// No assertion on KillPane — just ensure no panic
}
```

Note: add `"github.com/leeforge/agent-team/internal"` to imports.

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/ -run TestRunDelete -v`
Expected: FAIL — `RunDelete` not defined

**Step 3: Write implementation**

```go
// cmd/delete.go
package cmd

import (
	"fmt"
	"os"

	"github.com/leeforge/agent-team/internal"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Remove a role and its worktree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunDelete(args[0])
		},
	}
}

func (a *App) RunDelete(name string) error {
	root := a.Git.Root()
	wtPath := internal.WtPath(root, a.WtBase, name)
	configPath := internal.ConfigPath(root, a.WtBase, name)

	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		return fmt.Errorf("role '%s' not found", name)
	}

	fmt.Printf("Deleting role '%s'...\n", name)

	// Kill running pane if any
	cfg, err := internal.LoadRoleConfig(configPath)
	if err == nil && a.Session.PaneAlive(cfg.PaneID) {
		if killErr := a.Session.KillPane(cfg.PaneID); killErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close pane %s; continuing delete\n", cfg.PaneID)
		}
	}

	// Remove worktree
	if err := a.Git.WorktreeRemove(wtPath); err != nil {
		// Fallback: force remove directory
		os.RemoveAll(wtPath)
	}

	a.Git.DeleteBranch("team/" + name)
	fmt.Printf("✓ Deleted role '%s'\n", name)
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/ -run TestRunDelete -v`
Expected: PASS (3 tests)

**Step 5: Register and commit**

Add `rootCmd.AddCommand(newDeleteCmd())` to `RegisterCommands`.

```bash
git add cmd/delete.go cmd/delete_test.go cmd/root.go
git commit -m "feat: add delete command"
```

---

### Task 9: cmd/open + open-all

**Files:**
- Create: `cmd/open.go`
- Test: `cmd/open_test.go`

**Step 1: Write the failing test**

```go
// cmd/open_test.go
package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leeforge/agent-team/internal"
)

func TestRunOpen(t *testing.T) {
	app, dir := initTestApp(t)
	mock := &MockBackend{SpawnedID: "77"}
	app.Session = mock
	app.RunCreate("dev")

	if err := app.RunOpen("dev", "", ""); err != nil {
		t.Fatalf("RunOpen: %v", err)
	}

	// pane_id should be saved
	configPath := filepath.Join(dir, ".worktrees", "dev", "agents", "teams", "dev", "config.yaml")
	cfg, _ := internal.LoadRoleConfig(configPath)
	if cfg.PaneID != "77" {
		t.Errorf("PaneID = %q, want 77", cfg.PaneID)
	}

	// CLAUDE.md should exist
	claudeMD := filepath.Join(dir, ".worktrees", "dev", "CLAUDE.md")
	data, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatalf("CLAUDE.md not found: %v", err)
	}
	if !strings.Contains(string(data), "team/dev") {
		t.Error("CLAUDE.md missing worktree context")
	}

	// launch command should be sent
	if len(mock.SentTexts) == 0 {
		t.Error("no command sent to pane")
	}
	if !strings.Contains(mock.SentTexts[0], "claude") {
		t.Errorf("sent text = %q, want claude launch command", mock.SentTexts[0])
	}
}

func TestRunOpenAlreadyRunning(t *testing.T) {
	app, dir := initTestApp(t)
	app.RunCreate("running")

	// Set pane_id and mark alive
	configPath := filepath.Join(dir, ".worktrees", "running", "agents", "teams", "running", "config.yaml")
	cfg, _ := internal.LoadRoleConfig(configPath)
	cfg.PaneID = "55"
	cfg.Save(configPath)

	mock := &MockBackend{AlivePanes: map[string]bool{"55": true}}
	app.Session = mock

	// Should not error, just print message
	err := app.RunOpen("running", "", "")
	if err != nil {
		t.Fatalf("RunOpen on running role should not error: %v", err)
	}
}

func TestRunOpenWithProvider(t *testing.T) {
	app, _ := initTestApp(t)
	mock := &MockBackend{SpawnedID: "88"}
	app.Session = mock
	app.RunCreate("test")

	app.RunOpen("test", "codex", "gpt-5")

	if len(mock.SentTexts) == 0 {
		t.Fatal("no command sent")
	}
	if !strings.Contains(mock.SentTexts[0], "codex") {
		t.Errorf("sent = %q, want codex command", mock.SentTexts[0])
	}
	if !strings.Contains(mock.SentTexts[0], "--model gpt-5") {
		t.Errorf("sent = %q, want --model gpt-5", mock.SentTexts[0])
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/ -run TestRunOpen -v`
Expected: FAIL — `RunOpen` not defined

**Step 3: Write implementation**

```go
// cmd/open.go
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/leeforge/agent-team/internal"
	"github.com/spf13/cobra"
)

func newOpenCmd() *cobra.Command {
	var model string
	cmd := &cobra.Command{
		Use:   "open <name> [provider]",
		Short: "Open a role session in a new terminal tab",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := ""
			if len(args) > 1 {
				provider = args[1]
			}
			return GetApp(cmd).RunOpen(args[0], provider, model)
		},
	}
	cmd.Flags().StringVarP(&model, "model", "m", "", "AI model identifier")
	return cmd
}

func newOpenAllCmd() *cobra.Command {
	var model string
	cmd := &cobra.Command{
		Use:   "open-all [provider]",
		Short: "Open all role sessions",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := ""
			if len(args) > 0 {
				provider = args[0]
			}
			return GetApp(cmd).RunOpenAll(provider, model)
		},
	}
	cmd.Flags().StringVarP(&model, "model", "m", "", "AI model identifier")
	return cmd
}

func (a *App) RunOpen(name, provider, model string) error {
	root := a.Git.Root()
	teamsDir := internal.TeamsDir(root, a.WtBase, name)
	configPath := internal.ConfigPath(root, a.WtBase, name)
	wtPath := internal.WtPath(root, a.WtBase, name)

	if _, err := os.Stat(teamsDir); os.IsNotExist(err) {
		return fmt.Errorf("role '%s' not found", name)
	}

	cfg, err := internal.LoadRoleConfig(configPath)
	if err != nil {
		return err
	}

	if provider == "" {
		provider = cfg.DefaultProvider
		if provider == "" {
			provider = "claude"
		}
	}

	if a.Session.PaneAlive(cfg.PaneID) {
		fmt.Printf("Role '%s' is already running (pane %s)\n", name, cfg.PaneID)
		return nil
	}

	// Generate CLAUDE.md
	if err := internal.GenerateClaudeMD(wtPath, name, root); err != nil {
		return fmt.Errorf("generate CLAUDE.md: %w", err)
	}

	// Spawn pane
	paneID, err := a.Session.SpawnPane(wtPath)
	if err != nil || paneID == "" {
		return fmt.Errorf("failed to open session for '%s': %w", name, err)
	}

	a.Session.SetTitle(paneID, name)

	// Return focus (wezterm only)
	if currentPane := os.Getenv("WEZTERM_PANE"); currentPane != "" {
		a.Session.ActivatePane(currentPane)
	}

	// Save pane ID
	cfg.PaneID = paneID
	cfg.Save(configPath)

	// Wait for shell init, then launch AI
	fmt.Println("  Waiting for shell to initialize...")
	time.Sleep(2 * time.Second)

	launchCmd := internal.BuildLaunchCmd(provider, model)
	a.Session.PaneSend(paneID, launchCmd)

	fmt.Printf("✓ Opened role '%s' (%s) [pane %s]\n", name, provider, paneID)
	return nil
}

func (a *App) RunOpenAll(provider, model string) error {
	root := a.Git.Root()
	roles := internal.ListRoles(root, a.WtBase)
	if len(roles) == 0 {
		return fmt.Errorf("no roles found. Create one with: solo-ops create <name>")
	}
	for _, role := range roles {
		if err := a.RunOpen(role, provider, model); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to open '%s': %v\n", role, err)
		}
	}
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/ -run TestRunOpen -v`
Expected: PASS (3 tests)

**Step 5: Register and commit**

Add to `RegisterCommands`:
```go
rootCmd.AddCommand(newOpenCmd())
rootCmd.AddCommand(newOpenAllCmd())
```

```bash
git add cmd/open.go cmd/open_test.go cmd/root.go
git commit -m "feat: add open and open-all commands"
```

---

### Task 10: cmd/assign

**Files:**
- Create: `cmd/assign.go`
- Test: `cmd/assign_test.go`

**Step 1: Write the failing test**

```go
// cmd/assign_test.go
package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leeforge/agent-team/internal"
)

func TestRunAssign(t *testing.T) {
	app, dir := initTestApp(t)
	mock := &MockBackend{
		SpawnedID:  "99",
		AlivePanes: map[string]bool{"99": true},
	}
	app.Session = mock
	app.RunCreate("worker")

	// Set pane as running
	configPath := filepath.Join(dir, ".worktrees", "worker", "agents", "teams", "worker", "config.yaml")
	cfg, _ := internal.LoadRoleConfig(configPath)
	cfg.PaneID = "99"
	cfg.Save(configPath)

	err := app.RunAssign("worker", "Fix the login bug", "", "")
	if err != nil {
		t.Fatalf("RunAssign: %v", err)
	}

	// Task file should exist in pending
	pendingDir := filepath.Join(dir, ".worktrees", "worker", "agents", "teams", "worker", "tasks", "pending")
	entries, _ := os.ReadDir(pendingDir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 task file, got %d", len(entries))
	}

	data, _ := os.ReadFile(filepath.Join(pendingDir, entries[0].Name()))
	if !strings.Contains(string(data), "Fix the login bug") {
		t.Error("task file missing task description")
	}

	// Notification should be sent to pane
	if len(mock.SentTexts) == 0 {
		t.Error("no notification sent to pane")
	}
}

func TestRunAssignNotFound(t *testing.T) {
	app, _ := initTestApp(t)
	err := app.RunAssign("ghost", "task", "", "")
	if err == nil {
		t.Error("expected error for nonexistent role")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/ -run TestRunAssign -v`
Expected: FAIL

**Step 3: Write implementation**

```go
// cmd/assign.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/leeforge/agent-team/internal"
	"github.com/spf13/cobra"
)

func newAssignCmd() *cobra.Command {
	var model string
	cmd := &cobra.Command{
		Use:   `assign <name> "<task>" [provider]`,
		Short: "Write a task file and notify the role session",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := ""
			if len(args) > 2 {
				provider = args[2]
			}
			return GetApp(cmd).RunAssign(args[0], args[1], provider, model)
		},
	}
	cmd.Flags().StringVarP(&model, "model", "m", "", "AI model identifier")
	return cmd
}

func (a *App) RunAssign(name, task, provider, model string) error {
	root := a.Git.Root()
	teamsDir := internal.TeamsDir(root, a.WtBase, name)
	configPath := internal.ConfigPath(root, a.WtBase, name)

	if _, err := os.Stat(teamsDir); os.IsNotExist(err) {
		return fmt.Errorf("role '%s' not found", name)
	}

	// Create task file
	ts := time.Now().Format("2006-01-02-15-04-05")
	slug := internal.Slugify(task, 50)
	fileName := fmt.Sprintf("%s-%s.md", ts, slug)
	taskPath := filepath.Join(teamsDir, "tasks", "pending", fileName)

	nowUTC := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	content := fmt.Sprintf("# Task: %s\n\nAssigned: %s\nStatus: pending\n\n## Description\n\n%s\n\n## Notes\n\n_Add implementation notes here_\n",
		task, nowUTC, task)

	if err := os.WriteFile(taskPath, []byte(content), 0644); err != nil {
		return err
	}
	fmt.Printf("✓ Task file: %s\n", taskPath)

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
		// Reload config to get new pane ID
		cfg, err = internal.LoadRoleConfig(configPath)
		if err != nil {
			return err
		}
		fmt.Println("  Waiting for AI to initialize...")
		time.Sleep(3 * time.Second)
	}

	// Notify
	taskRel := fmt.Sprintf("agents/teams/%s/tasks/pending/%s", name, fileName)
	msg := fmt.Sprintf("New task assigned: %s\nPlease read the task file at: %s\nWhen complete, move it to agents/teams/%s/tasks/done/",
		task, taskRel, name)
	a.Session.PaneSend(cfg.PaneID, msg)

	fmt.Printf("✓ Assigned to '%s': %s\n", name, task)
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/ -run TestRunAssign -v`
Expected: PASS (2 tests)

**Step 5: Register and commit**

Add `rootCmd.AddCommand(newAssignCmd())` to `RegisterCommands`.

```bash
git add cmd/assign.go cmd/assign_test.go cmd/root.go
git commit -m "feat: add assign command"
```

---

### Task 11: cmd/reply, cmd/status, cmd/merge

**Files:**
- Create: `cmd/reply.go`, `cmd/status.go`, `cmd/merge.go`
- Test: `cmd/commands_test.go`

**Step 1: Write failing tests**

```go
// cmd/commands_test.go
package cmd

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/leeforge/agent-team/internal"
)

// --- reply ---

func TestRunReply(t *testing.T) {
	app, dir := initTestApp(t)
	mock := &MockBackend{
		SpawnedID:  "50",
		AlivePanes: map[string]bool{"50": true},
	}
	app.Session = mock
	app.RunCreate("dev")

	configPath := filepath.Join(dir, ".worktrees", "dev", "agents", "teams", "dev", "config.yaml")
	cfg, _ := internal.LoadRoleConfig(configPath)
	cfg.PaneID = "50"
	cfg.Save(configPath)

	err := app.RunReply("dev", "Use the factory pattern")
	if err != nil {
		t.Fatalf("RunReply: %v", err)
	}

	if len(mock.SentTexts) == 0 {
		t.Fatal("no text sent")
	}
	if !strings.Contains(mock.SentTexts[0], "[Main Controller Reply]") {
		t.Errorf("sent = %q, want [Main Controller Reply] prefix", mock.SentTexts[0])
	}
}

func TestRunReplyOffline(t *testing.T) {
	app, _ := initTestApp(t)
	app.RunCreate("offline")
	err := app.RunReply("offline", "hello")
	if err == nil {
		t.Error("expected error when role is offline")
	}
}

// --- status ---

func TestRunStatus(t *testing.T) {
	app, _ := initTestApp(t)
	app.RunCreate("alpha")
	app.RunCreate("beta")

	// Should not error
	err := app.RunStatus()
	if err != nil {
		t.Fatalf("RunStatus: %v", err)
	}
}

// --- merge ---

func TestRunMerge(t *testing.T) {
	app, dir := initTestApp(t)
	app.RunCreate("feature")

	// Make a commit in the worktree so merge has something
	wtPath := filepath.Join(dir, ".worktrees", "feature")
	commitFile := filepath.Join(wtPath, "test.txt")
	os.WriteFile(commitFile, []byte("test"), 0644)
	exec.Command("git", "-C", wtPath, "add", ".").Run()
	exec.Command("git", "-C", wtPath, "commit", "-m", "test commit").Run()

	err := app.RunMerge("feature")
	if err != nil {
		t.Fatalf("RunMerge: %v", err)
	}
}

func TestRunMergeNotFound(t *testing.T) {
	app, _ := initTestApp(t)
	err := app.RunMerge("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent role")
	}
}
```

Note: add `"os"` and `"os/exec"` to imports.

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/ -run "TestRunReply|TestRunStatus|TestRunMerge" -v`
Expected: FAIL

**Step 3: Write implementations**

```go
// cmd/reply.go
package cmd

import (
	"fmt"

	"github.com/leeforge/agent-team/internal"
	"github.com/spf13/cobra"
)

func newReplyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   `reply <name> "<answer>"`,
		Short: "Send a reply to a role's running session",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunReply(args[0], args[1])
		},
	}
}

func (a *App) RunReply(name, answer string) error {
	root := a.Git.Root()
	configPath := internal.ConfigPath(root, a.WtBase, name)

	cfg, err := internal.LoadRoleConfig(configPath)
	if err != nil {
		return fmt.Errorf("role '%s' not found", name)
	}

	if !a.Session.PaneAlive(cfg.PaneID) {
		return fmt.Errorf("role '%s' is not running", name)
	}

	a.Session.PaneSend(cfg.PaneID, "[Main Controller Reply] "+answer)
	fmt.Printf("✓ Replied to '%s'\n", name)
	return nil
}
```

```go
// cmd/status.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/leeforge/agent-team/internal"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show all roles, running state, and pending task count",
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
		fmt.Println("No roles found. Create one with: solo-ops create <name>")
		return nil
	}

	fmt.Printf("%-16s %-24s %s\n", "Role", "Status", "Pending Tasks")
	fmt.Printf("%-16s %-24s %s\n", "────────────────", "────────────────────────", "─────────────")

	for _, role := range roles {
		configPath := internal.ConfigPath(root, a.WtBase, role)
		pendingDir := filepath.Join(internal.TeamsDir(root, a.WtBase, role), "tasks", "pending")

		cfg, _ := internal.LoadRoleConfig(configPath)
		status := "✗ offline"
		if cfg != nil && a.Session.PaneAlive(cfg.PaneID) {
			status = fmt.Sprintf("✓ running [p:%s]", cfg.PaneID)
		}

		count := 0
		if entries, err := os.ReadDir(pendingDir); err == nil {
			for _, e := range entries {
				if filepath.Ext(e.Name()) == ".md" {
					count++
				}
			}
		}

		fmt.Printf("%-16s %-24s %d\n", role, status, count)
	}
	return nil
}
```

```go
// cmd/merge.go
package cmd

import (
	"fmt"
	"os"

	"github.com/leeforge/agent-team/internal"
	"github.com/spf13/cobra"
)

func newMergeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "merge <name>",
		Short: "Merge a role's branch into the current branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunMerge(args[0])
		},
	}
}

func (a *App) RunMerge(name string) error {
	root := a.Git.Root()
	wtPath := internal.WtPath(root, a.WtBase, name)
	branch := "team/" + name

	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		return fmt.Errorf("role '%s' not found", name)
	}

	mainBranch, _ := a.Git.CurrentBranch()

	fmt.Printf("Merging branch '%s' into '%s'...\n", branch, mainBranch)
	msg := fmt.Sprintf("merge: integrate work from team role '%s'", name)
	if err := a.Git.Merge(branch, msg); err != nil {
		return err
	}

	fmt.Printf("✓ Merged '%s' into %s\n", name, mainBranch)
	fmt.Printf("  → Run 'solo-ops delete %s' to remove the worktree when done\n", name)
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/ -run "TestRunReply|TestRunStatus|TestRunMerge" -v`
Expected: PASS (5 tests)

**Step 5: Register and commit**

```go
func RegisterCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(newCreateCmd())
	rootCmd.AddCommand(newDeleteCmd())
	rootCmd.AddCommand(newOpenCmd())
	rootCmd.AddCommand(newOpenAllCmd())
	rootCmd.AddCommand(newAssignCmd())
	rootCmd.AddCommand(newReplyCmd())
	rootCmd.AddCommand(newStatusCmd())
	rootCmd.AddCommand(newMergeCmd())
}
```

```bash
git add cmd/reply.go cmd/status.go cmd/merge.go cmd/commands_test.go cmd/root.go
git commit -m "feat: add reply, status, and merge commands"
```

---

### Task 12: Build Verification + Full Test Suite

**Step 1: Run full test suite**

```bash
go test ./... -v
```

Expected: All tests pass

**Step 2: Build binary**

```bash
make build
```

**Step 3: Smoke test**

```bash
./solo-ops --version
./solo-ops --help
./solo-ops create --help
./solo-ops status --help
```

**Step 4: Commit**

```bash
git add -A
git commit -m "chore: verify build and full test suite"
```

---

### Task 13: Distribution — goreleaser + SKILL.md Update

**Files:**
- Create: `.goreleaser.yaml`
- Modify: `skills/solo-ops/SKILL.md`

**Step 1: Create goreleaser config**

```yaml
# .goreleaser.yaml
version: 2
project_name: solo-ops

builds:
  - binary: solo-ops
    ldflags:
      - -s -w
      - -X github.com/leeforge/agent-team/cmd.Version={{.Version}}
    goos:
      - darwin
      - linux
    goarch:
      - amd64
      - arm64

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"

brews:
  - repository:
      owner: leeforge
      name: homebrew-tap
    homepage: https://github.com/leeforge/agent-team
    description: AI team role manager — git worktrees + terminal multiplexer
    install: |
      bin.install "solo-ops"

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc
```

**Step 2: Update SKILL.md**

Replace `python3 <base-dir>/scripts/solo_ops.py <command>` references with `solo-ops <command>`.

Key changes:
- Usage section: `solo-ops <command>` (no python prefix)
- Remove tmux-specific `solo_ops_tmux.py` references — backend is now selected via `SOLO_OPS_BACKEND=tmux`
- Install: `brew tap leeforge/tap && brew install solo-ops`

**Step 3: Commit**

```bash
git add .goreleaser.yaml skills/solo-ops/SKILL.md
git commit -m "chore: add goreleaser config and update SKILL.md for Go binary"
```

---

## Execution Notes

- **Test command**: `go test ./... -v` after each task
- **Build command**: `make build` to produce `./solo-ops`
- **Python reference**: Keep `skills/solo-ops/scripts/` until Go version is validated in production; clean up in a follow-up PR
- **MockBackend** in `cmd/create_test.go` is shared across all cmd tests via the same package
