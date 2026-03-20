package task

import "github.com/JsonLee12138/agent-team/internal"

type Service struct {
	Root string
}

func NewService(root string) Service {
	return Service{Root: root}
}

func (s Service) LoadChange(changeName string) (*internal.Change, error) {
	return internal.LoadChange(s.Root, changeName)
}

func (s Service) SaveChange(change *internal.Change) error {
	return internal.SaveChange(s.Root, change)
}

func (s Service) LoadConfig() (*internal.TaskConfig, error) {
	return internal.LoadTaskConfig(s.Root)
}
