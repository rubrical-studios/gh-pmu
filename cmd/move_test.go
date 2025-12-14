package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/rubrical-studios/gh-pmu/internal/api"
	"github.com/rubrical-studios/gh-pmu/internal/config"
	"github.com/spf13/cobra"
)

// mockMoveClient implements moveClient for testing
type mockMoveClient struct {
	issues       map[string]*api.Issue // "owner/repo#number" -> Issue
	project      *api.Project
	projectItems []api.ProjectItem
	subIssues    map[string][]api.SubIssue // "owner/repo#number" -> SubIssues
	fieldUpdates []fieldUpdate             // track field updates for verification

	// Microsprint support
	openIssuesByLabel map[string][]api.Issue // label -> issues

	// Error injection
	getIssueErr              error
	getProjectErr            error
	getProjectItemsErr       error
	getSubIssuesErr          error
	setProjectItemErr        error
	setProjectItemErrFor     map[string]error // itemID -> error
	getOpenIssuesByLabelErr  error
}

type fieldUpdate struct {
	projectID string
	itemID    string
	fieldName string
	value     string
}

func newMockMoveClient() *mockMoveClient {
	return &mockMoveClient{
		issues:               make(map[string]*api.Issue),
		subIssues:            make(map[string][]api.SubIssue),
		openIssuesByLabel:    make(map[string][]api.Issue),
		setProjectItemErrFor: make(map[string]error),
	}
}

func (m *mockMoveClient) GetIssue(owner, repo string, number int) (*api.Issue, error) {
	if m.getIssueErr != nil {
		return nil, m.getIssueErr
	}
	key := fmt.Sprintf("%s/%s#%d", owner, repo, number)
	if issue, ok := m.issues[key]; ok {
		return issue, nil
	}
	return nil, fmt.Errorf("issue not found: %s", key)
}

func (m *mockMoveClient) GetProject(owner string, number int) (*api.Project, error) {
	if m.getProjectErr != nil {
		return nil, m.getProjectErr
	}
	if m.project != nil {
		return m.project, nil
	}
	return nil, fmt.Errorf("project not found")
}

func (m *mockMoveClient) GetProjectItems(projectID string, filter *api.ProjectItemsFilter) ([]api.ProjectItem, error) {
	if m.getProjectItemsErr != nil {
		return nil, m.getProjectItemsErr
	}
	return m.projectItems, nil
}

func (m *mockMoveClient) GetSubIssues(owner, repo string, number int) ([]api.SubIssue, error) {
	if m.getSubIssuesErr != nil {
		return nil, m.getSubIssuesErr
	}
	key := fmt.Sprintf("%s/%s#%d", owner, repo, number)
	result := m.subIssues[key]
	// Debug: fmt.Printf("DEBUG GetSubIssues: key=%q, found=%d\n", key, len(result))
	return result, nil
}

func (m *mockMoveClient) SetProjectItemField(projectID, itemID, fieldName, value string) error {
	if m.setProjectItemErr != nil {
		return m.setProjectItemErr
	}
	if err, ok := m.setProjectItemErrFor[itemID]; ok {
		return err
	}
	m.fieldUpdates = append(m.fieldUpdates, fieldUpdate{
		projectID: projectID,
		itemID:    itemID,
		fieldName: fieldName,
		value:     value,
	})
	return nil
}

func (m *mockMoveClient) GetOpenIssuesByLabel(owner, repo, label string) ([]api.Issue, error) {
	if m.getOpenIssuesByLabelErr != nil {
		return nil, m.getOpenIssuesByLabelErr
	}
	return m.openIssuesByLabel[label], nil
}

// Test helpers

func testMoveConfig() *config.Config {
	return &config.Config{
		Project: config.Project{
			Owner:  "testowner",
			Number: 1,
		},
		Repositories: []string{"testowner/testrepo"},
		Fields: map[string]config.Field{
			"status": {
				Field: "Status",
				Values: map[string]string{
					"in_progress": "In Progress",
					"done":        "Done",
					"todo":        "Todo",
				},
			},
			"priority": {
				Field: "Priority",
				Values: map[string]string{
					"high":   "High",
					"medium": "Medium",
					"low":    "Low",
				},
			},
			"microsprint": {
				Field: "Microsprint",
			},
		},
	}
}

