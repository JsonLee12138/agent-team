// internal/role.go
package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
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

// --- v2 path functions ---

// RoleDir returns the path to a role definition: agents/teams/<role-name>/
func RoleDir(root, roleName string) string {
	return filepath.Join(root, "agents", "teams", roleName)
}

// RoleYAMLPath returns the path to a role's role.yaml.
func RoleYAMLPath(root, roleName string) string {
	return filepath.Join(RoleDir(root, roleName), "references", "role.yaml")
}

// RoleSystemMDPath returns the path to a role's system.md.
func RoleSystemMDPath(root, roleName string) string {
	return filepath.Join(RoleDir(root, roleName), "system.md")
}

// WorkerDir returns the path to a worker config directory: agents/workers/<worker-id>/
func WorkerDir(root, workerID string) string {
	return filepath.Join(root, "agents", "workers", workerID)
}

// WorkerConfigPath returns the path to a worker's config.yaml.
func WorkerConfigPath(root, workerID string) string {
	return filepath.Join(WorkerDir(root, workerID), "config.yaml")
}

// WorkerInfo holds summary info for a worker.
type WorkerInfo struct {
	WorkerID string
	Role     string
	Config   *WorkerConfig
}

// ListAvailableRoles scans agents/teams/ for directories containing SKILL.md.
func ListAvailableRoles(root string) []string {
	teamsDir := filepath.Join(root, "agents", "teams")
	entries, err := os.ReadDir(teamsDir)
	if err != nil {
		return nil
	}
	var roles []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		skillPath := filepath.Join(teamsDir, e.Name(), "SKILL.md")
		if _, err := os.Stat(skillPath); err == nil {
			roles = append(roles, e.Name())
		}
	}
	sort.Strings(roles)
	return roles
}

// ListWorkers scans agents/workers/ for directories containing config.yaml.
func ListWorkers(root string) []WorkerInfo {
	workersDir := filepath.Join(root, "agents", "workers")
	entries, err := os.ReadDir(workersDir)
	if err != nil {
		return nil
	}
	var workers []WorkerInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		configPath := filepath.Join(workersDir, e.Name(), "config.yaml")
		cfg, err := LoadWorkerConfig(configPath)
		if err != nil {
			continue
		}
		workers = append(workers, WorkerInfo{
			WorkerID: e.Name(),
			Role:     cfg.Role,
			Config:   cfg,
		})
	}
	return workers
}

// workerIDPattern matches <role-name>-<3-digit-number>
var workerIDPattern = regexp.MustCompile(`^(.+)-(\d{3})$`)

// NextWorkerID computes the next worker ID for a given role (e.g., frontend-dev-001).
func NextWorkerID(root, roleName string) string {
	workersDir := filepath.Join(root, "agents", "workers")
	entries, err := os.ReadDir(workersDir)
	if err != nil {
		return fmt.Sprintf("%s-001", roleName)
	}

	maxNum := 0
	prefix := roleName + "-"
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), prefix) {
			continue
		}
		m := workerIDPattern.FindStringSubmatch(e.Name())
		if m == nil || m[1] != roleName {
			continue
		}
		num, err := strconv.Atoi(m[2])
		if err != nil {
			continue
		}
		if num > maxNum {
			maxNum = num
		}
	}
	return fmt.Sprintf("%s-%03d", roleName, maxNum+1)
}

// WriteWorktreeGitignore writes a .gitignore to exclude worker-local files.
func WriteWorktreeGitignore(wtPath string) error {
	content := ".gitignore\n.claude/\n.codex/\nopenspec/\n"
	return os.WriteFile(filepath.Join(wtPath, ".gitignore"), []byte(content), 0644)
}

// --- Legacy v1 path functions (kept for backward compat during migration) ---

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

// --- Shared utilities ---

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

