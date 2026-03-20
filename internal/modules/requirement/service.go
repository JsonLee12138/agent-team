package requirement

import (
	"github.com/JsonLee12138/agent-team/internal"
	"github.com/JsonLee12138/agent-team/internal/governance"
)

type Service struct {
	Root string
}

func NewService(root string) Service {
	return Service{Root: root}
}

func (s Service) LoadGovernanceIndex() (governance.Index, error) {
	idx, err := internal.LoadRequirementIndex(s.Root)
	if err != nil {
		return governance.Index{}, err
	}
	entries := make([]governance.IndexEntry, 0, len(idx.Requirements))
	for _, requirement := range idx.Requirements {
		entries = append(entries, governance.IndexEntry{
			ID:       requirement.Name,
			Kind:     "requirement",
			Path:     internal.RequirementYAMLPath(s.Root, requirement.Name),
			Archived: requirement.Status == internal.RequirementStatusDone,
		})
	}
	return governance.Index{Entries: entries}, nil
}

func (s Service) EnsureIndexEntry(entry governance.IndexEntry) error {
	idx, err := internal.LoadRequirementIndex(s.Root)
	if err != nil {
		return err
	}

	status := internal.RequirementStatusOpen
	if entry.Archived {
		status = internal.RequirementStatusDone
	}

	found := false
	for i := range idx.Requirements {
		if idx.Requirements[i].Name == entry.ID {
			idx.Requirements[i].Status = status
			found = true
			break
		}
	}

	if !found {
		idx.Requirements = append(idx.Requirements, internal.RequirementIndexEntry{
			Name:         entry.ID,
			Status:       status,
			SubTaskCount: 0,
			DoneCount:    0,
		})
	}

	return internal.SaveRequirementIndex(s.Root, idx)
}
