//go:build integration

package cmd

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/rubrical-studios/gh-pmu/internal/testutil"
)

// TestRunEdit_Integration_BodyStdin tests --body-stdin flag with piped input
func TestRunEdit_Integration_BodyStdin(t *testing.T) {
	env := testutil.RequireTestEnv(t)

	// First create a test issue
	title := fmt.Sprintf("Test Issue - BodyStdin - %d", testUniqueID())
	createResult := testutil.RunCommand(t, "create", "--title", title, "--body", "Original body")
	testutil.AssertExitCode(t, createResult, 0)

	issueNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, issueNum)

	// Update using --body-stdin with piped content
	newBody := "Updated body via stdin"
	cmd := exec.Command("sh", "-c", fmt.Sprintf("echo '%s' | gh pmu edit %d --body-stdin", newBody, issueNum))
	cmd.Dir = env.WorkDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run edit with --body-stdin: %v\nOutput: %s", err, output)
	}

	// Verify the body was updated
	viewResult := testutil.RunCommand(t, "view", fmt.Sprintf("%d", issueNum), "--body-stdout")
	testutil.AssertExitCode(t, viewResult, 0)
	testutil.AssertContains(t, viewResult.Stdout, newBody)
}

// TestRunEdit_Integration_Title tests updating issue title
func TestRunEdit_Integration_Title(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create a test issue
	title := fmt.Sprintf("Test Issue - EditTitle - %d", testUniqueID())
	createResult := testutil.RunCommand(t, "create", "--title", title)
	testutil.AssertExitCode(t, createResult, 0)

	issueNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, issueNum)

	// Update the title
	newTitle := fmt.Sprintf("Updated Title - %d", testUniqueID())
	editResult := testutil.RunCommand(t, "edit", fmt.Sprintf("%d", issueNum), "--title", newTitle)
	testutil.AssertExitCode(t, editResult, 0)
	testutil.AssertContains(t, editResult.Stdout, "Updated issue")
	testutil.AssertContains(t, editResult.Stdout, "title")

	// Verify the title was updated
	viewResult := testutil.RunCommand(t, "view", fmt.Sprintf("%d", issueNum))
	testutil.AssertExitCode(t, viewResult, 0)
	testutil.AssertContains(t, viewResult.Stdout, newTitle)
}

// TestRunEdit_Integration_Body tests updating issue body with --body flag
func TestRunEdit_Integration_Body(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Create a test issue
	title := fmt.Sprintf("Test Issue - EditBody - %d", testUniqueID())
	createResult := testutil.RunCommand(t, "create", "--title", title, "--body", "Original body")
	testutil.AssertExitCode(t, createResult, 0)

	issueNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, issueNum)

	// Update the body
	newBody := "Updated body content"
	editResult := testutil.RunCommand(t, "edit", fmt.Sprintf("%d", issueNum), "--body", newBody)
	testutil.AssertExitCode(t, editResult, 0)
	testutil.AssertContains(t, editResult.Stdout, "Updated issue")
	testutil.AssertContains(t, editResult.Stdout, "body")

	// Verify the body was updated
	viewResult := testutil.RunCommand(t, "view", fmt.Sprintf("%d", issueNum), "--body-stdout")
	testutil.AssertExitCode(t, viewResult, 0)
	testutil.AssertContains(t, viewResult.Stdout, newBody)
}

// TestRunEdit_Integration_BodyStdinMutualExclusion tests --body-stdin exclusivity
func TestRunEdit_Integration_BodyStdinMutualExclusion(t *testing.T) {
	testutil.RequireTestEnv(t)

	// Try to use --body-stdin with --body (should fail)
	result := testutil.RunCommand(t, "edit", "1", "--body", "test", "--body-stdin")

	// Should fail with mutual exclusion error
	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code when using --body-stdin with --body")
	}
	testutil.AssertContains(t, result.Stderr, "cannot use --body-stdin")
}
