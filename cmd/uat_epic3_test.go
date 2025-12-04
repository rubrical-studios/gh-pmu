//go:build integration

package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/scooter-indie/gh-pmu/internal/testutil"
)

// UAT Epic 3: Enhanced Integration Tests
// These tests verify end-to-end user workflows for enhanced features

// uatEpic3TestID returns a unique identifier for UAT Epic 3 test issues
var uatEpic3Counter int

func uatEpic3TestID() int {
	uatEpic3Counter++
	return uatEpic3Counter + int(strings.Count("uat-epic3", "e")*1000000)
}

// =============================================================================
// UAT-3.1: Cross-repo sub-issues
// =============================================================================

// TestUAT_CrossRepoSubIssues_ViewWithRepoFlag tests viewing issues with --repo flag
func TestUAT_CrossRepoSubIssues_ViewWithRepoFlag(t *testing.T) {
	env := testutil.RequireTestEnv(t)

	// Create an issue in the test repo
	title := fmt.Sprintf("UAT CrossRepo View - %d", uatEpic3TestID())
	createResult := testutil.RunCommand(t, "create", "--title", title)
	testutil.AssertExitCode(t, createResult, 0)

	issueNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, issueNum)

	// View with explicit repo flag (same repo, but tests the flag works)
	repoFlag := fmt.Sprintf("%s/%s", env.RepoOwner, env.RepoName)
	viewResult := testutil.RunCommand(t, "view",
		fmt.Sprintf("%d", issueNum),
		"--repo", repoFlag,
	)
	testutil.AssertExitCode(t, viewResult, 0)
	testutil.AssertContains(t, viewResult.Stdout, title)
}

// TestUAT_SubAdd_WithHashPrefix tests sub add command with # prefix on issue numbers
func TestUAT_SubAdd_WithHashPrefix(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("UAT Hash Parent - %d", uatEpic3TestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)
	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create child issue
	childTitle := fmt.Sprintf("UAT Hash Child - %d", uatEpic3TestID())
	childResult := testutil.RunCommand(t, "create", "--title", childTitle)
	testutil.AssertExitCode(t, childResult, 0)
	childNum := testutil.ExtractIssueNumber(t, childResult.Stdout)
	defer testutil.DeleteTestIssue(t, childNum)

	// Add child using # prefix
	addResult := testutil.RunCommand(t, "sub", "add",
		fmt.Sprintf("#%d", parentNum),
		fmt.Sprintf("#%d", childNum),
	)
	testutil.AssertExitCode(t, addResult, 0)

	// Verify link was created
	listResult := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum))
	testutil.AssertExitCode(t, listResult, 0)
	testutil.AssertContains(t, listResult.Stdout, childTitle)
}

// =============================================================================
// UAT-3.2: Progress tracking (sub-issue completion %)
// =============================================================================

// TestUAT_ProgressTracking_NoSubIssues tests viewing issue with no sub-issues
func TestUAT_ProgressTracking_NoSubIssues(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create an issue with no sub-issues
	title := fmt.Sprintf("UAT Progress NoSubs - %d", uatEpic3TestID())
	createResult := testutil.RunCommand(t, "create", "--title", title)
	testutil.AssertExitCode(t, createResult, 0)

	issueNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, issueNum)

	// View JSON output
	viewResult := testutil.RunCommand(t, "view", fmt.Sprintf("%d", issueNum), "--json")
	testutil.AssertExitCode(t, viewResult, 0)

	// JSON should include subIssues info (even if empty)
	testutil.AssertContains(t, viewResult.Stdout, "subIssues")
}

// TestUAT_ProgressTracking_WithSubIssues tests viewing issue with sub-issues
func TestUAT_ProgressTracking_WithSubIssues(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("UAT Progress Parent - %d", uatEpic3TestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)
	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create sub-issues
	child1Title := fmt.Sprintf("UAT Progress Child1 - %d", uatEpic3TestID())
	child1Result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", child1Title,
	)
	testutil.AssertExitCode(t, child1Result, 0)
	child1Num := testutil.ExtractIssueNumber(t, child1Result.Stdout)
	defer testutil.DeleteTestIssue(t, child1Num)

	child2Title := fmt.Sprintf("UAT Progress Child2 - %d", uatEpic3TestID())
	child2Result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", child2Title,
	)
	testutil.AssertExitCode(t, child2Result, 0)
	child2Num := testutil.ExtractIssueNumber(t, child2Result.Stdout)
	defer testutil.DeleteTestIssue(t, child2Num)

	// View parent JSON output - should show sub-issue info
	viewResult := testutil.RunCommand(t, "view", fmt.Sprintf("%d", parentNum), "--json")
	testutil.AssertExitCode(t, viewResult, 0)
	testutil.AssertContains(t, viewResult.Stdout, "subIssues")
	testutil.AssertContains(t, viewResult.Stdout, "totalCount")

	// Sub list should show both children
	listResult := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum))
	testutil.AssertExitCode(t, listResult, 0)
	testutil.AssertContains(t, listResult.Stdout, child1Title)
	testutil.AssertContains(t, listResult.Stdout, child2Title)
}

