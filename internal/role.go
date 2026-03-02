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
	"sync"

	"gopkg.in/yaml.v3"
)

var deprecatedOnce sync.Once

// ResolveAgentsDir 优先返回 .agents/，回退到 agents/（加 deprecation 警告）
func ResolveAgentsDir(root string) string {
	newPath := filepath.Join(root, ".agents")
	if _, err := os.Stat(newPath); err == nil {
		return newPath
	}
	oldPath := filepath.Join(root, "agents")
	if _, err := os.Stat(oldPath); err == nil {
		deprecatedOnce.Do(func() {
			fmt.Fprintln(os.Stderr,
				"[DEPRECATED] Using agents/ directory. Run 'agent-team migrate' to update to .agents/")
		})
		return oldPath
	}
	return newPath // 两者都不存在时返回新路径（将由命令创建）
}

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

// RoleDir returns the path to a role definition: .agents/teams/<role-name>/
func RoleDir(root, roleName string) string {
	return filepath.Join(ResolveAgentsDir(root), "teams", roleName)
}

// RoleYAMLPath returns the path to a role's role.yaml.
func RoleYAMLPath(root, roleName string) string {
	return filepath.Join(RoleDir(root, roleName), "references", "role.yaml")
}

// RoleSystemMDPath returns the path to a role's system.md.
func RoleSystemMDPath(root, roleName string) string {
	return filepath.Join(RoleDir(root, roleName), "system.md")
}

// WorkerInfo holds summary info for a worker.
type WorkerInfo struct {
	WorkerID string
	Role     string
	Config   *WorkerConfig
}

// RoleMatch describes a resolved role location.
type RoleMatch struct {
	RoleName    string // 角色目录名
	Path        string // 绝对路径
	Scope       string // "project" | "global"
	Description string // from role.yaml
	MatchType   string // "exact" | "keyword"
}

// GlobalRolesDir returns the global roles directory (~/.agents/roles/).
func GlobalRolesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".agents", "roles"), nil
}

// fileExists returns true if the path exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// splitRoleKeywords splits a kebab-case name into keywords.
func splitRoleKeywords(name string) []string {
	parts := strings.Split(name, "-")
	var keywords []string
	for _, p := range parts {
		if p != "" {
			keywords = append(keywords, strings.ToLower(p))
		}
	}
	return keywords
}

// roleYAMLFull is a struct for reading name/description/scope from role.yaml.
type roleYAMLFull struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Scope       struct {
		InScope []string `yaml:"in_scope"`
	} `yaml:"scope"`
}

// readRoleYAMLFull reads role.yaml from the given role path.
// Returns zero value if file doesn't exist or can't be parsed.
func readRoleYAMLFull(rolePath string) roleYAMLFull {
	yamlPath := filepath.Join(rolePath, "references", "role.yaml")
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return roleYAMLFull{}
	}
	var ry roleYAMLFull
	if err := yaml.Unmarshal(data, &ry); err != nil {
		return roleYAMLFull{}
	}
	return ry
}

// isRoleDir checks if a directory contains SKILL.md or system.md (i.e., is a valid role).
func isRoleDir(path string) bool {
	return fileExists(filepath.Join(path, "SKILL.md")) || fileExists(filepath.Join(path, "system.md"))
}

// listRolesInDir scans a directory for role subdirectories and returns RoleMatch entries.
func listRolesInDir(dir, scope string) []RoleMatch {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var matches []RoleMatch
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		rolePath := filepath.Join(dir, e.Name())
		if !isRoleDir(rolePath) {
			continue
		}
		ry := readRoleYAMLFull(rolePath)
		matches = append(matches, RoleMatch{
			RoleName:    e.Name(),
			Path:        rolePath,
			Scope:       scope,
			Description: ry.Description,
			MatchType:   "exact",
		})
	}
	return matches
}

// matchesKeywords checks if a role directory or its YAML metadata matches the given keywords.
func matchesKeywords(dirName string, ry roleYAMLFull, keywords []string) bool {
	searchText := strings.ToLower(dirName + " " + ry.Description)
	for _, item := range ry.Scope.InScope {
		searchText += " " + strings.ToLower(item)
	}
	for _, kw := range keywords {
		if strings.Contains(searchText, kw) {
			return true
		}
	}
	return false
}

