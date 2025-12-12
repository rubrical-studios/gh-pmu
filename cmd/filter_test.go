package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rubrical-studios/gh-pmu/internal/api"
)

func TestFilterCommand_Exists(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"filter", "--help"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("filter command should exist: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "filter") {
		t.Error("Expected help output to mention 'filter'")
	}
}

func TestFilterCommand_HasStatusFlag(t *testing.T) {
	cmd := NewRootCommand()
	filterCmd, _, err := cmd.Find([]string{"filter"})
	if err != nil {
		t.Fatalf("filter command not found: %v", err)
	}

	flag := filterCmd.Flags().Lookup("status")
	if flag == nil {
		t.Fatal("Expected --status flag to exist")
	}
	if flag.Shorthand != "s" {
		t.Errorf("Expected shorthand 's', got '%s'", flag.Shorthand)
	}
}

func TestFilterCommand_HasPriorityFlag(t *testing.T) {
	cmd := NewRootCommand()
	filterCmd, _, err := cmd.Find([]string{"filter"})
	if err != nil {
		t.Fatalf("filter command not found: %v", err)
	}

	flag := filterCmd.Flags().Lookup("priority")
	if flag == nil {
		t.Fatal("Expected --priority flag to exist")
	}
	if flag.Shorthand != "p" {
		t.Errorf("Expected shorthand 'p', got '%s'", flag.Shorthand)
	}
}

func TestFilterCommand_HasAssigneeFlag(t *testing.T) {
	cmd := NewRootCommand()
	filterCmd, _, err := cmd.Find([]string{"filter"})
	if err != nil {
		t.Fatalf("filter command not found: %v", err)
	}

	flag := filterCmd.Flags().Lookup("assignee")
	if flag == nil {
		t.Fatal("Expected --assignee flag to exist")
	}
	if flag.Shorthand != "a" {
		t.Errorf("Expected shorthand 'a', got '%s'", flag.Shorthand)
	}
}

func TestFilterCommand_HasLabelFlag(t *testing.T) {
	cmd := NewRootCommand()
	filterCmd, _, err := cmd.Find([]string{"filter"})
	if err != nil {
		t.Fatalf("filter command not found: %v", err)
	}

	flag := filterCmd.Flags().Lookup("label")
	if flag == nil {
		t.Fatal("Expected --label flag to exist")
	}
	if flag.Shorthand != "l" {
		t.Errorf("Expected shorthand 'l', got '%s'", flag.Shorthand)
	}
}

func TestFilterCommand_HasJSONFlag(t *testing.T) {
	cmd := NewRootCommand()
	filterCmd, _, err := cmd.Find([]string{"filter"})
	if err != nil {
		t.Fatalf("filter command not found: %v", err)
	}

	flag := filterCmd.Flags().Lookup("json")
	if flag == nil {
		t.Fatal("Expected --json flag to exist")
	}
	if flag.Value.Type() != "bool" {
		t.Errorf("Expected --json to be bool, got %s", flag.Value.Type())
	}
}

func TestFilterCommand_HelpText(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"filter", "--help"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("filter help failed: %v", err)
	}

	output := buf.String()

	// Should mention piping from gh issue list
	if !strings.Contains(output, "gh issue list") {
		t.Error("Expected help to mention 'gh issue list'")
	}

	// Should have example usage
	if !strings.Contains(output, "Example") {
		t.Error("Expected help to have Example section")
	}
}

// ============================================================================
// hasFieldValue Tests
// ============================================================================

