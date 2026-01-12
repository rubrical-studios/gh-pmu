package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rubrical-studios/gh-pmu/internal/api"
	"github.com/rubrical-studios/gh-pmu/internal/ui"
	"github.com/spf13/cobra"
)

// mockViewClient implements viewClient for testing
type mockViewClient struct {
	issue       *api.Issue
	fieldValues []api.FieldValue
	subIssues   []api.SubIssue
	parentIssue *api.Issue
	comments    []api.Comment

	// Error injection
	getIssueErr                  error
	getIssueWithProjectFieldsErr error
	getSubIssuesErr              error
	getParentIssueErr            error
	getIssueCommentsErr          error
}

func newMockViewClient() *mockViewClient {
	return &mockViewClient{
		issue: &api.Issue{
			Number: 42,
			Title:  "Test Issue",
			State:  "OPEN",
			URL:    "https://github.com/owner/repo/issues/42",
			Author: api.Actor{Login: "testuser"},
		},
		fieldValues: []api.FieldValue{},
		subIssues:   []api.SubIssue{},
	}
}

func (m *mockViewClient) GetIssue(owner, repo string, number int) (*api.Issue, error) {
	if m.getIssueErr != nil {
		return nil, m.getIssueErr
	}
	return m.issue, nil
}

func (m *mockViewClient) GetIssueWithProjectFields(owner, repo string, number int) (*api.Issue, []api.FieldValue, error) {
	if m.getIssueWithProjectFieldsErr != nil {
		return nil, nil, m.getIssueWithProjectFieldsErr
	}
	return m.issue, m.fieldValues, nil
}

func (m *mockViewClient) GetSubIssues(owner, repo string, number int) ([]api.SubIssue, error) {
	if m.getSubIssuesErr != nil {
		return nil, m.getSubIssuesErr
	}
	return m.subIssues, nil
}

func (m *mockViewClient) GetParentIssue(owner, repo string, number int) (*api.Issue, error) {
	if m.getParentIssueErr != nil {
		return nil, m.getParentIssueErr
	}
	return m.parentIssue, nil
}

func (m *mockViewClient) GetIssueComments(owner, repo string, number int) ([]api.Comment, error) {
	if m.getIssueCommentsErr != nil {
		return nil, m.getIssueCommentsErr
	}
	return m.comments, nil
}

// ============================================================================
// runViewWithDeps Tests
// ============================================================================

