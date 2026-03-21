package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const projectRulesDirName = "project"
const projectCommandsFileName = "project-commands.md"

var requiredCoreRuleFiles = []string{
	"debugging.md",
	"agent-team-commands.md",
	"merge-workflow.md",
	"context-management.md",
	"worktree.md",
	"skill-resolution.md",
}

var defaultCoreRuleFiles = map[string]string{
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
- MUST NOT bypass ` + "`agent-team worker open`" + `, ` + "`agent-team task assign`" + `, or ` + "`agent-team reply-main`" + ` with ad hoc shell commands.
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
- Context cleanup is a standalone session reset and file re-anchoring flow.
- Controller/main MUST read ` + "`.agent-team/rules/index.md`" + ` first, then open only the matching rule files, then the current workflow/task artifacts.
- Worker MUST read ` + "`worker.yaml`" + ` first, then ` + "`task.yaml`" + `, and only then read ` + "`context.md`" + ` or referenced materials when needed.
- NEVER jump directly to rule bodies, ` + "`context.md`" + `, or other detail files before reading the required entry file.
- NEVER default to scanning every context file; expand only what the index or current task explicitly requires.

## Provider Handling

- Claude, Codex, Gemini, and other providers MUST follow the same context-cleanup and index-first recovery strategy.
- Provider-specific prompt injections MUST point to this rule for context-cleanup guidance.
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
	"skill-resolution.md": `# Skill Resolution Rules

## Trigger

Apply this rule whenever a role needs a skill that is missing locally at runtime.

## Required Flow

1. MUST run ` + "`find-skills`" + ` first to search for a matching skill.
2. MUST allow only project-level installation for runtime skill resolution.
3. MUST NOT trigger any global skill installation path from a worker/runtime flow.
4. If project-level installation fails, MUST print a warning with the failure reason and a suggested next step, then continue the current task.
5. After the task, MAY suggest regenerating the role with ` + "`agent-team role create`" + ` so the skill is declared in ` + "`references/role.yaml`" + `.
`,
}

var defaultRuleFiles = map[string]string{
	"index.md": defaultRulesIndexContent,
}

const defaultRulesIndexContent = `# Rules Index

Read this file first. It is the single entry point for project rules.

## Core Rules

- ` + "`core/debugging.md`" + `: use for bugs, flaky tests, runtime errors, build failures, and unexpected behavior.
- ` + "`core/agent-team-commands.md`" + `: use for worker lifecycle, task lifecycle, and agent-team CLI boundaries.
- ` + "`core/merge-workflow.md`" + `: use for controller-side rebase, synchronization, and merge sequencing.
- ` + "`core/context-management.md`" + `: use for context-cleanup, index-first recovery, and resume rules.
- ` + "`core/worktree.md`" + `: use for branch safety, worktree limits, and file placement.
- ` + "`core/skill-resolution.md`" + `: use when runtime skill lookup or installation is needed.

## Project Rules

- Read the relevant files under ` + "`project/`" + ` before running project-specific commands or workflows.
- The ` + "`project/`" + ` directory is AI-generated during ` + "`agent-team init`" + ` and refreshed by ` + "`agent-team rules sync`" + `.
- Project rules must stay split by single responsibility. Do not collapse them back into this index.
`

type ProjectRuleFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type projectRulesEnvelope struct {
	Files []ProjectRuleFile `json:"files"`
}

type ProjectRulesGenerator func(root string, scan *BuildScriptScan) (string, error)

var projectRulesGenerator ProjectRulesGenerator = generateProjectRulesWithCodex

func SetProjectRulesGenerator(fn ProjectRulesGenerator) func() {
	prev := projectRulesGenerator
	projectRulesGenerator = fn
	return func() {
		projectRulesGenerator = prev
	}
}

func SetProjectCommandsGenerator(fn func(root string, scan *BuildScriptScan) (string, error)) func() {
	return SetProjectRulesGenerator(func(root string, scan *BuildScriptScan) (string, error) {
		content, err := fn(root, scan)
		if err != nil {
			return "", err
		}
		env := projectRulesEnvelope{
			Files: []ProjectRuleFile{{
				Path:    projectCommandsFileName,
				Content: content,
			}},
		}
		data, err := json.Marshal(env)
		if err != nil {
			return "", err
		}
		return string(data), nil
	})
}

func RulesRootDir(root string) string {
	return filepath.Join(ResolveAgentsDir(root), "rules")
}

func RulesIndexPath(root string) string {
	return filepath.Join(RulesRootDir(root), "index.md")
}

func RulesCoreDir(root string) string {
	return filepath.Join(RulesRootDir(root), "core")
}