func TestHasFieldValue_Filter(t *testing.T) {
	tests := []struct {
		name      string
		item      api.ProjectItem
		fieldName string
		value     string
		want      bool
	}{
		{
			name: "exact match",
			item: api.ProjectItem{
				FieldValues: []api.FieldValue{
					{Field: "Status", Value: "In Progress"},
				},
			},
			fieldName: "Status",
			value:     "In Progress",
			want:      true,
		},
		{
			name: "case insensitive field",
			item: api.ProjectItem{
				FieldValues: []api.FieldValue{
					{Field: "Status", Value: "Ready"},
				},
			},
			fieldName: "status",
			value:     "Ready",
			want:      true,
		},
		{
			name: "case insensitive value",
			item: api.ProjectItem{
				FieldValues: []api.FieldValue{
					{Field: "Status", Value: "In Progress"},
				},
			},
			fieldName: "Status",
			value:     "in progress",
			want:      true,
		},
		{
			name: "no match",
			item: api.ProjectItem{
				FieldValues: []api.FieldValue{
					{Field: "Status", Value: "Done"},
				},
			},
			fieldName: "Status",
			value:     "Ready",
			want:      false,
		},
		{
			name:      "empty field values",
			item:      api.ProjectItem{},
			fieldName: "Status",
			value:     "Ready",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasFieldValue(tt.item, tt.fieldName, tt.value)
			if got != tt.want {
				t.Errorf("hasFieldValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// hasAssignee Tests
// ============================================================================

func TestHasAssignee(t *testing.T) {
	tests := []struct {
		name     string
		issue    FilterInput
		assignee string
		want     bool
	}{
		{
			name: "exact match",
			issue: FilterInput{
				Assignees: []User{{Login: "user1"}},
			},
			assignee: "user1",
			want:     true,
		},
		{
			name: "case insensitive",
			issue: FilterInput{
				Assignees: []User{{Login: "User1"}},
			},
			assignee: "user1",
			want:     true,
		},
		{
			name: "multiple assignees",
			issue: FilterInput{
				Assignees: []User{{Login: "user1"}, {Login: "user2"}},
			},
			assignee: "user2",
			want:     true,
		},
		{
			name: "no match",
			issue: FilterInput{
				Assignees: []User{{Login: "user1"}},
			},
			assignee: "user3",
			want:     false,
		},
		{
			name:     "empty assignees",
			issue:    FilterInput{},
			assignee: "user1",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasAssignee(tt.issue, tt.assignee)
			if got != tt.want {
				t.Errorf("hasAssignee() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// hasLabel Tests
// ============================================================================

func TestHasLabel(t *testing.T) {
	tests := []struct {
		name  string
		issue FilterInput
		label string
		want  bool
	}{
		{
			name: "exact match",
			issue: FilterInput{
				Labels: []Label{{Name: "bug"}},
			},
			label: "bug",
			want:  true,
		},
		{
			name: "case insensitive",
			issue: FilterInput{
				Labels: []Label{{Name: "Bug"}},
			},
			label: "bug",
			want:  true,
		},
		{
			name: "multiple labels",
			issue: FilterInput{
				Labels: []Label{{Name: "bug"}, {Name: "enhancement"}},
			},
			label: "enhancement",
			want:  true,
		},
		{
			name: "no match",
			issue: FilterInput{
				Labels: []Label{{Name: "bug"}},
			},
			label: "feature",
			want:  false,
		},
		{
			name:  "empty labels",
			issue: FilterInput{},
			label: "bug",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasLabel(tt.issue, tt.label)
			if got != tt.want {
				t.Errorf("hasLabel() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// outputFilterTable Tests
// ============================================================================

func TestOutputFilterTable_EmptyIssues(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := createTestCmd(buf)

	err := outputFilterTable(cmd, []FilterInput{})
	if err != nil {
		t.Fatalf("outputFilterTable() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No matching issues found") {
		t.Errorf("Expected 'No matching issues found', got: %s", output)
	}
}

// ============================================================================
// outputFilterJSON Tests
// ============================================================================

func TestOutputFilterJSON_EmptyIssues(t *testing.T) {
	// outputFilterJSON writes to os.Stdout
	// We can verify it doesn't error
	err := outputFilterJSON([]FilterInput{})
	if err != nil {
		t.Fatalf("outputFilterJSON() error = %v", err)
	}
}

func TestOutputFilterJSON_WithIssues(t *testing.T) {
	issues := []FilterInput{
		{
			Number: 42,
			Title:  "Test Issue",
			State:  "open",
			URL:    "https://github.com/owner/repo/issues/42",
		},
	}

	err := outputFilterJSON(issues)
	if err != nil {
		t.Fatalf("outputFilterJSON() error = %v", err)
	}
}
