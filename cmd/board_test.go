package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rubrical-studios/gh-pmu/internal/api"
	"github.com/rubrical-studios/gh-pmu/internal/config"
)

func TestNewBoardCommand(t *testing.T) {
	cmd := newBoardCommand()

	// Verify command basics
	if cmd.Use != "board" {
		t.Errorf("unexpected Use: %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	// Verify flags exist
	flags := []struct {
		name      string
		shorthand string
	}{
		{"status", "s"},
		{"priority", "p"},
		{"limit", "n"},
		{"no-border", ""},
		{"json", ""},
	}

	for _, f := range flags {
		flag := cmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("flag --%s not found", f.name)
			continue
		}
		if f.shorthand != "" && flag.Shorthand != f.shorthand {
			t.Errorf("flag --%s shorthand = %q, want %q", f.name, flag.Shorthand, f.shorthand)
		}
	}
}

func TestGetStatusColumns_FromConfig(t *testing.T) {
	cfg := &config.Config{
		Fields: map[string]config.Field{
			"status": {
				Field: "Status",
				Values: map[string]string{
					"backlog":     "Backlog",
					"in_progress": "In progress",
					"done":        "Done",
				},
			},
		},
	}

	columns := getStatusColumns(cfg)

	// Should have 3 columns
	if len(columns) != 3 {
		t.Errorf("expected 3 columns, got %d", len(columns))
	}

	// Verify order follows preferredOrder
	expectedOrder := []string{"backlog", "in_progress", "done"}
	for i, expected := range expectedOrder {
		if i >= len(columns) {
			break
		}
		if columns[i].alias != expected {
			t.Errorf("column %d: expected alias %q, got %q", i, expected, columns[i].alias)
		}
	}
}

func TestGetStatusColumns_Fallback(t *testing.T) {
	cfg := &config.Config{
		Fields: map[string]config.Field{},
	}

	columns := getStatusColumns(cfg)

	// Should have default fallback columns
	if len(columns) == 0 {
		t.Error("expected fallback columns, got none")
	}

	// Check that backlog is first
	if columns[0].alias != "backlog" {
		t.Errorf("expected first column to be 'backlog', got %q", columns[0].alias)
	}
}

func TestGroupByStatus(t *testing.T) {
	columns := []statusColumn{
		{alias: "backlog", value: "Backlog"},
		{alias: "done", value: "Done"},
	}

	items := []api.ProjectItem{
		{
			ID:    "1",
			Issue: &api.Issue{Number: 1, Title: "Issue 1"},
			FieldValues: []api.FieldValue{
				{Field: "Status", Value: "Backlog"},
			},
		},
		{
			ID:    "2",
			Issue: &api.Issue{Number: 2, Title: "Issue 2"},
			FieldValues: []api.FieldValue{
				{Field: "Status", Value: "Done"},
			},
		},
		{
			ID:    "3",
			Issue: &api.Issue{Number: 3, Title: "Issue 3"},
			FieldValues: []api.FieldValue{
				{Field: "Status", Value: "Backlog"},
			},
		},
	}

	grouped := groupByStatus(items, columns)

	if len(grouped["Backlog"]) != 2 {
		t.Errorf("expected 2 items in Backlog, got %d", len(grouped["Backlog"]))
	}

	if len(grouped["Done"]) != 1 {
		t.Errorf("expected 1 item in Done, got %d", len(grouped["Done"]))
	}
}

func TestGroupByStatus_SkipsNilIssues(t *testing.T) {
	columns := []statusColumn{
		{alias: "backlog", value: "Backlog"},
	}

	items := []api.ProjectItem{
		{
			ID:    "1",
			Issue: nil, // nil issue should be skipped
			FieldValues: []api.FieldValue{
				{Field: "Status", Value: "Backlog"},
			},
		},
		{
			ID:    "2",
			Issue: &api.Issue{Number: 2, Title: "Issue 2"},
			FieldValues: []api.FieldValue{
				{Field: "Status", Value: "Backlog"},
			},
		},
	}

	grouped := groupByStatus(items, columns)

	if len(grouped["Backlog"]) != 1 {
		t.Errorf("expected 1 item in Backlog (nil issue skipped), got %d", len(grouped["Backlog"]))
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a long string", 10, "this is..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc"},
		{"", 5, ""},
	}

	for _, tt := range tests {
		result := truncateString(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

func TestOutputBoardSimple(t *testing.T) {
	columns := []statusColumn{
		{alias: "backlog", value: "Backlog"},
		{alias: "done", value: "Done"},
	}

	grouped := map[string][]api.ProjectItem{
		"Backlog": {
			{
				ID:    "1",
				Issue: &api.Issue{Number: 1, Title: "Test Issue"},
			},
		},
		"Done": {},
	}

	cmd := newBoardCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := outputBoardSimple(cmd, grouped, columns)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Check that headers are present
	if !strings.Contains(output, "## Backlog (1)") {
		t.Error("expected Backlog header with count")
	}

	if !strings.Contains(output, "## Done (0)") {
		t.Error("expected Done header with count")
	}

	// Check that issue is listed
	if !strings.Contains(output, "#1 Test Issue") {
		t.Error("expected issue #1 in output")
	}

	// Check that empty column shows (empty)
	if !strings.Contains(output, "(empty)") {
		t.Error("expected (empty) for Done column")
	}
}

func TestOutputBoardJSON(t *testing.T) {
	columns := []statusColumn{
		{alias: "backlog", value: "Backlog"},
	}

	grouped := map[string][]api.ProjectItem{
		"Backlog": {
			{
				ID:    "1",
				Issue: &api.Issue{Number: 42, Title: "JSON Test"},
				FieldValues: []api.FieldValue{
					{Field: "Priority", Value: "P1"},
				},
			},
		},
	}

	cmd := newBoardCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := outputBoardJSON(cmd, grouped, columns)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Check JSON structure
	if !strings.Contains(output, `"status": "Backlog"`) {
		t.Error("expected status in JSON")
	}

	if !strings.Contains(output, `"number": 42`) {
		t.Error("expected issue number in JSON")
	}

	if !strings.Contains(output, `"title": "JSON Test"`) {
		t.Error("expected issue title in JSON")
	}

	if !strings.Contains(output, `"priority": "P1"`) {
		t.Error("expected priority in JSON")
	}
}

func TestOutputBoardBox(t *testing.T) {
	columns := []statusColumn{
		{alias: "backlog", value: "Backlog"},
	}

	grouped := map[string][]api.ProjectItem{
		"Backlog": {
			{
				ID:    "1",
				Issue: &api.Issue{Number: 1, Title: "Box Test"},
			},
		},
	}

	cmd := newBoardCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := outputBoardBox(cmd, grouped, columns, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Check box characters are present
	if !strings.Contains(output, "┌") {
		t.Error("expected top-left corner")
	}

	if !strings.Contains(output, "┐") {
		t.Error("expected top-right corner")
	}

	if !strings.Contains(output, "└") {
		t.Error("expected bottom-left corner")
	}

	if !strings.Contains(output, "┘") {
		t.Error("expected bottom-right corner")
	}

	// Check content
	if !strings.Contains(output, "Backlog") {
		t.Error("expected Backlog header")
	}

	if !strings.Contains(output, "#1") {
		t.Error("expected issue #1")
	}
}
