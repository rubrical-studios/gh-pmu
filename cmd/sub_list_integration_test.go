//go:build integration

package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/rubrical-studios/gh-pmu/internal/testutil"
)

// TestRunSubList_Integration_ListSubIssues tests listing sub-issues of a parent
func TestRunSubList_Integration_ListSubIssues(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("Test SubList Parent - List - %d", subListTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create first sub-issue
	sub1Title := fmt.Sprintf("Test SubList Child1 - List - %d", subListTestID())
	sub1Result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", sub1Title,
	)
	testutil.AssertExitCode(t, sub1Result, 0)

	sub1Num := testutil.ExtractIssueNumber(t, sub1Result.Stdout)
	defer testutil.DeleteTestIssue(t, sub1Num)

	// Create second sub-issue
	sub2Title := fmt.Sprintf("Test SubList Child2 - List - %d", subListTestID())
	sub2Result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", sub2Title,
	)
	testutil.AssertExitCode(t, sub2Result, 0)

	sub2Num := testutil.ExtractIssueNumber(t, sub2Result.Stdout)
	defer testutil.DeleteTestIssue(t, sub2Num)

	// List sub-issues
	result := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum))
	testutil.AssertExitCode(t, result, 0)

	// Verify both sub-issues are listed
	testutil.AssertContains(t, result.Stdout, sub1Title)
	testutil.AssertContains(t, result.Stdout, sub2Title)
	testutil.AssertContains(t, result.Stdout, fmt.Sprintf("#%d", sub1Num))
	testutil.AssertContains(t, result.Stdout, fmt.Sprintf("#%d", sub2Num))
}

// TestRunSubList_Integration_ShowCompletionCount tests that completion count is shown
func TestRunSubList_Integration_ShowCompletionCount(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("Test SubList Parent - Completion - %d", subListTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create two sub-issues
	sub1Title := fmt.Sprintf("Test SubList Child1 - Completion - %d", subListTestID())
	sub1Result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", sub1Title,
	)
	testutil.AssertExitCode(t, sub1Result, 0)

	sub1Num := testutil.ExtractIssueNumber(t, sub1Result.Stdout)
	defer testutil.DeleteTestIssue(t, sub1Num)

	sub2Title := fmt.Sprintf("Test SubList Child2 - Completion - %d", subListTestID())
	sub2Result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", sub2Title,
	)
	testutil.AssertExitCode(t, sub2Result, 0)

	sub2Num := testutil.ExtractIssueNumber(t, sub2Result.Stdout)
	defer testutil.DeleteTestIssue(t, sub2Num)

	// List sub-issues - should show 0/2 complete
	result := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum))
	testutil.AssertExitCode(t, result, 0)
	testutil.AssertContains(t, result.Stdout, "0/2 complete")
}

// TestRunSubList_Integration_JSONOutput tests --json output format
func TestRunSubList_Integration_JSONOutput(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("Test SubList Parent - JSON - %d", subListTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create a sub-issue
	subTitle := fmt.Sprintf("Test SubList Child - JSON - %d", subListTestID())
	subResult := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", subTitle,
	)
	testutil.AssertExitCode(t, subResult, 0)

	subNum := testutil.ExtractIssueNumber(t, subResult.Stdout)
	defer testutil.DeleteTestIssue(t, subNum)

	// List with --json
	result := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum), "--json")
	testutil.AssertExitCode(t, result, 0)

	// Verify it's valid JSON
	var jsonOutput map[string]interface{}
	err := json.Unmarshal([]byte(result.Stdout), &jsonOutput)
	if err != nil {
		t.Errorf("Expected valid JSON output, got error: %v\nOutput: %s", err, result.Stdout)
	}

	// Verify expected JSON structure (uses "issue" and "children", not "parent" and "subIssues")
	if _, ok := jsonOutput["issue"]; !ok {
		t.Error("JSON output missing 'issue' field")
	}
	if _, ok := jsonOutput["children"]; !ok {
		t.Error("JSON output missing 'children' field")
	}
	if _, ok := jsonOutput["summary"]; !ok {
		t.Error("JSON output missing 'summary' field")
	}

	// Verify issue contains expected data
	issue, ok := jsonOutput["issue"].(map[string]interface{})
	if !ok {
		t.Error("JSON 'issue' field is not an object")
	} else {
		if issue["number"] == nil {
			t.Error("JSON issue missing 'number' field")
		}
		if issue["title"] == nil {
			t.Error("JSON issue missing 'title' field")
		}
	}

	// Verify summary contains expected data
	summary, ok := jsonOutput["summary"].(map[string]interface{})
	if !ok {
		t.Error("JSON 'summary' field is not an object")
	} else {
		if summary["total"] == nil {
			t.Error("JSON summary missing 'total' field")
		}
		if summary["open"] == nil {
			t.Error("JSON summary missing 'open' field")
		}
		if summary["closed"] == nil {
			t.Error("JSON summary missing 'closed' field")
		}
	}
}

