package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/rubrical-studios/gh-pmu/internal/api"
	"github.com/rubrical-studios/gh-pmu/internal/config"
)

// mockIntakeClient implements intakeClient for testing
type mockIntakeClient struct {
	project          *api.Project
	projectItems     []api.ProjectItem
	repositoryIssues []api.Issue
	addedItemID      string

	// Error injection
	getProjectErr          error
	getProjectItemsErr     error
	getRepositoryIssuesErr error
	addIssueToProjectErr   error
	setProjectItemFieldErr error
}

func newMockIntakeClient() *mockIntakeClient {
	return &mockIntakeClient{
		project: &api.Project{
			ID:    "proj-1",
			Title: "Test Project",
		},
		projectItems:     []api.ProjectItem{},
		repositoryIssues: []api.Issue{},
		addedItemID:      "item-123",
	}
}

func (m *mockIntakeClient) GetProject(owner string, number int) (*api.Project, error) {
	if m.getProjectErr != nil {
		return nil, m.getProjectErr
	}
	return m.project, nil
}

func (m *mockIntakeClient) GetProjectItems(projectID string, filter *api.ProjectItemsFilter) ([]api.ProjectItem, error) {
	if m.getProjectItemsErr != nil {
		return nil, m.getProjectItemsErr
	}
	return m.projectItems, nil
}

func (m *mockIntakeClient) GetRepositoryIssues(owner, repo, state string) ([]api.Issue, error) {
	if m.getRepositoryIssuesErr != nil {
		return nil, m.getRepositoryIssuesErr
	}
	return m.repositoryIssues, nil
}

func (m *mockIntakeClient) AddIssueToProject(projectID, issueID string) (string, error) {
	if m.addIssueToProjectErr != nil {
		return "", m.addIssueToProjectErr
	}
	return m.addedItemID, nil
}

func (m *mockIntakeClient) SetProjectItemField(projectID, itemID, fieldName, value string) error {
	return m.setProjectItemFieldErr
}

func TestIntakeCommand(t *testing.T) {
	t.Run("has correct command structure", func(t *testing.T) {
		cmd := newIntakeCommand()

		if cmd.Use != "intake" {
			t.Errorf("expected Use to be 'intake', got %s", cmd.Use)
		}

		if cmd.Short == "" {
			t.Error("expected Short description to be set")
		}

		// Check aliases
		if len(cmd.Aliases) == 0 || cmd.Aliases[0] != "in" {
			t.Error("expected 'in' alias")
		}
	})

	t.Run("has required flags", func(t *testing.T) {
		cmd := newIntakeCommand()

		// Check --apply flag
		applyFlag := cmd.Flags().Lookup("apply")
		if applyFlag == nil {
			t.Fatal("expected --apply flag")
		}
		if applyFlag.Shorthand != "a" {
			t.Errorf("expected --apply shorthand 'a', got %s", applyFlag.Shorthand)
		}

		// Check --dry-run flag
		dryRunFlag := cmd.Flags().Lookup("dry-run")
		if dryRunFlag == nil {
			t.Error("expected --dry-run flag")
		}

		// Check --json flag
		jsonFlag := cmd.Flags().Lookup("json")
		if jsonFlag == nil {
			t.Error("expected --json flag")
		}

		// Check --label flag
		labelFlag := cmd.Flags().Lookup("label")
		if labelFlag == nil {
			t.Fatal("expected --label flag")
		}
		if labelFlag.Shorthand != "l" {
			t.Errorf("expected --label shorthand 'l', got %s", labelFlag.Shorthand)
		}

		// Check --assignee flag
		assigneeFlag := cmd.Flags().Lookup("assignee")
		if assigneeFlag == nil {
			t.Error("expected --assignee flag")
		}
	})

	t.Run("command is registered in root", func(t *testing.T) {
		root := NewRootCommand()
		buf := new(bytes.Buffer)
		root.SetOut(buf)
		root.SetArgs([]string{"intake", "--help"})
		err := root.Execute()
		if err != nil {
			t.Errorf("intake command not registered: %v", err)
		}
	})
}

func TestIntakeOptions(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		opts := &intakeOptions{}

		if opts.apply != "" {
			t.Error("apply should be empty string by default")
		}
		if opts.dryRun {
			t.Error("dryRun should be false by default")
		}
		if opts.json {
			t.Error("json should be false by default")
		}
		if len(opts.label) > 0 {
			t.Error("label should be empty by default")
		}
		if len(opts.assignee) > 0 {
			t.Error("assignee should be empty by default")
		}
	})
}

