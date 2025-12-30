package cmd

import (
	"bytes"
	"testing"

	"github.com/rubrical-studios/gh-pmu/internal/api"
	"github.com/rubrical-studios/gh-pmu/internal/config"
	"github.com/spf13/cobra"
)

// mockEditClient implements editClient for testing
type mockEditClient struct {
	issue             *api.Issue
	getIssueErr       error
	updateBodyErr     error
	updateTitleErr    error
	addLabelErr       error
	updateBodyCalls   []string
	updateTitleCalls  []string
	addLabelCalls     []string
}

func (m *mockEditClient) GetIssueByNumber(owner, repo string, number int) (*api.Issue, error) {
	if m.getIssueErr != nil {
		return nil, m.getIssueErr
	}
	return m.issue, nil
}

func (m *mockEditClient) UpdateIssueBody(issueID, body string) error {
	m.updateBodyCalls = append(m.updateBodyCalls, body)
	return m.updateBodyErr
}

func (m *mockEditClient) UpdateIssueTitle(issueID, title string) error {
	m.updateTitleCalls = append(m.updateTitleCalls, title)
	return m.updateTitleErr
}

func (m *mockEditClient) AddLabelToIssue(issueID, labelName string) error {
	m.addLabelCalls = append(m.addLabelCalls, labelName)
	return m.addLabelErr
}

func setupMockForEdit() *mockEditClient {
	return &mockEditClient{
		issue: &api.Issue{
			ID:     "ISSUE_123",
			Number: 123,
			Title:  "Test Issue",
			URL:    "https://github.com/testowner/testrepo/issues/123",
		},
	}
}

func testEditConfig() *config.Config {
	return &config.Config{
		Project: config.Project{
			Owner:  "testowner",
			Number: 1,
		},
		Repositories: []string{"testowner/testrepo"},
	}
}

func newTestEditCmd() (*cobra.Command, *bytes.Buffer) {
	cmd := &cobra.Command{Use: "edit"}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	return cmd, buf
}

// ============================================================================
// Edit Command Structure Tests
// ============================================================================

func TestEditCommand_Structure(t *testing.T) {
	cmd := NewRootCommand()
	editCmd, _, err := cmd.Find([]string{"edit"})
	if err != nil {
		t.Fatalf("edit command not found: %v", err)
	}

	if editCmd.Use != "edit <issue-number>" {
		t.Errorf("Expected Use 'edit <issue-number>', got '%s'", editCmd.Use)
	}

	// Requires exactly 1 argument
	if err := editCmd.Args(editCmd, []string{}); err == nil {
		t.Error("Expected error when no arguments provided")
	}
	if err := editCmd.Args(editCmd, []string{"123"}); err != nil {
		t.Errorf("Unexpected error with one argument: %v", err)
	}
}

func TestEditCommand_Flags(t *testing.T) {
	cmd := NewRootCommand()
	editCmd, _, err := cmd.Find([]string{"edit"})
	if err != nil {
		t.Fatalf("edit command not found: %v", err)
	}

	tests := []struct {
		flag      string
		shorthand string
	}{
		{"title", "t"},
		{"body", "b"},
		{"body-file", "F"},
		{"label", "l"},
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			flag := editCmd.Flags().Lookup(tt.flag)
			if flag == nil {
				t.Fatalf("Expected --%s flag to exist", tt.flag)
			}
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("Expected --%s shorthand to be '%s', got '%s'", tt.flag, tt.shorthand, flag.Shorthand)
			}
		})
	}
}

// ============================================================================
// Edit Command Behavior Tests
// ============================================================================

func TestRunEditWithDeps_UpdatesTitle(t *testing.T) {
	// ARRANGE
	mock := setupMockForEdit()
	cfg := testEditConfig()
	cmd, buf := newTestEditCmd()
	opts := &editOptions{
		issueNumber: 123,
		title:       "New Title",
	}

	// ACT
	err := runEditWithDeps(cmd, opts, cfg, mock, "testowner", "testrepo")

	// ASSERT
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(mock.updateTitleCalls) != 1 || mock.updateTitleCalls[0] != "New Title" {
		t.Errorf("Expected title update to 'New Title', got: %v", mock.updateTitleCalls)
	}
	output := buf.String()
	if !contains(output, "Updated issue #123") {
		t.Errorf("Expected output to contain 'Updated issue #123', got: %s", output)
	}
}

