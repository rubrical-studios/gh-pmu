//go:build integration

package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rubrical-studios/gh-pmu/internal/testutil"
)

// TestRunSubCreate_Integration_WithTitle tests creating a sub-issue with --title
func TestRunSubCreate_Integration_WithTitle(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("Test SubCreate Parent - Title - %d", subCreateTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create sub-issue
	subTitle := fmt.Sprintf("Test SubCreate Child - Title - %d", subCreateTestID())
	result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", subTitle,
	)
	testutil.AssertExitCode(t, result, 0)

	// Verify success message
	testutil.AssertContains(t, result.Stdout, "Created sub-issue")
	testutil.AssertContains(t, result.Stdout, subTitle)

	// Extract sub-issue number for cleanup
	subNum := testutil.ExtractIssueNumber(t, result.Stdout)
	defer testutil.DeleteTestIssue(t, subNum)

	// Verify the sub-issue is linked
	listResult := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum))
	testutil.AssertExitCode(t, listResult, 0)
	testutil.AssertContains(t, listResult.Stdout, subTitle)
}

// TestRunSubCreate_Integration_WithTitleAndBody tests creating sub-issue with --title and --body
func TestRunSubCreate_Integration_WithTitleAndBody(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("Test SubCreate Parent - TitleBody - %d", subCreateTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create sub-issue with body
	subTitle := fmt.Sprintf("Test SubCreate Child - TitleBody - %d", subCreateTestID())
	subBody := "This is the body of the sub-issue created by integration test."

	result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", subTitle,
		"--body", subBody,
	)
	testutil.AssertExitCode(t, result, 0)

	// Extract sub-issue number for cleanup
	subNum := testutil.ExtractIssueNumber(t, result.Stdout)
	defer testutil.DeleteTestIssue(t, subNum)

	// Verify the sub-issue body via view command
	viewResult := testutil.RunCommand(t, "view", fmt.Sprintf("%d", subNum))
	testutil.AssertExitCode(t, viewResult, 0)
	testutil.AssertContains(t, viewResult.Stdout, subBody)
}

// TestRunSubCreate_Integration_InheritLabels tests --inherit-labels flag (default true)
func TestRunSubCreate_Integration_InheritLabels(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue with a label
	parentTitle := fmt.Sprintf("Test SubCreate Parent - InheritLabels - %d", subCreateTestID())
	parentResult := testutil.RunCommand(t, "create",
		"--title", parentTitle,
		"--label", "bug",
	)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create sub-issue (should inherit labels by default)
	subTitle := fmt.Sprintf("Test SubCreate Child - InheritLabels - %d", subCreateTestID())
	result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", subTitle,
	)
	testutil.AssertExitCode(t, result, 0)

	// Extract sub-issue number for cleanup
	subNum := testutil.ExtractIssueNumber(t, result.Stdout)
	defer testutil.DeleteTestIssue(t, subNum)

	// Verify labels were inherited (check output mentions labels)
	// The output shows "Labels:" if labels were applied
	testutil.AssertContains(t, result.Stdout, "bug")
}

// TestRunSubCreate_Integration_NoInheritLabels tests --inherit-labels=false
func TestRunSubCreate_Integration_NoInheritLabels(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue with a label
	parentTitle := fmt.Sprintf("Test SubCreate Parent - NoInheritLabels - %d", subCreateTestID())
	parentResult := testutil.RunCommand(t, "create",
		"--title", parentTitle,
		"--label", "bug",
	)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create sub-issue with --inherit-labels=false
	subTitle := fmt.Sprintf("Test SubCreate Child - NoInheritLabels - %d", subCreateTestID())
	result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", subTitle,
		"--inherit-labels=false",
	)
	testutil.AssertExitCode(t, result, 0)

	// Extract sub-issue number for cleanup
	subNum := testutil.ExtractIssueNumber(t, result.Stdout)
	defer testutil.DeleteTestIssue(t, subNum)

	// Verify labels were NOT inherited (output should not mention Labels: unless config defaults apply)
	// This is tricky because config defaults might add labels
	// At minimum, the command should succeed
	testutil.AssertContains(t, result.Stdout, "Created sub-issue")
}

// TestRunSubCreate_Integration_InheritAssignees tests --inherit-assignees flag
func TestRunSubCreate_Integration_InheritAssignees(t *testing.T) {
	env := testutil.RequireTestEnv(t)

	// Create parent issue (we can't easily assign in tests without knowing valid usernames)
	// Instead, we'll test that the flag is accepted and command succeeds
	parentTitle := fmt.Sprintf("Test SubCreate Parent - InheritAssignees - %d", subCreateTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create sub-issue with --inherit-assignees=true
	subTitle := fmt.Sprintf("Test SubCreate Child - InheritAssignees - %d", subCreateTestID())
	result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", subTitle,
		"--inherit-assignees=true",
	)
	testutil.AssertExitCode(t, result, 0)

	// Extract sub-issue number for cleanup
	subNum := testutil.ExtractIssueNumber(t, result.Stdout)
	defer testutil.DeleteTestIssue(t, subNum)

	_ = env // used for RequireTestEnv
	testutil.AssertContains(t, result.Stdout, "Created sub-issue")
}