func RulesProjectDir(root string) string {
	return filepath.Join(RulesRootDir(root), projectRulesDirName)
}

func InitRulesDirV2(root string) (created int, err error) {
	for _, dir := range []string{RulesRootDir(root), RulesCoreDir(root), RulesProjectDir(root)} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return created, fmt.Errorf("create rules directory %s: %w", dir, err)
		}
	}
	if _, err := os.Stat(RulesIndexPath(root)); os.IsNotExist(err) {
		if err := os.WriteFile(RulesIndexPath(root), []byte(defaultRulesIndexContent), 0644); err != nil {
			return created, fmt.Errorf("write index.md: %w", err)
		}
		created++
	}
	for name, content := range defaultCoreRuleFiles {
		path := filepath.Join(RulesCoreDir(root), name)
		if _, err := os.Stat(path); err == nil {
			continue
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return created, fmt.Errorf("write %s: %w", name, err)
		}
		created++
	}
	if err := cleanupLegacyRuleArtifactsV2(root); err != nil {
		return created, err
	}
	return created, nil
}

func SyncRulesDirV2(root string) (written int, err error) {
	for _, dir := range []string{RulesRootDir(root), RulesCoreDir(root), RulesProjectDir(root)} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return written, fmt.Errorf("create rules directory %s: %w", dir, err)
		}
	}
	if err := os.WriteFile(RulesIndexPath(root), []byte(defaultRulesIndexContent), 0644); err != nil {
		return written, fmt.Errorf("write index.md: %w", err)
	}
	written++
	for name, content := range defaultCoreRuleFiles {
		path := filepath.Join(RulesCoreDir(root), name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return written, fmt.Errorf("write %s: %w", name, err)
		}
		written++
	}
	if err := cleanupLegacyRuleArtifactsV2(root); err != nil {
		return written, err
	}
	return written, nil
}

func RebuildProjectRules(root string) (*BuildScriptScan, error) {
	projectDir := RulesProjectDir(root)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return nil, fmt.Errorf("create %s: %w", projectDir, err)
	}
	if projectRulesGenerator == nil {
		return nil, fmt.Errorf("generate project rules: no generator configured")
	}

	scan, err := DetectProjectBuildScripts(root)
	if err != nil {
		return nil, err
	}

	raw, err := projectRulesGenerator(root, scan)
	if err != nil {
		return nil, err
	}
	files, err := normalizeProjectRulesContent(raw)
	if err != nil {
		return nil, err
	}

	if err := os.RemoveAll(projectDir); err != nil {
		return nil, fmt.Errorf("clear %s: %w", projectDir, err)
	}
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return nil, fmt.Errorf("recreate %s: %w", projectDir, err)
	}

	for _, file := range files {
		path := filepath.Join(projectDir, file.Path)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, fmt.Errorf("create project rule directory for %s: %w", file.Path, err)
		}
		if err := os.WriteFile(path, []byte(file.Content), 0644); err != nil {
			return nil, fmt.Errorf("write project rule %s: %w", file.Path, err)
		}
	}

	if err := cleanupLegacyRuleArtifactsV2(root); err != nil {
		return nil, err
	}
	return scan, nil
}

func normalizeProjectRulesContent(content string) ([]ProjectRuleFile, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil, fmt.Errorf("generate project rules: empty response")
	}
	if strings.HasPrefix(trimmed, "```") {
		lines := strings.Split(trimmed, "\n")
		if len(lines) >= 3 && strings.HasPrefix(lines[len(lines)-1], "```") {
			trimmed = strings.TrimSpace(strings.Join(lines[1:len(lines)-1], "\n"))
			trimmed = strings.TrimPrefix(trimmed, "json\n")
			trimmed = strings.TrimSpace(trimmed)
		}
	}

	var env projectRulesEnvelope
	if err := json.Unmarshal([]byte(trimmed), &env); err != nil {
		return nil, fmt.Errorf("generate project rules: invalid JSON response: %w", err)
	}
	if len(env.Files) == 0 {
		return nil, fmt.Errorf("generate project rules: response contains no files")
	}

	seen := map[string]struct{}{}
	files := make([]ProjectRuleFile, 0, len(env.Files))
	for _, file := range env.Files {
		rel, err := sanitizeProjectRulePath(file.Path)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[rel]; ok {
			return nil, fmt.Errorf("generate project rules: duplicate path %s", rel)
		}
		seen[rel] = struct{}{}
		trimmedContent := strings.TrimSpace(file.Content)
		if trimmedContent == "" {
			return nil, fmt.Errorf("generate project rules: empty content for %s", rel)
		}
		if !strings.HasPrefix(trimmedContent, "# ") {
			return nil, fmt.Errorf("generate project rules: %s must start with a level-1 markdown heading", rel)
		}
		if !strings.HasSuffix(trimmedContent, "\n") {
			trimmedContent += "\n"
		}
		files = append(files, ProjectRuleFile{Path: rel, Content: trimmedContent})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

func sanitizeProjectRulePath(path string) (string, error) {
	path = filepath.ToSlash(strings.TrimSpace(path))
	path = strings.TrimPrefix(path, "./")
	path = strings.TrimPrefix(path, "project/")
	if path == "" {
		return "", fmt.Errorf("generate project rules: empty file path")
	}
	if strings.HasPrefix(path, "/") {
		return "", fmt.Errorf("generate project rules: absolute path %s is not allowed", path)
	}
	cleaned := filepath.ToSlash(filepath.Clean(path))
	if cleaned == "." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") {
		return "", fmt.Errorf("generate project rules: path %s escapes project rules directory", path)
	}
	if filepath.Ext(cleaned) != ".md" {
		return "", fmt.Errorf("generate project rules: %s must be a markdown file", cleaned)
	}
	return cleaned, nil
}