// TestRunSubList_Integration_NoSubIssues tests case when parent has no sub-issues
func TestRunSubList_Integration_NoSubIssues(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue with no sub-issues
	parentTitle := fmt.Sprintf("Test SubList Parent - NoSubs - %d", subListTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// List sub-issues (should be empty)
	result := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum))
	testutil.AssertExitCode(t, result, 0)

	// Should show no children message or empty
	testutil.AssertContains(t, result.Stdout, parentTitle)
	// The output should indicate there are no sub-issues
	testutil.AssertContains(t, result.Stdout, "No sub-issues found")
}

// TestRunSubList_Integration_ParentNotFound tests error when parent doesn't exist
func TestRunSubList_Integration_ParentNotFound(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Try to list sub-issues for non-existent parent
	result := testutil.RunCommand(t, "sub", "list", "999999")

	// Should fail
	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code for non-existent parent")
	}

	testutil.AssertContains(t, result.Stderr, "failed to get issue")
}

// TestRunSubList_Integration_WithHashPrefix tests using # prefix on issue number
func TestRunSubList_Integration_WithHashPrefix(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("Test SubList Parent - Hash - %d", subListTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create a sub-issue
	subTitle := fmt.Sprintf("Test SubList Child - Hash - %d", subListTestID())
	subResult := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", subTitle,
	)
	testutil.AssertExitCode(t, subResult, 0)

	subNum := testutil.ExtractIssueNumber(t, subResult.Stdout)
	defer testutil.DeleteTestIssue(t, subNum)

	// List using # prefix
	result := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("#%d", parentNum))
	testutil.AssertExitCode(t, result, 0)

	testutil.AssertContains(t, result.Stdout, subTitle)
}

// TestRunSubList_Integration_StateFilter tests --state filter
func TestRunSubList_Integration_StateFilter(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("Test SubList Parent - State - %d", subListTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create a sub-issue
	subTitle := fmt.Sprintf("Test SubList Child - State - %d", subListTestID())
	subResult := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", subTitle,
	)
	testutil.AssertExitCode(t, subResult, 0)

	subNum := testutil.ExtractIssueNumber(t, subResult.Stdout)
	defer testutil.DeleteTestIssue(t, subNum)

	// List with --state open (should include the sub-issue)
	openResult := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum), "--state", "open")
	testutil.AssertExitCode(t, openResult, 0)
	testutil.AssertContains(t, openResult.Stdout, subTitle)

	// List with --state closed (should not include the sub-issue)
	closedResult := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum), "--state", "closed")
	testutil.AssertExitCode(t, closedResult, 0)
	testutil.AssertNotContains(t, closedResult.Stdout, subTitle)

	// List with --state all (should include the sub-issue)
	allResult := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum), "--state", "all")
	testutil.AssertExitCode(t, allResult, 0)
	testutil.AssertContains(t, allResult.Stdout, subTitle)
}

// TestRunSubList_Integration_LimitOption tests --limit option
func TestRunSubList_Integration_LimitOption(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("Test SubList Parent - Limit - %d", subListTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create 3 sub-issues
	var subNums []int
	for i := 1; i <= 3; i++ {
		subTitle := fmt.Sprintf("Test SubList Child%d - Limit - %d", i, subListTestID())
		subResult := testutil.RunCommand(t, "sub", "create",
			"--parent", fmt.Sprintf("%d", parentNum),
			"--title", subTitle,
		)
		testutil.AssertExitCode(t, subResult, 0)

		subNum := testutil.ExtractIssueNumber(t, subResult.Stdout)
		subNums = append(subNums, subNum)
		defer testutil.DeleteTestIssue(t, subNum)
	}

	// List with --limit 2 (should show only 2 sub-issues)
	result := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum), "--limit", "2")
	testutil.AssertExitCode(t, result, 0)

	// Count how many sub-issues are shown
	// Should contain at most 2 issue references
	lines := strings.Split(result.Stdout, "\n")
	issueCount := 0
	for _, line := range lines {
		for _, subNum := range subNums {
			if strings.Contains(line, fmt.Sprintf("#%d", subNum)) {
				issueCount++
				break
			}
		}
	}

	if issueCount > 2 {
		t.Errorf("Expected at most 2 sub-issues with --limit 2, got %d", issueCount)
	}
}

