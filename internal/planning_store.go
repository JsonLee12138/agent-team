package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func PlanningRootDir(root string) string {
	return filepath.Join(AgentTeamDir(root), "planning")
}

func PlanningArchiveRootDir(root string) string {
	return filepath.Join(AgentTeamDir(root), "archive")
}

func PlanningDeprecatedRootDir(root string) string {
	return filepath.Join(AgentTeamDir(root), "deprecated")
}

func PlanningKindRootDir(root string, kind PlanningKind, lifecycle PlanningLifecycle) string {
	base := PlanningRootDir(root)
	switch lifecycle {
	case PlanningLifecycleArchived:
		base = PlanningArchiveRootDir(root)
	case PlanningLifecycleDeprecated:
		base = PlanningDeprecatedRootDir(root)
	}
	return filepath.Join(base, planningDirName(kind))
}

func PlanningDir(root string, kind PlanningKind, id string, lifecycle PlanningLifecycle) string {
	return filepath.Join(PlanningKindRootDir(root, kind, lifecycle), id)
}

func PlanningYAMLPath(root string, kind PlanningKind, id string, lifecycle PlanningLifecycle) string {
	return filepath.Join(PlanningDir(root, kind, id, lifecycle), planningFileName(kind))
}

func PlanningRelPath(kind PlanningKind, id string, lifecycle PlanningLifecycle) string {
	segments := []string{".agent-team"}
	switch lifecycle {
	case PlanningLifecycleArchived:
		segments = append(segments, "archive")
	case PlanningLifecycleDeprecated:
		segments = append(segments, "deprecated")
	default:
		segments = append(segments, "planning")
	}
	segments = append(segments, planningDirName(kind), id)
	return filepath.ToSlash(filepath.Join(segments...))
}

func GeneratePlanningID(kind PlanningKind, title string, now time.Time) string {
	return fmt.Sprintf("%s-%s", string(kind), GenerateTaskID(title, now))
}

func EnsurePlanningDirs(root string) error {
	for _, lifecycle := range []PlanningLifecycle{PlanningLifecycleActive, PlanningLifecycleArchived, PlanningLifecycleDeprecated} {
		for _, kind := range []PlanningKind{PlanningKindRoadmap, PlanningKindMilestone, PlanningKindPhase} {
			if err := os.MkdirAll(PlanningKindRootDir(root, kind, lifecycle), 0755); err != nil {
				return fmt.Errorf("create %s %s directory: %w", lifecycle, kind, err)
			}
		}
	}
	return nil
}

func CreatePlanningRecord(root string, kind PlanningKind, title string, now time.Time) (*PlanningRecord, error) {
	if err := EnsurePlanningDirs(root); err != nil {
		return nil, err
	}
	id := GeneratePlanningID(kind, title, now)
	record := &PlanningRecord{
		ID:        id,
		Kind:      kind,
		Title:     title,
		Status:    "proposed",
		Lifecycle: PlanningLifecycleActive,
		Path:      PlanningRelPath(kind, id, PlanningLifecycleActive),
		CreatedAt: now.UTC().Format(time.RFC3339),
		UpdatedAt: now.UTC().Format(time.RFC3339),
	}
	if err := savePlanningRecord(root, record); err != nil {
		return nil, err
	}
	return record, nil
}

func LoadPlanningRecord(root, id string) (*PlanningRecord, error) {
	for _, lifecycle := range []PlanningLifecycle{PlanningLifecycleActive, PlanningLifecycleArchived, PlanningLifecycleDeprecated} {
		for _, kind := range []PlanningKind{PlanningKindRoadmap, PlanningKindMilestone, PlanningKindPhase} {
			path := PlanningYAMLPath(root, kind, id, lifecycle)
			data, err := os.ReadFile(path)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return nil, fmt.Errorf("read planning record: %w", err)
			}
			var record PlanningRecord
			if err := yaml.Unmarshal(data, &record); err != nil {
				return nil, fmt.Errorf("parse planning record: %w", err)
			}
			if !ValidPlanningKind(record.Kind) {
				return nil, fmt.Errorf("invalid planning kind: %s", record.Kind)
			}
			if !ValidPlanningLifecycle(record.Lifecycle) {
				return nil, fmt.Errorf("invalid planning lifecycle: %s", record.Lifecycle)
			}
			return &record, nil
		}
	}
	return nil, fmt.Errorf("planning '%s' not found", id)
}

