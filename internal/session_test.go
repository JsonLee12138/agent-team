// internal/session_test.go
package internal

import (
	"testing"
)

func TestNewSessionBackendDefault(t *testing.T) {
	t.Setenv("AGENT_TEAM_BACKEND", "")
	b := NewSessionBackend()
	if _, ok := b.(*WeztermBackend); !ok {
		t.Errorf("expected WeztermBackend, got %T", b)
	}
}

func TestNewSessionBackendTmux(t *testing.T) {
	t.Setenv("AGENT_TEAM_BACKEND", "tmux")
	b := NewSessionBackend()
	if _, ok := b.(*TmuxBackend); !ok {
		t.Errorf("expected TmuxBackend, got %T", b)
	}
}

func TestNewSessionBackendCaseInsensitive(t *testing.T) {
	t.Setenv("AGENT_TEAM_BACKEND", "  TMUX  ")
	b := NewSessionBackend()
	if _, ok := b.(*TmuxBackend); !ok {
		t.Errorf("expected TmuxBackend for 'TMUX', got %T", b)
	}
}

func TestPaneAliveEmptyID(t *testing.T) {
	t.Setenv("AGENT_TEAM_BACKEND", "")
	b := NewSessionBackend()
	if b.PaneAlive("") {
		t.Error("empty pane ID should not be alive")
	}
}