func generateProjectRulesWithCodex(root string, scan *BuildScriptScan) (string, error) {
	outputFile, err := os.CreateTemp("", "agent-team-project-rules-*.json")
	if err != nil {
		return "", fmt.Errorf("create temp output for project rules: %w", err)
	}
	outputPath := outputFile.Name()
	if err := outputFile.Close(); err != nil {
		return "", fmt.Errorf("close temp output for project rules: %w", err)
	}
	defer os.Remove(outputPath)

	prompt := buildProjectRulesPrompt(scan)
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
		return "", fmt.Errorf("generate project rules with codex exec: %w\n%s", err, strings.TrimSpace(string(output)))
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		return "", fmt.Errorf("read generated project rules: %w", err)
	}
	return string(data), nil
}

func buildProjectRulesPrompt(scan *BuildScriptScan) string {
	var b strings.Builder
	b.WriteString("Generate JSON only for `.agent-team/rules/project/` for the current repository.\n\n")
	b.WriteString("Return a JSON object with this shape exactly:\n")
	b.WriteString("{\n  \"files\": [\n    {\"path\": \"<relative-path>.md\", \"content\": \"<full markdown>\"}\n  ]\n}\n\n")
	b.WriteString("Output requirements:\n")
	b.WriteString("- Return JSON only. Do not wrap the answer in code fences.\n")
	b.WriteString("- Generate multiple markdown files under `.agent-team/rules/project/`, not a single oversized file.\n")
	b.WriteString("- Each file must follow single-responsibility.\n")
	b.WriteString("- Do not hardcode a fixed file count unless the repository evidence demands it.\n")
	b.WriteString("- Cover project-specific command entry points, working-directory expectations, failure handling, and project-specific constraints or cautions.\n")
	b.WriteString("- Do not invent commands that are not grounded in the repository. If information is missing, say how to inspect it instead of guessing.\n")
	b.WriteString("- The markdown in each file must start with a level-1 heading.\n\n")
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
	b.WriteString("\nInspect the repository directly before finalizing the rule files. Ground the result in the current repo only.\n")
	return b.String()
}

type RulesValidationIssue struct {
	Path    string
	Message string
}

type RulesValidationError struct {
	Issues []RulesValidationIssue
}

