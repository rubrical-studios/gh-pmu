//go:build integration

package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rubrical-studios/gh-pmu/internal/testutil"
)

// TestRunSubRemove_Integration_UnlinkSubIssue tests unlinking a sub-issue from parent
func TestRunSubRemove_Integration_UnlinkSubIssue(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("Test SubRemove Parent - Unlink - %d", subRemoveTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create child issue and link it
	childTitle := fmt.Sprintf("Test SubRemove Child - Unlink - %d", subRemoveTestID())
	childResult := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", childTitle,
	)
	testutil.AssertExitCode(t, childResult, 0)

	childNum := testutil.ExtractIssueNumber(t, childResult.Stdout)
	defer testutil.DeleteTestIssue(t, childNum)

	// Verify child is linked
	listBefore := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum))
	testutil.AssertExitCode(t, listBefore, 0)
	testutil.AssertContains(t, listBefore.Stdout, childTitle)

	// Remove the sub-issue link
	result := testutil.RunCommand(t, "sub", "remove", fmt.Sprintf("%d", parentNum), fmt.Sprintf("%d", childNum))
	testutil.AssertExitCode(t, result, 0)

	// Verify success message
	testutil.AssertContains(t, result.Stdout, "Removed sub-issue link")
	testutil.AssertContains(t, result.Stdout, fmt.Sprintf("#%d", childNum))

	// Verify child is no longer linked
	listAfter := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum))
	testutil.AssertExitCode(t, listAfter, 0)
	testutil.AssertContains(t, listAfter.Stdout, "No sub-issues found")
}

// TestRunSubRemove_Integration_NotLinkedIsIdempotent tests that removing unlinked issue is idempotent
// Note: The GitHub API's removeSubIssue mutation is idempotent - it succeeds even if the
// link doesn't exist. This is a valid API design choice for "remove" operations.
func TestRunSubRemove_Integration_NotLinkedIsIdempotent(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("Test SubRemove Parent - NotLinked - %d", subRemoveTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create child issue but DON'T link it
	childTitle := fmt.Sprintf("Test SubRemove Child - NotLinked - %d", subRemoveTestID())
	childResult := testutil.RunCommand(t, "create", "--title", childTitle)
	testutil.AssertExitCode(t, childResult, 0)

	childNum := testutil.ExtractIssueNumber(t, childResult.Stdout)
	defer testutil.DeleteTestIssue(t, childNum)

	// Remove a non-existent link - GitHub API is idempotent, so this succeeds
	result := testutil.RunCommand(t, "sub", "remove", fmt.Sprintf("%d", parentNum), fmt.Sprintf("%d", childNum))

	// The operation succeeds (idempotent behavior)
	testutil.AssertExitCode(t, result, 0)
	testutil.AssertContains(t, result.Stdout, "Removed sub-issue link")
}

// TestRunSubRemove_Integration_ParentNotFound tests error when parent doesn't exist
func TestRunSubRemove_Integration_ParentNotFound(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create a valid child issue
	childTitle := fmt.Sprintf("Test SubRemove Child - ParentNotFound - %d", subRemoveTestID())
	childResult := testutil.RunCommand(t, "create", "--title", childTitle)
	testutil.AssertExitCode(t, childResult, 0)

	childNum := testutil.ExtractIssueNumber(t, childResult.Stdout)
	defer testutil.DeleteTestIssue(t, childNum)

	// Try to remove from a non-existent parent
	result := testutil.RunCommand(t, "sub", "remove", "999999", fmt.Sprintf("%d", childNum))

	// Should fail
	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code for non-existent parent")
	}

	testutil.AssertContains(t, result.Stderr, "failed to get parent issue")
}

// TestRunSubRemove_Integration_ChildNotFound tests error when child doesn't exist
func TestRunSubRemove_Integration_ChildNotFound(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create a valid parent issue
	parentTitle := fmt.Sprintf("Test SubRemove Parent - ChildNotFound - %d", subRemoveTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Try to remove a non-existent child
	result := testutil.RunCommand(t, "sub", "remove", fmt.Sprintf("%d", parentNum), "999999")

	// Should fail
	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code for non-existent child")
	}

	testutil.AssertContains(t, result.Stdout, "failed to get issue")
}

// TestRunSubRemove_Integration_WithHashPrefix tests using # prefix on issue numbers
func TestRunSubRemove_Integration_WithHashPrefix(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("Test SubRemove Parent - Hash - %d", subRemoveTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create child issue and link it
	childTitle := fmt.Sprintf("Test SubRemove Child - Hash - %d", subRemoveTestID())
	childResult := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", childTitle,
	)
	testutil.AssertExitCode(t, childResult, 0)

	childNum := testutil.ExtractIssueNumber(t, childResult.Stdout)
	defer testutil.DeleteTestIssue(t, childNum)

	// Remove using # prefix
	result := testutil.RunCommand(t, "sub", "remove", fmt.Sprintf("#%d", parentNum), fmt.Sprintf("#%d", childNum))
	testutil.AssertExitCode(t, result, 0)

	testutil.AssertContains(t, result.Stdout, "Removed sub-issue link")
}

