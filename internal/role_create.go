// internal/role_create.go
package internal

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

//go:embed templates/*.tmpl
var roleTemplateFS embed.FS

var (
	kebabCasePattern      = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	checkboxLinePattern   = regexp.MustCompile(`^\s*(\d+)\.\s*\[([xX ])\]`)
	numericSelectPattern  = regexp.MustCompile(`^\d+(?:\s*,\s*\d+)*$`)
	nonAlnumPattern       = regexp.MustCompile(`[^a-z0-9]+`)
	multiDashPattern      = regexp.MustCompile(`-{2,}`)
)

// managedFiles maps output path → template filename.
var managedFiles = map[string]string{
	"SKILL.md":             "SKILL.md.tmpl",
	"references/role.yaml": "role.yaml.tmpl",
	"system.md":            "system.md.tmpl",
}

// RoleConfig holds all inputs needed to generate a role skill package.
type RoleConfig struct {
	RoleName    string
	Description string
	SystemGoal  string
	InScope     []string
	OutOfScope  []string
	Skills      []string
}

// GenerationResult describes what was created.
type GenerationResult struct {
	TargetDir string
}

// IsKebabCase reports whether s is valid kebab-case.
func IsKebabCase(s string) bool {
	return kebabCasePattern.MatchString(s)
}

// NormalizeRoleName converts an arbitrary string to kebab-case.
func NormalizeRoleName(s string) string {
	lowered := strings.ToLower(strings.TrimSpace(s))
	normalized := nonAlnumPattern.ReplaceAllString(lowered, "-")
	normalized = multiDashPattern.ReplaceAllString(normalized, "-")
	return strings.Trim(normalized, "-")
}

// ValidateRoleName returns the name unchanged if valid, otherwise returns an
// error that includes a kebab-case suggestion.
func ValidateRoleName(roleName string) (string, error) {
	if IsKebabCase(roleName) {
		return roleName, nil
	}
	suggestion := NormalizeRoleName(roleName)
	if suggestion == "" {
		suggestion = "role-name"
	}
	return "", fmt.Errorf("role name %q must be kebab-case (example: 'frontend-dev'). Try %q.", roleName, suggestion)
}