func TestRunViewWithDeps_Success(t *testing.T) {
	mock := newMockViewClient()
	mock.issue = &api.Issue{
		Number: 42,
		Title:  "Test Issue",
		State:  "OPEN",
		URL:    "https://github.com/owner/repo/issues/42",
		Author: api.Actor{Login: "testuser"},
	}
	mock.fieldValues = []api.FieldValue{
		{Field: "Status", Value: "In Progress"},
	}

	cmd := newViewCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	opts := &viewOptions{}
	err := runViewWithDeps(cmd, opts, mock, "owner", "repo", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunViewWithDeps_GetIssueError(t *testing.T) {
	mock := newMockViewClient()
	mock.getIssueWithProjectFieldsErr = errors.New("issue not found")

	cmd := newViewCommand()
	opts := &viewOptions{}
	err := runViewWithDeps(cmd, opts, mock, "owner", "repo", 42)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to get issue") {
		t.Errorf("expected 'failed to get issue' error, got: %v", err)
	}
}

func TestRunViewWithDeps_JSONOutput(t *testing.T) {
	mock := newMockViewClient()
	mock.issue = &api.Issue{
		Number: 42,
		Title:  "JSON Test Issue",
		State:  "OPEN",
		URL:    "https://github.com/owner/repo/issues/42",
		Author: api.Actor{Login: "testuser"},
	}

	cmd := newViewCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	opts := &viewOptions{json: true}
	err := runViewWithDeps(cmd, opts, mock, "owner", "repo", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunViewWithDeps_WithSubIssues(t *testing.T) {
	mock := newMockViewClient()
	mock.issue = &api.Issue{
		Number: 42,
		Title:  "Parent Issue",
		State:  "OPEN",
		URL:    "https://github.com/owner/repo/issues/42",
		Author: api.Actor{Login: "testuser"},
	}
	mock.subIssues = []api.SubIssue{
		{Number: 43, Title: "Sub 1", State: "CLOSED"},
		{Number: 44, Title: "Sub 2", State: "OPEN"},
	}

	cmd := newViewCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	opts := &viewOptions{}
	err := runViewWithDeps(cmd, opts, mock, "owner", "repo", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunViewWithDeps_WithParentIssue(t *testing.T) {
	mock := newMockViewClient()
	mock.issue = &api.Issue{
		Number: 43,
		Title:  "Sub-Issue",
		State:  "OPEN",
		URL:    "https://github.com/owner/repo/issues/43",
		Author: api.Actor{Login: "testuser"},
	}
	mock.parentIssue = &api.Issue{
		Number: 42,
		Title:  "Parent Issue",
		URL:    "https://github.com/owner/repo/issues/42",
	}

	cmd := newViewCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	opts := &viewOptions{}
	err := runViewWithDeps(cmd, opts, mock, "owner", "repo", 43)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunViewWithDeps_WithComments(t *testing.T) {
	mock := newMockViewClient()
	mock.issue = &api.Issue{
		Number: 42,
		Title:  "Test Issue",
		State:  "OPEN",
		URL:    "https://github.com/owner/repo/issues/42",
		Author: api.Actor{Login: "testuser"},
	}
	mock.comments = []api.Comment{
		{Author: "user1", Body: "Comment 1", CreatedAt: "2024-01-01T10:00:00Z"},
		{Author: "user2", Body: "Comment 2", CreatedAt: "2024-01-02T11:00:00Z"},
	}

	cmd := newViewCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	opts := &viewOptions{comments: true}
	err := runViewWithDeps(cmd, opts, mock, "owner", "repo", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestViewCommand_Exists(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"view", "--help"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("view command should exist: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("view")) {
		t.Error("Expected help output to mention 'view'")
	}
}

func TestViewCommand_RequiresIssueNumber(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"view"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when issue number not provided")
	}
}

func TestViewCommand_HasJSONFlag(t *testing.T) {
	cmd := NewRootCommand()
	viewCmd, _, err := cmd.Find([]string{"view"})
	if err != nil {
		t.Fatalf("view command not found: %v", err)
	}

	flag := viewCmd.Flags().Lookup("json")
	if flag == nil {
		t.Fatal("Expected --json flag to exist")
	}
}

func TestViewCommand_HasWebFlag(t *testing.T) {
	cmd := NewRootCommand()
	viewCmd, _, err := cmd.Find([]string{"view"})
	if err != nil {
		t.Fatalf("view command not found: %v", err)
	}

	flag := viewCmd.Flags().Lookup("web")
	if flag == nil {
		t.Fatal("Expected --web flag to exist")
	}

	// Check shorthand
	if flag.Shorthand != "w" {
		t.Errorf("Expected --web shorthand to be 'w', got %s", flag.Shorthand)
	}
}

func TestViewCommand_HasCommentsFlag(t *testing.T) {
	cmd := NewRootCommand()
	viewCmd, _, err := cmd.Find([]string{"view"})
	if err != nil {
		t.Fatalf("view command not found: %v", err)
	}

	flag := viewCmd.Flags().Lookup("comments")
	if flag == nil {
		t.Fatal("Expected --comments flag to exist")
	}

	// Check shorthand
	if flag.Shorthand != "c" {
		t.Errorf("Expected --comments shorthand to be 'c', got %s", flag.Shorthand)
	}
}

func TestViewCommand_HasRepoFlag(t *testing.T) {
	cmd := NewRootCommand()
	viewCmd, _, err := cmd.Find([]string{"view"})
	if err != nil {
		t.Fatalf("view command not found: %v", err)
	}

	flag := viewCmd.Flags().Lookup("repo")
	if flag == nil {
		t.Fatal("Expected --repo flag to exist")
	}

	// Check shorthand
	if flag.Shorthand != "R" {
		t.Errorf("Expected --repo shorthand to be 'R', got %s", flag.Shorthand)
	}

	// Check type
	if flag.Value.Type() != "string" {
		t.Errorf("Expected --repo to be string, got %s", flag.Value.Type())
	}
}

func TestViewCommand_AcceptsIssueNumber(t *testing.T) {
	cmd := NewRootCommand()
	viewCmd, _, err := cmd.Find([]string{"view"})
	if err != nil {
		t.Fatalf("view command not found: %v", err)
	}

	// Verify the command accepts exactly 1 argument
	if viewCmd.Args == nil {
		t.Error("Expected Args validator to be set")
	}
}

func TestViewCommand_ParsesIssueNumber(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		wantErr bool
	}{
		{"valid number", "123", false},
		{"with hash", "#123", false},
		{"invalid string", "abc", true},
		{"negative number", "-1", true},
		{"zero", "0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseIssueNumber(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseIssueNumber(%q) error = %v, wantErr %v", tt.arg, err, tt.wantErr)
			}
		})
	}
}

func TestViewCommand_ParsesIssueReference(t *testing.T) {
	tests := []struct {
		name       string
		arg        string
		wantOwner  string
		wantRepo   string
		wantNumber int
		wantErr    bool
	}{
		{"number only", "123", "", "", 123, false},
		{"with hash", "#123", "", "", 123, false},
		{"full reference", "owner/repo#123", "owner", "repo", 123, false},
		{"invalid", "invalid", "", "", 0, true},
		// URL formats
		{"https URL", "https://github.com/owner/repo/issues/123", "owner", "repo", 123, false},
		{"http URL", "http://github.com/owner/repo/issues/123", "owner", "repo", 123, false},
		{"URL with anchor", "https://github.com/owner/repo/issues/123#issuecomment-456", "owner", "repo", 123, false},
		{"invalid URL - not issues", "https://github.com/owner/repo/pulls/123", "", "", 0, true},
		{"invalid URL - too short", "https://github.com/owner", "", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, number, err := parseIssueReference(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseIssueReference(%q) error = %v, wantErr %v", tt.arg, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if owner != tt.wantOwner {
					t.Errorf("parseIssueReference(%q) owner = %v, want %v", tt.arg, owner, tt.wantOwner)
				}
				if repo != tt.wantRepo {
					t.Errorf("parseIssueReference(%q) repo = %v, want %v", tt.arg, repo, tt.wantRepo)
				}
				if number != tt.wantNumber {
					t.Errorf("parseIssueReference(%q) number = %v, want %v", tt.arg, number, tt.wantNumber)
				}
			}
		})
	}
}

// Progress bar tests

func TestRenderProgressBar(t *testing.T) {
	tests := []struct {
		name      string
		completed int
		total     int
		width     int
		want      string
	}{
		{"empty", 0, 10, 10, "[░░░░░░░░░░]"},
		{"half", 5, 10, 10, "[█████░░░░░]"},
		{"full", 10, 10, 10, "[██████████]"},
		{"quarter", 1, 4, 8, "[██░░░░░░]"},
		{"zero total", 0, 0, 10, "[░░░░░░░░░░]"},
		{"60 percent", 3, 5, 10, "[██████░░░░]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderProgressBar(tt.completed, tt.total, tt.width)
			if got != tt.want {
				t.Errorf("renderProgressBar(%d, %d, %d) = %q, want %q",
					tt.completed, tt.total, tt.width, got, tt.want)
			}
		})
	}
}

