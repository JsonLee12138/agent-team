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