// ResolveRole looks up a role by name with project-first priority.
// Search order: .agents/teams/<role> → ~/.agents/roles/<role>
func ResolveRole(root, roleName string) (*RoleMatch, error) {
	// 1. Project-level
	projectDir := RoleDir(root, roleName)
	if isRoleDir(projectDir) {
		ry := readRoleYAMLFull(projectDir)
		return &RoleMatch{
			RoleName:    roleName,
			Path:        projectDir,
			Scope:       "project",
			Description: ry.Description,
			MatchType:   "exact",
		}, nil
	}

	// 2. Global
	globalDir, err := GlobalRolesDir()
	if err != nil {
		return nil, fmt.Errorf("role '%s' not found in project, and cannot resolve global dir: %w", roleName, err)
	}
	globalRolePath := filepath.Join(globalDir, roleName)
	if isRoleDir(globalRolePath) {
		ry := readRoleYAMLFull(globalRolePath)
		return &RoleMatch{
			RoleName:    roleName,
			Path:        globalRolePath,
			Scope:       "global",
			Description: ry.Description,
			MatchType:   "exact",
		}, nil
	}

	return nil, fmt.Errorf("role '%s' not found in .agents/teams/ or ~/.agents/roles/.\nCreate it first using the role-creator skill", roleName)
}

// SearchGlobalRoles searches global roles by exact name match and keyword matching.
// Exact matches are sorted before keyword matches.
func SearchGlobalRoles(roleName string) ([]RoleMatch, error) {
	globalDir, err := GlobalRolesDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(globalDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	keywords := splitRoleKeywords(roleName)
	var exact, keyword []RoleMatch

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		rolePath := filepath.Join(globalDir, e.Name())
		if !isRoleDir(rolePath) {
			continue
		}
		ry := readRoleYAMLFull(rolePath)

		if e.Name() == roleName {
			exact = append(exact, RoleMatch{
				RoleName:    e.Name(),
				Path:        rolePath,
				Scope:       "global",
				Description: ry.Description,
				MatchType:   "exact",
			})
			continue
		}

		if matchesKeywords(e.Name(), ry, keywords) {
			keyword = append(keyword, RoleMatch{
				RoleName:    e.Name(),
				Path:        rolePath,
				Scope:       "global",
				Description: ry.Description,
				MatchType:   "keyword",
			})
		}
	}

	sort.Slice(keyword, func(i, j int) bool {
		return keyword[i].RoleName < keyword[j].RoleName
	})

	return append(exact, keyword...), nil
}

// ListGlobalRoles lists all roles in ~/.agents/roles/.
func ListGlobalRoles() ([]RoleMatch, error) {
	globalDir, err := GlobalRolesDir()
	if err != nil {
		return nil, err
	}
	roles := listRolesInDir(globalDir, "global")
	sort.Slice(roles, func(i, j int) bool {
		return roles[i].RoleName < roles[j].RoleName
	})
	return roles, nil
}