func TestRenderProgressBar_OverflowProtection(t *testing.T) {
	// Test that completed > total doesn't overflow
	result := renderProgressBar(15, 10, 10)
	// Should cap at full
	if result != "[██████████]" {
		t.Errorf("renderProgressBar with overflow should cap at full, got %q", result)
	}
}

// ============================================================================
// outputViewTable Tests
// ============================================================================

// createViewTestCmd creates a cobra command for testing view output
func createViewTestCmd(buf *bytes.Buffer) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	return cmd
}

func TestOutputViewTable_BasicIssue(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := createViewTestCmd(buf)

	issue := &api.Issue{
		Number: 42,
		Title:  "Test Issue Title",
		State:  "OPEN",
		URL:    "https://github.com/owner/repo/issues/42",
		Author: api.Actor{Login: "testuser"},
	}

	err := outputViewTable(cmd, issue, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("outputViewTable() error = %v", err)
	}

	// Note: outputViewTable writes to os.Stdout, not cmd buffer
	// We verify no error occurred
}

func TestOutputViewTable_WithAssignees(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := createViewTestCmd(buf)

	issue := &api.Issue{
		Number: 42,
		Title:  "Test Issue",
		State:  "OPEN",
		URL:    "https://github.com/owner/repo/issues/42",
		Author: api.Actor{Login: "author"},
		Assignees: []api.Actor{
			{Login: "user1"},
			{Login: "user2"},
		},
	}

	err := outputViewTable(cmd, issue, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("outputViewTable() error = %v", err)
	}
}

