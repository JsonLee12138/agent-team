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
	"text/template"

	"gopkg.in/yaml.v3"
)

// ResolveAgentsDir returns the project-local agent-team directory.
func ResolveAgentsDir(root string) string {
	return AgentTeamDir(root)
}

var SupportedProviders = map[string]bool{
	"claude":   true,
	"codex":    true,
	"opencode": true,
	"gemini":   true,
}

var launchCommands = map[string]string{
	"claude":   "claude --dangerously-skip-permissions",
	"codex":    "codex --dangerously-bypass-approvals-and-sandbox",
	"opencode": "opencode",
	"gemini":   "gemini --approval-mode yolo",
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

// RoleDir returns the path to a role definition: .agent-team/teams/<role-name>/
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
// Search order: .agent-team/teams/<role> → ~/.agents/roles/<role>
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

	return nil, fmt.Errorf("role '%s' not found in .agent-team/teams/ or ~/.agents/roles/.\nTry searching and installing it first with:\n  agent-team role-repo find %s\nIf no matching role exists in role-repo, then create it using the role-creator skill.", roleName, roleName)
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

// ListAvailableRoles scans .agent-team/teams/ for directories containing SKILL.md.
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

// ListWorkers scans worktree-local worker configs.
func ListWorkers(root, wtBase string) []WorkerInfo {
	var workers []WorkerInfo

	wtDir := filepath.Join(root, wtBase)
	if entries, err := os.ReadDir(wtDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			workerID := e.Name()
			cfg, err := LoadWorkerConfig(WorkerYAMLPath(filepath.Join(wtDir, workerID)))
			if err != nil {
				continue
			}
			workers = append(workers, WorkerInfo{WorkerID: workerID, Role: cfg.Role, Config: cfg})
		}
	}

	sort.Slice(workers, func(i, j int) bool {
		return workers[i].WorkerID < workers[j].WorkerID
	})
	return workers
}

// workerIDPattern matches <role-name>-<3-digit-number>
var workerIDPattern = regexp.MustCompile(`^(.+)-(\d{3})$`)

// NextWorkerID computes the next worker ID for a given role (e.g., frontend-dev-001).
func NextWorkerID(root, wtBase, roleName string) string {
	maxNum := 0
	prefix := roleName + "-"
	consider := func(workerID string) {
		if !strings.HasPrefix(workerID, prefix) {
			return
		}
		m := workerIDPattern.FindStringSubmatch(workerID)
		if m == nil || m[1] != roleName {
			return
		}
		num, err := strconv.Atoi(m[2])
		if err != nil {
			return
		}
		if num > maxNum {
			maxNum = num
		}
	}

	wtDir := filepath.Join(root, wtBase)
	if entries, err := os.ReadDir(wtDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				consider(e.Name())
			}
		}
	}

	return fmt.Sprintf("%s-%03d", roleName, maxNum+1)
}