// InjectSection injects content into a file within <!-- {tag}:START --> ... <!-- {tag}:END --> markers.
func InjectSection(filePath, tag, content string) error {
	startMarker := fmt.Sprintf("<!-- %s:START -->", tag)
	endMarker := fmt.Sprintf("<!-- %s:END -->", tag)
	section := startMarker + "\n" + content + "\n" + endMarker

	existing, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return os.WriteFile(filePath, []byte(section+"\n"), 0644)
		}
		return err
	}

	fileContent := string(existing)

	re := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(startMarker) + `.*?` + regexp.QuoteMeta(endMarker))
	if re.MatchString(fileContent) {
		fileContent = re.ReplaceAllString(fileContent, section)
		return os.WriteFile(filePath, []byte(fileContent), 0644)
	}

	fileContent = section + "\n\n" + fileContent
	return os.WriteFile(filePath, []byte(fileContent), 0644)
}

// buildRoleSection builds the AGENT_TEAM section content from system.md (v2) or prompt.md (v1 fallback).
func buildRoleSection(wtPath, workerID, roleName, root string) (string, error) {
	// v2: try reading system.md from role definition in main repo
	roleSystemPath := RoleSystemMDPath(root, roleName)
	prompt, err := os.ReadFile(roleSystemPath)
	if err != nil {
		// v1 fallback: try prompt.md in worktree
		teamsDir := filepath.Join(wtPath, "agents", "teams", workerID)
		promptPath := filepath.Join(teamsDir, "prompt.md")
		prompt, err = os.ReadFile(promptPath)
		if err != nil {
			if os.IsNotExist(err) {
				return "", nil
			}
			return "", err
		}
	}

	var b strings.Builder
	b.Write(prompt)
	b.WriteString("\n## Development Environment\n\n")
	b.WriteString("You are working in an **isolated git worktree**. All development MUST happen here:\n\n")
	fmt.Fprintf(&b, "- **Working directory**: `%s`\n", wtPath)
	fmt.Fprintf(&b, "- **Git branch**: `team/%s` (your dedicated branch)\n", workerID)
	fmt.Fprintf(&b, "- **Main project root**: `%s`\n\n", root)
	b.WriteString("### Git Rules\n\n")
	fmt.Fprintf(&b, "- All changes and commits go to the `team/%s` branch — this is already checked out\n", workerID)
	b.WriteString("- **Never** run `git checkout`, `git switch`, or change branches\n")
	b.WriteString("- **Never** merge or rebase from within this worktree\n")
	b.WriteString("- Commit regularly with clear messages as you complete work\n\n")
	b.WriteString("The main controller will merge your branch back to main when ready.\n\n")
	b.WriteString("### Task Completion Protocol\n\n")
	b.WriteString("When you finish a task:\n")
	b.WriteString("1. Run `/openspec archive` to archive the completed change\n")
	b.WriteString("2. Notify the main controller:\n")
	b.WriteString("   ```bash\n")
	b.WriteString("   agent-team reply-main \"<summary of completed work>\"\n")
	b.WriteString("   ```\n")

	return b.String(), nil
}

// InjectRolePrompt injects the role prompt into CLAUDE.md and AGENTS.md using tagged sections.
func InjectRolePrompt(wtPath, workerID, roleName, root string) error {
	content, err := buildRoleSection(wtPath, workerID, roleName, root)
	if err != nil {
		return err
	}
	if content == "" {
		return nil
	}

	claudePath := filepath.Join(wtPath, "CLAUDE.md")
	if err := InjectSection(claudePath, "AGENT_TEAM", content); err != nil {
		return fmt.Errorf("inject CLAUDE.md: %w", err)
	}

	agentsPath := filepath.Join(wtPath, "AGENTS.md")
	if err := InjectSection(agentsPath, "AGENT_TEAM", content); err != nil {
		return fmt.Errorf("inject AGENTS.md: %w", err)
	}

	return nil
}

// PromptMDContent generates a default prompt.md for legacy v1 roles.
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
	b.WriteString("When you need clarification or have a question for the main controller, run:\n\n")
	b.WriteString("```bash\n")
	b.WriteString("agent-team reply-main \"<your question here>\"\n")
	b.WriteString("```\n\n")
	b.WriteString("Wait for the main controller to reply. Replies will appear as:\n")
	b.WriteString("`[Main Controller Reply]`\n\n")
	b.WriteString("Do NOT proceed on blocked tasks until you receive a reply.\n")
	return b.String()
}
