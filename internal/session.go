// internal/session.go
package internal

import (
	"os"
	"os/exec"
	"strings"
	"time"
)

// SessionBackend abstracts terminal multiplexer operations (Strategy pattern).
type SessionBackend interface {
	PaneAlive(paneID string) bool
	PaneSend(paneID string, text string) error
	SpawnPane(cwd string, newWindow bool) (paneID string, err error)
	KillPane(paneID string) error
	SetTitle(paneID string, title string) error
	ActivatePane(paneID string) error
}

func NewSessionBackend() SessionBackend {
	backend := strings.TrimSpace(strings.ToLower(os.Getenv("AGENT_TEAM_BACKEND")))
	if backend == "tmux" {
		return &TmuxBackend{}
	}
	return &WeztermBackend{}
}

// --- WeztermBackend ---

type WeztermBackend struct{}

func (w *WeztermBackend) PaneAlive(paneID string) bool {
	if paneID == "" {
		return false
	}
	out, err := exec.Command("wezterm", "cli", "list").Output()
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(out), "\n")[1:] {
		parts := strings.Fields(line)
		if len(parts) >= 3 && parts[2] == paneID {
			return true
		}
	}
	return false
}

func (w *WeztermBackend) PaneSend(paneID string, text string) error {
	cmd := exec.Command("wezterm", "cli", "send-text", "--pane-id", paneID, "--no-paste")
	cmd.Stdin = strings.NewReader(text)
	if err := cmd.Run(); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	cmd2 := exec.Command("wezterm", "cli", "send-text", "--pane-id", paneID, "--no-paste")
	cmd2.Stdin = strings.NewReader("\r")
	return cmd2.Run()
}

func (w *WeztermBackend) SpawnPane(cwd string, newWindow bool) (string, error) {
	args := []string{"cli", "spawn", "--cwd", cwd}
	if newWindow {
		args = append(args, "--new-window")
	}
	out, err := exec.Command("wezterm", args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (w *WeztermBackend) KillPane(paneID string) error {
	return exec.Command("wezterm", "cli", "kill-pane", "--pane-id", paneID).Run()
}

func (w *WeztermBackend) SetTitle(paneID string, title string) error {
	return exec.Command("wezterm", "cli", "set-tab-title", "--pane-id", paneID, title).Run()
}

func (w *WeztermBackend) ActivatePane(paneID string) error {
	return exec.Command("wezterm", "cli", "activate-pane", "--pane-id", paneID).Run()
}

// --- TmuxBackend ---

type TmuxBackend struct{}

func (t *TmuxBackend) PaneAlive(paneID string) bool {
	if paneID == "" {
		return false
	}
	out, err := exec.Command("tmux", "list-panes", "-a", "-F", "#{pane_id}").Output()
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == paneID {
			return true
		}
	}
	return false
}

func (t *TmuxBackend) PaneSend(paneID string, text string) error {
	if err := exec.Command("tmux", "send-keys", "-t", paneID, "-l", text).Run(); err != nil {
		return err
	}
	return exec.Command("tmux", "send-keys", "-t", paneID, "Enter").Run()
}

func (t *TmuxBackend) SpawnPane(cwd string, _ bool) (string, error) {
	out, err := exec.Command("tmux", "new-session", "-d", "-P", "-F", "#{pane_id}", "-c", cwd).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (t *TmuxBackend) KillPane(paneID string) error {
	return exec.Command("tmux", "kill-pane", "-t", paneID).Run()
}

func (t *TmuxBackend) SetTitle(paneID string, title string) error {
	return exec.Command("tmux", "rename-window", "-t", paneID, title).Run()
}

func (t *TmuxBackend) ActivatePane(_ string) error {
	return nil // tmux does not steal focus
}