func TestOutputViewTable_WithLabels(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := createViewTestCmd(buf)

	issue := &api.Issue{
		Number: 42,
		Title:  "Test Issue",
		State:  "OPEN",
		URL:    "https://github.com/owner/repo/issues/42",
		Author: api.Actor{Login: "author"},
		Labels: []api.Label{
			{Name: "bug"},
			{Name: "priority:high"},
		},
	}

	err := outputViewTable(cmd, issue, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("outputViewTable() error = %v", err)
	}
}

func TestOutputViewTable_WithMilestone(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := createViewTestCmd(buf)

	issue := &api.Issue{
		Number:    42,
		Title:     "Test Issue",
		State:     "OPEN",
		URL:       "https://github.com/owner/repo/issues/42",
		Author:    api.Actor{Login: "author"},
		Milestone: &api.Milestone{Title: "v1.0.0"},
	}

	err := outputViewTable(cmd, issue, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("outputViewTable() error = %v", err)
	}
}

func TestOutputViewTable_WithFieldValues(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := createViewTestCmd(buf)

	issue := &api.Issue{
		Number: 42,
		Title:  "Test Issue",
		State:  "OPEN",
		URL:    "https://github.com/owner/repo/issues/42",
		Author: api.Actor{Login: "author"},
	}

	fieldValues := []api.FieldValue{
		{Field: "Status", Value: "In Progress"},
		{Field: "Priority", Value: "High"},
	}

	err := outputViewTable(cmd, issue, fieldValues, nil, nil, nil)
	if err != nil {
		t.Fatalf("outputViewTable() error = %v", err)
	}
}

func TestOutputViewTable_WithParentIssue(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := createViewTestCmd(buf)

	issue := &api.Issue{
		Number: 42,
		Title:  "Sub-Issue",
		State:  "OPEN",
		URL:    "https://github.com/owner/repo/issues/42",
		Author: api.Actor{Login: "author"},
	}

	parentIssue := &api.Issue{
		Number: 10,
		Title:  "Parent Issue",
		URL:    "https://github.com/owner/repo/issues/10",
	}

	err := outputViewTable(cmd, issue, nil, nil, parentIssue, nil)
	if err != nil {
		t.Fatalf("outputViewTable() error = %v", err)
	}
}

func TestOutputViewTable_WithSubIssues(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := createViewTestCmd(buf)

	issue := &api.Issue{
		Number: 42,
		Title:  "Parent Issue",
		State:  "OPEN",
		URL:    "https://github.com/owner/repo/issues/42",
		Author: api.Actor{Login: "author"},
		Repository: api.Repository{
			Owner: "owner",
			Name:  "repo",
		},
	}

	subIssues := []api.SubIssue{
		{Number: 43, Title: "Sub 1", State: "CLOSED", URL: "https://github.com/owner/repo/issues/43"},
		{Number: 44, Title: "Sub 2", State: "OPEN", URL: "https://github.com/owner/repo/issues/44"},
		{Number: 45, Title: "Sub 3", State: "CLOSED", URL: "https://github.com/owner/repo/issues/45"},
	}

	err := outputViewTable(cmd, issue, nil, subIssues, nil, nil)
	if err != nil {
		t.Fatalf("outputViewTable() error = %v", err)
	}
}

func TestOutputViewTable_WithCrossRepoSubIssues(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := createViewTestCmd(buf)

	issue := &api.Issue{
		Number: 42,
		Title:  "Parent Issue",
		State:  "OPEN",
		URL:    "https://github.com/owner/repo/issues/42",
		Author: api.Actor{Login: "author"},
		Repository: api.Repository{
			Owner: "owner",
			Name:  "repo",
		},
	}

	subIssues := []api.SubIssue{
		{
			Number: 43,
			Title:  "Same Repo Sub",
			State:  "OPEN",
			URL:    "https://github.com/owner/repo/issues/43",
			Repository: api.Repository{
				Owner: "owner",
				Name:  "repo",
			},
		},
		{
			Number: 10,
			Title:  "Cross Repo Sub",
			State:  "CLOSED",
			URL:    "https://github.com/owner/other-repo/issues/10",
			Repository: api.Repository{
				Owner: "owner",
				Name:  "other-repo",
			},
		},
	}

	err := outputViewTable(cmd, issue, nil, subIssues, nil, nil)
	if err != nil {
		t.Fatalf("outputViewTable() error = %v", err)
	}
}