// TestRunSubRemove_Integration_MultipleChildren tests removing multiple children at once
func TestRunSubRemove_Integration_MultipleChildren(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("Test SubRemove Parent - Multiple - %d", subRemoveTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create first child
	child1Title := fmt.Sprintf("Test SubRemove Child1 - Multiple - %d", subRemoveTestID())
	child1Result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", child1Title,
	)
	testutil.AssertExitCode(t, child1Result, 0)

	child1Num := testutil.ExtractIssueNumber(t, child1Result.Stdout)
	defer testutil.DeleteTestIssue(t, child1Num)

	// Create second child
	child2Title := fmt.Sprintf("Test SubRemove Child2 - Multiple - %d", subRemoveTestID())
	child2Result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", child2Title,
	)
	testutil.AssertExitCode(t, child2Result, 0)

	child2Num := testutil.ExtractIssueNumber(t, child2Result.Stdout)
	defer testutil.DeleteTestIssue(t, child2Num)

	// Verify both are linked
	listBefore := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum))
	testutil.AssertExitCode(t, listBefore, 0)
	testutil.AssertContains(t, listBefore.Stdout, child1Title)
	testutil.AssertContains(t, listBefore.Stdout, child2Title)

	// Remove both children at once
	result := testutil.RunCommand(t, "sub", "remove",
		fmt.Sprintf("%d", parentNum),
		fmt.Sprintf("%d", child1Num),
		fmt.Sprintf("%d", child2Num),
	)
	testutil.AssertExitCode(t, result, 0)

	// Verify batch output format
	testutil.AssertContains(t, result.Stdout, "Removing sub-issues from parent")
	testutil.AssertContains(t, result.Stdout, "2 succeeded")

	// Verify both are unlinked
	listAfter := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum))
	testutil.AssertExitCode(t, listAfter, 0)
	testutil.AssertContains(t, listAfter.Stdout, "No sub-issues found")
}

// TestRunSubRemove_Integration_OutputShowsParentTitle tests that output shows parent title
func TestRunSubRemove_Integration_OutputShowsParentTitle(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("Test SubRemove Parent - Output - %d", subRemoveTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create child issue and link it
	childTitle := fmt.Sprintf("Test SubRemove Child - Output - %d", subRemoveTestID())
	childResult := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", childTitle,
	)
	testutil.AssertExitCode(t, childResult, 0)

	childNum := testutil.ExtractIssueNumber(t, childResult.Stdout)
	defer testutil.DeleteTestIssue(t, childNum)

	// Remove the sub-issue link
	result := testutil.RunCommand(t, "sub", "remove", fmt.Sprintf("%d", parentNum), fmt.Sprintf("%d", childNum))
	testutil.AssertExitCode(t, result, 0)

	// Verify output shows parent title
	testutil.AssertContains(t, result.Stdout, "Former parent")
	testutil.AssertContains(t, result.Stdout, parentTitle)
}

// TestRunSubRemove_Integration_DoubleRemoveIsIdempotent tests that removing twice is idempotent
// Note: The GitHub API's removeSubIssue mutation is idempotent
func TestRunSubRemove_Integration_DoubleRemoveIsIdempotent(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("Test SubRemove Parent - Double - %d", subRemoveTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create child issue and link it
	childTitle := fmt.Sprintf("Test SubRemove Child - Double - %d", subRemoveTestID())
	childResult := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", childTitle,
	)
	testutil.AssertExitCode(t, childResult, 0)

	childNum := testutil.ExtractIssueNumber(t, childResult.Stdout)
	defer testutil.DeleteTestIssue(t, childNum)

	// Remove the sub-issue link (first time - should succeed)
	result1 := testutil.RunCommand(t, "sub", "remove", fmt.Sprintf("%d", parentNum), fmt.Sprintf("%d", childNum))
	testutil.AssertExitCode(t, result1, 0)

	// Remove again - idempotent operation succeeds
	result2 := testutil.RunCommand(t, "sub", "remove", fmt.Sprintf("%d", parentNum), fmt.Sprintf("%d", childNum))
	testutil.AssertExitCode(t, result2, 0)
	testutil.AssertContains(t, result2.Stdout, "Removed sub-issue link")
}

// subRemoveTestID returns a unique identifier for sub remove test issues
var subRemoveTestCounter int

func subRemoveTestID() int {
	subRemoveTestCounter++
	return subRemoveTestCounter + int(strings.Count("sub-remove-integration", "e")*1000000)
}
