//go:build integration

package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rubrical-studios/gh-pmu/internal/testutil"
)

// UAT Epic 1: Core Unification Tests
// These tests verify end-to-end user workflows for core functionality

// uatTestID returns a unique identifier for UAT test issues
var uatTestCounter int

func uatTestID() int {
	uatTestCounter++
	return uatTestCounter + int(strings.Count("uat-epic1", "a")*1000000)
}

// =============================================================================
// UAT-1.2: List and filter issues (gh pmu list)
// =============================================================================

// TestUAT_ListAndFilter_Workflow tests the list command with filtering
func TestUAT_ListAndFilter_Workflow(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create test issues with different statuses
	backlogTitle := fmt.Sprintf("UAT List - Backlog - %d", uatTestID())
	inProgressTitle := fmt.Sprintf("UAT List - InProgress - %d", uatTestID())

	// Create issue in backlog
	backlogResult := testutil.RunCommand(t, "create",
		"--title", backlogTitle,
		"--status", "backlog",
	)
	testutil.AssertExitCode(t, backlogResult, 0)
	backlogNum := testutil.ExtractIssueNumber(t, backlogResult.Stdout)
	defer testutil.DeleteTestIssue(t, backlogNum)

	// Create issue in progress
	inProgressResult := testutil.RunCommand(t, "create",
		"--title", inProgressTitle,
		"--status", "in_progress",
	)
	testutil.AssertExitCode(t, inProgressResult, 0)
	inProgressNum := testutil.ExtractIssueNumber(t, inProgressResult.Stdout)
	defer testutil.DeleteTestIssue(t, inProgressNum)

	// Test: List all issues shows both
	listAllResult := testutil.RunCommand(t, "list")
	testutil.AssertExitCode(t, listAllResult, 0)
	testutil.AssertContains(t, listAllResult.Stdout, backlogTitle)
	testutil.AssertContains(t, listAllResult.Stdout, inProgressTitle)

	// Test: Filter by status=backlog shows only backlog issue
	listBacklogResult := testutil.RunCommand(t, "list", "--status", "backlog")
	testutil.AssertExitCode(t, listBacklogResult, 0)
	testutil.AssertContains(t, listBacklogResult.Stdout, backlogTitle)
	testutil.AssertNotContains(t, listBacklogResult.Stdout, inProgressTitle)

	// Test: Filter by status=in_progress shows only in_progress issue
	listInProgressResult := testutil.RunCommand(t, "list", "--status", "in_progress")
	testutil.AssertExitCode(t, listInProgressResult, 0)
	testutil.AssertContains(t, listInProgressResult.Stdout, inProgressTitle)
	testutil.AssertNotContains(t, listInProgressResult.Stdout, backlogTitle)
}

// =============================================================================
// UAT-1.3: Create issue with fields (gh pmu create)
// =============================================================================

// TestUAT_CreateWithFields_Workflow tests creating issues with various field values
func TestUAT_CreateWithFields_Workflow(t *testing.T) {
	testutil.RequireTestEnv(t)

	title := fmt.Sprintf("UAT Create Fields - %d", uatTestID())
	body := "This is a UAT test issue with custom fields set."

	// Create issue with all major fields
	result := testutil.RunCommand(t, "create",
		"--title", title,
		"--body", body,
		"--status", "in_progress",
		"--priority", "p0",
	)
	testutil.AssertExitCode(t, result, 0)
	testutil.AssertContains(t, result.Stdout, "Created issue #")

	issueNum := testutil.ExtractIssueNumber(t, result.Stdout)
	defer testutil.DeleteTestIssue(t, issueNum)

	// Verify all fields were set correctly
	viewResult := testutil.RunCommand(t, "view", fmt.Sprintf("%d", issueNum), "--json")
	testutil.AssertExitCode(t, viewResult, 0)
	testutil.AssertContains(t, viewResult.Stdout, title)
	testutil.AssertContains(t, viewResult.Stdout, body)
	testutil.AssertContains(t, viewResult.Stdout, "In progress")
	testutil.AssertContains(t, viewResult.Stdout, "P0")
}

// =============================================================================
// UAT-1.4: Move issue through workflow (gh pmu move)
// =============================================================================