func TestOutputViewTable_WithBody(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := createViewTestCmd(buf)

	issue := &api.Issue{
		Number: 42,
		Title:  "Test Issue",
		State:  "OPEN",
		URL:    "https://github.com/owner/repo/issues/42",
		Author: api.Actor{Login: "author"},
		Body:   "This is the issue body with some content.\n\nMultiple paragraphs.",
	}

	err := outputViewTable(cmd, issue, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("outputViewTable() error = %v", err)
	}
}

func TestOutputViewTable_FullIssue(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := createViewTestCmd(buf)

	issue := &api.Issue{
		Number:    42,
		Title:     "Full Featured Issue",
		State:     "OPEN",
		URL:       "https://github.com/owner/repo/issues/42",
		Body:      "Issue body content",
		Author:    api.Actor{Login: "author"},
		Assignees: []api.Actor{{Login: "dev1"}, {Login: "dev2"}},
		Labels:    []api.Label{{Name: "bug"}, {Name: "urgent"}},
		Milestone: &api.Milestone{Title: "v2.0"},
		Repository: api.Repository{
			Owner: "owner",
			Name:  "repo",
		},
	}

	fieldValues := []api.FieldValue{
		{Field: "Status", Value: "In Progress"},
		{Field: "Priority", Value: "P1"},
	}

	subIssues := []api.SubIssue{
		{Number: 43, Title: "Task 1", State: "CLOSED"},
		{Number: 44, Title: "Task 2", State: "OPEN"},
	}

	parentIssue := &api.Issue{
		Number: 10,
		Title:  "Epic",
		URL:    "https://github.com/owner/repo/issues/10",
	}

	err := outputViewTable(cmd, issue, fieldValues, subIssues, parentIssue, nil)
	if err != nil {
		t.Fatalf("outputViewTable() error = %v", err)
	}
}

// ============================================================================
// outputViewJSON Tests
// ============================================================================

func TestOutputViewJSON_BasicIssue(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := createViewTestCmd(buf)

	issue := &api.Issue{
		Number: 42,
		Title:  "Test Issue",
		State:  "OPEN",
		URL:    "https://github.com/owner/repo/issues/42",
		Author: api.Actor{Login: "testuser"},
	}

	err := outputViewJSON(cmd, issue, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("outputViewJSON() error = %v", err)
	}
}

func TestOutputViewJSON_WithAllFields(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := createViewTestCmd(buf)

	issue := &api.Issue{
		Number:    42,
		Title:     "Full Issue",
		State:     "OPEN",
		Body:      "Issue description",
		URL:       "https://github.com/owner/repo/issues/42",
		Author:    api.Actor{Login: "author"},
		Assignees: []api.Actor{{Login: "dev1"}, {Login: "dev2"}},
		Labels:    []api.Label{{Name: "bug"}, {Name: "priority:high"}},
		Milestone: &api.Milestone{Title: "v1.0"},
	}

	fieldValues := []api.FieldValue{
		{Field: "Status", Value: "In Progress"},
		{Field: "Priority", Value: "High"},
	}

	err := outputViewJSON(cmd, issue, fieldValues, nil, nil, nil)
	if err != nil {
		t.Fatalf("outputViewJSON() error = %v", err)
	}
}

func TestOutputViewJSON_WithSubIssues(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := createViewTestCmd(buf)

	issue := &api.Issue{
		Number: 42,
		Title:  "Parent Issue",
		State:  "OPEN",
		URL:    "https://github.com/owner/repo/issues/42",
		Author: api.Actor{Login: "author"},
	}

	subIssues := []api.SubIssue{
		{Number: 43, Title: "Sub 1", State: "CLOSED", URL: "https://github.com/owner/repo/issues/43"},
		{Number: 44, Title: "Sub 2", State: "OPEN", URL: "https://github.com/owner/repo/issues/44"},
		{Number: 45, Title: "Sub 3", State: "CLOSED", URL: "https://github.com/owner/repo/issues/45"},
	}

	err := outputViewJSON(cmd, issue, nil, subIssues, nil, nil)
	if err != nil {
		t.Fatalf("outputViewJSON() error = %v", err)
	}
}

