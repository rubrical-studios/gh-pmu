//go:build integration

package cmd

import (
	"fmt"
	"testing"

	"github.com/rubrical-studios/gh-pmu/internal/testutil"
)

// TestRunMove_Integration_ChangeStatus tests changing --status
func TestRunMove_Integration_ChangeStatus(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create a test issue
	title := fmt.Sprintf("Test Issue - MoveStatus - %d", testUniqueID())
	createResult := testutil.RunCommand(t, "create", "--title", title, "--status", "backlog")
	testutil.AssertExitCode(t, createResult, 0)

	issueNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, issueNum)

	// Move to in_progress
	moveResult := testutil.RunCommand(t, "move", fmt.Sprintf("%d", issueNum), "--status", "in_progress")
	testutil.AssertExitCode(t, moveResult, 0)
	testutil.AssertContains(t, moveResult.Stdout, "Updated issue")
	testutil.AssertContains(t, moveResult.Stdout, "Status")
	testutil.AssertContains(t, moveResult.Stdout, "In progress")

	// Verify change via view
	viewResult := testutil.RunCommand(t, "view", fmt.Sprintf("%d", issueNum), "--json")
	testutil.AssertExitCode(t, viewResult, 0)
	testutil.AssertContains(t, viewResult.Stdout, "In progress")
}

// TestRunMove_Integration_ChangePriority tests changing --priority
func TestRunMove_Integration_ChangePriority(t *testing.T) {
	testutil.RequireTestEnv(t)

	title := fmt.Sprintf("Test Issue - MovePriority - %d", testUniqueID())
	createResult := testutil.RunCommand(t, "create", "--title", title, "--priority", "p2")
	testutil.AssertExitCode(t, createResult, 0)

	issueNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, issueNum)

	// Move to P0
	moveResult := testutil.RunCommand(t, "move", fmt.Sprintf("%d", issueNum), "--priority", "p0")
	testutil.AssertExitCode(t, moveResult, 0)
	testutil.AssertContains(t, moveResult.Stdout, "Priority")
	testutil.AssertContains(t, moveResult.Stdout, "P0")

	// Verify change
	viewResult := testutil.RunCommand(t, "view", fmt.Sprintf("%d", issueNum), "--json")
	testutil.AssertExitCode(t, viewResult, 0)
	testutil.AssertContains(t, viewResult.Stdout, "P0")
}

// TestRunMove_Integration_MultipleFields tests changing multiple fields
func TestRunMove_Integration_MultipleFields(t *testing.T) {
	testutil.RequireTestEnv(t)

	title := fmt.Sprintf("Test Issue - MoveMultiple - %d", testUniqueID())
	createResult := testutil.RunCommand(t, "create", "--title", title, "--status", "backlog", "--priority", "p2")
	testutil.AssertExitCode(t, createResult, 0)

	issueNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, issueNum)

	// Move both status and priority
	moveResult := testutil.RunCommand(t, "move", fmt.Sprintf("%d", issueNum),
		"--status", "in_progress",
		"--priority", "p1",
	)
	testutil.AssertExitCode(t, moveResult, 0)
	testutil.AssertContains(t, moveResult.Stdout, "Status")
	testutil.AssertContains(t, moveResult.Stdout, "Priority")

	// Verify changes
	viewResult := testutil.RunCommand(t, "view", fmt.Sprintf("%d", issueNum), "--json")
	testutil.AssertExitCode(t, viewResult, 0)
	testutil.AssertContains(t, viewResult.Stdout, "In progress")
	testutil.AssertContains(t, viewResult.Stdout, "P1")
}