// TestUAT_MoveWorkflow_CompleteLifecycle tests moving an issue through all workflow states
func TestUAT_MoveWorkflow_CompleteLifecycle(t *testing.T) {
	testutil.RequireTestEnv(t)

	title := fmt.Sprintf("UAT Move Lifecycle - %d", uatTestID())

	// Create issue in backlog
	createResult := testutil.RunCommand(t, "create",
		"--title", title,
		"--status", "backlog",
	)
	testutil.AssertExitCode(t, createResult, 0)
	issueNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, issueNum)

	// Move to in_progress
	moveResult1 := testutil.RunCommand(t, "move", fmt.Sprintf("%d", issueNum), "--status", "in_progress")
	testutil.AssertExitCode(t, moveResult1, 0)
	testutil.AssertContains(t, moveResult1.Stdout, "In progress")

	// Verify status changed
	viewResult1 := testutil.RunCommand(t, "view", fmt.Sprintf("%d", issueNum), "--json")
	testutil.AssertExitCode(t, viewResult1, 0)
	testutil.AssertContains(t, viewResult1.Stdout, "In progress")

	// Move to in_review
	moveResult2 := testutil.RunCommand(t, "move", fmt.Sprintf("%d", issueNum), "--status", "in_review")
	testutil.AssertExitCode(t, moveResult2, 0)
	testutil.AssertContains(t, moveResult2.Stdout, "In review")

	// Move to done
	moveResult3 := testutil.RunCommand(t, "move", fmt.Sprintf("%d", issueNum), "--status", "done")
	testutil.AssertExitCode(t, moveResult3, 0)
	testutil.AssertContains(t, moveResult3.Stdout, "Done")

	// Final verification
	viewResultFinal := testutil.RunCommand(t, "view", fmt.Sprintf("%d", issueNum), "--json")
	testutil.AssertExitCode(t, viewResultFinal, 0)
	testutil.AssertContains(t, viewResultFinal.Stdout, "Done")
}

// TestUAT_MovePriority tests changing priority via move command
func TestUAT_MovePriority(t *testing.T) {
	testutil.RequireTestEnv(t)

	title := fmt.Sprintf("UAT Move Priority - %d", uatTestID())

	// Create issue with low priority
	createResult := testutil.RunCommand(t, "create",
		"--title", title,
		"--priority", "p2",
	)
	testutil.AssertExitCode(t, createResult, 0)
	issueNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, issueNum)

	// Escalate to high priority
	moveResult := testutil.RunCommand(t, "move", fmt.Sprintf("%d", issueNum), "--priority", "p0")
	testutil.AssertExitCode(t, moveResult, 0)
	testutil.AssertContains(t, moveResult.Stdout, "P0")

	// Verify priority changed
	viewResult := testutil.RunCommand(t, "view", fmt.Sprintf("%d", issueNum), "--json")
	testutil.AssertExitCode(t, viewResult, 0)
	testutil.AssertContains(t, viewResult.Stdout, "P0")
}

// =============================================================================
// UAT-1.7: Manage sub-issues (gh pmu sub *)
// =============================================================================

// TestUAT_SubIssueManagement_CompleteWorkflow tests the full sub-issue lifecycle
func TestUAT_SubIssueManagement_CompleteWorkflow(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("UAT Parent Issue - %d", uatTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)
	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create child issue directly linked to parent
	child1Title := fmt.Sprintf("UAT Sub Create Child - %d", uatTestID())
	child1Result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", child1Title,
	)
	testutil.AssertExitCode(t, child1Result, 0)
	child1Num := testutil.ExtractIssueNumber(t, child1Result.Stdout)
	defer testutil.DeleteTestIssue(t, child1Num)

	// Verify child is linked using sub list
	listResult := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum))
	testutil.AssertExitCode(t, listResult, 0)
	testutil.AssertContains(t, listResult.Stdout, child1Title)

	// Create another issue and add it as sub-issue
	child2Title := fmt.Sprintf("UAT Sub Add Child - %d", uatTestID())
	child2Result := testutil.RunCommand(t, "create", "--title", child2Title)
	testutil.AssertExitCode(t, child2Result, 0)
	child2Num := testutil.ExtractIssueNumber(t, child2Result.Stdout)
	defer testutil.DeleteTestIssue(t, child2Num)

	// Add existing issue as sub-issue
	addResult := testutil.RunCommand(t, "sub", "add",
		fmt.Sprintf("%d", parentNum),
		fmt.Sprintf("%d", child2Num),
	)
	testutil.AssertExitCode(t, addResult, 0)

	// Verify both children are now linked
	listResult2 := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum))
	testutil.AssertExitCode(t, listResult2, 0)
	testutil.AssertContains(t, listResult2.Stdout, child1Title)
	testutil.AssertContains(t, listResult2.Stdout, child2Title)

	// Remove one sub-issue
	removeResult := testutil.RunCommand(t, "sub", "remove",
		fmt.Sprintf("%d", parentNum),
		fmt.Sprintf("%d", child2Num),
	)
	testutil.AssertExitCode(t, removeResult, 0)

	// Verify only child1 remains
	listResult3 := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum))
	testutil.AssertExitCode(t, listResult3, 0)
	testutil.AssertContains(t, listResult3.Stdout, child1Title)
	testutil.AssertNotContains(t, listResult3.Stdout, child2Title)
}