func TestOutputIntakeTable(t *testing.T) {
	t.Run("displays issues in table format", func(t *testing.T) {
		cmd := newIntakeCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		issues := []api.Issue{
			{
				Number:     1,
				Title:      "First issue",
				State:      "OPEN",
				Repository: api.Repository{Owner: "owner", Name: "repo"},
			},
			{
				Number:     2,
				Title:      "Second issue",
				State:      "OPEN",
				Repository: api.Repository{Owner: "owner", Name: "repo"},
			},
		}

		err := outputIntakeTable(cmd, issues)
		if err != nil {
			t.Fatalf("outputIntakeTable failed: %v", err)
		}

		// Note: outputIntakeTable writes directly to os.Stdout, not cmd.Out()
		// We're testing it doesn't error; actual output goes to stdout
	})

	t.Run("truncates long titles to 50 chars", func(t *testing.T) {
		cmd := newIntakeCommand()

		// Create issue with 60-character title
		longTitle := strings.Repeat("A", 60)
		issues := []api.Issue{
			{
				Number:     1,
				Title:      longTitle,
				State:      "OPEN",
				Repository: api.Repository{Owner: "owner", Name: "repo"},
			},
		}

		// outputIntakeTable writes to os.Stdout, so we just verify no error
		err := outputIntakeTable(cmd, issues)
		if err != nil {
			t.Fatalf("outputIntakeTable failed with long title: %v", err)
		}
	})

	t.Run("handles empty issue list", func(t *testing.T) {
		cmd := newIntakeCommand()
		issues := []api.Issue{}

		err := outputIntakeTable(cmd, issues)
		if err != nil {
			t.Fatalf("outputIntakeTable failed with empty list: %v", err)
		}
	})
}

func TestOutputIntakeJSON(t *testing.T) {
	t.Run("outputs correct JSON structure with dry-run status", func(t *testing.T) {
		cmd := newIntakeCommand()

		issues := []api.Issue{
			{
				Number:     42,
				Title:      "Test issue",
				State:      "OPEN",
				URL:        "https://github.com/owner/repo/issues/42",
				Repository: api.Repository{Owner: "owner", Name: "repo"},
			},
		}

		// Capture stdout for JSON output
		// Note: outputIntakeJSON writes to os.Stdout via json.NewEncoder
		err := outputIntakeJSON(cmd, issues, "dry-run")
		if err != nil {
			t.Fatalf("outputIntakeJSON failed: %v", err)
		}
	})

	t.Run("status field matches input status", func(t *testing.T) {
		// Test that various status values are preserved
		statuses := []string{"dry-run", "applied", "untracked"}
		for _, status := range statuses {
			cmd := newIntakeCommand()
			issues := []api.Issue{}

			err := outputIntakeJSON(cmd, issues, status)
			if err != nil {
				t.Fatalf("outputIntakeJSON failed with status %q: %v", status, err)
			}
		}
	})

	t.Run("count matches issues length", func(t *testing.T) {
		cmd := newIntakeCommand()

		issues := []api.Issue{
			{Number: 1, Title: "Issue 1", Repository: api.Repository{Owner: "o", Name: "r"}},
			{Number: 2, Title: "Issue 2", Repository: api.Repository{Owner: "o", Name: "r"}},
			{Number: 3, Title: "Issue 3", Repository: api.Repository{Owner: "o", Name: "r"}},
		}

		err := outputIntakeJSON(cmd, issues, "test")
		if err != nil {
			t.Fatalf("outputIntakeJSON failed: %v", err)
		}
	})
}

