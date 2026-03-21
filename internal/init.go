package internal

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// DetectedProvider holds info about a detected agent provider binary.
type DetectedProvider struct {
	Name string
	Path string
}

// DetectInstalledProviders checks for known agent provider binaries on PATH.
func DetectInstalledProviders() []DetectedProvider {
	names := []string{"claude", "gemini", "opencode", "codex"}
	var detected []DetectedProvider
	for _, name := range names {
		if path, err := exec.LookPath(name); err == nil {
			detected = append(detected, DetectedProvider{Name: name, Path: path})
		}
	}
	return detected
}

// PluginRoleCandidate represents a role found in the plugin's skills directory.
type PluginRoleCandidate struct {
	Name    string
	Path    string
	DirHash string
}

// ScanPluginRoles scans $CLAUDE_PLUGIN_ROOT/skills/ for directories containing references/role.yaml.
func ScanPluginRoles() []PluginRoleCandidate {
	pluginRoot := os.Getenv("CLAUDE_PLUGIN_ROOT")
	if pluginRoot == "" {
		return nil
	}
	skillsDir := filepath.Join(pluginRoot, "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil
	}

	var candidates []PluginRoleCandidate
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		rolePath := filepath.Join(skillsDir, e.Name())
		yamlPath := filepath.Join(rolePath, "references", "role.yaml")
		if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
			continue
		}
		hash, _ := HashLocalRoleDir(rolePath)
		candidates = append(candidates, PluginRoleCandidate{
			Name:    e.Name(),
			Path:    rolePath,
			DirHash: hash,
		})
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Name < candidates[j].Name
	})
	return candidates
}

// HashLocalRoleDir computes a SHA-256 hash of all files in a role directory.
// Files are sorted by relative path for deterministic output.
func HashLocalRoleDir(dir string) (string, error) {
	h := sha256.New()
	var paths []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		paths = append(paths, rel)
		return nil
	})
	if err != nil {
		return "", err
	}

	sort.Strings(paths)
	for _, rel := range paths {
		data, err := os.ReadFile(filepath.Join(dir, rel))
		if err != nil {
			return "", err
		}
		fmt.Fprintf(h, "%s\n%d\n", rel, len(data))
		h.Write(data)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// InstallAction describes what happened during a plugin role install.
type InstallAction string

const (
	InstallActionSkipped   InstallAction = "skipped"
	InstallActionInstalled InstallAction = "installed"
	InstallActionUpdated   InstallAction = "updated"
)

// InstallPluginRoleToGlobal copies a plugin role to the global roles directory.
// It compares hashes to decide: skip (identical), install (new), or update (changed).
func InstallPluginRoleToGlobal(candidate PluginRoleCandidate, globalDir string) (InstallAction, error) {
	destDir := filepath.Join(globalDir, candidate.Name)

	// Check if destination already exists
	if _, err := os.Stat(destDir); err == nil {
		// Destination exists — compare hashes
		existingHash, err := HashLocalRoleDir(destDir)
		if err != nil {
			return "", fmt.Errorf("hash existing role %s: %w", candidate.Name, err)
		}
		if existingHash == candidate.DirHash {
			return InstallActionSkipped, nil
		}
		// Hash differs — update (remove old, copy new)
		if err := os.RemoveAll(destDir); err != nil {
			return "", fmt.Errorf("remove old role %s: %w", candidate.Name, err)
		}
		if err := copyDir(candidate.Path, destDir); err != nil {
			return "", fmt.Errorf("update role %s: %w", candidate.Name, err)
		}
		return InstallActionUpdated, nil
	}

	// New installation
	if err := copyDir(candidate.Path, destDir); err != nil {
		return "", fmt.Errorf("install role %s: %w", candidate.Name, err)
	}
	return InstallActionInstalled, nil
}

// InitProject creates the .agent-team/teams/ directory with a .gitkeep file.
func InitProject(root string) error {
	teamsDir := filepath.Join(ResolveAgentsDir(root), "teams")
	if err := os.MkdirAll(teamsDir, 0755); err != nil {
		return fmt.Errorf("create .agent-team/teams/: %w", err)
	}
	gitkeep := filepath.Join(teamsDir, ".gitkeep")
	if _, err := os.Stat(gitkeep); os.IsNotExist(err) {
		if err := os.WriteFile(gitkeep, []byte(""), 0644); err != nil {
			return fmt.Errorf("create .gitkeep: %w", err)
		}
	}
	return nil
}

// EnsureGlobalRolesDir creates ~/.agents/roles/ if it doesn't exist.
func EnsureGlobalRolesDir() (string, error) {
	dir, err := GlobalRolesDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create global roles dir: %w", err)
	}
	return dir, nil
}