// =============================================================================
// UAT-1.8: Split issue into tasks (gh pmu split)
// =============================================================================

// TestUAT_SplitIssue_Workflow tests splitting an issue into multiple tasks
func TestUAT_SplitIssue_Workflow(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("UAT Split Parent - %d", uatTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)
	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Split into tasks
	task1 := fmt.Sprintf("Task 1 - %d", uatTestID())
	task2 := fmt.Sprintf("Task 2 - %d", uatTestID())

	splitResult := testutil.RunCommand(t, "split",
		fmt.Sprintf("%d", parentNum),
		"--task", task1,
		"--task", task2,
	)
	testutil.AssertExitCode(t, splitResult, 0)
	testutil.AssertContains(t, splitResult.Stdout, "Created 2 sub-issue")

	// Extract created issue numbers for cleanup
	// The split command outputs the created issues
	lines := strings.Split(splitResult.Stdout, "\n")
	var createdNums []int
	for _, line := range lines {
		if strings.Contains(line, "Created #") {
			num := testutil.ExtractIssueNumber(t, line)
			if num > 0 {
				createdNums = append(createdNums, num)
				defer testutil.DeleteTestIssue(t, num)
			}
		}
	}

	// Verify tasks are linked as sub-issues
	listResult := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum))
	testutil.AssertExitCode(t, listResult, 0)
	testutil.AssertContains(t, listResult.Stdout, task1)
	testutil.AssertContains(t, listResult.Stdout, task2)
}

// =============================================================================
// UAT Integration Workflow: Complete Issue Lifecycle
// =============================================================================

// TestUAT_CompleteIssueLifecycle tests an issue from creation to completion
func TestUAT_CompleteIssueLifecycle(t *testing.T) {
	testutil.RequireTestEnv(t)

	title := fmt.Sprintf("UAT Complete Lifecycle - %d", uatTestID())

	// Step 1: Create issue in backlog
	createResult := testutil.RunCommand(t, "create",
		"--title", title,
		"--status", "backlog",
		"--priority", "p1",
	)
	testutil.AssertExitCode(t, createResult, 0)
	issueNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, issueNum)

	// Step 2: Verify it appears in list with correct status
	listResult := testutil.RunCommand(t, "list", "--status", "backlog")
	testutil.AssertExitCode(t, listResult, 0)
	testutil.AssertContains(t, listResult.Stdout, title)

	// Step 3: Start work - move to in_progress
	moveResult1 := testutil.RunCommand(t, "move", fmt.Sprintf("%d", issueNum), "--status", "in_progress")
	testutil.AssertExitCode(t, moveResult1, 0)

	// Step 4: Verify it moved out of backlog filter
	listBacklogResult := testutil.RunCommand(t, "list", "--status", "backlog")
	testutil.AssertExitCode(t, listBacklogResult, 0)
	testutil.AssertNotContains(t, listBacklogResult.Stdout, title)

	// Step 5: Verify it appears in in_progress filter
	listInProgressResult := testutil.RunCommand(t, "list", "--status", "in_progress")
	testutil.AssertExitCode(t, listInProgressResult, 0)
	testutil.AssertContains(t, listInProgressResult.Stdout, title)

	// Step 6: Complete work - move to done
	moveResult2 := testutil.RunCommand(t, "move", fmt.Sprintf("%d", issueNum), "--status", "done")
	testutil.AssertExitCode(t, moveResult2, 0)

	// Step 7: Final verification
	viewResult := testutil.RunCommand(t, "view", fmt.Sprintf("%d", issueNum), "--json")
	testutil.AssertExitCode(t, viewResult, 0)
	testutil.AssertContains(t, viewResult.Stdout, "Done")
}