// TestRunMove_Integration_FieldAliases tests field value aliases
func TestRunMove_Integration_FieldAliases(t *testing.T) {
	testutil.RequireTestEnv(t)

	title := fmt.Sprintf("Test Issue - MoveAliases - %d", testUniqueID())
	createResult := testutil.RunCommand(t, "create", "--title", title)
	testutil.AssertExitCode(t, createResult, 0)

	issueNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, issueNum)

	// Use aliases: in_review -> "In review"
	moveResult := testutil.RunCommand(t, "move", fmt.Sprintf("%d", issueNum), "--status", "in_review")
	testutil.AssertExitCode(t, moveResult, 0)
	testutil.AssertContains(t, moveResult.Stdout, "In review")

	// Verify alias resolved correctly
	viewResult := testutil.RunCommand(t, "view", fmt.Sprintf("%d", issueNum), "--json")
	testutil.AssertExitCode(t, viewResult, 0)
	testutil.AssertContains(t, viewResult.Stdout, "In review")
}

// TestRunMove_Integration_NotInProject tests issue not in project error
func TestRunMove_Integration_NotInProject(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Try to move a non-existent issue number
	result := testutil.RunCommand(t, "move", "99999", "--status", "backlog")

	// Should fail
	if result.ExitCode == 0 {
		t.Error("expected non-zero exit code for non-existent issue")
	}

	// Should show error message
	if result.Stderr == "" && result.Stdout == "" {
		t.Error("expected error message")
	}
}

// TestRunMove_Integration_NoFlags tests error when no flags provided
func TestRunMove_Integration_NoFlags(t *testing.T) {
	testutil.RequireTestEnv(t)

	result := testutil.RunCommand(t, "move", "1")

	// Should fail
	if result.ExitCode == 0 {
		t.Error("expected non-zero exit code when no flags provided")
	}

	testutil.AssertContains(t, result.Stderr, "at least one of --status or --priority is required")
}

// TestRunMove_Integration_DryRun tests --dry-run flag
func TestRunMove_Integration_DryRun(t *testing.T) {
	testutil.RequireTestEnv(t)

	title := fmt.Sprintf("Test Issue - MoveDryRun - %d", testUniqueID())
	createResult := testutil.RunCommand(t, "create", "--title", title, "--status", "backlog")
	testutil.AssertExitCode(t, createResult, 0)

	issueNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, issueNum)

	// Dry run - should not change anything
	moveResult := testutil.RunCommand(t, "move", fmt.Sprintf("%d", issueNum),
		"--status", "done",
		"--dry-run",
	)
	testutil.AssertExitCode(t, moveResult, 0)
	testutil.AssertContains(t, moveResult.Stdout, "Dry run")

	// Verify status is still backlog
	viewResult := testutil.RunCommand(t, "view", fmt.Sprintf("%d", issueNum), "--json")
	testutil.AssertExitCode(t, viewResult, 0)
	testutil.AssertContains(t, viewResult.Stdout, "Backlog")
	testutil.AssertNotContains(t, viewResult.Stdout, "\"Status\": \"Done\"")
}

// TestRunMove_Integration_Recursive tests --recursive flag with sub-issues
func TestRunMove_Integration_Recursive(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Use seed issue #4 which has sub-issue #5
	// Move with dry-run to avoid modifying seed data
	result := testutil.RunCommand(t, "move", "4",
		"--status", "done",
		"--recursive",
		"--dry-run",
	)

	testutil.AssertExitCode(t, result, 0)
	testutil.AssertContains(t, result.Stdout, "Dry run")
	testutil.AssertContains(t, result.Stdout, "Issues to update")
	testutil.AssertContains(t, result.Stdout, "#4")
	testutil.AssertContains(t, result.Stdout, "#5")
}

// TestRunMove_Integration_SeedIssue tests moving a seed issue (read-only test)
func TestRunMove_Integration_SeedIssue(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Use dry-run on seed issue to verify command works
	result := testutil.RunCommand(t, "move", "1", "--status", "done", "--dry-run")

	testutil.AssertExitCode(t, result, 0)
	testutil.AssertContains(t, result.Stdout, "Dry run")
	testutil.AssertContains(t, result.Stdout, "#1")
	testutil.AssertContains(t, result.Stdout, "Done")
}