func setupMockWithIssue(number int, title string, itemID string) *mockMoveClient {
	mock := newMockMoveClient()
	mock.project = &api.Project{
		ID:     "proj-1",
		Number: 1,
		Title:  "Test Project",
	}
	mock.issues[fmt.Sprintf("testowner/testrepo#%d", number)] = &api.Issue{
		ID:     fmt.Sprintf("issue-%d", number),
		Number: number,
		Title:  title,
		Repository: api.Repository{
			Owner: "testowner",
			Name:  "testrepo",
		},
	}
	mock.projectItems = []api.ProjectItem{
		{
			ID: itemID,
			Issue: &api.Issue{
				Number: number,
				Repository: api.Repository{
					Owner: "testowner",
					Name:  "testrepo",
				},
			},
		},
	}
	return mock
}

// ============================================================================
// Command Flag Tests
// ============================================================================

func TestMoveCommand_Exists(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"move", "--help"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("move command should exist: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("move")) {
		t.Error("Expected help output to mention 'move'")
	}
}

func TestMoveCommand_RequiresIssueNumber(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"move"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when issue number not provided")
	}
}

func TestMoveCommand_HasStatusFlag(t *testing.T) {
	cmd := NewRootCommand()
	moveCmd, _, err := cmd.Find([]string{"move"})
	if err != nil {
		t.Fatalf("move command not found: %v", err)
	}

	flag := moveCmd.Flags().Lookup("status")
	if flag == nil {
		t.Fatal("Expected --status flag to exist")
	}
}

func TestMoveCommand_HasPriorityFlag(t *testing.T) {
	cmd := NewRootCommand()
	moveCmd, _, err := cmd.Find([]string{"move"})
	if err != nil {
		t.Fatalf("move command not found: %v", err)
	}

	flag := moveCmd.Flags().Lookup("priority")
	if flag == nil {
		t.Fatal("Expected --priority flag to exist")
	}
}

func TestMoveCommand_RequiresAtLeastOneFlag(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"move", "123"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when no field flags provided")
	}
}

func TestMoveCommand_HasRecursiveFlag(t *testing.T) {
	cmd := NewRootCommand()
	moveCmd, _, err := cmd.Find([]string{"move"})
	if err != nil {
		t.Fatalf("move command not found: %v", err)
	}

	flag := moveCmd.Flags().Lookup("recursive")
	if flag == nil {
		t.Fatal("Expected --recursive flag to exist")
	}

	if flag.Shorthand != "r" {
		t.Errorf("Expected --recursive shorthand to be 'r', got '%s'", flag.Shorthand)
	}
}

func TestMoveCommand_HasDepthFlag(t *testing.T) {
	cmd := NewRootCommand()
	moveCmd, _, err := cmd.Find([]string{"move"})
	if err != nil {
		t.Fatalf("move command not found: %v", err)
	}

	flag := moveCmd.Flags().Lookup("depth")
	if flag == nil {
		t.Fatal("Expected --depth flag to exist")
	}

	if flag.DefValue != "10" {
		t.Errorf("Expected --depth default to be 10, got '%s'", flag.DefValue)
	}
}

func TestMoveCommand_HasDryRunFlag(t *testing.T) {
	cmd := NewRootCommand()
	moveCmd, _, err := cmd.Find([]string{"move"})
	if err != nil {
		t.Fatalf("move command not found: %v", err)
	}

	flag := moveCmd.Flags().Lookup("dry-run")
	if flag == nil {
		t.Fatal("Expected --dry-run flag to exist")
	}
}

func TestMoveCommand_HasYesFlag(t *testing.T) {
	cmd := NewRootCommand()
	moveCmd, _, err := cmd.Find([]string{"move"})
	if err != nil {
		t.Fatalf("move command not found: %v", err)
	}

	flag := moveCmd.Flags().Lookup("yes")
	if flag == nil {
		t.Fatal("Expected --yes flag to exist")
	}

	if flag.Shorthand != "y" {
		t.Errorf("Expected --yes shorthand to be 'y', got '%s'", flag.Shorthand)
	}
}

func TestMoveCommand_RecursiveHelpText(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"move", "--help"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("move help failed: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("--recursive")) {
		t.Error("Expected help to mention --recursive flag")
	}
	if !bytes.Contains([]byte(output), []byte("sub-issues")) {
		t.Error("Expected help to mention sub-issues")
	}
}

func TestMoveCommand_HelpHasRecursiveExamples(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"move", "--help"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("move help failed: %v", err)
	}

	output := buf.String()

	// Verify recursive examples are documented
	tests := []struct {
		name     string
		expected string
	}{
		{"basic recursive example", "--status in_progress --recursive"},
		{"dry-run example", "--recursive --dry-run"},
		{"yes flag example", "--recursive --yes"},
		{"depth flag example", "--recursive --depth"},
	}

	for _, tt := range tests {
		if !strings.Contains(output, tt.expected) {
			t.Errorf("Expected help to contain %s example: %s", tt.name, tt.expected)
		}
	}
}

