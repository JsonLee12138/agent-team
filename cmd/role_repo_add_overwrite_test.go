package cmd

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestDecideRoleOverwriteOnConflict_OverwriteFlagWins(t *testing.T) {
	called := false
	got, err := decideRoleOverwriteOnConflict(
		strings.NewReader(""),
		&bytes.Buffer{},
		"frontend",
		true,
		&roleOverwritePolicy{},
		func(_ io.Reader) bool { return true },
		func(_ io.Reader, _ io.Writer, _ string, _ []string, _ string) (string, error) {
			called = true
			return "No", nil
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Fatal("expected overwrite=true when -y is set")
	}
	if called {
		t.Fatal("chooser should not be called when overwrite flag is set")
	}
}

func TestDecideRoleOverwriteOnConflict_NonInteractiveDefaultsNo(t *testing.T) {
	called := false
	got, err := decideRoleOverwriteOnConflict(
		strings.NewReader(""),
		&bytes.Buffer{},
		"frontend",
		false,
		&roleOverwritePolicy{},
		func(_ io.Reader) bool { return false },
		func(_ io.Reader, _ io.Writer, _ string, _ []string, _ string) (string, error) {
			called = true
			return "Yes", nil
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Fatal("expected overwrite=false for non-interactive mode")
	}
	if called {
		t.Fatal("chooser should not be called in non-interactive mode")
	}
}

func TestDecideRoleOverwriteOnConflict_InteractiveYes(t *testing.T) {
	got, err := decideRoleOverwriteOnConflict(
		strings.NewReader(""),
		&bytes.Buffer{},
		"frontend",
		false,
		&roleOverwritePolicy{},
		func(_ io.Reader) bool { return true },
		func(_ io.Reader, _ io.Writer, _ string, _ []string, _ string) (string, error) { return "Yes", nil },
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Fatal("expected overwrite=true when user chooses Yes")
	}
}

func TestDecideRoleOverwriteOnConflict_InteractivePropagatesError(t *testing.T) {
	wantErr := errors.New("prompt failed")
	_, err := decideRoleOverwriteOnConflict(
		strings.NewReader(""),
		&bytes.Buffer{},
		"frontend",
		false,
		&roleOverwritePolicy{},
		func(_ io.Reader) bool { return true },
		func(_ io.Reader, _ io.Writer, _ string, _ []string, _ string) (string, error) { return "", wantErr },
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestDecideRoleOverwriteOnConflict_AllSelectionAppliesToSubsequentRoles(t *testing.T) {
	policy := &roleOverwritePolicy{}
	calls := 0
	chooser := func(_ io.Reader, _ io.Writer, _ string, _ []string, _ string) (string, error) {
		calls++
		return "All", nil
	}

	first, err := decideRoleOverwriteOnConflict(
		strings.NewReader(""),
		&bytes.Buffer{},
		"frontend",
		false,
		policy,
		func(_ io.Reader) bool { return true },
		chooser,
	)
	if err != nil {
		t.Fatalf("first decision error: %v", err)
	}
	if !first {
		t.Fatal("first decision should overwrite after choosing All")
	}

	second, err := decideRoleOverwriteOnConflict(
		strings.NewReader(""),
		&bytes.Buffer{},
		"backend",
		false,
		policy,
		func(_ io.Reader) bool { return true },
		chooser,
	)
	if err != nil {
		t.Fatalf("second decision error: %v", err)
	}
	if !second {
		t.Fatal("second decision should inherit overwrite=true from All")
	}
	if calls != 1 {
		t.Fatalf("chooser called %d times, want 1", calls)
	}
}

func TestDecideRoleOverwriteOnConflict_NoneSelectionAppliesToSubsequentRoles(t *testing.T) {
	policy := &roleOverwritePolicy{}
	calls := 0
	chooser := func(_ io.Reader, _ io.Writer, _ string, _ []string, _ string) (string, error) {
		calls++
		return "None", nil
	}

	first, err := decideRoleOverwriteOnConflict(
		strings.NewReader(""),
		&bytes.Buffer{},
		"frontend",
		false,
		policy,
		func(_ io.Reader) bool { return true },
		chooser,
	)
	if err != nil {
		t.Fatalf("first decision error: %v", err)
	}
	if first {
		t.Fatal("first decision should skip after choosing None")
	}

	second, err := decideRoleOverwriteOnConflict(
		strings.NewReader(""),
		&bytes.Buffer{},
		"backend",
		false,
		policy,
		func(_ io.Reader) bool { return true },
		chooser,
	)
	if err != nil {
		t.Fatalf("second decision error: %v", err)
	}
	if second {
		t.Fatal("second decision should inherit overwrite=false from None")
	}
	if calls != 1 {
		t.Fatalf("chooser called %d times, want 1", calls)
	}
}