// FormatProviderList returns a comma-separated string of provider names.
func FormatProviderList(providers []DetectedProvider) string {
	names := make([]string, len(providers))
	for i, p := range providers {
		names[i] = p.Name
	}
	return strings.Join(names, ", ")
}


type BuildCommand struct {
	Name    string
	Command string
}

type BuildPackage struct {
	Path    string
	Name    string
	Scripts []BuildCommand
}

type BuildModule struct {
	Path      string
	Module    string
	GoVersion string
}

type BuildScriptScan struct {
	Files        []string
	Hash         string
	MakeTargets  []string
	GoModules    []BuildModule
	NodePackages []BuildPackage
}

type packageJSONBuildInfo struct {
	Name    string            `json:"name"`
	Scripts map[string]string `json:"scripts"`
}

var makeTargetPattern = regexp.MustCompile(`^([A-Za-z0-9_.-]+):(?:\s|$)`)

// DetectProjectBuildScripts scans known build-script files and generates a deterministic summary.
func DetectProjectBuildScripts(root string) (*BuildScriptScan, error) {
	files, err := collectBuildScriptFiles(root)
	if err != nil {
		return nil, err
	}

	hash, err := hashFiles(root, files)
	if err != nil {
		return nil, err
	}

	scan := &BuildScriptScan{
		Files: files,
		Hash:  hash,
	}

	for _, rel := range files {
		abs := filepath.Join(root, rel)
		switch filepath.Base(rel) {
		case "Makefile", "makefile", "GNUmakefile":
			targets, err := parseMakefileTargets(abs)
			if err != nil {
				return nil, err
			}
			scan.MakeTargets = append(scan.MakeTargets, targets...)
		case "go.mod":
			mod, err := parseGoModFile(abs, rel)
			if err != nil {
				return nil, err
			}
			scan.GoModules = append(scan.GoModules, mod)
		case "package.json":
			pkg, err := parsePackageJSONFile(abs, rel)
			if err != nil {
				return nil, err
			}
			scan.NodePackages = append(scan.NodePackages, pkg)
		}
	}

	scan.MakeTargets = uniqueStrings(scan.MakeTargets)
	sort.Strings(scan.MakeTargets)
	sort.Slice(scan.GoModules, func(i, j int) bool { return scan.GoModules[i].Path < scan.GoModules[j].Path })
	sort.Slice(scan.NodePackages, func(i, j int) bool { return scan.NodePackages[i].Path < scan.NodePackages[j].Path })

	return scan, nil
}


func collectBuildScriptFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}

		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}

		if d.IsDir() {
			switch d.Name() {
			case ".git", ".worktrees", "worktrees", "node_modules", "vendor":
				return filepath.SkipDir
			}
			return nil
		}

		switch filepath.Base(path) {
		case "Makefile", "makefile", "GNUmakefile", "go.mod", "package.json":
			files = append(files, rel)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan build scripts: %w", err)
	}

	sort.Strings(files)
	return files, nil
}

