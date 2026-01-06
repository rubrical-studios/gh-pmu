package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/rubrical-studios/gh-pmu/internal/api"
)

// mockCommentClient implements commentClient for testing
type mockCommentClient struct {
	issue   *api.Issue
	comment *api.Comment

	// Error injection
	getIssueErr      error
	addCommentErr    error
}

func newMockCommentClient() *mockCommentClient {
	return &mockCommentClient{
		issue: &api.Issue{
			ID:     "issue-1",
			Number: 42,
			Title:  "Test Issue",
			URL:    "https://github.com/owner/repo/issues/42",
		},
		comment: &api.Comment{
			ID:        "IC_comment123",
			Body:      "Test comment",
			Author:    "testuser",
			CreatedAt: "2026-01-05T10:00:00Z",
		},
	}
}

func (m *mockCommentClient) GetIssueByNumber(owner, repo string, number int) (*api.Issue, error) {
	if m.getIssueErr != nil {
		return nil, m.getIssueErr
	}
	return m.issue, nil
}

func (m *mockCommentClient) AddIssueComment(issueID, body string) (*api.Comment, error) {
	if m.addCommentErr != nil {
		return nil, m.addCommentErr
	}
	m.comment.Body = body
	return m.comment, nil
}

// ============================================================================
// Command Flag Tests
// ============================================================================

func TestCommentCommand_Exists(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"comment", "--help"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("comment command should exist: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "comment") {
		t.Error("Expected help output to mention 'comment'")
	}
}

func TestCommentCommand_HasBodyFlag(t *testing.T) {
	cmd := NewRootCommand()
	commentCmd, _, err := cmd.Find([]string{"comment"})
	if err != nil {
		t.Fatalf("comment command not found: %v", err)
	}

	flag := commentCmd.Flags().Lookup("body")
	if flag == nil {
		t.Fatal("Expected --body flag to exist")
	}
	if flag.Shorthand != "b" {
		t.Errorf("Expected shorthand 'b', got '%s'", flag.Shorthand)
	}
}

func TestCommentCommand_HasBodyFileFlag(t *testing.T) {
	cmd := NewRootCommand()
	commentCmd, _, err := cmd.Find([]string{"comment"})
	if err != nil {
		t.Fatalf("comment command not found: %v", err)
	}

	flag := commentCmd.Flags().Lookup("body-file")
	if flag == nil {
		t.Fatal("Expected --body-file flag to exist")
	}
	if flag.Shorthand != "F" {
		t.Errorf("Expected shorthand 'F', got '%s'", flag.Shorthand)
	}
}

func TestCommentCommand_HasBodyStdinFlag(t *testing.T) {
	cmd := NewRootCommand()
	commentCmd, _, err := cmd.Find([]string{"comment"})
	if err != nil {
		t.Fatalf("comment command not found: %v", err)
	}

	flag := commentCmd.Flags().Lookup("body-stdin")
	if flag == nil {
		t.Fatal("Expected --body-stdin flag to exist")
	}
	if flag.Value.Type() != "bool" {
		t.Errorf("Expected --body-stdin to be bool, got %s", flag.Value.Type())
	}
}

func TestCommentCommand_HasRepoFlag(t *testing.T) {
	cmd := NewRootCommand()
	commentCmd, _, err := cmd.Find([]string{"comment"})
	if err != nil {
		t.Fatalf("comment command not found: %v", err)
	}

	flag := commentCmd.Flags().Lookup("repo")
	if flag == nil {
		t.Fatal("Expected --repo flag to exist")
	}
	if flag.Shorthand != "R" {
		t.Errorf("Expected shorthand 'R', got '%s'", flag.Shorthand)
	}
}

func TestCommentCommand_RequiresIssueNumber(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"comment", "--body", "test"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when no issue number provided")
	}
}

// ============================================================================
// runCommentWithDeps Tests
// ============================================================================

func TestRunCommentWithDeps_Success(t *testing.T) {
	mock := newMockCommentClient()
	cmd := newCommentCommand()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	opts := &commentOptions{issueNumber: 42, body: "Test comment"}
	err := runCommentWithDeps(cmd, opts, mock, "owner", "repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Added comment to issue #42") {
		t.Errorf("Expected success message, got: %s", output)
	}
}