func TestRunEditWithDeps_UpdatesBody(t *testing.T) {
	// ARRANGE
	mock := setupMockForEdit()
	cfg := testEditConfig()
	cmd, _ := newTestEditCmd()
	opts := &editOptions{
		issueNumber: 123,
		body:        "New body content",
	}

	// ACT
	err := runEditWithDeps(cmd, opts, cfg, mock, "testowner", "testrepo")

	// ASSERT
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(mock.updateBodyCalls) != 1 || mock.updateBodyCalls[0] != "New body content" {
		t.Errorf("Expected body update to 'New body content', got: %v", mock.updateBodyCalls)
	}
}

func TestRunEditWithDeps_AddsLabels(t *testing.T) {
	// ARRANGE
	mock := setupMockForEdit()
	cfg := testEditConfig()
	cmd, buf := newTestEditCmd()
	opts := &editOptions{
		issueNumber: 123,
		addLabels:   []string{"bug", "urgent"},
	}

	// ACT
	err := runEditWithDeps(cmd, opts, cfg, mock, "testowner", "testrepo")

	// ASSERT
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(mock.addLabelCalls) != 2 {
		t.Errorf("Expected 2 label calls, got: %d", len(mock.addLabelCalls))
	}
	output := buf.String()
	if !contains(output, "2 label(s)") {
		t.Errorf("Expected output to contain '2 label(s)', got: %s", output)
	}
}

func TestRunEditWithDeps_MultipleUpdates(t *testing.T) {
	// ARRANGE
	mock := setupMockForEdit()
	cfg := testEditConfig()
	cmd, buf := newTestEditCmd()
	opts := &editOptions{
		issueNumber: 123,
		title:       "New Title",
		body:        "New body",
		addLabels:   []string{"fix"},
	}

	// ACT
	err := runEditWithDeps(cmd, opts, cfg, mock, "testowner", "testrepo")

	// ASSERT
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(mock.updateTitleCalls) != 1 {
		t.Errorf("Expected 1 title update, got: %d", len(mock.updateTitleCalls))
	}
	if len(mock.updateBodyCalls) != 1 {
		t.Errorf("Expected 1 body update, got: %d", len(mock.updateBodyCalls))
	}
	if len(mock.addLabelCalls) != 1 {
		t.Errorf("Expected 1 label call, got: %d", len(mock.addLabelCalls))
	}
	output := buf.String()
	if !contains(output, "title") || !contains(output, "body") || !contains(output, "label") {
		t.Errorf("Expected output to contain all update types, got: %s", output)
	}
}

func TestRunEditWithDeps_RequiresAtLeastOneOption(t *testing.T) {
	// ARRANGE
	mock := setupMockForEdit()
	cfg := testEditConfig()
	cmd, _ := newTestEditCmd()
	opts := &editOptions{
		issueNumber: 123,
		// No title, body, bodyFile, or labels
	}

	// ACT
	err := runEditWithDeps(cmd, opts, cfg, mock, "testowner", "testrepo")

	// ASSERT
	if err == nil {
		t.Fatal("Expected error when no options provided")
	}
	if !contains(err.Error(), "at least one of") {
		t.Errorf("Expected 'at least one of' error, got: %s", err.Error())
	}
}

func TestRunEditWithDeps_CannotUseBothBodyAndBodyFile(t *testing.T) {
	// ARRANGE
	mock := setupMockForEdit()
	cfg := testEditConfig()
	cmd, _ := newTestEditCmd()
	opts := &editOptions{
		issueNumber: 123,
		body:        "inline body",
		bodyFile:    "file.md",
	}

	// ACT
	err := runEditWithDeps(cmd, opts, cfg, mock, "testowner", "testrepo")

	// ASSERT
	if err == nil {
		t.Fatal("Expected error when both body and body-file provided")
	}
	if !contains(err.Error(), "cannot use --body and --body-file together") {
		t.Errorf("Expected 'cannot use --body and --body-file together' error, got: %s", err.Error())
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