func hashFiles(root string, relPaths []string) (string, error) {
	h := sha256.New()
	for _, rel := range relPaths {
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			return "", fmt.Errorf("read %s: %w", rel, err)
		}
		fmt.Fprintf(h, "%s\n%d\n", rel, len(data))
		h.Write(data)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func parseMakefileTargets(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var targets []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ".") {
			continue
		}
		match := makeTargetPattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		target := match[1]
		if strings.Contains(target, "%") {
			continue
		}
		targets = append(targets, target)
	}
	return uniqueStrings(targets), nil
}

func parseGoModFile(path, rel string) (BuildModule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return BuildModule{}, fmt.Errorf("read %s: %w", path, err)
	}

	mod := BuildModule{Path: rel}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "module "):
			mod.Module = strings.TrimSpace(strings.TrimPrefix(line, "module "))
		case strings.HasPrefix(line, "go "):
			mod.GoVersion = strings.TrimSpace(strings.TrimPrefix(line, "go "))
		}
	}
	return mod, nil
}

func parsePackageJSONFile(path, rel string) (BuildPackage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return BuildPackage{}, fmt.Errorf("read %s: %w", path, err)
	}

	var meta packageJSONBuildInfo
	if err := json.Unmarshal(data, &meta); err != nil {
		return BuildPackage{}, fmt.Errorf("parse %s: %w", rel, err)
	}

	pkg := BuildPackage{
		Path: rel,
		Name: meta.Name,
	}

	if len(meta.Scripts) == 0 {
		return pkg, nil
	}

	names := make([]string, 0, len(meta.Scripts))
	for name := range meta.Scripts {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		pkg.Scripts = append(pkg.Scripts, BuildCommand{
			Name:    name,
			Command: strings.TrimSpace(meta.Scripts[name]),
		})
	}
	return pkg, nil
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	var result []string
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

// --- Rules directory initialization ---

// InitRulesDir keeps the legacy public name but now delegates to the v2 rules layout.
func InitRulesDir(root string) (created int, err error) {
	return InitRulesDirV2(root)
}

// SyncRulesDir keeps the legacy public name but now delegates to the v2 rules layout.
func SyncRulesDir(root string) (written int, err error) {
	return SyncRulesDirV2(root)
}

// providerFileTag is the tag used for injecting rules references into provider files.
const providerFileTag = "AGENT_TEAM"

type ClaudeSettings map[string]json.RawMessage

type ClaudeHookMatcher struct {
	Matcher string      `json:"matcher,omitempty"`
	Hooks   []ClaudeHook `json:"hooks"`
}

type ClaudeHook struct {
	Name    string `json:"name,omitempty"`
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"`
}

// defaultProviderInstructions returns the content to inject into provider files.
func defaultProviderInstructions(root string) string {
	agentsDir := ResolveAgentsDir(root)
	// Use relative path from root for portability
	relAgents, err := filepath.Rel(root, agentsDir)
	if err != nil {
		relAgents = ".agents"
	}
	rulesRel := filepath.Join(relAgents, "rules")

	var b strings.Builder
	b.WriteString("# Claude Instructions\n\n")
	b.WriteString("Use this file when working in Claude Code on this repository.\n\n")
	b.WriteString("- MUST read `" + rulesRel + "/index.md` at task start and load only the rule files required by the task.\n")
	b.WriteString("- MUST follow `" + rulesRel + "/context-management.md` for context-cleanup and index-first recovery whenever context drifts, phases change, or work resumes.\n")
	b.WriteString("- MUST keep status updates concise.\n")
	b.WriteString("- MUST obey `" + rulesRel + "/worktree.md` for branch and git safety.\n")
	b.WriteString("- MUST read `" + rulesRel + "/index.md` at task start and then open only the relevant files under `" + rulesRel + "/core/` and `" + rulesRel + "/project/`.\n")
	b.WriteString("- MUST follow `" + rulesRel + "/core/context-management.md` for context-cleanup and index-first recovery whenever context drifts, phases change, or work resumes.\n")
	b.WriteString("- MUST keep status updates concise.\n")
	b.WriteString("- MUST obey `" + rulesRel + "/core/worktree.md` for branch and git safety.\n")
	b.WriteString("- MUST follow `" + rulesRel + "/core/agent-team-commands.md` for worker lifecycle and generated-file boundaries.\n")
	b.WriteString("- MUST follow `" + rulesRel + "/core/merge-workflow.md` for controller-side rebase and merge sequencing.\n")
	b.WriteString("\n## Rules Reference\n\n")
	b.WriteString("Load `" + rulesRel + "/index.md` first, then load only the matching rule files:\n\n")
	b.WriteString("- `" + rulesRel + "/core/debugging.md` for bugs, flaky tests, runtime errors, or unexpected behavior\n")
	b.WriteString("- `" + rulesRel + "/core/agent-team-commands.md` for agent-team CLI boundaries and worker lifecycle operations\n")
	b.WriteString("- `" + rulesRel + "/core/merge-workflow.md` for controller-side rebase, merge ordering, and generated file safety\n")
	b.WriteString("- `" + rulesRel + "/core/context-management.md` for context-cleanup triggers, session reset boundaries, and index-first file recovery\n")
	b.WriteString("- `" + rulesRel + "/core/worktree.md` for branch safety, worktree limits, and ignored path handling\n")
	b.WriteString("- Read the relevant files under `" + rulesRel + "/project/` before running project-specific commands or workflows\n")
	return b.String()
}