// ListAvailableRoles scans .agents/teams/ for directories containing SKILL.md.
func ListAvailableRoles(root string) []string {
	teamsDir := filepath.Join(ResolveAgentsDir(root), "teams")
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

// ListWorkers scans worktrees for directories containing worker.yaml.
func ListWorkers(root, wtBase string) []WorkerInfo {
	wtDir := filepath.Join(root, wtBase)
	entries, err := os.ReadDir(wtDir)
	if err != nil {
		return nil
	}
	var workers []WorkerInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		configPath := WorkerYAMLPath(filepath.Join(wtDir, e.Name()))
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
func NextWorkerID(root, wtBase, roleName string) string {
	wtDir := filepath.Join(root, wtBase)
	entries, err := os.ReadDir(wtDir)
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
	content := ".gitignore\n.claude/\n.codex/\n.tasks/\nworker.yaml\n"
	return os.WriteFile(filepath.Join(wtPath, ".gitignore"), []byte(content), 0644)
}

// --- Shared utilities ---

// WtPath returns the path to a worker's worktree directory.
func WtPath(root, wtBase, name string) string {
	return filepath.Join(root, wtBase, name)
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

// buildRoleSectionFromPath builds the AGENT_TEAM section content from the role at rolePath.
func buildRoleSectionFromPath(wtPath, workerID, roleName, rolePath, root string) (string, error) {
	roleSystemPath := filepath.Join(rolePath, "system.md")
	prompt, err := os.ReadFile(roleSystemPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	var b strings.Builder
	b.Write(prompt)

	// Inject skill-first workflow and available skills list
	b.WriteString("\n## Skill-First Workflow\n\n")
	fmt.Fprintf(&b, "**Role skill (MUST use):** `%s`\n\n", roleName)
	fmt.Fprintf(&b, "When you receive ANY task, you MUST first invoke the `%s` skill (via `/%s` or the Skill tool). ", roleName, roleName)
	b.WriteString("This is your primary skill that defines how you approach work. Never skip it.\n\n")

	skills, skillErr := ReadRoleSkillsFromPath(rolePath)
	if skillErr == nil && len(skills) > 0 {
		b.WriteString("**Dependency skills:**\n\n")
		for _, skill := range skills {
			shortName := parseSkillName(skill)
			fmt.Fprintf(&b, "- `%s`\n", shortName)
		}
		b.WriteString("\n")
	}

	b.WriteString("**Workflow:**\n\n")
	b.WriteString("1. **Match skills first** — Check which of your available skills are relevant to the task before doing any direct work.\n")
	b.WriteString("2. **Invoke matched skills** — For each relevant skill, invoke it (via `/skill-name` or the Skill tool). Let the skill guide execution.\n")
	b.WriteString("3. **Combine skill outputs** — If a task spans multiple skills, invoke them in logical order and integrate their outputs.\n")
	b.WriteString("4. **Direct work only as fallback** — Only work directly when no available skill covers the requirement.\n")
	b.WriteString("5. **Dynamic skill discovery** — If no current skill matches, invoke `find-skills` to search for one. If found, use it and suggest adding it to the role for future sessions.\n")

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
	b.WriteString("Use the `agent-team` skill's **Reply to main controller (used by workers)** protocol for worker-to-main communication.\n")
	b.WriteString("For EVERY completed task, you MUST send a completion message to main controller.\n")
	b.WriteString("When any task is done:\n")
	b.WriteString("1. Run `agent-team task archive <worker-id> <change-name>` to archive the completed change\n")
	b.WriteString("2. After the archive attempt (success or failure), ALWAYS notify main controller:\n")
	b.WriteString("   ```bash\n")
	b.WriteString("   agent-team reply-main \"Task completed: <summary>; change archived: <change-name>\"\n")
	b.WriteString("   ```\n")
	b.WriteString("3. If archive fails, you may still report completion, but MUST include the failure details:\n")
	b.WriteString("   ```bash\n")
	b.WriteString("   agent-team reply-main \"Task completed: <summary>; archive failed for <change-name>: <error>\"\n")
	b.WriteString("   ```\n")
	b.WriteString("4. If you have blockers, questions, or implementation options, report them to main controller:\n")
	b.WriteString("   ```bash\n")
	b.WriteString("   agent-team reply-main \"Need decision: <problem or options>\"\n")
	b.WriteString("   ```\n")
	b.WriteString("5. Do not start the next task until the completion summary has been sent.\n")

	return b.String(), nil
}

// buildRoleSection builds the AGENT_TEAM section content from the role's system.md.
func buildRoleSection(wtPath, workerID, roleName, root string) (string, error) {
	return buildRoleSectionFromPath(wtPath, workerID, roleName, RoleDir(root, roleName), root)
}

// InjectRolePromptWithPath injects the role prompt using an explicit rolePath.
func InjectRolePromptWithPath(wtPath, workerID, roleName, rolePath, root string) error {
	content, err := buildRoleSectionFromPath(wtPath, workerID, roleName, rolePath, root)
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

// InjectRolePrompt injects the role prompt into CLAUDE.md and AGENTS.md using tagged sections.
func InjectRolePrompt(wtPath, workerID, roleName, root string) error {
	return InjectRolePromptWithPath(wtPath, workerID, roleName, RoleDir(root, roleName), root)
}