func TestOutputViewJSON_WithParentIssue(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := createViewTestCmd(buf)

	issue := &api.Issue{
		Number: 42,
		Title:  "Sub-Issue",
		State:  "OPEN",
		URL:    "https://github.com/owner/repo/issues/42",
		Author: api.Actor{Login: "author"},
	}

	parentIssue := &api.Issue{
		Number: 10,
		Title:  "Parent Issue",
		URL:    "https://github.com/owner/repo/issues/10",
	}

	err := outputViewJSON(cmd, issue, nil, nil, parentIssue, nil)
	if err != nil {
		t.Fatalf("outputViewJSON() error = %v", err)
	}
}

func TestOutputViewJSON_SubIssueProgress(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := createViewTestCmd(buf)

	issue := &api.Issue{
		Number: 42,
		Title:  "Parent Issue",
		State:  "OPEN",
		URL:    "https://github.com/owner/repo/issues/42",
		Author: api.Actor{Login: "author"},
	}

	// 3 closed out of 5 = 60%
	subIssues := []api.SubIssue{
		{Number: 1, Title: "Task 1", State: "CLOSED"},
		{Number: 2, Title: "Task 2", State: "CLOSED"},
		{Number: 3, Title: "Task 3", State: "OPEN"},
		{Number: 4, Title: "Task 4", State: "CLOSED"},
		{Number: 5, Title: "Task 5", State: "OPEN"},
	}

	err := outputViewJSON(cmd, issue, nil, subIssues, nil, nil)
	if err != nil {
		t.Fatalf("outputViewJSON() error = %v", err)
	}
}

func TestOpenInBrowser(t *testing.T) {
	// Test that function exists and handles URL parameter
	// We can't actually test browser opening in unit tests
	_ = ui.OpenInBrowser
}

func TestOutputViewTable_WithComments(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := createViewTestCmd(buf)

	issue := &api.Issue{
		Number: 42,
		Title:  "Test Issue",
		State:  "OPEN",
		URL:    "https://github.com/owner/repo/issues/42",
		Author: api.Actor{Login: "author"},
	}

	comments := []api.Comment{
		{Author: "user1", Body: "First comment", CreatedAt: "2024-01-01T10:00:00Z"},
		{Author: "user2", Body: "Second comment", CreatedAt: "2024-01-02T11:00:00Z"},
	}

	err := outputViewTable(cmd, issue, nil, nil, nil, comments)
	if err != nil {
		t.Fatalf("outputViewTable() error = %v", err)
	}
}

func TestOutputViewJSON_WithComments(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := createViewTestCmd(buf)

	issue := &api.Issue{
		Number: 42,
		Title:  "Test Issue",
		State:  "OPEN",
		URL:    "https://github.com/owner/repo/issues/42",
		Author: api.Actor{Login: "author"},
	}

	comments := []api.Comment{
		{Author: "user1", Body: "First comment", CreatedAt: "2024-01-01T10:00:00Z"},
		{Author: "user2", Body: "Second comment", CreatedAt: "2024-01-02T11:00:00Z"},
	}

	err := outputViewJSON(cmd, issue, nil, nil, nil, comments)
	if err != nil {
		t.Fatalf("outputViewJSON() error = %v", err)
	}
}

// ============================================================================
// ViewJSONOutput Structure Tests
// ============================================================================

func TestViewJSONOutput_Structure(t *testing.T) {
	output := ViewJSONOutput{
		Number:    42,
		Title:     "Test Issue",
		State:     "OPEN",
		Body:      "Issue body",
		URL:       "https://github.com/owner/repo/issues/42",
		Author:    "testuser",
		Assignees: []string{"user1", "user2"},
		Labels:    []string{"bug", "urgent"},
		Milestone: "v1.0",
		FieldValues: map[string]string{
			"Status":   "In Progress",
			"Priority": "High",
		},
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal ViewJSONOutput: %v", err)
	}

	var parsed ViewJSONOutput
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal ViewJSONOutput: %v", err)
	}

	if parsed.Number != 42 {
		t.Errorf("Expected Number 42, got %d", parsed.Number)
	}
	if parsed.Title != "Test Issue" {
		t.Errorf("Expected Title 'Test Issue', got %s", parsed.Title)
	}
	if len(parsed.Assignees) != 2 {
		t.Errorf("Expected 2 assignees, got %d", len(parsed.Assignees))
	}
	if parsed.FieldValues["Status"] != "In Progress" {
		t.Errorf("Expected Status 'In Progress', got %s", parsed.FieldValues["Status"])
	}
}