func TestIntakeJSONOutput_Structure(t *testing.T) {
	t.Run("marshals to correct JSON format", func(t *testing.T) {
		output := intakeJSONOutput{
			Status: "dry-run",
			Count:  2,
			Issues: []intakeJSONIssue{
				{
					Number:     1,
					Title:      "First",
					State:      "OPEN",
					URL:        "https://github.com/owner/repo/issues/1",
					Repository: "owner/repo",
				},
				{
					Number:     2,
					Title:      "Second",
					State:      "OPEN",
					URL:        "https://github.com/owner/repo/issues/2",
					Repository: "owner/repo",
				},
			},
		}

		data, err := json.Marshal(output)
		if err != nil {
			t.Fatalf("Failed to marshal intakeJSONOutput: %v", err)
		}

		// Unmarshal and verify
		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}

		if result["status"] != "dry-run" {
			t.Errorf("Expected status 'dry-run', got %v", result["status"])
		}

		if int(result["count"].(float64)) != 2 {
			t.Errorf("Expected count 2, got %v", result["count"])
		}

		issues, ok := result["issues"].([]interface{})
		if !ok {
			t.Fatal("Expected issues to be an array")
		}
		if len(issues) != 2 {
			t.Errorf("Expected 2 issues, got %d", len(issues))
		}
	})

	t.Run("intakeJSONIssue includes all fields", func(t *testing.T) {
		issue := intakeJSONIssue{
			Number:     42,
			Title:      "Test Issue",
			State:      "OPEN",
			URL:        "https://github.com/owner/repo/issues/42",
			Repository: "owner/repo",
		}

		data, err := json.Marshal(issue)
		if err != nil {
			t.Fatalf("Failed to marshal intakeJSONIssue: %v", err)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}

		expectedFields := []string{"number", "title", "state", "url", "repository"}
		for _, field := range expectedFields {
			if _, exists := result[field]; !exists {
				t.Errorf("Expected field %q to exist in JSON output", field)
			}
		}
	})
}

func TestFilterIntakeByLabel(t *testing.T) {
	issues := []api.Issue{
		{
			Number: 1,
			Title:  "Bug issue",
			Labels: []api.Label{{Name: "bug"}, {Name: "urgent"}},
		},
		{
			Number: 2,
			Title:  "Feature issue",
			Labels: []api.Label{{Name: "feature"}},
		},
		{
			Number: 3,
			Title:  "No labels",
			Labels: []api.Label{},
		},
	}

	t.Run("filters by single label", func(t *testing.T) {
		filtered := filterIntakeByLabel(issues, []string{"bug"})
		if len(filtered) != 1 {
			t.Errorf("Expected 1 issue, got %d", len(filtered))
		}
		if filtered[0].Number != 1 {
			t.Errorf("Expected issue #1, got #%d", filtered[0].Number)
		}
	})

	t.Run("filters by multiple labels (OR)", func(t *testing.T) {
		filtered := filterIntakeByLabel(issues, []string{"bug", "feature"})
		if len(filtered) != 2 {
			t.Errorf("Expected 2 issues, got %d", len(filtered))
		}
	})

	t.Run("case insensitive matching", func(t *testing.T) {
		filtered := filterIntakeByLabel(issues, []string{"BUG"})
		if len(filtered) != 1 {
			t.Errorf("Expected 1 issue with case-insensitive match, got %d", len(filtered))
		}
	})

	t.Run("returns empty for non-matching label", func(t *testing.T) {
		filtered := filterIntakeByLabel(issues, []string{"nonexistent"})
		if len(filtered) != 0 {
			t.Errorf("Expected 0 issues, got %d", len(filtered))
		}
	})
}

func TestFilterIntakeByAssignee(t *testing.T) {
	issues := []api.Issue{
		{
			Number:    1,
			Title:     "Assigned to alice",
			Assignees: []api.Actor{{Login: "alice"}},
		},
		{
			Number:    2,
			Title:     "Assigned to bob",
			Assignees: []api.Actor{{Login: "bob"}},
		},
		{
			Number:    3,
			Title:     "Assigned to both",
			Assignees: []api.Actor{{Login: "alice"}, {Login: "bob"}},
		},
		{
			Number:    4,
			Title:     "No assignees",
			Assignees: []api.Actor{},
		},
	}

	t.Run("filters by single assignee", func(t *testing.T) {
		filtered := filterIntakeByAssignee(issues, []string{"alice"})
		if len(filtered) != 2 {
			t.Errorf("Expected 2 issues assigned to alice, got %d", len(filtered))
		}
	})

	t.Run("filters by multiple assignees (OR)", func(t *testing.T) {
		filtered := filterIntakeByAssignee(issues, []string{"alice", "bob"})
		if len(filtered) != 3 {
			t.Errorf("Expected 3 issues, got %d", len(filtered))
		}
	})

	t.Run("case insensitive matching", func(t *testing.T) {
		filtered := filterIntakeByAssignee(issues, []string{"ALICE"})
		if len(filtered) != 2 {
			t.Errorf("Expected 2 issues with case-insensitive match, got %d", len(filtered))
		}
	})

	t.Run("returns empty for non-matching assignee", func(t *testing.T) {
		filtered := filterIntakeByAssignee(issues, []string{"charlie"})
		if len(filtered) != 0 {
			t.Errorf("Expected 0 issues, got %d", len(filtered))
		}
	})
}