func TestRunCommentWithDeps_NoBodySource(t *testing.T) {
	mock := newMockCommentClient()
	cmd := newCommentCommand()

	opts := &commentOptions{issueNumber: 42} // No body source
	err := runCommentWithDeps(cmd, opts, mock, "owner", "repo")

	if err == nil {
		t.Fatal("expected error for no body source")
	}
	if !strings.Contains(err.Error(), "one of --body, --body-file (-F), or --body-stdin is required") {
		t.Errorf("expected body source required error, got: %v", err)
	}
}

func TestRunCommentWithDeps_MultipleBodySources(t *testing.T) {
	mock := newMockCommentClient()
	cmd := newCommentCommand()

	opts := &commentOptions{
		issueNumber: 42,
		body:        "Body from flag",
		bodyStdin:   true,
	}
	err := runCommentWithDeps(cmd, opts, mock, "owner", "repo")

	if err == nil {
		t.Fatal("expected error for multiple body sources")
	}
	if !strings.Contains(err.Error(), "only one of --body, --body-file (-F), or --body-stdin can be used") {
		t.Errorf("expected mutual exclusivity error, got: %v", err)
	}
}

func TestRunCommentWithDeps_EmptyBody(t *testing.T) {
	mock := newMockCommentClient()
	cmd := newCommentCommand()

	opts := &commentOptions{issueNumber: 42, body: "   "} // Whitespace only
	err := runCommentWithDeps(cmd, opts, mock, "owner", "repo")

	if err == nil {
		t.Fatal("expected error for empty body")
	}
	if !strings.Contains(err.Error(), "comment body cannot be empty") {
		t.Errorf("expected empty body error, got: %v", err)
	}
}

func TestRunCommentWithDeps_GetIssueError(t *testing.T) {
	mock := newMockCommentClient()
	mock.getIssueErr = errors.New("issue not found")
	cmd := newCommentCommand()

	opts := &commentOptions{issueNumber: 42, body: "Test comment"}
	err := runCommentWithDeps(cmd, opts, mock, "owner", "repo")

	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "failed to get issue #42") {
		t.Errorf("expected 'failed to get issue' error, got: %v", err)
	}
}

func TestRunCommentWithDeps_AddCommentError(t *testing.T) {
	mock := newMockCommentClient()
	mock.addCommentErr = errors.New("API error")
	cmd := newCommentCommand()

	opts := &commentOptions{issueNumber: 42, body: "Test comment"}
	err := runCommentWithDeps(cmd, opts, mock, "owner", "repo")

	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "failed to add comment") {
		t.Errorf("expected 'failed to add comment' error, got: %v", err)
	}
}

func TestRunCommentWithDeps_BodyAndBodyFileMutuallyExclusive(t *testing.T) {
	mock := newMockCommentClient()
	cmd := newCommentCommand()

	opts := &commentOptions{
		issueNumber: 42,
		body:        "Body from flag",
		bodyFile:    "somefile.md",
	}
	err := runCommentWithDeps(cmd, opts, mock, "owner", "repo")

	if err == nil {
		t.Fatal("expected error for multiple body sources")
	}
	if !strings.Contains(err.Error(), "only one of") {
		t.Errorf("expected mutual exclusivity error, got: %v", err)
	}
}

func TestRunCommentWithDeps_BodyFileAndBodyStdinMutuallyExclusive(t *testing.T) {
	mock := newMockCommentClient()
	cmd := newCommentCommand()

	opts := &commentOptions{
		issueNumber: 42,
		bodyFile:    "somefile.md",
		bodyStdin:   true,
	}
	err := runCommentWithDeps(cmd, opts, mock, "owner", "repo")

	if err == nil {
		t.Fatal("expected error for multiple body sources")
	}
	if !strings.Contains(err.Error(), "only one of") {
		t.Errorf("expected mutual exclusivity error, got: %v", err)
	}
}

// ============================================================================
// extractCommentID Tests
// ============================================================================

func TestExtractCommentID_ReturnsNodeID(t *testing.T) {
	// For now, extractCommentID just returns the node ID
	nodeID := "IC_kwDOLF2XKM5nQi8x"
	result := extractCommentID(nodeID)
	if result != nodeID {
		t.Errorf("Expected %q, got %q", nodeID, result)
	}
}

func TestExtractCommentID_EmptyString(t *testing.T) {
	result := extractCommentID("")
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}
