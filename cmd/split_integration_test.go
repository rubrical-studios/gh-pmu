//go:build integration

package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/rubrical-studios/gh-pmu/internal/testutil"
)

// TestRunSplit_Integration_FromBodyChecklist tests splitting from issue body checklist
func TestRunSplit_Integration_FromBodyChecklist(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create a parent issue with a checklist in the body
	parentTitle := fmt.Sprintf("Test Split Parent - Body - %d", splitTestID())
	parentBody := `## Epic Description

This is an epic that needs to be split.

## Tasks
- [ ] First sub-task from body
- [ ] Second sub-task from body
- [x] Already completed task (should be skipped)
- [ ] Third sub-task from body
`

	// Create parent issue using gh CLI directly
	createResult := testutil.RunCommand(t, "create", "--title", parentTitle, "--body", parentBody)
	testutil.AssertExitCode(t, createResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Split from body
	result := testutil.RunCommand(t, "split", fmt.Sprintf("%d", parentNum), "--from=body")
	testutil.AssertExitCode(t, result, 0)

	// Verify output mentions created sub-issues
	testutil.AssertContains(t, result.Stdout, "Created sub-issue #")
	testutil.AssertContains(t, result.Stdout, "First sub-task from body")
	testutil.AssertContains(t, result.Stdout, "Second sub-task from body")
	testutil.AssertContains(t, result.Stdout, "Third sub-task from body")

	// Should NOT contain the completed task
	testutil.AssertNotContains(t, result.Stdout, "Already completed task")

	// Verify 3 sub-issues were created
	testutil.AssertContains(t, result.Stdout, "3 sub-issue(s) created")

	// Clean up created sub-issues by extracting their numbers
	lines := strings.Split(result.Stdout, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Created sub-issue #") {
			subNum := testutil.ExtractIssueNumber(t, line)
			defer testutil.DeleteTestIssue(t, subNum)
		}
	}
}

// TestRunSplit_Integration_FromArguments tests splitting from command line arguments
func TestRunSplit_Integration_FromArguments(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create a parent issue
	parentTitle := fmt.Sprintf("Test Split Parent - Args - %d", splitTestID())

	createResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, createResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Split using arguments
	result := testutil.RunCommand(t, "split", fmt.Sprintf("%d", parentNum),
		"Task from args 1",
		"Task from args 2",
	)
	testutil.AssertExitCode(t, result, 0)

	// Verify output
	testutil.AssertContains(t, result.Stdout, "Created sub-issue #")
	testutil.AssertContains(t, result.Stdout, "Task from args 1")
	testutil.AssertContains(t, result.Stdout, "Task from args 2")
	testutil.AssertContains(t, result.Stdout, "2 sub-issue(s) created")

	// Clean up created sub-issues
	lines := strings.Split(result.Stdout, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Created sub-issue #") {
			subNum := testutil.ExtractIssueNumber(t, line)
			defer testutil.DeleteTestIssue(t, subNum)
		}
	}
}

// TestRunSplit_Integration_DryRun tests --dry-run shows what would be created
func TestRunSplit_Integration_DryRun(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create a parent issue with checklist
	parentTitle := fmt.Sprintf("Test Split Parent - DryRun - %d", splitTestID())
	parentBody := `## Tasks
- [ ] Dry run task 1
- [ ] Dry run task 2
`

	createResult := testutil.RunCommand(t, "create", "--title", parentTitle, "--body", parentBody)
	testutil.AssertExitCode(t, createResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Run with --dry-run
	result := testutil.RunCommand(t, "split", fmt.Sprintf("%d", parentNum), "--from=body", "--dry-run")
	testutil.AssertExitCode(t, result, 0)

	// Verify it shows what WOULD be created
	testutil.AssertContains(t, result.Stdout, "Would create")
	testutil.AssertContains(t, result.Stdout, "2 sub-issue(s)")
	testutil.AssertContains(t, result.Stdout, "Dry run task 1")
	testutil.AssertContains(t, result.Stdout, "Dry run task 2")

	// Verify no actual sub-issues were created by listing sub-issues
	listResult := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum))
	testutil.AssertExitCode(t, listResult, 0)
	testutil.AssertContains(t, listResult.Stdout, "No sub-issues found")
}

// TestRunSplit_Integration_SubIssuesLinked tests that sub-issues are properly linked to parent
func TestRunSplit_Integration_SubIssuesLinked(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create a parent issue
	parentTitle := fmt.Sprintf("Test Split Parent - Linked - %d", splitTestID())

	createResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, createResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Split using arguments
	result := testutil.RunCommand(t, "split", fmt.Sprintf("%d", parentNum),
		"Linked sub-task 1",
		"Linked sub-task 2",
	)
	testutil.AssertExitCode(t, result, 0)

	// Extract created sub-issue numbers for cleanup
	var subNums []int
	lines := strings.Split(result.Stdout, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Created sub-issue #") {
			subNum := testutil.ExtractIssueNumber(t, line)
			subNums = append(subNums, subNum)
			defer testutil.DeleteTestIssue(t, subNum)
		}
	}

	// Verify sub-issues are linked by listing them
	listResult := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum))
	testutil.AssertExitCode(t, listResult, 0)

	// Should show the created sub-issues
	testutil.AssertContains(t, listResult.Stdout, "Linked sub-task 1")
	testutil.AssertContains(t, listResult.Stdout, "Linked sub-task 2")

	// Should show progress (0/2 complete since all are open)
	testutil.AssertContains(t, listResult.Stdout, "0/2 complete")
}

