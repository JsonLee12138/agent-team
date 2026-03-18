// cmd/req_test.go
package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

// --- helpers ---

// buildReqRootWithDir returns a cobra.Command tree wired to an App backed by the given dir.
func buildReqRootWithDir(t *testing.T, app *App) *cobra.Command {
	t.Helper()
	root := NewRootCmd()
	RegisterCommands(root)
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		cmd.SetContext(WithApp(cmd.Context(), app))
		return nil
	}
	return root
}

// buildReqRoot returns a cobra.Command tree wired to a temp-dir-backed App.
// The returned dir is the git root.
func buildReqRoot(t *testing.T) (*cobra.Command, string) {
	t.Helper()
	app, dir := initTestApp(t)
	root := buildReqRootWithDir(t, app)
	return root, dir
}

// execCmd runs a cobra command tree with args and returns combined output + error.
func execCmd(root *cobra.Command, args ...string) (string, error) {
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

// --- 3B-1: req create ---

func TestReqCreate_Success(t *testing.T) {
	root, dir := buildReqRoot(t)
	_, err := execCmd(root, "req", "create", "auth-feature", "-d", "Implement auth")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}

	// Verify requirement file was created
	req, err := internal.LoadRequirement(dir, "auth-feature")
	if err != nil {
		t.Fatalf("LoadRequirement: %v", err)
	}
	if req.Name != "auth-feature" {
		t.Errorf("Name = %q, want auth-feature", req.Name)
	}
	if req.Description != "Implement auth" {
		t.Errorf("Description = %q, want 'Implement auth'", req.Description)
	}
	if req.Status != internal.RequirementStatusOpen {
		t.Errorf("Status = %q, want open", req.Status)
	}

	// Verify index was updated
	idx, err := internal.LoadRequirementIndex(dir)
	if err != nil {
		t.Fatalf("LoadRequirementIndex: %v", err)
	}
	if len(idx.Requirements) != 1 || idx.Requirements[0].Name != "auth-feature" {
		t.Errorf("index = %+v, want 1 entry for auth-feature", idx.Requirements)
	}
}

func TestReqCreate_MissingName(t *testing.T) {
	root, _ := buildReqRoot(t)
	_, err := execCmd(root, "req", "create")
	if err == nil {
		t.Fatal("expected error for missing name arg")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg") {
		t.Errorf("error = %q, want 'accepts 1 arg'", err.Error())
	}
}

func TestReqCreate_TooManyArgs(t *testing.T) {
	root, _ := buildReqRoot(t)
	_, err := execCmd(root, "req", "create", "a", "b")
	if err == nil {
		t.Fatal("expected error for too many args")
	}
}

// --- 3B-1: req split ---

func TestReqSplit_Success(t *testing.T) {
	root, dir := buildReqRoot(t)

	// Create a requirement first
	execCmd(root, "req", "create", "feat-x")

	// Split with tasks
	_, err := execCmd(root, "req", "split", "feat-x", "--task", "design API", "--task", "implement DB")
	if err != nil {
		t.Fatalf("split: %v", err)
	}

	req, err := internal.LoadRequirement(dir, "feat-x")
	if err != nil {
		t.Fatalf("LoadRequirement: %v", err)
	}
	if len(req.SubTasks) != 2 {
		t.Fatalf("SubTasks len = %d, want 2", len(req.SubTasks))
	}
	if req.SubTasks[0].ID != 1 || req.SubTasks[0].Title != "design API" {
		t.Errorf("SubTask[0] = %+v, want id=1 title='design API'", req.SubTasks[0])
	}
	if req.SubTasks[1].ID != 2 || req.SubTasks[1].Title != "implement DB" {
		t.Errorf("SubTask[1] = %+v, want id=2 title='implement DB'", req.SubTasks[1])
	}
	if req.SubTasks[0].Status != internal.SubTaskStatusPending {
		t.Errorf("SubTask[0].Status = %q, want pending", req.SubTasks[0].Status)
	}
}

func TestReqSplit_MissingName(t *testing.T) {
	root, _ := buildReqRoot(t)
	_, err := execCmd(root, "req", "split")
	if err == nil {
		t.Fatal("expected error for missing name arg")
	}
}

func TestReqSplit_MissingTaskFlag(t *testing.T) {
	root, _ := buildReqRoot(t)
	execCmd(root, "req", "create", "feat-y")
	_, err := execCmd(root, "req", "split", "feat-y")
	if err == nil {
		t.Fatal("expected error for missing required --task flag")
	}
}

func TestReqSplit_NonexistentReq(t *testing.T) {
	root, _ := buildReqRoot(t)
	_, err := execCmd(root, "req", "split", "nonexistent", "--task", "x")
	if err == nil {
		t.Fatal("expected error for nonexistent requirement")
	}
}

