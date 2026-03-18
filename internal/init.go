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

const buildHashFileName = ".build-hash"

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

// BuildHashPath returns the repository-level build hash file path.
func BuildHashPath(root string) string {
	return filepath.Join(root, buildHashFileName)
}

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

// BuildRulesNeedRebuild compares the current build-script hash against the stored record.
func BuildRulesNeedRebuild(root string) (needsRebuild bool, currentHash, storedHash string, err error) {
	scan, err := DetectProjectBuildScripts(root)
	if err != nil {
		return false, "", "", err
	}

	storedHash, err = ReadBuildHash(root)
	if err != nil {
		if os.IsNotExist(err) {
			return true, scan.Hash, "", nil
		}
		return false, "", "", err
	}

	return storedHash != scan.Hash, scan.Hash, storedHash, nil
}

// ReadBuildHash reads the stored build-script hash from the repository root.
func ReadBuildHash(root string) (string, error) {
	data, err := os.ReadFile(BuildHashPath(root))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// EnsureBuildHashFile writes the hash file if it does not already exist.
func EnsureBuildHashFile(root string) error {
	if _, err := os.Stat(BuildHashPath(root)); err == nil {
		return nil
	}

	scan, err := DetectProjectBuildScripts(root)
	if err != nil {
		return err
	}
	return os.WriteFile(BuildHashPath(root), []byte(scan.Hash+"\n"), 0644)
}

// RebuildBuildRules regenerates build-verification.md and refreshes the build hash file.
func RebuildBuildRules(root string) (*BuildScriptScan, error) {
	rulesDir := filepath.Join(ResolveAgentsDir(root), "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		return nil, fmt.Errorf("create .agents/rules/: %w", err)
	}

	scan, err := DetectProjectBuildScripts(root)
	if err != nil {
		return nil, err
	}

	buildRulesPath := filepath.Join(rulesDir, "build-verification.md")
	if err := os.WriteFile(buildRulesPath, []byte(scan.RenderBuildVerificationRules()), 0644); err != nil {
		return nil, fmt.Errorf("write build-verification.md: %w", err)
	}
	if err := os.WriteFile(BuildHashPath(root), []byte(scan.Hash+"\n"), 0644); err != nil {
		return nil, fmt.Errorf("write %s: %w", buildHashFileName, err)
	}

	return scan, nil
}

// RenderBuildVerificationRules returns the generated markdown for build-verification.md.
func (s *BuildScriptScan) RenderBuildVerificationRules() string {
	var b strings.Builder
	b.WriteString("# Build Verification Rules\n\n")
	b.WriteString("## Trigger\n\n")
	b.WriteString("Apply this rule before `go build`, before commit, before review handoff, and before reporting task completion.\n\n")
	b.WriteString("## Project Build Scan\n\n")
	if len(s.Files) == 0 {
		b.WriteString("No supported build-script files were detected. Refresh with `agent-team rules sync --rebuild` after adding `Makefile`, `go.mod`, or `package.json` files.\n\n")
	} else {
		b.WriteString("Detected build-script inputs:\n")
		for _, file := range s.Files {
			b.WriteString("- `" + file + "`\n")
		}
		b.WriteString("\nCurrent build-script hash: `" + s.Hash + "`\n")
		b.WriteString("Refresh this file after build-script changes with `agent-team rules sync --rebuild`.\n\n")
	}

	b.WriteString("## Recommended Verification Commands\n\n")
	if len(s.MakeTargets) > 0 {
		b.WriteString("Primary make targets:\n")
		for _, target := range s.MakeTargets {
			b.WriteString("- `make " + target + "`\n")
		}
		b.WriteString("\n")
	}
	if len(s.GoModules) > 0 {
		b.WriteString("Go module checks:\n")
		for _, mod := range s.GoModules {
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
				b.WriteString("- `" + label + "`: " + strings.Join(meta, ", ") + "\n")
			} else {
				b.WriteString("- `" + label + "`\n")
			}
		}
		b.WriteString("- `go build ./...`\n")
		b.WriteString("- `go vet ./...`\n")
		b.WriteString("- `go test ./...`\n\n")
	}
	if len(s.NodePackages) > 0 {
		b.WriteString("Node package scripts:\n")
		for _, pkg := range s.NodePackages {
			label := pkg.Path
			if pkg.Name != "" {
				label += " (" + pkg.Name + ")"
			}
			b.WriteString("- `" + label + "`\n")
			for _, script := range pkg.Scripts {
				prefix := "npm"
				if pkg.Path != "package.json" {
					prefix = "npm --prefix " + filepath.Dir(pkg.Path)
				}
				b.WriteString("  - `" + prefix + " run " + script.Name + "`")
				if script.Command != "" {
					b.WriteString(" -> `" + script.Command + "`")
				}
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
	}
	if len(s.MakeTargets) == 0 && len(s.GoModules) == 0 && len(s.NodePackages) == 0 {
		b.WriteString("- Inspect the repository for missing build scripts before assuming generic commands.\n\n")
	}

	b.WriteString("## Pre-Build Checks\n\n")
	b.WriteString("- MUST confirm the target package list and working directory before running build commands.\n")
	b.WriteString("- MUST ensure required environment variables, generated files, and module metadata are present.\n")
	b.WriteString("- MUST avoid building from a dirty or partial state when unrelated edits would invalidate results.\n\n")
	b.WriteString("## Required Verification Commands\n\n")
	b.WriteString("- MUST run `go build ./...` for repository-wide Go changes unless the task scope clearly limits the package set.\n")
	b.WriteString("- MUST run `go vet ./...` for the affected scope before commit.\n")
	b.WriteString("- MUST run `go test ./...` for the affected scope; use `./...` for shared or cross-package changes.\n")
	b.WriteString("- MUST rerun the exact failing build or test command when the task is a fix.\n\n")
	b.WriteString("## Pre-Commit Checklist\n\n")
	b.WriteString("- ALWAYS confirm the changed files match the task scope.\n")
	b.WriteString("- ALWAYS review command failures before retrying; NEVER loop without reading output.\n")
	b.WriteString("- MUST verify that build, vet, and test results are current for the final diff.\n")
	b.WriteString("- MUST mention any skipped verification explicitly in the completion message with the reason.\n")
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
- ` + "`build-verification.md`" + `: before build, test, commit, PR
- ` + "`communication.md`" + `: ` + "`reply-main`" + `, blocker escalation, progress update
- ` + "`context-management.md`" + `: context pressure, handoff, provider switch, compact
- ` + "`task-protocol.md`" + `: task start, verify, completion, archive
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
	"build-verification.md": `# Build Verification Rules

## Trigger

Apply this rule before build, before commit, before review handoff, and before reporting task completion.

## Required Verification Commands

- MUST run build for repository-wide changes unless the task scope clearly limits the package set.
- MUST run lint/vet for the affected scope before commit.
- MUST run tests for the affected scope.
- MUST rerun the exact failing build or test command when the task is a fix.

## Pre-Commit Checklist

- ALWAYS confirm the changed files match the task scope.
- ALWAYS review command failures before retrying; NEVER loop without reading output.
- MUST verify that build, lint, and test results are current for the final diff.
`,
	"communication.md": `# Communication Rules

## Trigger

Apply this rule for all worker-to-controller updates, blockers, handoffs, and completion messages.

## ` + "`reply-main`" + ` Format

- MUST use ` + "`agent-team reply-main \"Task completed: <summary>; change archived: <change-name>\"`" + ` after a successful archive.
- MUST use ` + "`agent-team reply-main \"Task completed: <summary>; archive failed for <change-name>: <error>\"`" + ` if archive fails.
- MUST use ` + "`agent-team reply-main \"Need decision: <problem or options>\"`" + ` for blockers or ambiguity.
- ALWAYS keep messages factual, single-purpose, and short enough to scan quickly.

## Escalation Protocol

- MUST report blockers immediately when progress depends on a user or controller decision.
- NEVER hide failed verification, skipped checks, or archive errors.
`,
	"context-management.md": `# Context Management Rules

## Trigger

Apply this rule whenever context grows, the task changes phase, or a provider session is degrading.

## Compact Triggers

1. MUST compact before starting a new logical phase after finishing the current one.
2. MUST compact before reading or pasting large outputs, logs, or diffs that are not yet necessary.
3. MUST compact when the active thread can no longer hold the task goal, constraints, and next actions clearly.
4. MUST compact before handoff, provider switch, session restart, or resumed work after a long pause.

## Provider Handling

- Claude MUST use ` + "`/compact`" + ` when any trigger above fires.
- Codex and Gemini MUST create a manual summary of goal, constraints, changed files, verification state, and next step.
`,
	"task-protocol.md": `# Task Protocol Rules

## Trigger

Apply this rule when a change is assigned, implemented, verified, completed, or handed back to the controller.

## Required Completion Chain

- MUST finish implementation and run the required verification before preparing the final handoff.
- MUST review ` + "`git status`" + ` and stage only task-scoped files.
- MUST commit task-scoped changes before archive when uncommitted task work exists.
- MUST run ` + "`agent-team task archive <worker-id> <change-name>`" + ` after the commit step.
- MUST run ` + "`agent-team reply-main`" + ` after every archive attempt, including failure cases.
- MUST NOT start another task before the completion message has been sent.

## Failure Handling

- MUST report verify failures explicitly and include the failing command or reason.
- MUST report archive failures explicitly and still notify main with the failure details.
- NEVER claim completion while the change is still uncommitted or unreported.
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
		if name == "build-verification.md" {
			scan, rebuildErr := RebuildBuildRules(root)
			if rebuildErr != nil {
				return created, rebuildErr
			}
			if scan != nil {
				created++
				continue
			}
		}
		if err := os.WriteFile(fp, []byte(content), 0644); err != nil {
			return created, fmt.Errorf("write %s: %w", name, err)
		}
		created++
	}
	if err := EnsureBuildHashFile(root); err != nil {
		return created, err
	}
	return created, nil
}

// providerFileTag is the tag used for injecting rules references into provider files.
const providerFileTag = "AGENT_TEAM"

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
	b.WriteString("- MUST read `" + rulesRel + "/index.md` at task start and load the rule files required by the task.\n")
	b.WriteString("- MUST call `/compact` whenever any trigger in `" + rulesRel + "/context-management.md` fires.\n")
	b.WriteString("- MUST keep status updates concise and use `agent-team reply-main` formats from `" + rulesRel + "/communication.md`.\n")
	b.WriteString("- MUST follow `" + rulesRel + "/task-protocol.md` before reporting completion.\n")
	b.WriteString("- MUST obey `" + rulesRel + "/worktree.md` for branch and git safety.\n")
	b.WriteString("\n## Rules Reference\n\n")
	b.WriteString("Load `" + rulesRel + "/index.md` first, then load only the matching rule files:\n\n")
	b.WriteString("- `" + rulesRel + "/debugging.md` for bugs, flaky tests, runtime errors, or unexpected behavior\n")
	b.WriteString("- `" + rulesRel + "/build-verification.md` before `go build`, `go vet`, `go test`, commit, or PR handoff\n")
	b.WriteString("- `" + rulesRel + "/communication.md` for `reply-main`, blocker escalation, and progress updates\n")
	b.WriteString("- `" + rulesRel + "/context-management.md` for `/compact` decisions, handoff summaries, and provider-specific context control\n")
	b.WriteString("- `" + rulesRel + "/task-protocol.md` for task execution, verify, commit, archive, and completion reporting\n")
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
	return nil
}