// ============================================================================
// runMoveWithDeps Tests
// ============================================================================

func TestRunMoveWithDeps_InvalidIssueReference(t *testing.T) {
	mock := newMockMoveClient()
	cfg := testMoveConfig()
	cfg.Repositories = []string{} // No repos configured

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	opts := &moveOptions{status: "in_progress"}

	// Invalid issue reference with no repos
	err := runMoveWithDeps(cmd, []string{"invalid"}, opts, cfg, mock)
	if err == nil {
		t.Error("Expected error for invalid issue reference")
	}
}

func TestRunMoveWithDeps_NoRepoConfigured(t *testing.T) {
	mock := newMockMoveClient()
	cfg := testMoveConfig()
	cfg.Repositories = []string{}

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	opts := &moveOptions{status: "in_progress"}

	err := runMoveWithDeps(cmd, []string{"123"}, opts, cfg, mock)
	if err == nil {
		t.Error("Expected error when no repository configured")
	}
	if !strings.Contains(err.Error(), "no repository specified and none configured") {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestMoveCommand_HasRepoFlag(t *testing.T) {
	cmd := NewRootCommand()
	moveCmd, _, err := cmd.Find([]string{"move"})
	if err != nil {
		t.Fatalf("move command not found: %v", err)
	}

	flag := moveCmd.Flags().Lookup("repo")
	if flag == nil {
		t.Fatal("Expected --repo flag to exist")
	}

	if flag.Shorthand != "R" {
		t.Errorf("Expected --repo shorthand to be 'R', got '%s'", flag.Shorthand)
	}
}

func TestRunMoveWithDeps_RepoFlagOverridesConfig(t *testing.T) {
	mock := newMockMoveClient()
	mock.project = &api.Project{ID: "proj-1", Number: 1, Title: "Test Project"}
	mock.issues["other/repo#456"] = &api.Issue{
		ID:     "issue-456",
		Number: 456,
		Title:  "Issue in other repo",
		Repository: api.Repository{
			Owner: "other",
			Name:  "repo",
		},
	}
	mock.projectItems = []api.ProjectItem{
		{
			ID: "item-456",
			Issue: &api.Issue{
				Number: 456,
				Repository: api.Repository{
					Owner: "other",
					Name:  "repo",
				},
			},
		},
	}
	cfg := testMoveConfig() // Uses testowner/testrepo

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Use --repo flag to override config
	opts := &moveOptions{status: "in_progress", repo: "other/repo"}

	err := runMoveWithDeps(cmd, []string{"456"}, opts, cfg, mock)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(mock.fieldUpdates) != 1 {
		t.Fatalf("Expected 1 field update, got %d", len(mock.fieldUpdates))
	}
}

func TestRunMoveWithDeps_InvalidRepoFlagFormat(t *testing.T) {
	mock := newMockMoveClient()
	cfg := testMoveConfig()

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	opts := &moveOptions{status: "in_progress", repo: "invalid-format"}

	err := runMoveWithDeps(cmd, []string{"123"}, opts, cfg, mock)
	if err == nil {
		t.Error("Expected error for invalid --repo format")
	}
	if !strings.Contains(err.Error(), "invalid --repo format") {
		t.Errorf("Expected 'invalid --repo format' error, got: %v", err)
	}
}

func TestRunMoveWithDeps_RepoFlagWithNoConfiguredRepos(t *testing.T) {
	mock := newMockMoveClient()
	mock.project = &api.Project{ID: "proj-1", Number: 1, Title: "Test Project"}
	mock.issues["other/repo#123"] = &api.Issue{
		ID:     "issue-123",
		Number: 123,
		Title:  "Test Issue",
		Repository: api.Repository{
			Owner: "other",
			Name:  "repo",
		},
	}
	mock.projectItems = []api.ProjectItem{
		{
			ID: "item-123",
			Issue: &api.Issue{
				Number: 123,
				Repository: api.Repository{
					Owner: "other",
					Name:  "repo",
				},
			},
		},
	}
	cfg := testMoveConfig()
	cfg.Repositories = []string{} // No repos configured

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// --repo flag should work even without configured repos
	opts := &moveOptions{status: "in_progress", repo: "other/repo"}

	err := runMoveWithDeps(cmd, []string{"123"}, opts, cfg, mock)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(mock.fieldUpdates) != 1 {
		t.Fatalf("Expected 1 field update, got %d", len(mock.fieldUpdates))
	}
}

func TestRunMoveWithDeps_InvalidRepoFormat(t *testing.T) {
	mock := newMockMoveClient()
	cfg := testMoveConfig()
	cfg.Repositories = []string{"invalid-repo-format"} // Missing slash

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	opts := &moveOptions{status: "in_progress"}

	err := runMoveWithDeps(cmd, []string{"123"}, opts, cfg, mock)
	if err == nil {
		t.Error("Expected error for invalid repo format")
	}
}

func TestRunMoveWithDeps_GetIssueFails(t *testing.T) {
	mock := newMockMoveClient()
	mock.getIssueErr = fmt.Errorf("API error")
	cfg := testMoveConfig()

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	opts := &moveOptions{status: "in_progress"}

	err := runMoveWithDeps(cmd, []string{"123"}, opts, cfg, mock)
	if err == nil {
		t.Error("Expected error when GetIssue fails")
	}
}

func TestRunMoveWithDeps_GetProjectFails(t *testing.T) {
	mock := newMockMoveClient()
	mock.issues["testowner/testrepo#123"] = &api.Issue{
		ID:     "issue-123",
		Number: 123,
		Title:  "Test Issue",
	}
	mock.getProjectErr = fmt.Errorf("project API error")
	cfg := testMoveConfig()

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	opts := &moveOptions{status: "in_progress"}

	err := runMoveWithDeps(cmd, []string{"123"}, opts, cfg, mock)
	if err == nil {
		t.Error("Expected error when GetProject fails")
	}
}

func TestRunMoveWithDeps_GetProjectItemsFails(t *testing.T) {
	mock := newMockMoveClient()
	mock.issues["testowner/testrepo#123"] = &api.Issue{
		ID:     "issue-123",
		Number: 123,
		Title:  "Test Issue",
	}
	mock.project = &api.Project{ID: "proj-1", Number: 1, Title: "Test Project"}
	mock.getProjectItemsErr = fmt.Errorf("items API error")
	cfg := testMoveConfig()

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	opts := &moveOptions{status: "in_progress"}

	err := runMoveWithDeps(cmd, []string{"123"}, opts, cfg, mock)
	if err == nil {
		t.Error("Expected error when GetProjectItems fails")
	}
}

func TestRunMoveWithDeps_IssueNotInProject(t *testing.T) {
	mock := newMockMoveClient()
	mock.issues["testowner/testrepo#123"] = &api.Issue{
		ID:     "issue-123",
		Number: 123,
		Title:  "Test Issue",
	}
	mock.project = &api.Project{ID: "proj-1", Number: 1, Title: "Test Project"}
	mock.projectItems = []api.ProjectItem{} // Empty - issue not in project
	cfg := testMoveConfig()

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	opts := &moveOptions{status: "in_progress"}

	err := runMoveWithDeps(cmd, []string{"123"}, opts, cfg, mock)
	if err == nil {
		t.Error("Expected error when issue not in project")
	}
	if err.Error() != "issue #123 is not in the project" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestRunMoveWithDeps_SingleIssueStatusUpdate(t *testing.T) {
	mock := setupMockWithIssue(123, "Test Issue", "item-123")
	cfg := testMoveConfig()

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	opts := &moveOptions{status: "in_progress"}

	err := runMoveWithDeps(cmd, []string{"123"}, opts, cfg, mock)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify field update
	if len(mock.fieldUpdates) != 1 {
		t.Fatalf("Expected 1 field update, got %d", len(mock.fieldUpdates))
	}
	update := mock.fieldUpdates[0]
	if update.fieldName != "Status" {
		t.Errorf("Expected fieldName 'Status', got '%s'", update.fieldName)
	}
	if update.value != "In Progress" {
		t.Errorf("Expected value 'In Progress', got '%s'", update.value)
	}
}

func TestRunMoveWithDeps_SingleIssuePriorityUpdate(t *testing.T) {
	mock := setupMockWithIssue(123, "Test Issue", "item-123")
	cfg := testMoveConfig()

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	opts := &moveOptions{priority: "high"}

	err := runMoveWithDeps(cmd, []string{"123"}, opts, cfg, mock)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(mock.fieldUpdates) != 1 {
		t.Fatalf("Expected 1 field update, got %d", len(mock.fieldUpdates))
	}
	update := mock.fieldUpdates[0]
	if update.fieldName != "Priority" {
		t.Errorf("Expected fieldName 'Priority', got '%s'", update.fieldName)
	}
	if update.value != "High" {
		t.Errorf("Expected value 'High', got '%s'", update.value)
	}
}

func TestRunMoveWithDeps_BothStatusAndPriority(t *testing.T) {
	mock := setupMockWithIssue(123, "Test Issue", "item-123")
	cfg := testMoveConfig()

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	opts := &moveOptions{status: "done", priority: "low"}

	err := runMoveWithDeps(cmd, []string{"123"}, opts, cfg, mock)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(mock.fieldUpdates) != 2 {
		t.Fatalf("Expected 2 field updates, got %d", len(mock.fieldUpdates))
	}
}

func TestRunMoveWithDeps_DryRunNoChanges(t *testing.T) {
	mock := setupMockWithIssue(123, "Test Issue", "item-123")
	cfg := testMoveConfig()

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	opts := &moveOptions{status: "in_progress", dryRun: true}

	err := runMoveWithDeps(cmd, []string{"123"}, opts, cfg, mock)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Dry run should not make any changes - this is the key assertion
	if len(mock.fieldUpdates) != 0 {
		t.Errorf("Expected no field updates in dry run, got %d", len(mock.fieldUpdates))
	}
}

func TestRunMoveWithDeps_StatusUpdateFails(t *testing.T) {
	mock := setupMockWithIssue(123, "Test Issue", "item-123")
	mock.setProjectItemErr = fmt.Errorf("update failed")
	cfg := testMoveConfig()

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	opts := &moveOptions{status: "in_progress"}

	// Should not return error, just print warning
	err := runMoveWithDeps(cmd, []string{"123"}, opts, cfg, mock)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestRunMoveWithDeps_FullIssueReference(t *testing.T) {
	mock := newMockMoveClient()
	mock.project = &api.Project{ID: "proj-1", Number: 1, Title: "Test Project"}
	mock.issues["other/repo#456"] = &api.Issue{
		ID:     "issue-456",
		Number: 456,
		Title:  "Other Repo Issue",
		Repository: api.Repository{
			Owner: "other",
			Name:  "repo",
		},
	}
	mock.projectItems = []api.ProjectItem{
		{
			ID: "item-456",
			Issue: &api.Issue{
				Number: 456,
				Repository: api.Repository{
					Owner: "other",
					Name:  "repo",
				},
			},
		},
	}
	cfg := testMoveConfig()

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	opts := &moveOptions{status: "in_progress"}

	err := runMoveWithDeps(cmd, []string{"other/repo#456"}, opts, cfg, mock)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(mock.fieldUpdates) != 1 {
		t.Fatalf("Expected 1 field update, got %d", len(mock.fieldUpdates))
	}
}

// ============================================================================
// Recursive Operation Tests
// ============================================================================

func TestRunMoveWithDeps_RecursiveCollectSubIssues(t *testing.T) {
	mock := newMockMoveClient()
	mock.project = &api.Project{ID: "proj-1", Number: 1, Title: "Test Project"}

	// Parent issue
	mock.issues["testowner/testrepo#1"] = &api.Issue{
		ID:     "issue-1",
		Number: 1,
		Title:  "Parent Issue",
		Repository: api.Repository{
			Owner: "testowner",
			Name:  "testrepo",
		},
	}

	// Project items for parent and sub-issues
	mock.projectItems = []api.ProjectItem{
		{
			ID: "item-1",
			Issue: &api.Issue{
				Number: 1,
				Repository: api.Repository{
					Owner: "testowner",
					Name:  "testrepo",
				},
			},
		},
		{
			ID: "item-2",
			Issue: &api.Issue{
				Number: 2,
				Repository: api.Repository{
					Owner: "testowner",
					Name:  "testrepo",
				},
			},
		},
		{
			ID: "item-3",
			Issue: &api.Issue{
				Number: 3,
				Repository: api.Repository{
					Owner: "testowner",
					Name:  "testrepo",
				},
			},
		},
	}

	// Sub-issues - these are returned when GetSubIssues is called for issue #1
	mock.subIssues["testowner/testrepo#1"] = []api.SubIssue{
		{
			ID:     "issue-2",
			Number: 2,
			Title:  "Sub Issue 1",
			Repository: api.Repository{
				Owner: "testowner",
				Name:  "testrepo",
			},
		},
		{
			ID:     "issue-3",
			Number: 3,
			Title:  "Sub Issue 2",
			Repository: api.Repository{
				Owner: "testowner",
				Name:  "testrepo",
			},
		},
	}

	cfg := testMoveConfig()

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	opts := &moveOptions{status: "in_progress", recursive: true, yes: true, depth: 10}

	err := runMoveWithDeps(cmd, []string{"1"}, opts, cfg, mock)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should update parent + 2 sub-issues = 3 issues
	// Each issue gets 1 status update
	if len(mock.fieldUpdates) != 3 {
		t.Errorf("Expected 3 field updates (1 parent + 2 sub-issues), got %d. Updates: %+v", len(mock.fieldUpdates), mock.fieldUpdates)
	}
}

func TestRunMoveWithDeps_RecursiveDryRun(t *testing.T) {
	mock := newMockMoveClient()
	mock.project = &api.Project{ID: "proj-1", Number: 1, Title: "Test Project"}

	mock.issues["testowner/testrepo#1"] = &api.Issue{
		ID:     "issue-1",
		Number: 1,
		Title:  "Parent Issue",
	}

	mock.projectItems = []api.ProjectItem{
		{
			ID: "item-1",
			Issue: &api.Issue{
				Number: 1,
				Repository: api.Repository{
					Owner: "testowner",
					Name:  "testrepo",
				},
			},
		},
	}

	mock.subIssues["testowner/testrepo#1"] = []api.SubIssue{
		{
			ID:     "issue-2",
			Number: 2,
			Title:  "Sub Issue",
			Repository: api.Repository{
				Owner: "testowner",
				Name:  "testrepo",
			},
		},
	}

	cfg := testMoveConfig()

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	opts := &moveOptions{status: "in_progress", recursive: true, dryRun: true, depth: 10}

	err := runMoveWithDeps(cmd, []string{"1"}, opts, cfg, mock)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Dry run should not make any changes - this is the key assertion
	if len(mock.fieldUpdates) != 0 {
		t.Errorf("Expected no field updates in dry run, got %d", len(mock.fieldUpdates))
	}
}

func TestRunMoveWithDeps_RecursiveSubIssueNotInProject(t *testing.T) {
	mock := newMockMoveClient()
	mock.project = &api.Project{ID: "proj-1", Number: 1, Title: "Test Project"}

	mock.issues["testowner/testrepo#1"] = &api.Issue{
		ID:     "issue-1",
		Number: 1,
		Title:  "Parent Issue",
	}

	// Only parent is in project, sub-issue is not
	mock.projectItems = []api.ProjectItem{
		{
			ID: "item-1",
			Issue: &api.Issue{
				Number: 1,
				Repository: api.Repository{
					Owner: "testowner",
					Name:  "testrepo",
				},
			},
		},
	}

	mock.subIssues["testowner/testrepo#1"] = []api.SubIssue{
		{
			ID:     "issue-2",
			Number: 2,
			Title:  "Sub Issue Not In Project",
			Repository: api.Repository{
				Owner: "testowner",
				Name:  "testrepo",
			},
		},
	}

	cfg := testMoveConfig()

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	opts := &moveOptions{status: "in_progress", recursive: true, yes: true, depth: 10}

	err := runMoveWithDeps(cmd, []string{"1"}, opts, cfg, mock)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Only parent should be updated (sub-issue skipped because not in project)
	if len(mock.fieldUpdates) != 1 {
		t.Errorf("Expected 1 field update (parent only), got %d", len(mock.fieldUpdates))
	}
}

func TestRunMoveWithDeps_RecursiveGetSubIssuesFails(t *testing.T) {
	mock := newMockMoveClient()
	mock.project = &api.Project{ID: "proj-1", Number: 1, Title: "Test Project"}

	mock.issues["testowner/testrepo#1"] = &api.Issue{
		ID:     "issue-1",
		Number: 1,
		Title:  "Parent Issue",
	}

	mock.projectItems = []api.ProjectItem{
		{
			ID: "item-1",
			Issue: &api.Issue{
				Number: 1,
				Repository: api.Repository{
					Owner: "testowner",
					Name:  "testrepo",
				},
			},
		},
	}

	mock.getSubIssuesErr = fmt.Errorf("sub-issues API error")

	cfg := testMoveConfig()

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	opts := &moveOptions{status: "in_progress", recursive: true, yes: true, depth: 10}

	err := runMoveWithDeps(cmd, []string{"1"}, opts, cfg, mock)
	// Should return error when collecting sub-issues fails
	if err == nil {
		t.Error("Expected error when GetSubIssues fails")
	}
}

// ============================================================================
// collectSubIssuesRecursive Tests
// ============================================================================

func TestCollectSubIssuesRecursive_RespectsDepthLimit(t *testing.T) {
	mock := newMockMoveClient()

	// Create a deep hierarchy: 1 -> 2 -> 3 -> 4 -> 5
	mock.subIssues["testowner/testrepo#1"] = []api.SubIssue{
		{Number: 2, Title: "Level 1", Repository: api.Repository{Owner: "testowner", Name: "testrepo"}},
	}
	mock.subIssues["testowner/testrepo#2"] = []api.SubIssue{
		{Number: 3, Title: "Level 2", Repository: api.Repository{Owner: "testowner", Name: "testrepo"}},
	}
	mock.subIssues["testowner/testrepo#3"] = []api.SubIssue{
		{Number: 4, Title: "Level 3", Repository: api.Repository{Owner: "testowner", Name: "testrepo"}},
	}
	mock.subIssues["testowner/testrepo#4"] = []api.SubIssue{
		{Number: 5, Title: "Level 4", Repository: api.Repository{Owner: "testowner", Name: "testrepo"}},
	}

	itemIDMap := map[string]string{
		"testowner/testrepo#2": "item-2",
		"testowner/testrepo#3": "item-3",
		"testowner/testrepo#4": "item-4",
		"testowner/testrepo#5": "item-5",
	}

	// Collect with maxDepth=2 (should get levels 1 and 2, i.e., issues 2 and 3)
	result, err := collectSubIssuesRecursive(mock, "testowner", "testrepo", 1, itemIDMap, 1, 2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 issues (depth 1 and 2), got %d", len(result))
	}

	// Verify depths
	depths := make(map[int]int) // issue number -> depth
	for _, info := range result {
		depths[info.Number] = info.Depth
	}

	if depths[2] != 1 {
		t.Errorf("Expected issue #2 at depth 1, got %d", depths[2])
	}
	if depths[3] != 2 {
		t.Errorf("Expected issue #3 at depth 2, got %d", depths[3])
	}
}

func TestCollectSubIssuesRecursive_HandlesEmptySubIssues(t *testing.T) {
	mock := newMockMoveClient()
	// No sub-issues for any issue

	itemIDMap := map[string]string{}

	result, err := collectSubIssuesRecursive(mock, "testowner", "testrepo", 1, itemIDMap, 1, 10)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 sub-issues, got %d", len(result))
	}
}

func TestCollectSubIssuesRecursive_HandlesCrossRepoSubIssues(t *testing.T) {
	mock := newMockMoveClient()

	// Parent in repo A has sub-issue in repo B
	mock.subIssues["owner-a/repo-a#1"] = []api.SubIssue{
		{
			Number: 100,
			Title:  "Cross-repo sub-issue",
			Repository: api.Repository{
				Owner: "owner-b",
				Name:  "repo-b",
			},
		},
	}

	itemIDMap := map[string]string{
		"owner-b/repo-b#100": "item-100",
	}

	result, err := collectSubIssuesRecursive(mock, "owner-a", "repo-a", 1, itemIDMap, 1, 10)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 sub-issue, got %d", len(result))
	}

	// Verify cross-repo handling
	if result[0].Owner != "owner-b" {
		t.Errorf("Expected owner 'owner-b', got '%s'", result[0].Owner)
	}
	if result[0].Repo != "repo-b" {
		t.Errorf("Expected repo 'repo-b', got '%s'", result[0].Repo)
	}
	if result[0].Number != 100 {
		t.Errorf("Expected number 100, got %d", result[0].Number)
	}
}

func TestCollectSubIssuesRecursive_InheritsRepoWhenEmpty(t *testing.T) {
	mock := newMockMoveClient()

	// Sub-issue with empty repository (should inherit parent's repo)
	mock.subIssues["testowner/testrepo#1"] = []api.SubIssue{
		{
			Number: 2,
			Title:  "Same-repo sub-issue",
			Repository: api.Repository{
				Owner: "", // Empty - should inherit
				Name:  "", // Empty - should inherit
			},
		},
	}

	itemIDMap := map[string]string{
		"testowner/testrepo#2": "item-2",
	}

	result, err := collectSubIssuesRecursive(mock, "testowner", "testrepo", 1, itemIDMap, 1, 10)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 sub-issue, got %d", len(result))
	}

	// Should inherit parent's repo
	if result[0].Owner != "testowner" {
		t.Errorf("Expected owner 'testowner', got '%s'", result[0].Owner)
	}
	if result[0].Repo != "testrepo" {
		t.Errorf("Expected repo 'testrepo', got '%s'", result[0].Repo)
	}
}

func TestCollectSubIssuesRecursive_SubIssueNotInProject(t *testing.T) {
	mock := newMockMoveClient()

	mock.subIssues["testowner/testrepo#1"] = []api.SubIssue{
		{
			Number: 2,
			Title:  "Not in project",
			Repository: api.Repository{
				Owner: "testowner",
				Name:  "testrepo",
			},
		},
	}

	// Empty itemIDMap - sub-issue not in project
	itemIDMap := map[string]string{}

	result, err := collectSubIssuesRecursive(mock, "testowner", "testrepo", 1, itemIDMap, 1, 10)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 sub-issue, got %d", len(result))
	}

	// ItemID should be empty since not in project
	if result[0].ItemID != "" {
		t.Errorf("Expected empty ItemID, got '%s'", result[0].ItemID)
	}
}

