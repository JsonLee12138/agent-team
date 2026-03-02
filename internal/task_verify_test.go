// internal/task_verify_test.go
package internal

import (
	"testing"
	"time"
)

func TestRunVerifySkipped(t *testing.T) {
	wtPath := t.TempDir()
	if err := InitTasksDir(wtPath); err != nil {
		t.Fatalf("InitTasksDir failed: %v", err)
	}

	change := &Change{
		Name:   "test-change",
		Verify: VerifyConfig{Skip: true},
	}

	result, err := RunVerify(wtPath, change)
	if err != nil {
		t.Fatalf("RunVerify failed: %v", err)
	}

	if !result.Passed {
		t.Error("Skipped verification should be passed")
	}
	if result.ExitCode != 0 {
		t.Errorf("Exit code should be 0, got %d", result.ExitCode)
	}
}

func TestRunVerifyExitZero(t *testing.T) {
	wtPath := t.TempDir()
	if err := InitTasksDir(wtPath); err != nil {
		t.Fatalf("InitTasksDir failed: %v", err)
	}

	change := &Change{
		Name: "test-change",
		Verify: VerifyConfig{
			Command: "exit 0",
			Timeout: "5s",
		},
	}

	result, err := RunVerify(wtPath, change)
	if err != nil {
		t.Fatalf("RunVerify failed: %v", err)
	}

	if !result.Passed {
		t.Error("Exit 0 should be passed")
	}
	if result.ExitCode != 0 {
		t.Errorf("Exit code should be 0, got %d", result.ExitCode)
	}
}

func TestRunVerifyExitOne(t *testing.T) {
	wtPath := t.TempDir()
	if err := InitTasksDir(wtPath); err != nil {
		t.Fatalf("InitTasksDir failed: %v", err)
	}

	change := &Change{
		Name: "test-change",
		Verify: VerifyConfig{
			Command: "exit 1",
			Timeout: "5s",
		},
	}

	result, err := RunVerify(wtPath, change)
	if err != nil {
		// Error is expected for non-zero exit
		if result == nil {
			t.Fatalf("RunVerify should return result even on error")
		}
	}

	if result.Passed {
		t.Error("Exit 1 should not be passed")
	}
	if result.ExitCode != 1 {
		t.Errorf("Exit code should be 1, got %d", result.ExitCode)
	}
}

func TestRunVerifyTimeout(t *testing.T) {
	wtPath := t.TempDir()
	if err := InitTasksDir(wtPath); err != nil {
		t.Fatalf("InitTasksDir failed: %v", err)
	}

	change := &Change{
		Name: "test-change",
		Verify: VerifyConfig{
			Command: "sleep 999",
			Timeout: "50ms",
		},
	}

	result, err := RunVerify(wtPath, change)
	if err != nil {
		// Timeout error is expected
		if result == nil {
			t.Fatalf("RunVerify should return result even on timeout")
		}
	}

	if result.Passed {
		t.Error("Timeout should not be passed")
	}
	if result.ExitCode != 124 {
		t.Errorf("Timeout exit code should be 124, got %d", result.ExitCode)
	}
}

func TestRunVerifyNoCommand(t *testing.T) {
	wtPath := t.TempDir()
	if err := InitTasksDir(wtPath); err != nil {
		t.Fatalf("InitTasksDir failed: %v", err)
	}

	change := &Change{
		Name:   "test-change",
		Verify: VerifyConfig{}, // empty command
	}

	result, err := RunVerify(wtPath, change)
	if err != nil {
		t.Fatalf("RunVerify failed: %v", err)
	}

	if !result.Passed {
		t.Error("No command should be passed")
	}
}

func TestResolveVerifyCommandChangeLevel(t *testing.T) {
	wtPath := t.TempDir()
	if err := InitTasksDir(wtPath); err != nil {
		t.Fatalf("InitTasksDir failed: %v", err)
	}

	cv := VerifyConfig{
		Command: "go test ./...",
		Timeout: "10s",
	}

	cmd, timeout, err := resolveVerifyCommand(wtPath, cv)
	if err != nil {
		t.Fatalf("resolveVerifyCommand failed: %v", err)
	}

	if cmd != "go test ./..." {
		t.Errorf("Command mismatch: %s", cmd)
	}
	if timeout != 10*time.Second {
		t.Errorf("Timeout mismatch: %v", timeout)
	}
}

func TestResolveVerifyCommandDefault(t *testing.T) {
	wtPath := t.TempDir()

	// Create config with default verify command
	if err := InitTasksDir(wtPath); err != nil {
		t.Fatalf("InitTasksDir failed: %v", err)
	}

	cfg, err := LoadTaskConfig(wtPath)
	if err != nil {
		t.Fatalf("LoadTaskConfig failed: %v", err)
	}

	cfg.Defaults.Verify.Command = "make test"
	cfg.Defaults.Verify.Timeout = "3m"
	if err := SaveTaskConfig(wtPath, cfg); err != nil {
		t.Fatalf("SaveTaskConfig failed: %v", err)
	}

	// Empty change-level verify config should use default
	cv := VerifyConfig{}
	cmd, timeout, err := resolveVerifyCommand(wtPath, cv)
	if err != nil {
		t.Fatalf("resolveVerifyCommand failed: %v", err)
	}

	if cmd != "make test" {
		t.Errorf("Command should be from config: %s", cmd)
	}
	if timeout != 3*time.Minute {
		t.Errorf("Timeout should be from config: %v", timeout)
	}
}

func TestVerifyResultDuration(t *testing.T) {
	wtPath := t.TempDir()
	if err := InitTasksDir(wtPath); err != nil {
		t.Fatalf("InitTasksDir failed: %v", err)
	}

	change := &Change{
		Name: "test-change",
		Verify: VerifyConfig{
			Command: "sleep 0.1",
			Timeout: "1s",
		},
	}

	result, err := RunVerify(wtPath, change)
	if err != nil {
		t.Fatalf("RunVerify failed: %v", err)
	}

	if result.Duration < 100*time.Millisecond {
		t.Errorf("Duration should be >= 100ms, got %v", result.Duration)
	}
}

func TestRunVerifyOutputCapture(t *testing.T) {
	wtPath := t.TempDir()
	if err := InitTasksDir(wtPath); err != nil {
		t.Fatalf("InitTasksDir failed: %v", err)
	}

	change := &Change{
		Name: "test-change",
		Verify: VerifyConfig{
			Command: "echo 'test output'",
			Timeout: "5s",
		},
	}

	result, err := RunVerify(wtPath, change)
	if err != nil {
		t.Fatalf("RunVerify failed: %v", err)
	}

	if result.Passed && result.Output == "" {
		t.Error("Output should be captured")
	}
}