// ParseCSVList splits a comma-separated string into trimmed non-empty items.
func ParseCSVList(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(s, ",") {
		if t := strings.TrimSpace(part); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// DedupeKeepOrder removes duplicates while preserving insertion order.
func DedupeKeepOrder(items []string) []string {
	seen := make(map[string]bool, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			out = append(out, item)
		}
	}
	return out
}

// ParseCheckboxIndices parses a reply that may contain checkbox lines.
// Returns nil if no checkbox lines are found; otherwise returns the 1-based
// indices of checked items (limited to [1, maxIndex]).
func ParseCheckboxIndices(reply string, maxIndex int) []int {
	hasCheckbox := false
	var checked []int
	seen := make(map[int]bool)
	for _, line := range strings.Split(reply, "\n") {
		m := checkboxLinePattern.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil {
			continue
		}
		hasCheckbox = true
		var idx int
		fmt.Sscan(m[1], &idx)
		mark := strings.ToLower(m[2])
		if mark != "x" {
			continue
		}
		if idx >= 1 && idx <= maxIndex && !seen[idx] {
			seen[idx] = true
			checked = append(checked, idx)
		}
	}
	if !hasCheckbox {
		return nil
	}
	return checked
}

// ParseNumericIndices parses a reply for comma-separated numeric index lines.
func ParseNumericIndices(reply string, maxIndex int) []int {
	seen := make(map[int]bool)
	var out []int
	for _, line := range strings.Split(reply, "\n") {
		stripped := strings.TrimSpace(line)
		if stripped == "" {
			continue
		}
		if checkboxLinePattern.MatchString(stripped) {
			continue
		}
		if !numericSelectPattern.MatchString(stripped) {
			continue
		}
		for _, part := range strings.Split(stripped, ",") {
			var idx int
			fmt.Sscan(strings.TrimSpace(part), &idx)
			if idx >= 1 && idx <= maxIndex && !seen[idx] {
				seen[idx] = true
				out = append(out, idx)
			}
		}
	}
	return out
}

// ParseSelectionReply parses the user's selection reply against a recommended
// skill list. Returns (selectedSkills, mode, checkboxPrecedence).
func ParseSelectionReply(reply string, recommendedSkills []string) ([]string, string, bool) {
	checkboxIdx := ParseCheckboxIndices(reply, len(recommendedSkills))
	numericIdx := ParseNumericIndices(reply, len(recommendedSkills))

	if checkboxIdx != nil {
		selected := make([]string, 0, len(checkboxIdx))
		for _, i := range checkboxIdx {
			selected = append(selected, recommendedSkills[i-1])
		}
		return selected, "checkbox", len(numericIdx) > 0
	}

	if len(numericIdx) > 0 {
		selected := make([]string, 0, len(numericIdx))
		for _, i := range numericIdx {
			selected = append(selected, recommendedSkills[i-1])
		}
		return selected, "numeric", false
	}

	return nil, "", false
}

// ResolveFinalSkills merges skill inputs with dedup and removal.
func ResolveFinalSkills(selectedSkills, recommendedSkills, addedSkills, removedSkills, manualSkills []string) []string {
	selected := DedupeKeepOrder(selectedSkills)
	recommended := DedupeKeepOrder(recommendedSkills)
	manual := DedupeKeepOrder(manualSkills)
	added := DedupeKeepOrder(addedSkills)
	removed := make(map[string]bool)
	for _, s := range DedupeKeepOrder(removedSkills) {
		removed[s] = true
	}

	base := selected
	if len(base) == 0 {
		base = recommended
	}
	if len(base) == 0 {
		base = manual
	}

	var final []string
	for _, s := range base {
		if !removed[s] {
			final = append(final, s)
		}
	}
	for _, s := range added {
		if !removed[s] {
			found := false
			for _, f := range final {
				if f == s {
					found = true
					break
				}
			}
			if !found {
				final = append(final, s)
			}
		}
	}
	if final == nil {
		final = []string{}
	}
	return final
}

// CollectScope merges values from --in-scope / --out-of-scope flags (each may
// be comma-separated) and returns the fallback if the result is empty.
func CollectScope(values []string, fallback string) []string {
	var merged []string
	for _, v := range values {
		merged = append(merged, ParseCSVList(v)...)
	}
	if len(merged) == 0 {
		return []string{fallback}
	}
	return merged
}

// templateData is the struct passed to Go templates.
type templateData struct {
	RoleName       string
	Description    string
	DescriptionYAML string
	SystemGoal     string
	InScopeMD      string
	OutOfScopeMD   string
	InScopeSummary string
	SkillsMD       string
	InScopeYAML    string
	OutOfScopeYAML string
	SkillsField    string
}

func mdBullets(items []string) string {
	lines := make([]string, len(items))
	for i, item := range items {
		lines[i] = "- " + item
	}
	return strings.Join(lines, "\n")
}

func yamlQuote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func yamlList(items []string, indent int) string {
	prefix := strings.Repeat(" ", indent)
	lines := make([]string, len(items))
	for i, item := range items {
		lines[i] = prefix + "- " + yamlQuote(item)
	}
	return strings.Join(lines, "\n")
}

func renderSkillsField(skills []string) string {
	if len(skills) == 0 {
		return "skills: []"
	}
	return "skills:\n" + yamlList(skills, 2)
}

// RenderFiles renders all three managed files and returns their contents.
func RenderFiles(config RoleConfig) (map[string]string, error) {
	data := templateData{
		RoleName:        config.RoleName,
		Description:     config.Description,
		DescriptionYAML: yamlQuote(config.Description),
		SystemGoal:      config.SystemGoal,
		InScopeMD:       mdBullets(config.InScope),
		OutOfScopeMD:    mdBullets(config.OutOfScope),
		InScopeSummary:  strings.ToLower(strings.Join(config.InScope, ", ")),
		SkillsMD:        mdBullets(config.Skills),
		InScopeYAML:     yamlList(config.InScope, 4),
		OutOfScopeYAML:  yamlList(config.OutOfScope, 4),
		SkillsField:     renderSkillsField(config.Skills),
	}

	rendered := make(map[string]string, len(managedFiles))
	for outputPath, tmplName := range managedFiles {
		tmplContent, err := fs.ReadFile(roleTemplateFS, "templates/"+tmplName)
		if err != nil {
			return nil, fmt.Errorf("read template %s: %w", tmplName, err)
		}
		t, err := template.New(tmplName).Parse(string(tmplContent))
		if err != nil {
			return nil, fmt.Errorf("parse template %s: %w", tmplName, err)
		}
		var buf bytes.Buffer
		if err := t.Execute(&buf, data); err != nil {
			return nil, fmt.Errorf("render template %s: %w", tmplName, err)
		}
		content := strings.TrimRight(buf.String(), "\n") + "\n"
		rendered[outputPath] = content
	}
	return rendered, nil
}

// CreateOrUpdateRole creates (or overwrites) a role skill package.
//
// confirmFn is called when overwriteMode == "ask". It receives the target
// directory path and should return true if the user confirms overwrite.
func CreateOrUpdateRole(
	repoRoot string,
	config RoleConfig,
	overwriteMode string, // "ask" | "yes" | "no"
	confirmFn func(targetDir string) (bool, error),
	targetDirName string,
) (GenerationResult, error) {
	if _, err := ValidateRoleName(config.RoleName); err != nil {
		return GenerationResult{}, err
	}

	var baseDir string
	if targetDirName == ".agent-team/teams" {
		baseDir = filepath.Join(ResolveAgentsDir(repoRoot), "teams")
	} else {
		baseDir = filepath.Join(repoRoot, targetDirName)
	}
	targetDir := filepath.Join(baseDir, config.RoleName)

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return GenerationResult{}, fmt.Errorf("create base dir: %w", err)
	}

	if info, err := os.Stat(targetDir); err == nil {
		if !info.IsDir() {
			return GenerationResult{}, fmt.Errorf("target path exists and is not a directory: %s", targetDir)
		}

		shouldOverwrite := overwriteMode == "yes"
		if overwriteMode == "ask" {
			if confirmFn == nil {
				return GenerationResult{}, fmt.Errorf("confirmFn required for overwrite mode 'ask'")
			}
			ok, err := confirmFn(targetDir)
			if err != nil {
				return GenerationResult{}, err
			}
			shouldOverwrite = ok
		}
		if overwriteMode == "no" || !shouldOverwrite {
			return GenerationResult{}, fmt.Errorf("role directory already exists and overwrite not confirmed: %s", targetDir)
		}
	} else if !os.IsNotExist(err) {
		return GenerationResult{}, err
	} else {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return GenerationResult{}, fmt.Errorf("create target dir: %w", err)
		}
	}

	rendered, err := RenderFiles(config)
	if err != nil {
		return GenerationResult{}, err
	}

	for outputPath, content := range rendered {
		fullPath := filepath.Join(targetDir, outputPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return GenerationResult{}, err
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return GenerationResult{}, err
		}
	}

	// Remove legacy root-level role.yaml if present.
	legacyYAML := filepath.Join(targetDir, "role.yaml")
	if info, err := os.Stat(legacyYAML); err == nil && !info.IsDir() {
		_ = os.Remove(legacyYAML)
	}

	return GenerationResult{TargetDir: targetDir}, nil
}
