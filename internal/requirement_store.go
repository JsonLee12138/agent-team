// internal/requirement_store.go
package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

// RequirementsDir returns the path to .tasks/requirements/.
func RequirementsDir(wtPath string) string {
	return filepath.Join(TasksDir(wtPath), "requirements")
}

// RequirementDir returns the path to a specific requirement directory.
func RequirementDir(wtPath, reqName string) string {
	return filepath.Join(RequirementsDir(wtPath), reqName)
}

// RequirementYAMLPath returns the path to a requirement's YAML file.
func RequirementYAMLPath(wtPath, reqName string) string {
	return filepath.Join(RequirementDir(wtPath, reqName), "requirement.yaml")
}

// RequirementIndexPath returns the path to the requirements index.yaml.
func RequirementIndexPath(wtPath string) string {
	return filepath.Join(RequirementsDir(wtPath), "index.yaml")
}

// SaveRequirement saves a requirement to its YAML file.
func SaveRequirement(wtPath string, req *Requirement) error {
	yamlPath := RequirementYAMLPath(wtPath, req.Name)

	if err := os.MkdirAll(filepath.Dir(yamlPath), 0755); err != nil {
		return fmt.Errorf("create requirement directory: %w", err)
	}

	data, err := yaml.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal requirement: %w", err)
	}

	if err := os.WriteFile(yamlPath, data, 0644); err != nil {
		return fmt.Errorf("write requirement.yaml: %w", err)
	}

	return nil
}

// LoadRequirement loads a requirement from its YAML file.
func LoadRequirement(wtPath, reqName string) (*Requirement, error) {
	yamlPath := RequirementYAMLPath(wtPath, reqName)
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("read requirement.yaml: %w", err)
	}

	var req Requirement
	if err := yaml.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("parse requirement.yaml: %w", err)
	}

	return &req, nil
}

// ListRequirements returns all requirements, sorted by name.
func ListRequirements(wtPath string) ([]*Requirement, error) {
	reqsDir := RequirementsDir(wtPath)

	if _, err := os.Stat(reqsDir); os.IsNotExist(err) {
		return []*Requirement{}, nil
	}

	entries, err := os.ReadDir(reqsDir)
	if err != nil {
		return nil, fmt.Errorf("read requirements directory: %w", err)
	}

	var reqs []*Requirement
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		req, err := LoadRequirement(wtPath, e.Name())
		if err != nil {
			// Skip broken requirements
			continue
		}

		reqs = append(reqs, req)
	}

	sort.Slice(reqs, func(i, j int) bool {
		return reqs[i].Name < reqs[j].Name
	})

	return reqs, nil
}

// CreateRequirement creates a new requirement with the given name and description.
func CreateRequirement(wtPath, name, description string, subTasks []SubTask) (*Requirement, error) {
	req := &Requirement{
		Name:        name,
		Description: description,
		Status:      RequirementStatusOpen,
		CreatedAt:   time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		SubTasks:    subTasks,
	}

	if err := SaveRequirement(wtPath, req); err != nil {
		return nil, err
	}

	return req, nil
}

// --- RequirementIndex CRUD ---

// LoadRequirementIndex loads the requirement index from index.yaml.
func LoadRequirementIndex(wtPath string) (*RequirementIndex, error) {
	idxPath := RequirementIndexPath(wtPath)
	data, err := os.ReadFile(idxPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &RequirementIndex{}, nil
		}
		return nil, fmt.Errorf("read index.yaml: %w", err)
	}

	var idx RequirementIndex
	if err := yaml.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("parse index.yaml: %w", err)
	}

	return &idx, nil
}

// SaveRequirementIndex saves the requirement index to index.yaml.
func SaveRequirementIndex(wtPath string, idx *RequirementIndex) error {
	idxPath := RequirementIndexPath(wtPath)

	if err := os.MkdirAll(filepath.Dir(idxPath), 0755); err != nil {
		return fmt.Errorf("create requirements directory: %w", err)
	}

	data, err := yaml.Marshal(idx)
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}

	if err := os.WriteFile(idxPath, data, 0644); err != nil {
		return fmt.Errorf("write index.yaml: %w", err)
	}

	return nil
}

// RebuildRequirementIndex rebuilds the index by scanning all requirement directories.
func RebuildRequirementIndex(wtPath string) (*RequirementIndex, error) {
	reqs, err := ListRequirements(wtPath)
	if err != nil {
		return nil, fmt.Errorf("list requirements: %w", err)
	}

	idx := &RequirementIndex{}
	for _, req := range reqs {
		idx.Requirements = append(idx.Requirements, buildIndexEntry(req))
	}

	if err := SaveRequirementIndex(wtPath, idx); err != nil {
		return nil, err
	}

	return idx, nil
}

// UpdateIndexEntry updates a single entry in the index for the given requirement.
// If the requirement is not found in the index, it is appended.
func UpdateIndexEntry(wtPath string, req *Requirement) error {
	idx, err := LoadRequirementIndex(wtPath)
	if err != nil {
		return err
	}

	entry := buildIndexEntry(req)
	found := false
	for i, e := range idx.Requirements {
		if e.Name == req.Name {
			idx.Requirements[i] = entry
			found = true
			break
		}
	}

	if !found {
		idx.Requirements = append(idx.Requirements, entry)
	}

	return SaveRequirementIndex(wtPath, idx)
}

// buildIndexEntry creates an index entry from a requirement.
func buildIndexEntry(req *Requirement) RequirementIndexEntry {
	doneCount := 0
	for _, st := range req.SubTasks {
		if st.Status == SubTaskStatusDone || st.Status == SubTaskStatusSkipped {
			doneCount++
		}
	}

	return RequirementIndexEntry{
		Name:         req.Name,
		Status:       req.Status,
		SubTaskCount: len(req.SubTasks),
		DoneCount:    doneCount,
	}
}