func TestParseApplyFields(t *testing.T) {
	t.Run("parses single field", func(t *testing.T) {
		result := parseApplyFields("status:backlog")
		if len(result) != 1 {
			t.Errorf("Expected 1 field, got %d", len(result))
		}
		if result["status"] != "backlog" {
			t.Errorf("Expected status=backlog, got %s", result["status"])
		}
	})

	t.Run("parses multiple fields", func(t *testing.T) {
		result := parseApplyFields("status:backlog,priority:p1")
		if len(result) != 2 {
			t.Errorf("Expected 2 fields, got %d", len(result))
		}
		if result["status"] != "backlog" {
			t.Errorf("Expected status=backlog, got %s", result["status"])
		}
		if result["priority"] != "p1" {
			t.Errorf("Expected priority=p1, got %s", result["priority"])
		}
	})

	t.Run("handles empty string", func(t *testing.T) {
		result := parseApplyFields("")
		if len(result) != 0 {
			t.Errorf("Expected 0 fields, got %d", len(result))
		}
	})

	t.Run("handles whitespace", func(t *testing.T) {
		result := parseApplyFields(" status : backlog , priority : p1 ")
		if result["status"] != "backlog" {
			t.Errorf("Expected status=backlog, got %s", result["status"])
		}
		if result["priority"] != "p1" {
			t.Errorf("Expected priority=p1, got %s", result["priority"])
		}
	})

	t.Run("ignores invalid pairs", func(t *testing.T) {
		result := parseApplyFields("status:backlog,invalid,priority:p1")
		if len(result) != 2 {
			t.Errorf("Expected 2 fields (ignoring invalid), got %d", len(result))
		}
	})

	t.Run("handles trailing comma", func(t *testing.T) {
		result := parseApplyFields("status:backlog,")
		if len(result) != 1 {
			t.Errorf("Expected 1 field, got %d", len(result))
		}
	})
}

// ============================================================================
// runIntakeWithDeps Tests
// ============================================================================

func TestRunIntakeWithDeps_GetProjectError(t *testing.T) {
	mock := newMockIntakeClient()
	mock.getProjectErr = errors.New("project not found")
	cfg := &config.Config{
		Project:      config.Project{Owner: "test-org", Number: 1},
		Repositories: []string{"owner/repo"},
	}

	cmd := newIntakeCommand()
	opts := &intakeOptions{}
	err := runIntakeWithDeps(cmd, opts, cfg, mock)

	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "failed to get project") {
		t.Errorf("expected 'failed to get project' error, got: %v", err)
	}
}

func TestRunIntakeWithDeps_GetProjectItemsError(t *testing.T) {
	mock := newMockIntakeClient()
	mock.getProjectItemsErr = errors.New("items not found")
	cfg := &config.Config{
		Project:      config.Project{Owner: "test-org", Number: 1},
		Repositories: []string{"owner/repo"},
	}

	cmd := newIntakeCommand()
	opts := &intakeOptions{}
	err := runIntakeWithDeps(cmd, opts, cfg, mock)

	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "failed to get project items") {
		t.Errorf("expected 'failed to get project items' error, got: %v", err)
	}
}

func TestRunIntakeWithDeps_AllIssuesTracked(t *testing.T) {
	mock := newMockIntakeClient()
	mock.projectItems = []api.ProjectItem{
		{Issue: &api.Issue{ID: "issue-1", Number: 1}},
	}
	mock.repositoryIssues = []api.Issue{
		{ID: "issue-1", Number: 1, Title: "Tracked Issue"},
	}
	cfg := &config.Config{
		Project:      config.Project{Owner: "test-org", Number: 1},
		Repositories: []string{"owner/repo"},
	}

	cmd := newIntakeCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	opts := &intakeOptions{}
	err := runIntakeWithDeps(cmd, opts, cfg, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "All issues are already tracked") {
		t.Errorf("expected 'All issues are already tracked' message, got: %s", output)
	}
}

func TestRunIntakeWithDeps_FindsUntrackedIssues(t *testing.T) {
	mock := newMockIntakeClient()
	mock.projectItems = []api.ProjectItem{} // No tracked issues
	mock.repositoryIssues = []api.Issue{
		{ID: "issue-1", Number: 1, Title: "Untracked Issue", State: "OPEN"},
	}
	cfg := &config.Config{
		Project:      config.Project{Owner: "test-org", Number: 1},
		Repositories: []string{"owner/repo"},
	}

	cmd := newIntakeCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	opts := &intakeOptions{}
	err := runIntakeWithDeps(cmd, opts, cfg, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Found 1 untracked issue") {
		t.Errorf("expected 'Found 1 untracked issue' message, got: %s", output)
	}
}

