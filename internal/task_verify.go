// internal/task_verify.go
package internal

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

// VerifyResult holds the outcome of running verification.
type VerifyResult struct {
	Passed   bool
	Output   string
	ExitCode int
	Duration time.Duration
	Error    error
}

// RunVerify runs the verification command for a change.
// If Skip=true in config, directly returns Passed=true.
// If change.Verify.Command is empty, uses the default from config.yaml.
func RunVerify(wtPath string, change *Change) (*VerifyResult, error) {
	startTime := time.Now()

	// Resolve verify command and timeout
	cmd, timeout, err := resolveVerifyCommand(wtPath, change.Verify)
	if err != nil {
		return nil, err
	}

	// Check if verification should be skipped
	if change.Verify.Skip {
		return &VerifyResult{
			Passed:   true,
			Output:   "(skipped)",
			ExitCode: 0,
			Duration: time.Since(startTime),
		}, nil
	}

	// If no command, return passed (nothing to verify)
	if cmd == "" {
		return &VerifyResult{
			Passed:   true,
			Output:   "(no command configured)",
			ExitCode: 0,
			Duration: time.Since(startTime),
		}, nil
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Execute bash command
	bashCmd := exec.CommandContext(ctx, "bash", "-c", cmd)
	bashCmd.Dir = wtPath

	// Capture output
	var stdout, stderr bytes.Buffer
	bashCmd.Stdout = &stdout
	bashCmd.Stderr = &stderr

	// Run command
	err = bashCmd.Run()
	exitCode := 0
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			exitCode = 124 // timeout exit code
		} else if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	// Combine output
	output := stdout.String()
	if stderr.Len() > 0 {
		output += stderr.String()
	}

	return &VerifyResult{
		Passed:   exitCode == 0,
		Output:   output,
		ExitCode: exitCode,
		Duration: time.Since(startTime),
		Error:    err,
	}, nil
}

// resolveVerifyCommand determines which verify command and timeout to use.
// Priority: change.Verify.Command → config.yaml default → empty
func resolveVerifyCommand(wtPath string, cv VerifyConfig) (string, time.Duration, error) {
	// Use change-level command if provided
	if cv.Command != "" {
		timeout, err := cv.ParseTimeout()
		if err != nil {
			return "", 0, fmt.Errorf("parse timeout: %w", err)
		}
		return cv.Command, timeout, nil
	}

	// Fall back to config default
	cfg, err := LoadTaskConfig(wtPath)
	if err != nil {
		// If config doesn't exist, use defaults
		cfg = &TaskConfig{
			Defaults: TaskConfigDefaults{
				Verify: VerifyConfig{
					Timeout: "5m",
				},
			},
		}
	}

	timeout, err := cfg.Defaults.Verify.ParseTimeout()
	if err != nil {
		return "", 0, fmt.Errorf("parse default timeout: %w", err)
	}

	return cfg.Defaults.Verify.Command, timeout, nil
}