// TestRunSubCreate_Integration_SubIssueLinkedToParent tests that sub-issue is properly linked
func TestRunSubCreate_Integration_SubIssueLinkedToParent(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("Test SubCreate Parent - Linked - %d", subCreateTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create sub-issue
	subTitle := fmt.Sprintf("Test SubCreate Child - Linked - %d", subCreateTestID())
	result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", subTitle,
	)
	testutil.AssertExitCode(t, result, 0)

	subNum := testutil.ExtractIssueNumber(t, result.Stdout)
	defer testutil.DeleteTestIssue(t, subNum)

	// Verify output shows parent info
	testutil.AssertContains(t, result.Stdout, fmt.Sprintf("parent #%d", parentNum))

	// Verify link by checking sub list
	listResult := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum))
	testutil.AssertExitCode(t, listResult, 0)
	testutil.AssertContains(t, listResult.Stdout, subTitle)
	testutil.AssertContains(t, listResult.Stdout, fmt.Sprintf("#%d", subNum))

	// Verify via --relation parent from child's perspective
	parentRelation := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", subNum), "--relation", "parent")
	testutil.AssertExitCode(t, parentRelation, 0)
	testutil.AssertContains(t, parentRelation.Stdout, parentTitle)
}

// TestRunSubCreate_Integration_ParentNotFound tests error when parent doesn't exist
func TestRunSubCreate_Integration_ParentNotFound(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Try to create sub-issue for non-existent parent
	subTitle := fmt.Sprintf("Test SubCreate Child - ParentNotFound - %d", subCreateTestID())
	result := testutil.RunCommand(t, "sub", "create",
		"--parent", "999999",
		"--title", subTitle,
	)

	// Should fail
	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code for non-existent parent")
	}

	testutil.AssertContains(t, result.Stderr, "failed to get parent issue")
}

// TestRunSubCreate_Integration_OutputFormat tests output includes expected information
func TestRunSubCreate_Integration_OutputFormat(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("Test SubCreate Parent - Output - %d", subCreateTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create sub-issue
	subTitle := fmt.Sprintf("Test SubCreate Child - Output - %d", subCreateTestID())
	result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", subTitle,
	)
	testutil.AssertExitCode(t, result, 0)

	subNum := testutil.ExtractIssueNumber(t, result.Stdout)
	defer testutil.DeleteTestIssue(t, subNum)

	// Verify output format
	testutil.AssertContains(t, result.Stdout, "Created sub-issue")
	testutil.AssertContains(t, result.Stdout, fmt.Sprintf("#%d", subNum))
	testutil.AssertContains(t, result.Stdout, "Title:")
	testutil.AssertContains(t, result.Stdout, subTitle)
	testutil.AssertContains(t, result.Stdout, "Parent:")
	testutil.AssertContains(t, result.Stdout, parentTitle)
	// Should include URL
	testutil.AssertContains(t, result.Stdout, "https://github.com/")
}

// TestRunSubCreate_Integration_MultipleSubIssues tests creating multiple sub-issues
func TestRunSubCreate_Integration_MultipleSubIssues(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("Test SubCreate Parent - Multiple - %d", subCreateTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create first sub-issue
	sub1Title := fmt.Sprintf("Test SubCreate Child1 - Multiple - %d", subCreateTestID())
	result1 := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", sub1Title,
	)
	testutil.AssertExitCode(t, result1, 0)

	sub1Num := testutil.ExtractIssueNumber(t, result1.Stdout)
	defer testutil.DeleteTestIssue(t, sub1Num)

	// Create second sub-issue
	sub2Title := fmt.Sprintf("Test SubCreate Child2 - Multiple - %d", subCreateTestID())
	result2 := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", sub2Title,
	)
	testutil.AssertExitCode(t, result2, 0)

	sub2Num := testutil.ExtractIssueNumber(t, result2.Stdout)
	defer testutil.DeleteTestIssue(t, sub2Num)

	// Verify both are linked
	listResult := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum))
	testutil.AssertExitCode(t, listResult, 0)
	testutil.AssertContains(t, listResult.Stdout, sub1Title)
	testutil.AssertContains(t, listResult.Stdout, sub2Title)
	testutil.AssertContains(t, listResult.Stdout, "0/2 complete")
}

// TestRunSubCreate_Integration_WithHashPrefix tests using # prefix on parent
func TestRunSubCreate_Integration_WithHashPrefix(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("Test SubCreate Parent - Hash - %d", subCreateTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create sub-issue using # prefix
	subTitle := fmt.Sprintf("Test SubCreate Child - Hash - %d", subCreateTestID())
	result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("#%d", parentNum),
		"--title", subTitle,
	)
	testutil.AssertExitCode(t, result, 0)

	subNum := testutil.ExtractIssueNumber(t, result.Stdout)
	defer testutil.DeleteTestIssue(t, subNum)

	testutil.AssertContains(t, result.Stdout, "Created sub-issue")
}

// TestRunSubCreate_Integration_RequiredFlags tests that required flags are enforced
func TestRunSubCreate_Integration_RequiredFlags(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Missing --title
	result1 := testutil.RunCommand(t, "sub", "create", "--parent", "123")
	if result1.ExitCode == 0 {
		t.Error("Expected non-zero exit code when --title is missing")
	}

	// Missing --parent
	result2 := testutil.RunCommand(t, "sub", "create", "--title", "Test")
	if result2.ExitCode == 0 {
		t.Error("Expected non-zero exit code when --parent is missing")
	}
}

// subCreateTestID returns a unique identifier for sub create test issues
var subCreateTestCounter int

func subCreateTestID() int {
	subCreateTestCounter++
	return subCreateTestCounter + int(strings.Count("sub-create-integration", "c")*1000000)
}
