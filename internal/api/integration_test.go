//go:build integration

package api

import (
	"testing"
)

// Integration tests use the real test project:
// - Owner: rubrical-studios
// - Project: gh-pm-test-project (#13)
// - Repo: rubrical-studios/gh-pm-test
//
// Run with: go test -tags=integration ./internal/api/...

const (
	testOwner         = "rubrical-studios"
	testProjectNumber = 13
	testRepo          = "gh-pm-test"
)

func TestIntegration_GetProject(t *testing.T) {
	client := NewClient()

	project, err := client.GetProject(testOwner, testProjectNumber)
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	if project.Number != testProjectNumber {
		t.Errorf("Expected project number %d, got %d", testProjectNumber, project.Number)
	}

	if project.Title == "" {
		t.Error("Expected project title to be non-empty")
	}

	if project.Owner.Login != testOwner {
		t.Errorf("Expected owner '%s', got '%s'", testOwner, project.Owner.Login)
	}

	t.Logf("Project: %s (ID: %s)", project.Title, project.ID)
}

func TestIntegration_GetProjectFields(t *testing.T) {
	client := NewClient()

	// First get the project to get its ID
	project, err := client.GetProject(testOwner, testProjectNumber)
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	fields, err := client.GetProjectFields(project.ID)
	if err != nil {
		t.Fatalf("Failed to get project fields: %v", err)
	}

	if len(fields) == 0 {
		t.Error("Expected at least one field")
	}

	// Check for expected fields (Priority, Size were created in test setup)
	foundPriority := false
	foundStatus := false
	for _, field := range fields {
		t.Logf("Field: %s (Type: %s, Options: %d)", field.Name, field.DataType, len(field.Options))
		if field.Name == "Priority" {
			foundPriority = true
		}
		if field.Name == "Status" {
			foundStatus = true
		}
	}

	if !foundPriority {
		t.Error("Expected to find 'Priority' field")
	}
	if !foundStatus {
		t.Error("Expected to find 'Status' field")
	}
}

func TestIntegration_GetIssue(t *testing.T) {
	client := NewClient()

	issue, err := client.GetIssue(testOwner, testRepo, 1)
	if err != nil {
		t.Fatalf("Failed to get issue: %v", err)
	}

	if issue.Number != 1 {
		t.Errorf("Expected issue number 1, got %d", issue.Number)
	}

	if issue.Title == "" {
		t.Error("Expected issue title to be non-empty")
	}

	t.Logf("Issue: #%d %s (State: %s)", issue.Number, issue.Title, issue.State)
}

func TestIntegration_GetProjectItems(t *testing.T) {
	client := NewClient()

	// First get the project to get its ID
	project, err := client.GetProject(testOwner, testProjectNumber)
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	items, err := client.GetProjectItems(project.ID, nil)
	if err != nil {
		t.Fatalf("Failed to get project items: %v", err)
	}

	t.Logf("Found %d items in project", len(items))

	for _, item := range items {
		if item.Issue != nil {
			t.Logf("  Issue #%d: %s", item.Issue.Number, item.Issue.Title)
			for _, fv := range item.FieldValues {
				t.Logf("    %s: %s", fv.Field, fv.Value)
			}
		}
	}
}