func ListPlanningRecords(root string, kind PlanningKind, lifecycle PlanningLifecycle) ([]*PlanningRecord, error) {
	var kinds []PlanningKind
	if kind != "" {
		kinds = []PlanningKind{kind}
	} else {
		kinds = []PlanningKind{PlanningKindRoadmap, PlanningKindMilestone, PlanningKindPhase}
	}

	var lifecycles []PlanningLifecycle
	if lifecycle != "" {
		lifecycles = []PlanningLifecycle{lifecycle}
	} else {
		lifecycles = []PlanningLifecycle{PlanningLifecycleActive}
	}

	var records []*PlanningRecord
	for _, lc := range lifecycles {
		for _, k := range kinds {
			base := PlanningKindRootDir(root, k, lc)
			entries, err := os.ReadDir(base)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return nil, fmt.Errorf("read planning directory %s: %w", base, err)
			}
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				record, err := LoadPlanningRecord(root, entry.Name())
				if err != nil {
					continue
				}
				records = append(records, record)
			}
		}
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].ID < records[j].ID
	})
	return records, nil
}

func MovePlanningRecord(root, id string, to PlanningLifecycle, reason string, now time.Time) (*PlanningRecord, error) {
	record, err := LoadPlanningRecord(root, id)
	if err != nil {
		return nil, err
	}
	if record.Lifecycle == to {
		return nil, fmt.Errorf("planning '%s' is already %s", id, to)
	}

	from := record.Lifecycle
	record.Lifecycle = to
	record.Path = PlanningRelPath(record.Kind, record.ID, to)
	record.UpdatedAt = now.UTC().Format(time.RFC3339)
	switch to {
	case PlanningLifecycleArchived:
		record.ArchivedAt = now.UTC().Format(time.RFC3339)
		record.DeprecatedAt = ""
		record.DeprecatedReason = ""
	case PlanningLifecycleDeprecated:
		record.DeprecatedAt = now.UTC().Format(time.RFC3339)
		record.DeprecatedReason = strings.TrimSpace(reason)
		record.ArchivedAt = ""
	default:
		record.ArchivedAt = ""
		record.DeprecatedAt = ""
		record.DeprecatedReason = ""
	}

	srcDir := PlanningDir(root, record.Kind, record.ID, from)
	dstDir := PlanningDir(root, record.Kind, record.ID, to)
	if err := os.MkdirAll(filepath.Dir(dstDir), 0755); err != nil {
		return nil, fmt.Errorf("create destination parent directory: %w", err)
	}
	if _, err := os.Stat(dstDir); err == nil {
		return nil, fmt.Errorf("destination already exists: %s", dstDir)
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("check destination: %w", err)
	}
	if err := os.Rename(srcDir, dstDir); err != nil {
		return nil, fmt.Errorf("move planning record: %w", err)
	}
	if err := savePlanningRecord(root, record); err != nil {
		if rollbackErr := os.Rename(dstDir, srcDir); rollbackErr != nil {
			return nil, fmt.Errorf("write moved planning metadata: %w (rollback failed: %v)", err, rollbackErr)
		}
		return nil, fmt.Errorf("write moved planning metadata: %w", err)
	}
	return record, nil
}

func savePlanningRecord(root string, record *PlanningRecord) error {
	dir := PlanningDir(root, record.Kind, record.ID, record.Lifecycle)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create planning directory: %w", err)
	}
	data, err := yaml.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal planning record: %w", err)
	}
	if err := os.WriteFile(PlanningYAMLPath(root, record.Kind, record.ID, record.Lifecycle), data, 0644); err != nil {
		return fmt.Errorf("write planning record: %w", err)
	}
	return nil
}