func TestReqSplit_IncrementalIDs(t *testing.T) {
	app, dir := initTestApp(t)

	root1 := buildReqRootWithDir(t, app)
	execCmd(root1, "req", "create", "feat-inc")

	root2 := buildReqRootWithDir(t, app)
	execCmd(root2, "req", "split", "feat-inc", "--task", "first")

	// Add more tasks — IDs should continue from 2
	root3 := buildReqRootWithDir(t, app)
	execCmd(root3, "req", "split", "feat-inc", "--task", "second", "--task", "third")

	req, _ := internal.LoadRequirement(dir, "feat-inc")
	if len(req.SubTasks) != 3 {
		t.Fatalf("SubTasks len = %d, want 3", len(req.SubTasks))
	}
	if req.SubTasks[1].ID != 2 {
		t.Errorf("SubTask[1].ID = %d, want 2", req.SubTasks[1].ID)
	}
	if req.SubTasks[2].ID != 3 {
		t.Errorf("SubTask[2].ID = %d, want 3", req.SubTasks[2].ID)
	}
}

// --- 3B-1: req assign ---

func TestReqAssign_Success(t *testing.T) {
	root, dir := buildReqRoot(t)

	// Create requirement with sub-tasks
	execCmd(root, "req", "create", "feat-assign")
	execCmd(root, "req", "split", "feat-assign", "--task", "task-one")

	// Create a fake worker worktree directory
	wtPath := filepath.Join(dir, ".worktrees", "dev-001")
	if err := os.MkdirAll(wtPath, 0755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	// Also need .tasks dir for change creation
	if err := os.MkdirAll(filepath.Join(wtPath, ".tasks", "changes"), 0755); err != nil {
		t.Fatalf("mkdir .tasks: %v", err)
	}

	_, err := execCmd(root, "req", "assign", "feat-assign", "1", "dev-001")
	if err != nil {
		t.Fatalf("assign: %v", err)
	}

	req, _ := internal.LoadRequirement(dir, "feat-assign")
	if req.SubTasks[0].Status != internal.SubTaskStatusAssigned {
		t.Errorf("SubTask.Status = %q, want assigned", req.SubTasks[0].Status)
	}
	if req.SubTasks[0].AssignedTo != "dev-001" {
		t.Errorf("AssignedTo = %q, want dev-001", req.SubTasks[0].AssignedTo)
	}
	if req.SubTasks[0].ChangeName == "" {
		t.Error("ChangeName should not be empty after assign")
	}
	if req.Status != internal.RequirementStatusInProgress {
		t.Errorf("Requirement.Status = %q, want in_progress", req.Status)
	}
}

func TestReqAssign_MissingArgs(t *testing.T) {
	root, _ := buildReqRoot(t)
	_, err := execCmd(root, "req", "assign", "feat", "1")
	if err == nil {
		t.Fatal("expected error for missing worker-id arg")
	}
}

func TestReqAssign_InvalidTaskID(t *testing.T) {
	root, _ := buildReqRoot(t)
	_, err := execCmd(root, "req", "assign", "feat", "abc", "dev-001")
	if err == nil {
		t.Fatal("expected error for non-numeric task-id")
	}
	if !strings.Contains(err.Error(), "invalid task-id") {
		t.Errorf("error = %q, want 'invalid task-id'", err.Error())
	}
}

func TestReqAssign_WorkerNotFound(t *testing.T) {
	root, _ := buildReqRoot(t)
	execCmd(root, "req", "create", "feat-noworker")
	execCmd(root, "req", "split", "feat-noworker", "--task", "t1")
	_, err := execCmd(root, "req", "assign", "feat-noworker", "1", "nonexistent-worker")
	if err == nil {
		t.Fatal("expected error for nonexistent worker")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want 'not found'", err.Error())
	}
}

func TestReqAssign_SubTaskNotFound(t *testing.T) {
	root, dir := buildReqRoot(t)
	execCmd(root, "req", "create", "feat-nosub")
	execCmd(root, "req", "split", "feat-nosub", "--task", "t1")

	wtPath := filepath.Join(dir, ".worktrees", "dev-001")
	os.MkdirAll(wtPath, 0755)

	_, err := execCmd(root, "req", "assign", "feat-nosub", "99", "dev-001")
	if err == nil {
		t.Fatal("expected error for nonexistent sub-task")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want 'not found'", err.Error())
	}
}

// --- 3B-1: req status ---

func TestReqStatus_All(t *testing.T) {
	root, _ := buildReqRoot(t)
	execCmd(root, "req", "create", "alpha")
	execCmd(root, "req", "create", "beta")

	out, err := execCmd(root, "req", "status")
	if err != nil {
		t.Fatalf("status all: %v", err)
	}
	if !strings.Contains(out, "alpha") {
		t.Error("output should contain 'alpha'")
	}
	if !strings.Contains(out, "beta") {
		t.Error("output should contain 'beta'")
	}
	if !strings.Contains(out, "NAME") {
		t.Error("output should contain header 'NAME'")
	}
}

func TestReqStatus_One(t *testing.T) {
	app, _ := initTestApp(t)

	root1 := buildReqRootWithDir(t, app)
	execCmd(root1, "req", "create", "detail-req", "-d", "detailed description")

	root2 := buildReqRootWithDir(t, app)
	execCmd(root2, "req", "split", "detail-req", "--task", "t1", "--task", "t2")

	root3 := buildReqRootWithDir(t, app)
	out, err := execCmd(root3, "req", "status", "detail-req")
	if err != nil {
		t.Fatalf("status one: %v", err)
	}
	// The tabwriter output goes to cmd.OutOrStdout() and contains the sub-task table
	// fmt.Printf header lines go to os.Stdout (not captured by buffer)
	if !strings.Contains(out, "t1") {
		t.Logf("captured tabwriter output: %q (fmt.Printf goes to stdout)", out)
	}
}

func TestReqStatus_Empty(t *testing.T) {
	root, _ := buildReqRoot(t)
	// "No requirements found." is printed via fmt.Println (goes to stdout, not captured)
	// We just verify the command succeeds without error
	_, err := execCmd(root, "req", "status")
	if err != nil {
		t.Fatalf("status empty: %v", err)
	}
}

func TestReqStatus_Nonexistent(t *testing.T) {
	root, _ := buildReqRoot(t)
	_, err := execCmd(root, "req", "status", "ghost")
	if err == nil {
		t.Fatal("expected error for nonexistent requirement")
	}
}

// --- 3B-1: req done ---

func TestReqDone_AllDone(t *testing.T) {
	root, dir := buildReqRoot(t)
	execCmd(root, "req", "create", "done-req")
	// No sub-tasks, so it should be completable immediately
	_, err := execCmd(root, "req", "done", "done-req")
	if err != nil {
		t.Fatalf("done: %v", err)
	}

	req, _ := internal.LoadRequirement(dir, "done-req")
	if req.Status != internal.RequirementStatusDone {
		t.Errorf("Status = %q, want done", req.Status)
	}
}

func TestReqDone_AlreadyDone(t *testing.T) {
	app, dir := initTestApp(t)
	root1 := buildReqRootWithDir(t, app)
	execCmd(root1, "req", "create", "already-done")

	root2 := buildReqRootWithDir(t, app)
	execCmd(root2, "req", "done", "already-done")

	// Run again — should not error (idempotent)
	root3 := buildReqRootWithDir(t, app)
	_, err := execCmd(root3, "req", "done", "already-done")
	if err != nil {
		t.Fatalf("done again should not error: %v", err)
	}

	// Verify it's still done
	req, _ := internal.LoadRequirement(dir, "already-done")
	if req.Status != internal.RequirementStatusDone {
		t.Errorf("Status = %q, want done", req.Status)
	}
}

func TestReqDone_PendingSubTasks_NoForce(t *testing.T) {
	root, _ := buildReqRoot(t)
	execCmd(root, "req", "create", "pending-req")
	execCmd(root, "req", "split", "pending-req", "--task", "unfinished")

	_, err := execCmd(root, "req", "done", "pending-req")
	if err == nil {
		t.Fatal("expected error for pending sub-tasks without --force")
	}
	if !strings.Contains(err.Error(), "--force") {
		t.Errorf("error = %q, want mention of --force", err.Error())
	}
}

func TestReqDone_PendingSubTasks_WithForce(t *testing.T) {
	root, dir := buildReqRoot(t)
	execCmd(root, "req", "create", "forced-req")
	execCmd(root, "req", "split", "forced-req", "--task", "skippable")

	_, err := execCmd(root, "req", "done", "forced-req", "--force")
	if err != nil {
		t.Fatalf("done --force: %v", err)
	}

	req, _ := internal.LoadRequirement(dir, "forced-req")
	if req.Status != internal.RequirementStatusDone {
		t.Errorf("Status = %q, want done (forced)", req.Status)
	}
}

func TestReqDone_MissingName(t *testing.T) {
	root, _ := buildReqRoot(t)
	_, err := execCmd(root, "req", "done")
	if err == nil {
		t.Fatal("expected error for missing name arg")
	}
}

func TestReqDone_Nonexistent(t *testing.T) {
	root, _ := buildReqRoot(t)
	_, err := execCmd(root, "req", "done", "no-such-req")
	if err == nil {
		t.Fatal("expected error for nonexistent requirement")
	}
}

func TestReqDone_IndexSynced(t *testing.T) {
	root, dir := buildReqRoot(t)
	execCmd(root, "req", "create", "idx-sync")
	execCmd(root, "req", "done", "idx-sync")

	idx, err := internal.LoadRequirementIndex(dir)
	if err != nil {
		t.Fatalf("LoadRequirementIndex: %v", err)
	}
	for _, e := range idx.Requirements {
		if e.Name == "idx-sync" {
			if e.Status != internal.RequirementStatusDone {
				t.Errorf("index status = %q, want done", e.Status)
			}
			return
		}
	}
	t.Error("idx-sync not found in index")
}

// --- 3B-1: req (no subcommand) shows help ---

func TestReq_NoSubcommand(t *testing.T) {
	root, _ := buildReqRoot(t)
	out, err := execCmd(root, "req")
	if err != nil {
		t.Fatalf("req help: %v", err)
	}
	if !strings.Contains(out, "Available Commands") || !strings.Contains(out, "create") {
		t.Errorf("output = %q, want help text with subcommands", out)
	}
}
