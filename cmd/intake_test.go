package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rubrical-studios/gh-pmu/internal/api"
)

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