// WriteWorktreeGitignore writes a .gitignore to exclude worker-local files.
func WriteWorktreeGitignore(wtPath string) error {
	content := ".gitignore\n.claude/\n.codex/\n.gemini/\n.opencode/\n.tasks/\nworker.yaml\nCLAUDE.md\nGEMINI.md\nAGENTS.md\n"
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

// roleSectionData holds the template data for the AGENT_TEAM section.
type roleSectionData struct {
	RoleName string
	WorkerID string
	WtPath   string
	Root     string
	Skills   []string
}

// legacyRoleSectionTmpl is the full template used when .agent-team/rules/ does not exist (backward compatible).
var legacyRoleSectionTmpl = template.Must(template.New("legacyRoleSection").Parse(`
## Skill-First Workflow

**Role skill (MUST use):** ` + "`{{.RoleName}}`" + `

When you receive ANY task, you MUST first invoke the ` + "`{{.RoleName}}`" + ` skill (via ` + "`/{{.RoleName}}`" + ` or the Skill tool). This is your primary skill that defines how you approach work. Never skip it.
{{if .Skills}}
**Dependency skills:**
{{range .Skills}}
- ` + "`{{.}}`" + `
{{- end}}
{{end}}
**Workflow:**

0. **Execute, do not plan** — Your task has already been planned and assigned by the main controller. Execute it directly.
   - **Never** enter plan mode or any equivalent planning/design phase before working.
   - **Never** expand the task scope beyond what is described in the assigned task.
   - If requirements are unclear, use ` + "`agent-team reply-main`" + ` to ask the controller — do not start a planning phase yourself.
1. **Match skills first** — Check which of your available skills are relevant to the task before doing any direct work.
2. **Invoke matched skills** — For each relevant skill, invoke it (via ` + "`/skill-name`" + ` or the Skill tool). Let the skill guide execution.
3. **Combine skill outputs** — If a task spans multiple skills, invoke them in logical order and integrate their outputs.
4. **Direct work only as fallback** — Only work directly when no available skill covers the requirement.
5. **Dynamic skill discovery** — If no current skill matches, invoke ` + "`find-skills`" + ` to search for one. If found, use it and suggest adding it to the role for future sessions.

## Development Environment

You are working in an **isolated git worktree**. All development MUST happen here:

- **Working directory**: ` + "`{{.WtPath}}`" + `
- **Git branch**: ` + "`team/{{.WorkerID}}`" + ` (your dedicated branch)
- **Main project root**: ` + "`{{.Root}}`" + `

### Git Rules

- All changes and commits go to the ` + "`team/{{.WorkerID}}`" + ` branch — this is already checked out
- **Never** run ` + "`git checkout`" + `, ` + "`git switch`" + `, or change branches
- **Never** merge or rebase from within this worktree
- Commit regularly with clear messages as you complete work
- The following paths are excluded by ` + "`.gitignore`" + ` and will NOT be committed: ` + "`.claude/`" + `, ` + "`.codex/`" + `, ` + "`.gemini/`" + `, ` + "`.opencode/`" + `, ` + "`.tasks/`" + `, ` + "`worker.yaml`" + `
- **Never** place output documents, reports, or any deliverables in git-ignored directories. All work products must be in tracked paths so they are included in commits

The main controller will merge your branch back to main when ready.

### Completion Protocol

Use ` + "`agent-team reply-main`" + ` for worker-to-main communication.
When work is done:
1. Commit task-scoped changes.
2. Notify main controller:
   ` + "```bash" + `
   agent-team reply-main "Task completed: <summary>"
   ` + "```" + `
3. If blocked, report immediately:
   ` + "```bash" + `
   agent-team reply-main "Need decision: <problem or options>"
   ` + "```" + `
4. Do not start the next task until completion or blocker message has been sent.
`))

// slimRoleSectionTmpl is the slim template used when .agent-team/rules/ exists.
// Completion details are delegated to external rule files.
var slimRoleSectionTmpl = template.Must(template.New("slimRoleSection").Parse(`
## Skill-First Workflow

**Role skill (MUST use):** ` + "`{{.RoleName}}`" + `

When you receive ANY task, you MUST first invoke the ` + "`{{.RoleName}}`" + ` skill (via ` + "`/{{.RoleName}}`" + ` or the Skill tool). This is your primary skill that defines how you approach work. Never skip it.
{{if .Skills}}
**Dependency skills:**
{{range .Skills}}
- ` + "`{{.}}`" + `
{{- end}}
{{end}}
**Workflow:**

0. **Execute, do not plan** — Your task has already been planned and assigned by the main controller. Execute it directly.
   - **Never** enter plan mode or any equivalent planning/design phase before working.
   - **Never** expand the task scope beyond what is described in the assigned task.
   - If requirements are unclear, use ` + "`agent-team reply-main`" + ` to ask the controller — do not start a planning phase yourself.
1. **Match skills first** — Check which of your available skills are relevant to the task before doing any direct work.
2. **Invoke matched skills** — For each relevant skill, invoke it (via ` + "`/skill-name`" + ` or the Skill tool). Let the skill guide execution.
3. **Combine skill outputs** — If a task spans multiple skills, invoke them in logical order and integrate their outputs.
4. **Direct work only as fallback** — Only work directly when no available skill covers the requirement.
5. **Dynamic skill discovery** — If no current skill matches, invoke ` + "`find-skills`" + ` to search for one. If found, use it and suggest adding it to the role for future sessions.

## Development Environment

You are working in an **isolated git worktree**. All development MUST happen here:

- **Working directory**: ` + "`{{.WtPath}}`" + `
- **Git branch**: ` + "`team/{{.WorkerID}}`" + ` (your dedicated branch)
- **Main project root**: ` + "`{{.Root}}`" + `

The main controller will merge your branch back to main when ready.
`))

// HasRulesDir checks if .agent-team/rules/ directory exists at the project root.
func HasRulesDir(root string) bool {
	rulesDir := filepath.Join(ResolveAgentsDir(root), "rules")
	info, err := os.Stat(rulesDir)
	return err == nil && info.IsDir()
}

// buildRoleIdentity returns a minimal role identity string (role name + description).
func buildRoleIdentity(roleName, rolePath string) string {
	ry := readRoleYAMLFull(rolePath)
	var b strings.Builder
	b.WriteString("# System Prompt: " + roleName + "\n\nYou are the " + roleName + " role.\n")
	if ry.Description != "" {
		b.WriteString("\nPrimary objective:\n" + ry.Description + "\n")
	}
	return b.String()
}

// buildRulesIndexSection reads .agent-team/rules/index.md and returns its content.
// Returns empty string if the file does not exist.
func buildRulesIndexSection(root string) string {
	indexPath := filepath.Join(ResolveAgentsDir(root), "rules", "index.md")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return ""
	}
	return "\n## Rules Reference\n\nLoad `.agent-team/rules/index.md` at task start, then load only the matching rule files:\n\n" + string(data)
}