func (e *RulesValidationError) Error() string {
	if e == nil || len(e.Issues) == 0 {
		return "rules validation failed"
	}
	var b strings.Builder
	b.WriteString("rules validation failed:\n")
	for _, issue := range e.Issues {
		if issue.Path != "" {
			b.WriteString("- " + issue.Path + ": " + issue.Message + "\n")
		} else {
			b.WriteString("- " + issue.Message + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func RebuildProjectCommands(root string) (*BuildScriptScan, error) {
	return RebuildProjectRules(root)
}

func ValidateRules(root string) error {
	var issues []RulesValidationIssue
	indexPath := RulesIndexPath(root)
	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		issues = append(issues, RulesValidationIssue{Path: filepath.ToSlash(filepath.Join(".agent-team", "rules", "index.md")), Message: "missing index.md"})
	} else {
		indexText := string(indexData)
		for _, name := range requiredCoreRuleFiles {
			rel := filepath.ToSlash(filepath.Join("core", name))
			if !strings.Contains(indexText, rel) {
				issues = append(issues, RulesValidationIssue{Path: filepath.ToSlash(filepath.Join(".agent-team", "rules", "index.md")), Message: fmt.Sprintf("missing reference to %s", rel)})
			}
		}
		if !strings.Contains(indexText, "project/") {
			issues = append(issues, RulesValidationIssue{Path: filepath.ToSlash(filepath.Join(".agent-team", "rules", "index.md")), Message: "missing project/ entry hint"})
		}
		if !strings.Contains(indexText, "core/skill-resolution.md") {
			issues = append(issues, RulesValidationIssue{Path: filepath.ToSlash(filepath.Join(".agent-team", "rules", "index.md")), Message: "missing reference to core/skill-resolution.md"})
		}
		validateMarkdownMetrics(indexPath, filepath.ToSlash(filepath.Join(".agent-team", "rules", "index.md")), 80, 4000, 8, &issues)
	}

	coreDir := RulesCoreDir(root)
	if info, err := os.Stat(coreDir); err != nil || !info.IsDir() {
		issues = append(issues, RulesValidationIssue{Path: filepath.ToSlash(filepath.Join(".agent-team", "rules", "core")), Message: "missing core/ directory"})
	} else {
		for _, name := range requiredCoreRuleFiles {
			path := filepath.Join(coreDir, name)
			rel := filepath.ToSlash(filepath.Join(".agent-team", "rules", "core", name))
			if _, err := os.Stat(path); err != nil {
				issues = append(issues, RulesValidationIssue{Path: rel, Message: "missing required core rule"})
				continue
			}
			validateMarkdownMetrics(path, rel, 180, 12000, 12, &issues)
		}
	}

	projectDir := RulesProjectDir(root)
	projectEntries, err := os.ReadDir(projectDir)
	if err != nil {
		issues = append(issues, RulesValidationIssue{Path: filepath.ToSlash(filepath.Join(".agent-team", "rules", "project")), Message: "missing project/ directory"})
	} else {
		var markdownFiles []string
		for _, entry := range projectEntries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
				continue
			}
			markdownFiles = append(markdownFiles, entry.Name())
		}
		if len(markdownFiles) == 0 {
			issues = append(issues, RulesValidationIssue{Path: filepath.ToSlash(filepath.Join(".agent-team", "rules", "project")), Message: "no project rule markdown files were generated"})
		}
		if len(markdownFiles) < 2 {
			issues = append(issues, RulesValidationIssue{Path: filepath.ToSlash(filepath.Join(".agent-team", "rules", "project")), Message: "project rules should be split into multiple files"})
		}
		for _, name := range markdownFiles {
			path := filepath.Join(projectDir, name)
			rel := filepath.ToSlash(filepath.Join(".agent-team", "rules", "project", name))
			validateMarkdownMetrics(path, rel, 260, 18000, 14, &issues)
		}
	}

	if len(issues) > 0 {
		return &RulesValidationError{Issues: issues}
	}
	return nil
}

func validateMarkdownMetrics(absPath, relPath string, maxLines, maxBytes, maxHeadings int, issues *[]RulesValidationIssue) {
	data, err := os.ReadFile(absPath)
	if err != nil {
		*issues = append(*issues, RulesValidationIssue{Path: relPath, Message: fmt.Sprintf("read failed: %v", err)})
		return
	}
	text := strings.TrimSpace(string(data))
	if text == "" {
		*issues = append(*issues, RulesValidationIssue{Path: relPath, Message: "file is empty"})
		return
	}
	if len(data) > maxBytes {
		*issues = append(*issues, RulesValidationIssue{Path: relPath, Message: fmt.Sprintf("file is too large (%d bytes > %d bytes)", len(data), maxBytes)})
	}
	lines := strings.Count(text, "\n") + 1
	if lines > maxLines {
		*issues = append(*issues, RulesValidationIssue{Path: relPath, Message: fmt.Sprintf("file is too long (%d lines > %d lines)", lines, maxLines)})
	}
	headingCount := 0
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "### ") {
			headingCount++
		}
	}
	if headingCount > maxHeadings {
		*issues = append(*issues, RulesValidationIssue{Path: relPath, Message: fmt.Sprintf("file has too many sections (%d > %d); split it by responsibility", headingCount, maxHeadings)})
	}
}

func cleanupLegacyRuleArtifactsV2(root string) error {
	rulesDir := RulesRootDir(root)
	legacyPaths := []string{
		filepath.Join(rulesDir, "build-verification.md"),
		filepath.Join(rulesDir, "debugging.md"),
		filepath.Join(rulesDir, "project-commands.md"),
		filepath.Join(rulesDir, "agent-team-commands.md"),
		filepath.Join(rulesDir, "merge-workflow.md"),
		filepath.Join(rulesDir, "context-management.md"),
		filepath.Join(rulesDir, "worktree.md"),
		filepath.Join(root, ".build-hash"),
	}
	for _, path := range legacyPaths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove legacy rule artifact %s: %w", path, err)
		}
	}
	return nil
}