// InitProviderFiles creates or updates CLAUDE.md, AGENTS.md, and GEMINI.md at root.
// If a file does not exist, creates it with full content.
// If it exists, only updates the tagged section (preserves user content).
func InitProviderFiles(root string) error {
	content := defaultProviderInstructions(root)
	for _, name := range []string{"CLAUDE.md", "AGENTS.md", "GEMINI.md"} {
		fp := filepath.Join(root, name)
		if err := InjectSection(fp, providerFileTag, content); err != nil {
			return fmt.Errorf("inject %s: %w", name, err)
		}
	}
	if err := InitClaudeLocalSettings(root); err != nil {
		return err
	}
	return nil
}

func InitClaudeLocalSettings(root string) error {
	settingsPath := filepath.Join(root, ".claude", "settings.local.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		return fmt.Errorf("create .claude directory: %w", err)
	}

	cfg := ClaudeSettings{}
	if data, err := os.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("parse %s: %w", settingsPath, err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", settingsPath, err)
	}

	hook := ClaudeHook{
		Name:    "record-main-pane",
		Type:    "command",
		Command: "./scripts/session-start-record-main-pane.sh",
		Timeout: 10000,
	}
	matcher := ClaudeHookMatcher{
		Matcher: "*",
		Hooks:   []ClaudeHook{hook},
	}

	var hooks map[string][]ClaudeHookMatcher
	if raw, ok := cfg["hooks"]; ok && len(raw) > 0 {
		if err := json.Unmarshal(raw, &hooks); err != nil {
			return fmt.Errorf("parse hooks in %s: %w", settingsPath, err)
		}
	}
	if hooks == nil {
		hooks = map[string][]ClaudeHookMatcher{}
	}
	hooks["SessionStart"] = upsertSessionStartMatcher(hooks["SessionStart"], matcher)

	hooksData, err := json.Marshal(hooks)
	if err != nil {
		return fmt.Errorf("marshal hooks for %s: %w", settingsPath, err)
	}
	cfg["hooks"] = hooksData

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", settingsPath, err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("write %s: %w", settingsPath, err)
	}
	return nil
}

func upsertSessionStartMatcher(existing []ClaudeHookMatcher, matcher ClaudeHookMatcher) []ClaudeHookMatcher {
	for i := range existing {
		if existing[i].Matcher != matcher.Matcher {
			continue
		}
		for _, want := range matcher.Hooks {
			found := false
			for _, got := range existing[i].Hooks {
				if got.Command == want.Command && got.Type == want.Type {
					found = true
					break
				}
			}
			if !found {
				existing[i].Hooks = append(existing[i].Hooks, want)
			}
		}
		return existing
	}
	return append(existing, matcher)
}