// TestRunSubList_Integration_RelationParent tests --relation parent option
func TestRunSubList_Integration_RelationParent(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("Test SubList Parent - RelParent - %d", subListTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create a sub-issue
	subTitle := fmt.Sprintf("Test SubList Child - RelParent - %d", subListTestID())
	subResult := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", subTitle,
	)
	testutil.AssertExitCode(t, subResult, 0)

	subNum := testutil.ExtractIssueNumber(t, subResult.Stdout)
	defer testutil.DeleteTestIssue(t, subNum)

	// List parent from child's perspective
	result := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", subNum), "--relation", "parent")
	testutil.AssertExitCode(t, result, 0)

	// Should show parent info
	testutil.AssertContains(t, result.Stdout, parentTitle)
	testutil.AssertContains(t, result.Stdout, fmt.Sprintf("#%d", parentNum))
}

// TestRunSubList_Integration_RelationSiblings tests --relation siblings option
func TestRunSubList_Integration_RelationSiblings(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("Test SubList Parent - Siblings - %d", subListTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Create first sub-issue
	sub1Title := fmt.Sprintf("Test SubList Child1 - Siblings - %d", subListTestID())
	sub1Result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", sub1Title,
	)
	testutil.AssertExitCode(t, sub1Result, 0)

	sub1Num := testutil.ExtractIssueNumber(t, sub1Result.Stdout)
	defer testutil.DeleteTestIssue(t, sub1Num)

	// Create second sub-issue (sibling)
	sub2Title := fmt.Sprintf("Test SubList Child2 - Siblings - %d", subListTestID())
	sub2Result := testutil.RunCommand(t, "sub", "create",
		"--parent", fmt.Sprintf("%d", parentNum),
		"--title", sub2Title,
	)
	testutil.AssertExitCode(t, sub2Result, 0)

	sub2Num := testutil.ExtractIssueNumber(t, sub2Result.Stdout)
	defer testutil.DeleteTestIssue(t, sub2Num)

	// List siblings from first child's perspective
	result := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", sub1Num), "--relation", "siblings")
	testutil.AssertExitCode(t, result, 0)

	// Should show second child as sibling in the Siblings section
	testutil.AssertContains(t, result.Stdout, sub2Title)
	// The output header will contain sub1Title, but the Siblings section should not list it
	// Check that sibling list has sub2 but not sub1
	lines := strings.Split(result.Stdout, "\n")
	inSiblingsSection := false
	for _, line := range lines {
		if strings.Contains(line, "Siblings:") {
			inSiblingsSection = true
			continue
		}
		if inSiblingsSection {
			// Stop if we hit another section or blank line followed by content
			if strings.TrimSpace(line) == "" {
				continue
			}
			// Should not find sub1 in siblings list
			if strings.Contains(line, sub1Title) {
				t.Errorf("Found self (%s) in siblings list", sub1Title)
			}
		}
	}
}

// TestRunSubList_Integration_InvalidState tests error for invalid state option
func TestRunSubList_Integration_InvalidState(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("Test SubList Parent - InvalidState - %d", subListTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Try with invalid state
	result := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum), "--state", "invalid")

	// Should fail
	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code for invalid state")
	}

	testutil.AssertContains(t, result.Stderr, "invalid state")
}

// TestRunSubList_Integration_InvalidRelation tests error for invalid relation option
func TestRunSubList_Integration_InvalidRelation(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create parent issue
	parentTitle := fmt.Sprintf("Test SubList Parent - InvalidRel - %d", subListTestID())
	parentResult := testutil.RunCommand(t, "create", "--title", parentTitle)
	testutil.AssertExitCode(t, parentResult, 0)

	parentNum := testutil.ExtractIssueNumber(t, parentResult.Stdout)
	defer testutil.DeleteTestIssue(t, parentNum)

	// Try with invalid relation
	result := testutil.RunCommand(t, "sub", "list", fmt.Sprintf("%d", parentNum), "--relation", "invalid")

	// Should fail
	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code for invalid relation")
	}

	testutil.AssertContains(t, result.Stderr, "invalid relation")
}

// subListTestID returns a unique identifier for sub list test issues
var subListTestCounter int

func subListTestID() int {
	subListTestCounter++
	return subListTestCounter + int(strings.Count("sub-list-integration", "i")*1000000)
}