// TestUAT_ProgressTracking_CompletedSubIssues tests progress with completed sub-issues
func TestUAT_ProgressTracking_CompletedSubIssues(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("UAT Progress Complete - %d", uatEpic3TestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)
	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create sub-issues
	child1Title := fmt.Sprintf("UAT Complete Child1 - %d", uatEpic3TestID())
	child1Result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", child1Title,
	)
	testutil.AssertExitCode(t, child1Result, 0)
	child1Num := testutil.ExtractIssueNumber(t, child1Result.Stdout)
	defer testutil.DeleteTestIssue(t, child1Num)

	child2Title := fmt.Sprintf("UAT Complete Child2 - %d", uatEpic3TestID())
	child2Result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", child2Title,
	)
	testutil.AssertExitCode(t, child2Result, 0)
	child2Num := testutil.ExtractIssueNumber(t, child2Result.Stdout)
	defer testutil.DeleteTestIssue(t, child2Num)

	// Complete one sub-issue (move to done)
	moveResult := testutil.RunCommand(t, "move", fmt.Sprintf("%d", child1Num), "--status", "done")
	testutil.AssertExitCode(t, moveResult, 0)

	// List sub-issues - should show state information
	listResult := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum), "--json")
	testutil.AssertExitCode(t, listResult, 0)

	// JSON output should contain state info for children
	testutil.AssertContains(t, listResult.Stdout, "state")
}

// =============================================================================
// UAT-3.3: Recursive operations on issue trees
// =============================================================================

// TestUAT_RecursiveOperations_SubListShowsTree tests that sub list shows the issue tree
func TestUAT_RecursiveOperations_SubListShowsTree(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create grandparent (root) issue
	rootTitle := fmt.Sprintf("UAT Tree Root - %d", uatEpic3TestID())
	rootResult := testutil.RunCommand(t, "create", "--title", rootTitle)
	testutil.AssertExitCode(t, rootResult, 0)
	rootNum := testutil.ExtractIssueNumber(t, rootResult.Stdout)
	defer testutil.DeleteTestIssue(t, rootNum)

	// Create child of root
	childTitle := fmt.Sprintf("UAT Tree Child - %d", uatEpic3TestID())
	childResult := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", rootNum),
		"--title", childTitle,
	)
	testutil.AssertExitCode(t, childResult, 0)
	childNum := testutil.ExtractIssueNumber(t, childResult.Stdout)
	defer testutil.DeleteTestIssue(t, childNum)

	// Create grandchild (child of child)
	grandchildTitle := fmt.Sprintf("UAT Tree Grandchild - %d", uatEpic3TestID())
	grandchildResult := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", childNum),
		"--title", grandchildTitle,
	)
	testutil.AssertExitCode(t, grandchildResult, 0)
	grandchildNum := testutil.ExtractIssueNumber(t, grandchildResult.Stdout)
	defer testutil.DeleteTestIssue(t, grandchildNum)

	// List root's sub-issues (direct children only)
	listRootResult := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", rootNum))
	testutil.AssertExitCode(t, listRootResult, 0)
	testutil.AssertContains(t, listRootResult.Stdout, childTitle)

	// List child's sub-issues
	listChildResult := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", childNum))
	testutil.AssertExitCode(t, listChildResult, 0)
	testutil.AssertContains(t, listChildResult.Stdout, grandchildTitle)
}

// TestUAT_RecursiveOperations_ViewShowsRelations tests view command shows relations
func TestUAT_RecursiveOperations_ViewShowsRelations(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("UAT Relations Parent - %d", uatEpic3TestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)
	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create child issue
	childTitle := fmt.Sprintf("UAT Relations Child - %d", uatEpic3TestID())
	childResult := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", childTitle,
	)
	testutil.AssertExitCode(t, childResult, 0)
	childNum := testutil.ExtractIssueNumber(t, childResult.Stdout)
	defer testutil.DeleteTestIssue(t, childNum)

	// View parent - should show sub-issues exist
	viewParentResult := testutil.RunCommand(t, "view", fmt.Sprintf("%d", parentNum))
	testutil.AssertExitCode(t, viewParentResult, 0)
	// Parent view should indicate it has sub-issues
	testutil.AssertContains(t, viewParentResult.Stdout, "Sub-issues")

	// View child - should show parent relation
	viewChildResult := testutil.RunCommand(t, "view", fmt.Sprintf("%d", childNum))
	testutil.AssertExitCode(t, viewChildResult, 0)
	// Child view should show parent information
	testutil.AssertContains(t, viewChildResult.Stdout, "Parent")
}