func TestCollectSubIssuesRecursive_MaxDepthZero(t *testing.T) {
	mock := newMockMoveClient()

	mock.subIssues["testowner/testrepo#1"] = []api.SubIssue{
		{Number: 2, Title: "Sub", Repository: api.Repository{Owner: "testowner", Name: "testrepo"}},
	}

	itemIDMap := map[string]string{"testowner/testrepo#2": "item-2"}

	// maxDepth=0, currentDepth=1 -> should return nothing
	result, err := collectSubIssuesRecursive(mock, "testowner", "testrepo", 1, itemIDMap, 1, 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 sub-issues with maxDepth=0, got %d", len(result))
	}
}

// =============================================================================
// REQ-005: Integration with Move Command
// =============================================================================

// AC-005-1: Given `gh pmu move 42 --microsprint 2025-12-13-a`, Then Microsprint field set to specified value
func TestRunMoveWithDeps_MicrosprintExplicitValue(t *testing.T) {
	mock := setupMockWithIssue(42, "Test Issue", "item-42")
	cfg := testMoveConfig()

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	opts := &moveOptions{microsprint: "2025-12-13-a"}

	err := runMoveWithDeps(cmd, []string{"42"}, opts, cfg, mock)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify Microsprint field was set
	found := false
	for _, update := range mock.fieldUpdates {
		if update.fieldName == "Microsprint" && update.value == "2025-12-13-a" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected Microsprint field to be set to '2025-12-13-a', updates: %+v", mock.fieldUpdates)
	}
}