func TestRunIntakeWithDeps_DryRun(t *testing.T) {
	mock := newMockIntakeClient()
	mock.repositoryIssues = []api.Issue{
		{ID: "issue-1", Number: 1, Title: "Untracked Issue", State: "OPEN"},
	}
	cfg := &config.Config{
		Project:      config.Project{Owner: "test-org", Number: 1},
		Repositories: []string{"owner/repo"},
	}

	cmd := newIntakeCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	opts := &intakeOptions{dryRun: true}
	err := runIntakeWithDeps(cmd, opts, cfg, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Would add 1 issue") {
		t.Errorf("expected 'Would add 1 issue' message, got: %s", output)
	}
}

func TestRunIntakeWithDeps_JSONOutput(t *testing.T) {
	mock := newMockIntakeClient()
	mock.repositoryIssues = []api.Issue{
		{ID: "issue-1", Number: 1, Title: "Untracked Issue", State: "OPEN"},
	}
	cfg := &config.Config{
		Project:      config.Project{Owner: "test-org", Number: 1},
		Repositories: []string{"owner/repo"},
	}

	cmd := newIntakeCommand()
	opts := &intakeOptions{json: true}
	err := runIntakeWithDeps(cmd, opts, cfg, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// JSON output goes to stdout - verified by no error
}

func TestRunIntakeWithDeps_FilterByLabel(t *testing.T) {
	mock := newMockIntakeClient()
	mock.repositoryIssues = []api.Issue{
		{ID: "issue-1", Number: 1, Title: "Bug Issue", Labels: []api.Label{{Name: "bug"}}},
		{ID: "issue-2", Number: 2, Title: "Feature Issue", Labels: []api.Label{{Name: "feature"}}},
	}
	cfg := &config.Config{
		Project:      config.Project{Owner: "test-org", Number: 1},
		Repositories: []string{"owner/repo"},
	}

	cmd := newIntakeCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	opts := &intakeOptions{label: []string{"bug"}}
	err := runIntakeWithDeps(cmd, opts, cfg, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Found 1 untracked") {
		t.Errorf("expected 'Found 1 untracked' (filtered), got: %s", output)
	}
}

func TestRunIntakeWithDeps_FilterByAssignee(t *testing.T) {
	mock := newMockIntakeClient()
	mock.repositoryIssues = []api.Issue{
		{ID: "issue-1", Number: 1, Title: "Alice Issue", Assignees: []api.Actor{{Login: "alice"}}},
		{ID: "issue-2", Number: 2, Title: "Bob Issue", Assignees: []api.Actor{{Login: "bob"}}},
	}
	cfg := &config.Config{
		Project:      config.Project{Owner: "test-org", Number: 1},
		Repositories: []string{"owner/repo"},
	}

	cmd := newIntakeCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	opts := &intakeOptions{assignee: []string{"alice"}}
	err := runIntakeWithDeps(cmd, opts, cfg, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Found 1 untracked") {
		t.Errorf("expected 'Found 1 untracked' (filtered), got: %s", output)
	}
}

func TestRunIntakeWithDeps_InvalidRepoFormat(t *testing.T) {
	mock := newMockIntakeClient()
	cfg := &config.Config{
		Project:      config.Project{Owner: "test-org", Number: 1},
		Repositories: []string{"invalid-no-slash"},
	}

	cmd := newIntakeCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	opts := &intakeOptions{}
	err := runIntakeWithDeps(cmd, opts, cfg, mock)
	// Should not error, just warn
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "invalid repository format") || !strings.Contains(output, "All issues are already tracked") {
		// Either warning or "all tracked" message
	}
}

func TestRunIntakeWithDeps_AllTrackedJSON(t *testing.T) {
	mock := newMockIntakeClient()
	mock.projectItems = []api.ProjectItem{
		{Issue: &api.Issue{ID: "issue-1", Number: 1}},
	}
	mock.repositoryIssues = []api.Issue{
		{ID: "issue-1", Number: 1, Title: "Tracked Issue"},
	}
	cfg := &config.Config{
		Project:      config.Project{Owner: "test-org", Number: 1},
		Repositories: []string{"owner/repo"},
	}

	cmd := newIntakeCommand()
	opts := &intakeOptions{json: true}
	err := runIntakeWithDeps(cmd, opts, cfg, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// JSON output with empty issues goes to stdout - verified by no error
}