// TestUAT_RecursiveOperations_ListWithHasSubIssues tests filtering by sub-issue presence
func TestUAT_RecursiveOperations_ListWithHasSubIssues(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create standalone issue (no sub-issues)
	standaloneTitle := fmt.Sprintf("UAT Standalone - %d", uatEpic3TestID())
	standaloneResult := testutil.RunCommand(t, "create", "--title", standaloneTitle)
	testutil.AssertExitCode(t, standaloneResult, 0)
	standaloneNum := testutil.ExtractIssueNumber(t, standaloneResult.Stdout)
	defer testutil.DeleteTestIssue(t, standaloneNum)

	// Create parent issue
	parentTitle := fmt.Sprintf("UAT HasSubs Parent - %d", uatEpic3TestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)
	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create child for parent
	childTitle := fmt.Sprintf("UAT HasSubs Child - %d", uatEpic3TestID())
	childResult := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", childTitle,
	)
	testutil.AssertExitCode(t, childResult, 0)
	childNum := testutil.ExtractIssueNumber(t, childResult.Stdout)
	defer testutil.DeleteTestIssue(t, childNum)

	// List with --has-sub-issues flag should show parent but not standalone
	listResult := testutil.RunCommand(t, "list", "--has-sub-issues")
	testutil.AssertExitCode(t, listResult, 0)
	testutil.AssertContains(t, listResult.Stdout, parentTitle)
	testutil.AssertNotContains(t, listResult.Stdout, standaloneTitle)
}

// =============================================================================
// UAT Integration: Multi-level Issue Tree Workflow
// =============================================================================

// TestUAT_MultiLevelTreeWorkflow tests creating and managing a multi-level issue tree
func TestUAT_MultiLevelTreeWorkflow(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create Epic (root)
	epicTitle := fmt.Sprintf("UAT Epic - %d", uatEpic3TestID())
	epicResult := testutil.RunCommand(t, "create",
		"--title", epicTitle,
		"--status", "backlog",
	)
	testutil.AssertExitCode(t, epicResult, 0)
	epicNum := testutil.ExtractIssueNumber(t, epicResult.Stdout)
	defer testutil.DeleteTestIssue(t, epicNum)

	// Create Stories under Epic
	story1Title := fmt.Sprintf("UAT Story 1 - %d", uatEpic3TestID())
	story1Result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", epicNum),
		"--title", story1Title,
	)
	testutil.AssertExitCode(t, story1Result, 0)
	story1Num := testutil.ExtractIssueNumber(t, story1Result.Stdout)
	defer testutil.DeleteTestIssue(t, story1Num)

	story2Title := fmt.Sprintf("UAT Story 2 - %d", uatEpic3TestID())
	story2Result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", epicNum),
		"--title", story2Title,
	)
	testutil.AssertExitCode(t, story2Result, 0)
	story2Num := testutil.ExtractIssueNumber(t, story2Result.Stdout)
	defer testutil.DeleteTestIssue(t, story2Num)

	// Create Tasks under Story 1
	task1Title := fmt.Sprintf("UAT Task 1 - %d", uatEpic3TestID())
	task1Result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", story1Num),
		"--title", task1Title,
	)
	testutil.AssertExitCode(t, task1Result, 0)
	task1Num := testutil.ExtractIssueNumber(t, task1Result.Stdout)
	defer testutil.DeleteTestIssue(t, task1Num)

	// Verify Epic has stories
	listEpicResult := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", epicNum))
	testutil.AssertExitCode(t, listEpicResult, 0)
	testutil.AssertContains(t, listEpicResult.Stdout, story1Title)
	testutil.AssertContains(t, listEpicResult.Stdout, story2Title)

	// Verify Story 1 has tasks
	listStory1Result := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", story1Num))
	testutil.AssertExitCode(t, listStory1Result, 0)
	testutil.AssertContains(t, listStory1Result.Stdout, task1Title)

	// Complete task
	moveTaskResult := testutil.RunCommand(t, "move", fmt.Sprintf("%d", task1Num), "--status", "done")
	testutil.AssertExitCode(t, moveTaskResult, 0)

	// Complete story 1
	moveStory1Result := testutil.RunCommand(t, "move", fmt.Sprintf("%d", story1Num), "--status", "done")
	testutil.AssertExitCode(t, moveStory1Result, 0)

	// Verify Epic view shows sub-issues
	viewEpicResult := testutil.RunCommand(t, "view", fmt.Sprintf("%d", epicNum), "--json")
	testutil.AssertExitCode(t, viewEpicResult, 0)
	testutil.AssertContains(t, viewEpicResult.Stdout, "subIssues")
}