// skillIndexEntry holds a skill's name and trigger description for slim injection.
type skillIndexEntry struct {
	Name    string
	Trigger string
}

// extractSkillTrigger reads a SKILL.md file and extracts the description from YAML frontmatter.
func extractSkillTrigger(skillPath string) string {
	skillMD := filepath.Join(skillPath, "SKILL.md")
	data, err := os.ReadFile(skillMD)
	if err != nil {
		return ""
	}
	content := string(data)
	// Parse YAML frontmatter between --- markers
	if !strings.HasPrefix(content, "---") {
		return ""
	}
	end := strings.Index(content[3:], "---")
	if end < 0 {
		return ""
	}
	frontmatter := content[3 : 3+end]
	var fm struct {
		Description string `yaml:"description"`
	}
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return ""
	}
	return strings.TrimSpace(fm.Description)
}

// buildSkillIndexSection builds a concise skill index listing skill names and triggers.
// Only used in slim mode. Returns empty string if no skills found.
func buildSkillIndexSection(root, roleName, rolePath string) string {
	var entries []skillIndexEntry

	// Role skill itself
	trigger := extractSkillTrigger(rolePath)
	if trigger != "" {
		entries = append(entries, skillIndexEntry{Name: roleName, Trigger: trigger})
	}

	// Dependency skills
	skills, err := ReadRoleSkillsFromPath(rolePath)
	if err != nil {
		return ""
	}
	for _, skillName := range skills {
		shortName := parseSkillName(skillName)
		sp := findSkillPath(root, skillName)
		if sp == "" {
			entries = append(entries, skillIndexEntry{Name: shortName, Trigger: ""})
			continue
		}
		t := extractSkillTrigger(sp)
		entries = append(entries, skillIndexEntry{Name: shortName, Trigger: t})
	}

	if len(entries) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n## Skill Index\n\n")
	for _, e := range entries {
		if e.Trigger != "" {
			b.WriteString("- **" + e.Name + "**: " + e.Trigger + "\n")
		} else {
			b.WriteString("- **" + e.Name + "**\n")
		}
	}
	return b.String()
}

// buildRoleSectionFromPath builds the AGENT_TEAM section content from the role at rolePath.
// When .agent-team/rules/ exists, uses slim mode (minimal identity + rules index + skill index).
// Otherwise falls back to legacy mode (full system.md + inline completion protocol).
func buildRoleSectionFromPath(wtPath, workerID, roleName, rolePath, root string) (string, error) {
	roleSystemPath := filepath.Join(rolePath, "system.md")
	if _, err := os.Stat(roleSystemPath); err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	var depSkills []string
	if s, sErr := ReadRoleSkillsFromPath(rolePath); sErr == nil {
		for _, skill := range s {
			depSkills = append(depSkills, parseSkillName(skill))
		}
	}

	data := roleSectionData{
		RoleName: roleName,
		WorkerID: workerID,
		WtPath:   wtPath,
		Root:     root,
		Skills:   depSkills,
	}

	var b strings.Builder

	if HasRulesDir(root) {
		// Slim mode: minimal identity + slim template + rules index + skill index
		b.WriteString(buildRoleIdentity(roleName, rolePath))
		if err := slimRoleSectionTmpl.Execute(&b, data); err != nil {
			return "", fmt.Errorf("execute slim role section template: %w", err)
		}
		b.WriteString(buildRulesIndexSection(root))
		b.WriteString(buildSkillIndexSection(root, roleName, rolePath))
	} else {
		// Legacy mode: full system.md + full inline template
		prompt, err := os.ReadFile(roleSystemPath)
		if err != nil {
			return "", err
		}
		b.Write(prompt)
		if err := legacyRoleSectionTmpl.Execute(&b, data); err != nil {
			return "", fmt.Errorf("execute legacy role section template: %w", err)
		}
	}

	return b.String(), nil
}

const injectTag = "AGENT_TEAM"

// InjectRolePromptWithPath injects the role prompt using an explicit rolePath.
func InjectRolePromptWithPath(wtPath, workerID, roleName, rolePath, root string) error {
	content, err := buildRoleSectionFromPath(wtPath, workerID, roleName, rolePath, root)
	if err != nil {
		return err
	}
	if content == "" {
		return nil
	}

	for _, name := range []string{"CLAUDE.md", "AGENTS.md", "GEMINI.md"} {
		fp := filepath.Join(wtPath, name)
		if err := InjectSection(fp, injectTag, content); err != nil {
			return fmt.Errorf("inject %s: %w", name, err)
		}
	}

	return nil
}

// InjectRolePrompt injects the role prompt into CLAUDE.md, AGENTS.md, and GEMINI.md using tagged sections.
func InjectRolePrompt(wtPath, workerID, roleName, root string) error {
	return InjectRolePromptWithPath(wtPath, workerID, roleName, RoleDir(root, roleName), root)
}
