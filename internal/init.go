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

// InitProject creates the .agents/teams/ directory with a .gitkeep file.
func InitProject(root string) error {
	teamsDir := filepath.Join(root, ".agents", "teams")
	if err := os.MkdirAll(teamsDir, 0755); err != nil {
		return fmt.Errorf("create .agents/teams/: %w", err)
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

const projectCommandsFileName = "project-commands.md"

type ProjectCommandsGenerator func(root string, scan *BuildScriptScan) (string, error)

var projectCommandsGenerator ProjectCommandsGenerator = generateProjectCommandsWithCodex

func SetProjectCommandsGenerator(fn ProjectCommandsGenerator) func() {
	prev := projectCommandsGenerator
	projectCommandsGenerator = fn
	return func() {
		projectCommandsGenerator = prev
	}
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

// RebuildProjectCommands regenerates project-commands.md using the configured AI generator.
func RebuildProjectCommands(root string) (*BuildScriptScan, error) {
	rulesDir := filepath.Join(ResolveAgentsDir(root), "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		return nil, fmt.Errorf("create .agents/rules/: %w", err)
	}
	if projectCommandsGenerator == nil {
		return nil, fmt.Errorf("generate %s: no generator configured", projectCommandsFileName)
	}

	scan, err := DetectProjectBuildScripts(root)
	if err != nil {
		return nil, err
	}

	content, err := projectCommandsGenerator(root, scan)
	if err != nil {
		return nil, err
	}

	normalized, err := normalizeProjectCommandsContent(content)
	if err != nil {
		return nil, err
	}

	projectCommandsPath := filepath.Join(rulesDir, projectCommandsFileName)
	if err := os.WriteFile(projectCommandsPath, []byte(normalized), 0644); err != nil {
		return nil, fmt.Errorf("write %s: %w", projectCommandsFileName, err)
	}
	if err := cleanupLegacyRuleArtifacts(root); err != nil {
		return nil, err
	}

	return scan, nil
}

func normalizeProjectCommandsContent(content string) (string, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return "", fmt.Errorf("generate %s: empty response", projectCommandsFileName)
	}

	if strings.HasPrefix(trimmed, "```") {
		lines := strings.Split(trimmed, "\n")
		if len(lines) >= 3 && strings.HasPrefix(lines[len(lines)-1], "```") {
			trimmed = strings.TrimSpace(strings.Join(lines[1:len(lines)-1], "\n"))
			trimmed = strings.TrimPrefix(trimmed, "markdown\n")
			trimmed = strings.TrimPrefix(trimmed, "md\n")
			trimmed = strings.TrimSpace(trimmed)
		}
	}

	if idx := strings.Index(trimmed, "# Project Commands Rules"); idx >= 0 {
		trimmed = strings.TrimSpace(trimmed[idx:])
	}
	if !strings.HasPrefix(trimmed, "# Project Commands Rules") {
		return "", fmt.Errorf("generate %s: response missing '# Project Commands Rules' header", projectCommandsFileName)
	}
	if !strings.HasSuffix(trimmed, "\n") {
		trimmed += "\n"
	}
	return trimmed, nil
}

func generateProjectCommandsWithCodex(root string, scan *BuildScriptScan) (string, error) {
	outputFile, err := os.CreateTemp("", "agent-team-project-commands-*.md")
	if err != nil {
		return "", fmt.Errorf("create temp output for %s: %w", projectCommandsFileName, err)
	}
	outputPath := outputFile.Name()
	if err := outputFile.Close(); err != nil {
		return "", fmt.Errorf("close temp output for %s: %w", projectCommandsFileName, err)
	}
	defer os.Remove(outputPath)

	prompt := buildProjectCommandsPrompt(scan)
	cmd := exec.Command(
		"codex", "exec",
		"--full-auto",
		"--ephemeral",
		"--sandbox", "read-only",
		"--skip-git-repo-check",
		"-C", root,
		"-o", outputPath,
		prompt,
	)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("generate %s with codex exec: %w\n%s", projectCommandsFileName, err, strings.TrimSpace(string(output)))
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		return "", fmt.Errorf("read generated %s: %w", projectCommandsFileName, err)
	}
	return string(data), nil
}

func buildProjectCommandsPrompt(scan *BuildScriptScan) string {
	var b strings.Builder
	b.WriteString("Generate the full markdown content for `.agents/rules/project-commands.md` for the current repository.\n\n")
	b.WriteString("Output requirements:\n")
	b.WriteString("- Return markdown only. Do not wrap the answer in code fences.\n")
	b.WriteString("- The title must be exactly `# Project Commands Rules`.\n")
	b.WriteString("- State clearly near the top that this file is AI-generated for the current project and is regenerated by `agent-team init` and `agent-team rules sync`.\n")
	b.WriteString("- This file must tell AI workers to read it before running any project command.\n")
	b.WriteString("- Cover the current project's real command entry points for build, test, lint, dev, format, codegen, and any other detected workflows.\n")
	b.WriteString("- If a command fails, require the AI to inspect the repository and determine the correct command, working directory, prerequisites, or alternative entry point before retrying.\n")
	b.WriteString("- If rule drift is confirmed, require the AI to ask the user whether to update `project-commands.md`.\n")
	b.WriteString("- Do not invent commands that are not grounded in the repository. If something is unclear, say how to inspect it instead of guessing.\n\n")
	b.WriteString("Repository command signals already detected:\n")
	if len(scan.Files) == 0 {
		b.WriteString("- No Makefile, go.mod, or package.json files were detected by the CLI scan.\n")
	} else {
		for _, file := range scan.Files {
			b.WriteString("- `" + file + "`\n")
		}
	}
	if len(scan.MakeTargets) > 0 {
		b.WriteString("- Make targets:\n")
		for _, target := range scan.MakeTargets {
			b.WriteString("  - `make " + target + "`\n")
		}
	}
	if len(scan.GoModules) > 0 {
		b.WriteString("- Go modules:\n")
		for _, mod := range scan.GoModules {
			label := mod.Path
			if label == "go.mod" {
				label = "repo root"
			}
			if mod.Module != "" || mod.GoVersion != "" {
				var meta []string
				if mod.Module != "" {
					meta = append(meta, "module `"+mod.Module+"`")
				}
				if mod.GoVersion != "" {
					meta = append(meta, "go `"+mod.GoVersion+"`")
				}
				b.WriteString("  - `" + label + "`: " + strings.Join(meta, ", ") + "\n")
			} else {
				b.WriteString("  - `" + label + "`\n")
			}
		}
	}
	if len(scan.NodePackages) > 0 {
		b.WriteString("- Node package scripts:\n")
		for _, pkg := range scan.NodePackages {
			label := pkg.Path
			if pkg.Name != "" {
				label += " (" + pkg.Name + ")"
			}
			b.WriteString("  - `" + label + "`\n")
			for _, script := range pkg.Scripts {
				prefix := "npm"
				if pkg.Path != "package.json" {
					prefix = "npm --prefix " + filepath.Dir(pkg.Path)
				}
				b.WriteString("    - `" + prefix + " run " + script.Name + "`")
				if script.Command != "" {
					b.WriteString(" -> `" + script.Command + "`")
				}
				b.WriteString("\n")
			}
		}
	}
	b.WriteString("\nInspect the repository directly before finalizing the rule file. Ground the document in the current repo only.\n")
	return b.String()
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

// defaultRuleFiles maps filename to default content for .agents/rules/.
var defaultRuleFiles = map[string]string{
	"index.md": `# Level 0 Rules Index

Read the matching rule first:
- ` + "`debugging.md`" + `: bug, flaky test, runtime error
- ` + "`project-commands.md`" + `: before running any project command
- ` + "`agent-team-commands.md`" + `: agent-team CLI boundaries, worker lifecycle commands
- ` + "`merge-workflow.md`" + `: controller-side rebase, merge sequencing, generated file safety
- ` + "`context-management.md`" + `: context cleanup, handoff, provider switch, index-first recovery
- ` + "`worktree.md`" + `: worktree safety, branch limits, ignored paths

If unsure, MUST open this index.
`,
	"debugging.md": `# Debugging Rules

## Trigger

Apply this rule for any bug, flaky test, runtime error, build failure, or unexpected behavior.

## Required Flow

MUST follow the ` + "`systematic-debugging`" + ` workflow in order. ALWAYS reproduce, inspect, isolate, test, then validate.

### 1. Reproduce First

- MUST capture the exact command, input, environment, and full error text before changing code.
- MUST retry intermittent failures at least 3 times to confirm the pattern.

### 2. Check Logs and Evidence

- ALWAYS read the full stack trace, build output, and related logs before forming a hypothesis.
- ALWAYS inspect recent relevant changes with ` + "`git diff`" + ` and ` + "`git log`" + ` when regression is possible.

### 3. Isolate the Cause

- MUST reduce the issue to the smallest reproducible case.
- MUST change one variable at a time when testing a hypothesis.

### 4. Validate the Fix

- MUST rerun the original reproduction steps after the fix.
- MUST run the targeted verification commands for the affected scope.
`,
	"agent-team-commands.md": `# Agent-Team Commands Rules

## Trigger

Apply this rule whenever a worker session needs project workflow operations such as create, open, assign, archive, reply, or other ` + "`agent-team`" + ` CLI actions.

## Command Boundary

- MUST use the ` + "`agent-team`" + ` CLI for worker lifecycle and task lifecycle operations when the repository provides it.
- MUST NOT bypass ` + "`agent-team worker open`" + `, ` + "`agent-team worker assign`" + `, or ` + "`agent-team reply-main`" + ` with ad hoc shell commands.
- MUST treat worker bootstrap files and provider prompt files as controller-managed artifacts.

## Generated File Safety

- MUST NOT commit generated worker-local prompt files such as ` + "`CLAUDE.md`" + `, ` + "`GEMINI.md`" + `, or ` + "`AGENTS.md`" + ` from a worker worktree.
- MUST NOT commit worker-local metadata such as ` + "`.tasks/`" + `, ` + "`.claude/`" + `, ` + "`.codex/`" + `, ` + "`.gemini/`" + `, ` + "`.opencode/`" + `, or ` + "`worker.yaml`" + ` from a worker worktree.
- MUST keep deliverable files in tracked repository paths managed by the assigned change.
`,
	"merge-workflow.md": `# Merge Workflow Rules

## Trigger

Apply this rule when preparing a worker branch for assignment, merge, or controller-side synchronization.

## Controller-Side Synchronization

- MUST keep worker-side sessions free of ` + "`git rebase`" + ` and ` + "`git merge`" + ` inside the worker worktree.
- MUST perform any required rebase from the controller side before assignment when the worker is idle.
- MUST stop and surface conflicts immediately if controller-side rebase fails.

## Merge Safety

- MUST merge worker branches back through the controller workflow.
- MUST review generated files and ignore-only artifacts before merge so worker-local prompts or metadata do not enter commits.
- MUST preserve the repository's tracked deliverables while excluding controller-managed bootstrap files.
`,
	"context-management.md": `# Context Management Rules

## Trigger

Apply this rule whenever context grows, the task changes phase, or resumed work needs a recovery anchor.

## Context-Cleanup Triggers

1. MUST enter context cleanup before starting a new logical phase after finishing the current one.
2. MUST enter context cleanup before reading or pasting large outputs, logs, replies, or diffs that may displace the current working context.
3. MUST enter context cleanup when the active thread can no longer hold the task goal, constraints, and next actions clearly.
4. MUST enter context cleanup before resumed work after a long pause, restart, handoff, or provider switch.

## Required Recovery Model

- Context cleanup resets session context; it MUST NOT rewrite, compress, or discard file artifacts.
- Context cleanup is NOT a synonym for ` + "`/compact`" + `.
- Controller/main MUST read ` + "`.agents/rules/index.md`" + ` first, then open only the matching rule files, then the current workflow/task artifacts.
- Worker MUST read ` + "`worker.yaml`" + ` first, then ` + "`task.yaml`" + `, and only then read ` + "`context.md`" + ` or referenced materials when needed.
- NEVER jump directly to rule bodies, ` + "`context.md`" + `, or other detail files before reading the required entry file.
- NEVER default to scanning every context file; expand only what the index or current task explicitly requires.

## Provider Handling

- Claude, Codex, Gemini, and other providers MUST follow the same context-cleanup and index-first recovery strategy.
- Provider-specific prompt injections MUST point to this rule instead of requiring ` + "`/compact`" + `.
`,
	"worktree.md": `# Worktree Rules

## Trigger

Apply this rule for any git command, branch action, file placement decision, or task work inside a worker worktree.

## Branch and Worktree Safety

- MUST work only inside the assigned worktree.
- MUST keep all task commits on the current ` + "`team/<worker-id>`" + ` branch.
- NEVER run ` + "`git checkout`" + `, ` + "`git switch`" + `, ` + "`git merge`" + `, or ` + "`git rebase`" + ` inside the worker worktree.

## File Placement

- MUST keep deliverables in tracked repository paths.
- MUST NOT place task outputs in ignored locations.

## Staging and Commit Scope

- MUST inspect ` + "`git status`" + ` before staging changes.
- MUST stage only files required for the assigned change.
- NEVER use blanket staging commands that may capture unrelated work.
`,
}

// InitRulesDir creates .agents/rules/ with default rule files.
// Idempotent: does not overwrite existing files.
func InitRulesDir(root string) (created int, err error) {
	rulesDir := filepath.Join(ResolveAgentsDir(root), "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		return 0, fmt.Errorf("create .agents/rules/: %w", err)
	}

	for name, content := range defaultRuleFiles {
		fp := filepath.Join(rulesDir, name)
		if _, err := os.Stat(fp); err == nil {
			continue // already exists, skip
		}
		if err := os.WriteFile(fp, []byte(content), 0644); err != nil {
			return created, fmt.Errorf("write %s: %w", name, err)
		}
		created++
	}
	if err := cleanupLegacyRuleArtifacts(root); err != nil {
		return created, err
	}
	return created, nil
}

// SyncRulesDir overwrites the built-in static rule files with the current templates.
func SyncRulesDir(root string) (written int, err error) {
	rulesDir := filepath.Join(ResolveAgentsDir(root), "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		return 0, fmt.Errorf("create .agents/rules/: %w", err)
	}

	for name, content := range defaultRuleFiles {
		fp := filepath.Join(rulesDir, name)
		if err := os.WriteFile(fp, []byte(content), 0644); err != nil {
			return written, fmt.Errorf("write %s: %w", name, err)
		}
		written++
	}
	if err := cleanupLegacyRuleArtifacts(root); err != nil {
		return written, err
	}
	return written, nil
}

func cleanupLegacyRuleArtifacts(root string) error {
	rulesDir := filepath.Join(ResolveAgentsDir(root), "rules")
	legacyPaths := []string{
		filepath.Join(rulesDir, "build-verification.md"),
		filepath.Join(root, ".build-hash"),
	}
	for _, path := range legacyPaths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove legacy rule artifact %s: %w", path, err)
		}
	}
	return nil
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
	b.WriteString("- MUST read `" + rulesRel + "/project-commands.md` before running any project command.\n")
	b.WriteString("- MUST follow `" + rulesRel + "/agent-team-commands.md` for worker lifecycle and generated-file boundaries.\n")
	b.WriteString("- MUST follow `" + rulesRel + "/merge-workflow.md` for controller-side rebase and merge sequencing.\n")
	b.WriteString("\n## Rules Reference\n\n")
	b.WriteString("Load `" + rulesRel + "/index.md` first, then load only the matching rule files:\n\n")
	b.WriteString("- `" + rulesRel + "/debugging.md` for bugs, flaky tests, runtime errors, or unexpected behavior\n")
	b.WriteString("- `" + rulesRel + "/project-commands.md` before running any project command\n")
	b.WriteString("- `" + rulesRel + "/agent-team-commands.md` for agent-team CLI boundaries and worker lifecycle operations\n")
	b.WriteString("- `" + rulesRel + "/merge-workflow.md` for controller-side rebase, merge ordering, and generated file safety\n")
	b.WriteString("- `" + rulesRel + "/context-management.md` for context-cleanup triggers, session reset boundaries, and index-first file recovery\n")
	b.WriteString("- `" + rulesRel + "/worktree.md` for branch safety, worktree limits, and ignored path handling\n")
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
