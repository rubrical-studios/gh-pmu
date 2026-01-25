//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// TestViewJSONWithStandardFields tests --json with standard fields
func TestViewJSONWithStandardFields(t *testing.T) {
	cfg := setupTestConfig(t)

	// Create a test issue
	var testIssueNum int
	defer func() {
		if testIssueNum > 0 {
			runCleanupAfterTest(t, testIssueNum)
		}
	}()

	t.Run("create test issue", func(t *testing.T) {
		testIssueNum = createTestIssue(t, cfg, "JSON View Test Issue")
	})

	t.Run("view with json standard fields", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "view", itoa(testIssueNum), "--json", "number,title,state")
		assertExitCode(t, result, 0)

		// Parse JSON output
		var output map[string]interface{}
		if err := json.Unmarshal([]byte(result.Stdout), &output); err != nil {
			t.Fatalf("failed to parse JSON output: %v\nOutput: %s", err, result.Stdout)
		}

		// Verify fields exist
		if output["number"] == nil {
			t.Error("expected 'number' field in JSON output")
		}
		if output["title"] == nil {
			t.Error("expected 'title' field in JSON output")
		}
		if output["state"] == nil {
			t.Error("expected 'state' field in JSON output")
		}

		// Verify title matches
		if title, ok := output["title"].(string); ok {
			if !strings.Contains(title, "JSON View Test Issue") {
				t.Errorf("expected title to contain 'JSON View Test Issue', got %q", title)
			}
		}
	})
}

// TestViewJSONWithProjectFields tests --json with project-specific fields
func TestViewJSONWithProjectFields(t *testing.T) {
	cfg := setupTestConfig(t)

	// Create a test issue and set project fields
	var testIssueNum int
	defer func() {
		if testIssueNum > 0 {
			runCleanupAfterTest(t, testIssueNum)
		}
	}()

	t.Run("create and configure test issue", func(t *testing.T) {
		testIssueNum = createTestIssue(t, cfg, "JSON Project Fields Test")

		// Move to in_progress to set status field
		result := runPMU(t, cfg.Dir, "move", itoa(testIssueNum), "--status", "in_progress")
		assertExitCode(t, result, 0)
	})

	t.Run("view with json project fields", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "view", itoa(testIssueNum), "--json", "fieldValues")
		assertExitCode(t, result, 0)

		// Parse JSON output
		var output map[string]interface{}
		if err := json.Unmarshal([]byte(result.Stdout), &output); err != nil {
			t.Fatalf("failed to parse JSON output: %v\nOutput: %s", err, result.Stdout)
		}

		// Verify fieldValues exists
		fieldValues, ok := output["fieldValues"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected 'fieldValues' to be a map, got %T", output["fieldValues"])
		}

		// Verify Status field is set
		if status, ok := fieldValues["Status"].(string); ok {
			if status == "" {
				t.Error("expected Status field to have a value")
			}
		} else {
			t.Error("expected 'Status' field in fieldValues")
		}
	})
}

// TestViewJSONFieldsList tests --json-fields to list available fields
func TestViewJSONFieldsList(t *testing.T) {
	cfg := setupTestConfig(t)

	t.Run("list available json fields", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "view", "--json-fields")
		assertExitCode(t, result, 0)

		// Verify some expected fields are listed
		assertContains(t, result.Stdout, "number")
		assertContains(t, result.Stdout, "title")
		assertContains(t, result.Stdout, "state")
		assertContains(t, result.Stdout, "fieldValues")
	})
}

// itoa converts int to string (helper for test readability)
func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}