// TestRunSplit_Integration_JSONOutput tests --json output format
func TestRunSplit_Integration_JSONOutput(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create a parent issue with checklist
	parentTitle := fmt.Sprintf("Test Split Parent - JSON - %d", splitTestID())
	parentBody := `- [ ] JSON task 1
- [ ] JSON task 2`

	createResult := testutil.RunCommand(t, "create", "--title", parentTitle, "--body", parentBody)
	testutil.AssertExitCode(t, createResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Test dry-run with JSON output
	dryRunResult := testutil.RunCommand(t, "split", fmt.Sprintf("%d", parentNum), "--from=body", "--dry-run", "--json")
	testutil.AssertExitCode(t, dryRunResult, 0)

	// Verify JSON structure for dry-run
	var dryRunJSON map[string]interface{}
	err := json.Unmarshal([]byte(dryRunResult.Stdout), &dryRunJSON)
	if err != nil {
		t.Fatalf("Failed to parse dry-run JSON output: %v\nOutput: %s", err, dryRunResult.Stdout)
	}

	if dryRunJSON["status"] != "dry-run" {
		t.Errorf("Expected status 'dry-run', got %v", dryRunJSON["status"])
	}

	parent, ok := dryRunJSON["parent"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected parent to be an object")
	}
	if int(parent["number"].(float64)) != parentNum {
		t.Errorf("Expected parent number %d, got %v", parentNum, parent["number"])
	}

	if int(dryRunJSON["taskCount"].(float64)) != 2 {
		t.Errorf("Expected taskCount 2, got %v", dryRunJSON["taskCount"])
	}

	// Now test actual split with JSON output
	splitResult := testutil.RunCommand(t, "split", fmt.Sprintf("%d", parentNum), "--from=body", "--json")
	testutil.AssertExitCode(t, splitResult, 0)

	// Parse and verify actual split JSON
	var splitJSON map[string]interface{}
	err = json.Unmarshal([]byte(splitResult.Stdout), &splitJSON)
	if err != nil {
		t.Fatalf("Failed to parse split JSON output: %v\nOutput: %s", err, splitResult.Stdout)
	}

	if splitJSON["status"] != "completed" {
		t.Errorf("Expected status 'completed', got %v", splitJSON["status"])
	}

	if int(splitJSON["createdCount"].(float64)) != 2 {
		t.Errorf("Expected createdCount 2, got %v", splitJSON["createdCount"])
	}

	created, ok := splitJSON["created"].([]interface{})
	if !ok {
		t.Fatal("Expected created to be an array")
	}
	if len(created) != 2 {
		t.Errorf("Expected 2 created items, got %d", len(created))
	}

	// Clean up created sub-issues
	for _, item := range created {
		itemMap := item.(map[string]interface{})
		subNum := int(itemMap["number"].(float64))
		defer testutil.DeleteTestIssue(t, subNum)
	}
}

// TestRunSplit_Integration_NoChecklist tests splitting an issue with no checklist
func TestRunSplit_Integration_NoChecklist(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create a parent issue with no checklist
	parentTitle := fmt.Sprintf("Test Split Parent - NoChecklist - %d", splitTestID())
	parentBody := `## Description

This issue has no checklist items at all.
Just plain text without any - [ ] items.
`

	createResult := testutil.RunCommand(t, "create", "--title", parentTitle, "--body", parentBody)
	testutil.AssertExitCode(t, createResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Try to split from body - should handle gracefully
	result := testutil.RunCommand(t, "split", fmt.Sprintf("%d", parentNum), "--from=body")
	testutil.AssertExitCode(t, result, 0)

	// Verify it reports no tasks found
	testutil.AssertContains(t, result.Stdout, "No tasks found")
}

// TestRunSplit_Integration_NoChecklistJSON tests --json output when no checklist
func TestRunSplit_Integration_NoChecklistJSON(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create a parent issue with no checklist
	parentTitle := fmt.Sprintf("Test Split Parent - NoChecklistJSON - %d", splitTestID())
	parentBody := "No checklist here"

	createResult := testutil.RunCommand(t, "create", "--title", parentTitle, "--body", parentBody)
	testutil.AssertExitCode(t, createResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Try to split from body with JSON output
	result := testutil.RunCommand(t, "split", fmt.Sprintf("%d", parentNum), "--from=body", "--json")
	testutil.AssertExitCode(t, result, 0)

	// Verify JSON structure
	var jsonResult map[string]interface{}
	err := json.Unmarshal([]byte(result.Stdout), &jsonResult)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, result.Stdout)
	}

	if jsonResult["status"] != "no-tasks" {
		t.Errorf("Expected status 'no-tasks', got %v", jsonResult["status"])
	}

	if int(jsonResult["taskCount"].(float64)) != 0 {
		t.Errorf("Expected taskCount 0, got %v", jsonResult["taskCount"])
	}
}

// TestRunSplit_Integration_InvalidIssue tests error when parent issue doesn't exist
func TestRunSplit_Integration_InvalidIssue(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Try to split a non-existent issue
	result := testutil.RunCommand(t, "split", "999999", "--from=body")

	// Should fail
	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code for non-existent issue")
	}

	testutil.AssertContains(t, result.Stderr, "failed to get issue")
}

// TestRunSplit_Integration_NoTasksSpecified tests error when no tasks source is specified
func TestRunSplit_Integration_NoTasksSpecified(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create a parent issue
	parentTitle := fmt.Sprintf("Test Split Parent - NoSource - %d", splitTestID())

	createResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, createResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Try to split without --from and without task arguments
	result := testutil.RunCommand(t, "split", fmt.Sprintf("%d", parentNum))

	// Should fail
	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code when no tasks specified")
	}

	testutil.AssertContains(t, result.Stderr, "no tasks specified")
}

// splitTestID returns a unique identifier for split test issues
var splitTestCounter int

func splitTestID() int {
	splitTestCounter++
	return splitTestCounter + int(strings.Count("split-integration", "i")*1000000)
}
