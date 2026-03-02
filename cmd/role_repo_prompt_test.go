package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestPromptSelectNames_NonInteractiveDefaultsAll(t *testing.T) {
	var out bytes.Buffer
	got, err := promptSelectNames(strings.NewReader(""), &out, "Select", []string{"a", "b"})
	if err != nil {
		t.Fatalf("promptSelectNames: %v", err)
	}
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("unexpected selection: %+v", got)
	}
}

func TestPromptConfirm_NonInteractiveDefaultsNo(t *testing.T) {
	var out bytes.Buffer
	ok, err := promptConfirm(strings.NewReader(""), &out, "Proceed?")
	if err != nil {
		t.Fatalf("promptConfirm: %v", err)
	}
	if ok {
		t.Fatal("expected false in non-interactive mode")
	}
}

func TestPromptSingleChoice_NonInteractiveDefault(t *testing.T) {
	var out bytes.Buffer
	got, err := promptSingleChoice(strings.NewReader(""), &out, "Select", []string{"A", "B"}, "B")
	if err != nil {
		t.Fatalf("promptSingleChoice: %v", err)
	}
	if got != "B" {
		t.Fatalf("selected=%q, want B", got)
	}
}