// AC-005-2: Given `gh pmu move 42 --microsprint current`, Then Microsprint field set to active microsprint name
func TestRunMoveWithDeps_MicrosprintCurrent(t *testing.T) {
	mock := setupMockWithIssue(42, "Test Issue", "item-42")
	// Add active microsprint - use today's date so test passes regardless of when it runs
	today := time.Now().Format("2006-01-02")
	microsprintName := today + "-a"
	mock.openIssuesByLabel["microsprint"] = []api.Issue{
		{
			ID:     "TRACKER_100",
			Number: 100,
			Title:  "Microsprint: " + microsprintName,
			State:  "OPEN",
		},
	}
	cfg := testMoveConfig()

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	opts := &moveOptions{microsprint: "current"}

	err := runMoveWithDeps(cmd, []string{"42"}, opts, cfg, mock)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify Microsprint field was set to active microsprint name
	found := false
	for _, update := range mock.fieldUpdates {
		if update.fieldName == "Microsprint" && update.value == microsprintName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected Microsprint field to be set to '%s', updates: %+v", microsprintName, mock.fieldUpdates)
	}
}

// AC-005-3: Given `--microsprint current` with no active microsprint, Then error with suggestion to run `microsprint start`
func TestRunMoveWithDeps_MicrosprintCurrentNoActive(t *testing.T) {
	mock := setupMockWithIssue(42, "Test Issue", "item-42")
	// No active microsprint
	mock.openIssuesByLabel["microsprint"] = []api.Issue{}
	cfg := testMoveConfig()

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	opts := &moveOptions{microsprint: "current"}

	err := runMoveWithDeps(cmd, []string{"42"}, opts, cfg, mock)
	if err == nil {
		t.Fatal("Expected error when no active microsprint")
	}

	if !strings.Contains(err.Error(), "no active microsprint") {
		t.Errorf("Expected error to mention 'no active microsprint', got: %v", err)
	}
	if !strings.Contains(err.Error(), "microsprint start") {
		t.Errorf("Expected error to suggest 'microsprint start', got: %v", err)
	}
}
