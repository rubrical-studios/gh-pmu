//go:build integration

package cmd

import (
	"fmt"
	"os/exec"
	"strings"
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

// TestRunEdit_Integration_AddLabel tests adding a label to an issue
func TestRunEdit_Integration_AddLabel(t *testing.T) {
	testutil.RequireTestEnv(t)

	title := fmt.Sprintf("Test Issue - AddLabel - %d", testUniqueID())
	createResult := testutil.RunCommand(t, "create", "--title", title, "--body", "Label test")
	testutil.AssertExitCode(t, createResult, 0)

	issueNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, issueNum)

	// Add a label
	editResult := testutil.RunCommand(t, "edit", fmt.Sprintf("%d", issueNum), "--label", "bug")
	testutil.AssertExitCode(t, editResult, 0)
	testutil.AssertContains(t, editResult.Stdout, "Updated issue")
	testutil.AssertContains(t, editResult.Stdout, "label")

	// Verify label was added via gh issue view
	env := testutil.RequireTestEnv(t)
	verifyCmd := exec.Command("gh", "issue", "view",
		"--repo", fmt.Sprintf("%s/%s", env.RepoOwner, env.RepoName),
		fmt.Sprintf("%d", issueNum),
		"--json", "labels",
		"--jq", ".labels[].name",
	)
	output, err := verifyCmd.Output()
	if err != nil {
		t.Fatalf("Failed to verify labels: %v", err)
	}
	if !strings.Contains(string(output), "bug") {
		t.Errorf("Expected issue to have 'bug' label, got labels: %s", output)
	}
}

// TestRunEdit_Integration_RemoveLabel tests removing a label from an issue
func TestRunEdit_Integration_RemoveLabel(t *testing.T) {
	env := testutil.RequireTestEnv(t)

	title := fmt.Sprintf("Test Issue - RemoveLabel - %d", testUniqueID())
	createResult := testutil.RunCommand(t, "create", "--title", title, "--body", "Label removal test")
	testutil.AssertExitCode(t, createResult, 0)

	issueNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, issueNum)

	// First add a label directly via gh CLI
	addCmd := exec.Command("gh", "issue", "edit",
		"--repo", fmt.Sprintf("%s/%s", env.RepoOwner, env.RepoName),
		fmt.Sprintf("%d", issueNum),
		"--add-label", "enhancement",
	)
	if output, err := addCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to pre-add label: %v\nOutput: %s", err, output)
	}

	// Remove the label via gh pmu edit
	editResult := testutil.RunCommand(t, "edit", fmt.Sprintf("%d", issueNum), "--remove-label", "enhancement")
	testutil.AssertExitCode(t, editResult, 0)
	testutil.AssertContains(t, editResult.Stdout, "Updated issue")
	testutil.AssertContains(t, editResult.Stdout, "label")

	// Verify label was removed
	verifyCmd := exec.Command("gh", "issue", "view",
		"--repo", fmt.Sprintf("%s/%s", env.RepoOwner, env.RepoName),
		fmt.Sprintf("%d", issueNum),
		"--json", "labels",
		"--jq", ".labels[].name",
	)
	output, err := verifyCmd.Output()
	if err != nil {
		t.Fatalf("Failed to verify labels: %v", err)
	}
	if strings.Contains(string(output), "enhancement") {
		t.Errorf("Expected 'enhancement' label to be removed, but still present: %s", output)
	}
}

// TestRunEdit_Integration_MultipleLabels tests adding multiple labels at once
func TestRunEdit_Integration_MultipleLabels(t *testing.T) {
	env := testutil.RequireTestEnv(t)

	title := fmt.Sprintf("Test Issue - MultiLabels - %d", testUniqueID())
	createResult := testutil.RunCommand(t, "create", "--title", title, "--body", "Multi-label test")
	testutil.AssertExitCode(t, createResult, 0)

	issueNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, issueNum)

	// Add multiple labels
	editResult := testutil.RunCommand(t, "edit", fmt.Sprintf("%d", issueNum),
		"--label", "bug", "--label", "documentation")
	testutil.AssertExitCode(t, editResult, 0)
	testutil.AssertContains(t, editResult.Stdout, "2 label(s) added")

	// Verify both labels present
	verifyCmd := exec.Command("gh", "issue", "view",
		"--repo", fmt.Sprintf("%s/%s", env.RepoOwner, env.RepoName),
		fmt.Sprintf("%d", issueNum),
		"--json", "labels",
		"--jq", ".labels[].name",
	)
	output, err := verifyCmd.Output()
	if err != nil {
		t.Fatalf("Failed to verify labels: %v", err)
	}
	labels := string(output)
	if !strings.Contains(labels, "bug") {
		t.Errorf("Expected 'bug' label, got: %s", labels)
	}
	if !strings.Contains(labels, "documentation") {
		t.Errorf("Expected 'documentation' label, got: %s", labels)
	}
}

// TestRunEdit_Integration_CrossRepo tests editing an issue with --repo flag
func TestRunEdit_Integration_CrossRepo(t *testing.T) {
	env := testutil.RequireTestEnv(t)

	// Create a test issue
	title := fmt.Sprintf("Test Issue - CrossRepo - %d", testUniqueID())
	createResult := testutil.RunCommand(t, "create", "--title", title, "--body", "Cross-repo test")
	testutil.AssertExitCode(t, createResult, 0)

	issueNum := testutil.ExtractIssueNumber(t, createResult.Stdout)
	defer testutil.DeleteTestIssue(t, issueNum)

	// Edit using explicit --repo flag (same repo, but exercises the code path)
	repoFlag := fmt.Sprintf("%s/%s", env.RepoOwner, env.RepoName)
	newTitle := fmt.Sprintf("CrossRepo Updated - %d", testUniqueID())
	editResult := testutil.RunCommand(t, "edit", fmt.Sprintf("%d", issueNum),
		"--repo", repoFlag, "--title", newTitle)
	testutil.AssertExitCode(t, editResult, 0)
	testutil.AssertContains(t, editResult.Stdout, "Updated issue")

	// Verify the update applied
	viewResult := testutil.RunCommand(t, "view", fmt.Sprintf("%d", issueNum))
	testutil.AssertExitCode(t, viewResult, 0)
	testutil.AssertContains(t, viewResult.Stdout, newTitle)
}