func TestViewJSONOutput_WithSubProgress(t *testing.T) {
	output := ViewJSONOutput{
		Number: 42,
		Title:  "Parent",
		State:  "OPEN",
		URL:    "https://example.com",
		Author: "user",
		SubIssues: []SubIssueJSON{
			{Number: 1, Title: "Sub 1", State: "CLOSED"},
			{Number: 2, Title: "Sub 2", State: "OPEN"},
		},
		SubProgress: &SubProgressJSON{
			Total:      2,
			Completed:  1,
			Percentage: 50,
		},
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var parsed ViewJSONOutput
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if parsed.SubProgress == nil {
		t.Fatal("Expected SubProgress to be present")
	}
	if parsed.SubProgress.Percentage != 50 {
		t.Errorf("Expected 50%% progress, got %d%%", parsed.SubProgress.Percentage)
	}
}

func TestViewJSONOutput_WithParentIssue(t *testing.T) {
	output := ViewJSONOutput{
		Number: 42,
		Title:  "Sub-Issue",
		State:  "OPEN",
		URL:    "https://example.com",
		Author: "user",
		ParentIssue: &ParentIssueJSON{
			Number: 10,
			Title:  "Parent Issue",
			URL:    "https://example.com/10",
		},
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var parsed ViewJSONOutput
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if parsed.ParentIssue == nil {
		t.Fatal("Expected ParentIssue to be present")
	}
	if parsed.ParentIssue.Number != 10 {
		t.Errorf("Expected parent number 10, got %d", parsed.ParentIssue.Number)
	}
}

func TestSubIssueJSON_Structure(t *testing.T) {
	sub := SubIssueJSON{
		Number: 43,
		Title:  "Sub-Issue Title",
		State:  "CLOSED",
		URL:    "https://github.com/owner/repo/issues/43",
	}

	data, err := json.Marshal(sub)
	if err != nil {
		t.Fatalf("Failed to marshal SubIssueJSON: %v", err)
	}

	jsonStr := string(data)
	expectedFields := []string{"number", "title", "state", "url"}
	for _, field := range expectedFields {
		if !bytes.Contains(data, []byte(field)) {
			t.Errorf("Expected JSON to contain field %q, got: %s", field, jsonStr)
		}
	}
}

func TestSubProgressJSON_Structure(t *testing.T) {
	progress := SubProgressJSON{
		Total:      10,
		Completed:  6,
		Percentage: 60,
	}

	data, err := json.Marshal(progress)
	if err != nil {
		t.Fatalf("Failed to marshal SubProgressJSON: %v", err)
	}

	var parsed SubProgressJSON
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal SubProgressJSON: %v", err)
	}

	if parsed.Total != 10 {
		t.Errorf("Expected Total 10, got %d", parsed.Total)
	}
	if parsed.Completed != 6 {
		t.Errorf("Expected Completed 6, got %d", parsed.Completed)
	}
	if parsed.Percentage != 60 {
		t.Errorf("Expected Percentage 60, got %d", parsed.Percentage)
	}
}

// ============================================================================
// Body File Tests
// ============================================================================

func TestViewCommand_HasBodyFileFlag(t *testing.T) {
	cmd := NewRootCommand()
	viewCmd, _, err := cmd.Find([]string{"view"})
	if err != nil {
		t.Fatalf("view command not found: %v", err)
	}

	flag := viewCmd.Flags().Lookup("body-file")
	if flag == nil {
		t.Fatal("Expected --body-file flag to exist")
	}

	// Check shorthand
	if flag.Shorthand != "b" {
		t.Errorf("Expected --body-file shorthand to be 'b', got %s", flag.Shorthand)
	}
}

func TestWriteBodyToFile_Success(t *testing.T) {
	// Create a temp directory for testing
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir to temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	err := writeBodyToFile(42, "Test body content\n\nWith multiple lines.")
	if err != nil {
		t.Fatalf("writeBodyToFile() error = %v", err)
	}

	// Verify file was created
	filePath := filepath.Join("tmp", "issue-42.md")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	expected := "Test body content\n\nWith multiple lines."
	if string(content) != expected {
		t.Errorf("File content = %q, want %q", string(content), expected)
	}
}

func TestWriteBodyToFile_CreatesTmpDirectory(t *testing.T) {
	// Create a temp directory for testing
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir to temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	// Verify tmp doesn't exist
	if _, err := os.Stat("tmp"); !os.IsNotExist(err) {
		t.Fatal("tmp directory should not exist before test")
	}

	err := writeBodyToFile(123, "Body content")
	if err != nil {
		t.Fatalf("writeBodyToFile() error = %v", err)
	}

	// Verify tmp directory was created
	info, err := os.Stat("tmp")
	if err != nil {
		t.Fatalf("tmp directory should exist after writeBodyToFile: %v", err)
	}
	if !info.IsDir() {
		t.Error("tmp should be a directory")
	}
}

func TestWriteBodyToFile_EmptyBody(t *testing.T) {
	// Create a temp directory for testing
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir to temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	err := writeBodyToFile(99, "")
	if err != nil {
		t.Fatalf("writeBodyToFile() error = %v", err)
	}

	// Verify file was created with empty content
	filePath := filepath.Join("tmp", "issue-99.md")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	if string(content) != "" {
		t.Errorf("File content = %q, want empty string", string(content))
	}
}

func TestRunViewWithDeps_BodyFile(t *testing.T) {
	// Create a temp directory for testing
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir to temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	mock := newMockViewClient()
	mock.issue = &api.Issue{
		Number: 42,
		Title:  "Test Issue",
		State:  "OPEN",
		URL:    "https://github.com/owner/repo/issues/42",
		Author: api.Actor{Login: "testuser"},
		Body:   "This is the issue body for testing.",
	}

	cmd := newViewCommand()
	opts := &viewOptions{bodyFile: true}
	err := runViewWithDeps(cmd, opts, mock, "owner", "repo", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created with correct content
	filePath := filepath.Join("tmp", "issue-42.md")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	expected := "This is the issue body for testing."
	if string(content) != expected {
		t.Errorf("File content = %q, want %q", string(content), expected)
	}
}

// ============================================================================
// Body Stdout Tests
// ============================================================================

func TestViewCommand_HasBodyStdoutFlag(t *testing.T) {
	cmd := NewRootCommand()
	viewCmd, _, err := cmd.Find([]string{"view"})
	if err != nil {
		t.Fatalf("view command not found: %v", err)
	}

	flag := viewCmd.Flags().Lookup("body-stdout")
	if flag == nil {
		t.Fatal("Expected --body-stdout flag to exist")
	}

	// body-stdout has no shorthand (b is taken by body-file)
	if flag.Shorthand != "" {
		t.Errorf("Expected --body-stdout to have no shorthand, got %s", flag.Shorthand)
	}
}

func TestRunViewWithDeps_BodyStdout(t *testing.T) {
	mock := newMockViewClient()
	mock.issue = &api.Issue{
		Number: 42,
		Title:  "Test Issue",
		State:  "OPEN",
		URL:    "https://github.com/owner/repo/issues/42",
		Author: api.Actor{Login: "testuser"},
		Body:   "This is the issue body for stdout testing.",
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newViewCommand()
	opts := &viewOptions{bodyStdout: true}
	err := runViewWithDeps(cmd, opts, mock, "owner", "repo", 42)

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "This is the issue body for stdout testing."
	if buf.String() != expected {
		t.Errorf("stdout output = %q, want %q", buf.String(), expected)
	}
}

func TestRunViewWithDeps_BodyStdout_EmptyBody(t *testing.T) {
	mock := newMockViewClient()
	mock.issue = &api.Issue{
		Number: 99,
		Title:  "Empty Body Issue",
		State:  "OPEN",
		URL:    "https://github.com/owner/repo/issues/99",
		Author: api.Actor{Login: "testuser"},
		Body:   "",
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newViewCommand()
	opts := &viewOptions{bodyStdout: true}
	err := runViewWithDeps(cmd, opts, mock, "owner", "repo", 99)

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if buf.String() != "" {
		t.Errorf("stdout output = %q, want empty string", buf.String())
	}
}
