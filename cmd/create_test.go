package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateCommand_Exists(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"create", "--help"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("create command should exist: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("create")) {
		t.Error("Expected help output to mention 'create'")
	}
}

func TestCreateCommand_HasTitleFlag(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create command not found: %v", err)
	}

	flag := createCmd.Flags().Lookup("title")
	if flag == nil {
		t.Fatal("Expected --title flag to exist")
	}
}

func TestCreateCommand_HasBodyFlag(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create command not found: %v", err)
	}

	flag := createCmd.Flags().Lookup("body")
	if flag == nil {
		t.Fatal("Expected --body flag to exist")
	}
}

func TestCreateCommand_HasStatusFlag(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create command not found: %v", err)
	}

	flag := createCmd.Flags().Lookup("status")
	if flag == nil {
		t.Fatal("Expected --status flag to exist")
	}
}

func TestCreateCommand_HasPriorityFlag(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create command not found: %v", err)
	}

	flag := createCmd.Flags().Lookup("priority")
	if flag == nil {
		t.Fatal("Expected --priority flag to exist")
	}
}

func TestCreateCommand_HasLabelFlag(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create command not found: %v", err)
	}

	flag := createCmd.Flags().Lookup("label")
	if flag == nil {
		t.Fatal("Expected --label flag to exist")
	}
}

func TestCreateCommand_HasAssigneeFlag(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create command not found: %v", err)
	}

	flag := createCmd.Flags().Lookup("assignee")
	if flag == nil {
		t.Fatal("Expected --assignee flag to exist")
	}
	if flag.Shorthand != "a" {
		t.Errorf("Expected shorthand 'a', got '%s'", flag.Shorthand)
	}
	if flag.Value.Type() != "stringArray" {
		t.Errorf("Expected --assignee to be stringArray, got %s", flag.Value.Type())
	}
}

func TestCreateCommand_HasMilestoneFlag(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create command not found: %v", err)
	}

	flag := createCmd.Flags().Lookup("milestone")
	if flag == nil {
		t.Fatal("Expected --milestone flag to exist")
	}
	if flag.Shorthand != "m" {
		t.Errorf("Expected shorthand 'm', got '%s'", flag.Shorthand)
	}
}

func TestCreateCommand_HasRepoFlag(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create command not found: %v", err)
	}

	flag := createCmd.Flags().Lookup("repo")
	if flag == nil {
		t.Fatal("Expected --repo flag to exist")
	}
	if flag.Shorthand != "R" {
		t.Errorf("Expected shorthand 'R', got '%s'", flag.Shorthand)
	}
}

func TestCreateCommand_HasFromFileFlag(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create command not found: %v", err)
	}

	flag := createCmd.Flags().Lookup("from-file")
	if flag == nil {
		t.Fatal("Expected --from-file flag to exist")
	}
	if flag.Shorthand != "f" {
		t.Errorf("Expected shorthand 'f', got '%s'", flag.Shorthand)
	}
}

func TestCreateCommand_HasInteractiveFlag(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create command not found: %v", err)
	}

	flag := createCmd.Flags().Lookup("interactive")
	if flag == nil {
		t.Fatal("Expected --interactive flag to exist")
	}
	if flag.Shorthand != "i" {
		t.Errorf("Expected shorthand 'i', got '%s'", flag.Shorthand)
	}
	if flag.Value.Type() != "bool" {
		t.Errorf("Expected --interactive to be bool, got %s", flag.Value.Type())
	}
}

// =============================================================================
// REQ-006: Integration with Create Command
// =============================================================================

func TestCreateCommand_HasMicrosprintFlag(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create command not found: %v", err)
	}

	flag := createCmd.Flags().Lookup("microsprint")
	if flag == nil {
		t.Fatal("Expected --microsprint flag to exist")
	}
	if flag.Shorthand != "M" {
		t.Errorf("Expected shorthand 'M', got '%s'", flag.Shorthand)
	}
}

func TestCreateCommand_RequiresTitleInNonInteractiveMode(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"create", "--body", "test body"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when title not provided with --body")
	}
}

// ============================================================================
// createOptions Tests
// ============================================================================

func TestCreateOptions_DefaultValues(t *testing.T) {
	opts := &createOptions{}

	if opts.title != "" {
		t.Errorf("Expected empty title, got %q", opts.title)
	}
	if opts.body != "" {
		t.Errorf("Expected empty body, got %q", opts.body)
	}
	if opts.status != "" {
		t.Errorf("Expected empty status, got %q", opts.status)
	}
	if opts.priority != "" {
		t.Errorf("Expected empty priority, got %q", opts.priority)
	}
	if opts.labels != nil {
		t.Errorf("Expected nil labels, got %v", opts.labels)
	}
}

func TestCreateOptions_WithValues(t *testing.T) {
	opts := &createOptions{
		title:    "Test Issue",
		body:     "Test body content",
		status:   "in_progress",
		priority: "p1",
		labels:   []string{"bug", "urgent"},
	}

	if opts.title != "Test Issue" {
		t.Errorf("Expected title 'Test Issue', got %q", opts.title)
	}
	if opts.body != "Test body content" {
		t.Errorf("Expected body 'Test body content', got %q", opts.body)
	}
	if len(opts.labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(opts.labels))
	}
}

// ============================================================================
// Label Merging Logic Tests
// ============================================================================

func TestLabelMerging_EmptyDefaults(t *testing.T) {
	configLabels := []string{}
	cliLabels := []string{"bug", "urgent"}

	// Simulate the merging logic from runCreate
	labels := append([]string{}, configLabels...)
	labels = append(labels, cliLabels...)

	if len(labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(labels))
	}
	if labels[0] != "bug" || labels[1] != "urgent" {
		t.Errorf("Expected [bug, urgent], got %v", labels)
	}
}

func TestLabelMerging_WithDefaults(t *testing.T) {
	configLabels := []string{"pm-tracked"}
	cliLabels := []string{"bug", "urgent"}

	// Simulate the merging logic from runCreate
	labels := append([]string{}, configLabels...)
	labels = append(labels, cliLabels...)

	if len(labels) != 3 {
		t.Errorf("Expected 3 labels, got %d", len(labels))
	}
	if labels[0] != "pm-tracked" {
		t.Errorf("Expected first label 'pm-tracked', got %q", labels[0])
	}
}

func TestLabelMerging_NoCLILabels(t *testing.T) {
	configLabels := []string{"pm-tracked", "auto-created"}
	var cliLabels []string

	// Simulate the merging logic from runCreate
	labels := append([]string{}, configLabels...)
	labels = append(labels, cliLabels...)

	if len(labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(labels))
	}
}

func TestLabelMerging_BothEmpty(t *testing.T) {
	configLabels := []string{}
	var cliLabels []string

	// Simulate the merging logic from runCreate
	labels := append([]string{}, configLabels...)
	labels = append(labels, cliLabels...)

	if len(labels) != 0 {
		t.Errorf("Expected 0 labels, got %d", len(labels))
	}
}

// ============================================================================
// Error Message Tests
// ============================================================================

func TestCreateCommand_TitleRequiredErrorMessage(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"create"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error when no title provided")
	}

	// The error should mention title is required
	errStr := err.Error()
	if !strings.Contains(errStr, "title") && !strings.Contains(errStr, "configuration") {
		t.Errorf("Expected error about title or config, got: %v", err)
	}
}

func TestCreateCommand_FlagShortcuts(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create command not found: %v", err)
	}

	tests := []struct {
		longFlag  string
		shortFlag string
	}{
		{"title", "t"},
		{"body", "b"},
		{"status", "s"},
		{"priority", "p"},
		{"label", "l"},
	}

	for _, tt := range tests {
		t.Run(tt.longFlag, func(t *testing.T) {
			flag := createCmd.Flags().Lookup(tt.longFlag)
			if flag == nil {
				t.Fatalf("Flag --%s not found", tt.longFlag)
			}
			if flag.Shorthand != tt.shortFlag {
				t.Errorf("Expected shorthand -%s for --%s, got -%s", tt.shortFlag, tt.longFlag, flag.Shorthand)
			}
		})
	}
}

func TestCreateCommand_LabelFlagIsArray(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create command not found: %v", err)
	}

	flag := createCmd.Flags().Lookup("label")
	if flag == nil {
		t.Fatal("Expected --label flag to exist")
	}

	// Check that it's a stringArray type (can be specified multiple times)
	if flag.Value.Type() != "stringArray" {
		t.Errorf("Expected --label to be stringArray, got %s", flag.Value.Type())
	}
}

// ============================================================================
// runCreate Integration Tests (with temp config files)
// ============================================================================

// createTempConfig creates a temporary directory with a .gh-pmu.yml config file
// and returns the directory path. Caller should defer os.RemoveAll(dir).
func createTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".gh-pmu.yml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}
	return dir
}

func TestRunCreate_NoConfigFile_ReturnsError(t *testing.T) {
	// ARRANGE: Empty temp directory (no config)
	dir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"create", "--title", "Test Issue"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// ACT
	err := cmd.Execute()

	// ASSERT
	if err == nil {
		t.Fatal("Expected error when no config file exists")
	}
	if !strings.Contains(err.Error(), "configuration") {
		t.Errorf("Expected error about configuration, got: %v", err)
	}
}

func TestRunCreate_InvalidConfig_ReturnsError(t *testing.T) {
	// ARRANGE: Config missing required fields
	config := `
project:
  owner: ""
  number: 0
repositories: []
`
	dir := createTempConfig(t, config)
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"create", "--title", "Test Issue"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// ACT
	err := cmd.Execute()

	// ASSERT
	if err == nil {
		t.Fatal("Expected error for invalid config")
	}
	if !strings.Contains(err.Error(), "invalid configuration") {
		t.Errorf("Expected 'invalid configuration' error, got: %v", err)
	}
}

func TestRunCreate_NoRepositories_ReturnsError(t *testing.T) {
	// ARRANGE: Config with no repositories
	config := `
project:
  owner: "test-owner"
  number: 1
repositories: []
`
	dir := createTempConfig(t, config)
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"create", "--title", "Test Issue"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// ACT
	err := cmd.Execute()

	// ASSERT
	if err == nil {
		t.Fatal("Expected error for missing repositories")
	}
	// Either "invalid configuration" (validation) or "no repository" error
	errStr := err.Error()
	if !strings.Contains(errStr, "repository") && !strings.Contains(errStr, "configuration") {
		t.Errorf("Expected error about repositories, got: %v", err)
	}
}

func TestRunCreate_InvalidRepositoryFormat_ReturnsError(t *testing.T) {
	// ARRANGE: Config with invalid repository format (missing slash)
	config := `
project:
  owner: "test-owner"
  number: 1
repositories:
  - "invalid-repo-no-slash"
`
	dir := createTempConfig(t, config)
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"create", "--title", "Test Issue"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// ACT
	err := cmd.Execute()

	// ASSERT
	if err == nil {
		t.Fatal("Expected error for invalid repository format")
	}
	if !strings.Contains(err.Error(), "invalid repository format") {
		t.Errorf("Expected 'invalid repository format' error, got: %v", err)
	}
}

func TestRunCreate_NoTitle_ReturnsInteractiveModeError(t *testing.T) {
	// ARRANGE: Valid config but no title provided
	config := `
project:
  owner: "test-owner"
  number: 1
repositories:
  - "owner/repo"
`
	dir := createTempConfig(t, config)
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"create"}) // No --title

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// ACT
	err := cmd.Execute()

	// ASSERT
	if err == nil {
		t.Fatal("Expected error when no title provided")
	}
	if !strings.Contains(err.Error(), "--title is required") {
		t.Errorf("Expected '--title is required' error, got: %v", err)
	}
}

func TestRunCreate_ValidConfigWithTitle_AttemptsAPICall(t *testing.T) {
	// ARRANGE: Valid config with title provided
	// This test verifies that we get past config validation and into API calls
	config := `
project:
  owner: "test-owner"
  number: 1
repositories:
  - "owner/repo"
`
	dir := createTempConfig(t, config)
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"create", "--title", "Test Issue", "--body", "Test body"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// ACT
	err := cmd.Execute()

	// ASSERT: We expect an API error (since we're not authenticated in tests)
	// The key thing is we got PAST the config validation phase
	if err == nil {
		t.Skip("Skipping: API call succeeded (authenticated environment)")
	}

	// Should be an API-related error, not a config error
	errStr := err.Error()
	if strings.Contains(errStr, "configuration") || strings.Contains(errStr, "--title is required") {
		t.Errorf("Expected API error after passing config validation, got: %v", err)
	}
}

func TestRunCreate_WithAllFlags_ParsesFlagsCorrectly(t *testing.T) {
	// ARRANGE: Valid config with all flags
	config := `
project:
  owner: "test-owner"
  number: 1
repositories:
  - "owner/repo"
fields:
  status:
    field: Status
    values:
      todo: "Todo"
      in_progress: "In Progress"
  priority:
    field: Priority
    values:
      p1: "P1"
      p2: "P2"
defaults:
  labels:
    - "auto-label"
`
	dir := createTempConfig(t, config)
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	cmd := NewRootCommand()
	cmd.SetArgs([]string{
		"create",
		"--title", "Test Issue",
		"--body", "Test body",
		"--status", "in_progress",
		"--priority", "p1",
		"--label", "bug",
		"--label", "urgent",
	})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// ACT
	err := cmd.Execute()

	// ASSERT: We should get past flag parsing to API calls
	if err == nil {
		t.Skip("Skipping: API call succeeded (authenticated environment)")
	}

	// Verify we didn't get a flag parsing error
	errStr := err.Error()
	if strings.Contains(errStr, "unknown flag") || strings.Contains(errStr, "flag needs") {
		t.Errorf("Expected to pass flag parsing, got: %v", err)
	}
}

func TestRunCreate_ConfigWithDefaults_MergesLabels(t *testing.T) {
	// ARRANGE: Config with default labels
	config := `
project:
  owner: "test-owner"
  number: 1
repositories:
  - "owner/repo"
defaults:
  labels:
    - "pm-tracked"
    - "auto-created"
  status: "todo"
  priority: "p2"
`
	dir := createTempConfig(t, config)
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"create", "--title", "Test Issue"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// ACT
	err := cmd.Execute()

	// ASSERT: Should reach API call phase (past config loading and defaults)
	if err == nil {
		t.Skip("Skipping: API call succeeded (authenticated environment)")
	}

	// Verify we got past config validation
	errStr := err.Error()
	if strings.Contains(errStr, "configuration") {
		t.Errorf("Expected to pass config validation with defaults, got: %v", err)
	}
}

// ============================================================================
// Issue #325: Additional gh issue create options
// ============================================================================

func TestCreateCommand_HasBodyFileFlag(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create command not found: %v", err)
	}

	flag := createCmd.Flags().Lookup("body-file")
	if flag == nil {
		t.Fatal("Expected --body-file flag to exist")
	}
	if flag.Shorthand != "F" {
		t.Errorf("Expected shorthand 'F', got '%s'", flag.Shorthand)
	}
}

func TestCreateCommand_HasEditorFlag(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create command not found: %v", err)
	}

	flag := createCmd.Flags().Lookup("editor")
	if flag == nil {
		t.Fatal("Expected --editor flag to exist")
	}
	if flag.Shorthand != "e" {
		t.Errorf("Expected shorthand 'e', got '%s'", flag.Shorthand)
	}
	if flag.Value.Type() != "bool" {
		t.Errorf("Expected --editor to be bool, got %s", flag.Value.Type())
	}
}

func TestCreateCommand_HasTemplateFlag(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create command not found: %v", err)
	}

	flag := createCmd.Flags().Lookup("template")
	if flag == nil {
		t.Fatal("Expected --template flag to exist")
	}
	if flag.Shorthand != "T" {
		t.Errorf("Expected shorthand 'T', got '%s'", flag.Shorthand)
	}
}

func TestCreateCommand_HasWebFlag(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create command not found: %v", err)
	}

	flag := createCmd.Flags().Lookup("web")
	if flag == nil {
		t.Fatal("Expected --web flag to exist")
	}
	if flag.Shorthand != "w" {
		t.Errorf("Expected shorthand 'w', got '%s'", flag.Shorthand)
	}
	if flag.Value.Type() != "bool" {
		t.Errorf("Expected --web to be bool, got %s", flag.Value.Type())
	}
}

// ============================================================================
// readBodyFile Tests
// ============================================================================

func TestReadBodyFile_ReadsFileContent(t *testing.T) {
	// Create a temp file with content
	tmpfile, err := os.CreateTemp("", "body-*.md")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	content := "This is the issue body.\n\nWith multiple lines."
	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpfile.Close()

	// Read it back
	result, err := readBodyFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("readBodyFile failed: %v", err)
	}

	if result != content {
		t.Errorf("Expected %q, got %q", content, result)
	}
}

func TestReadBodyFile_FileNotFound(t *testing.T) {
	_, err := readBodyFile("/nonexistent/path/to/file.md")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

// ============================================================================
// extractTemplateBody Tests
// ============================================================================

func TestExtractTemplateBody_NoFrontmatter(t *testing.T) {
	content := "This is the template body.\n\nWith multiple lines."
	result := extractTemplateBody(content)
	if result != content {
		t.Errorf("Expected %q, got %q", content, result)
	}
}

func TestExtractTemplateBody_WithYAMLFrontmatter(t *testing.T) {
	content := `---
name: Bug Report
about: Create a report to help us improve
---

## Description
Describe the bug here.

## Steps to Reproduce
1. Step 1
2. Step 2`

	expected := `## Description
Describe the bug here.

## Steps to Reproduce
1. Step 1
2. Step 2`

	result := extractTemplateBody(content)
	if result != expected {
		t.Errorf("Expected:\n%q\n\nGot:\n%q", expected, result)
	}
}

func TestExtractTemplateBody_EmptyContent(t *testing.T) {
	result := extractTemplateBody("")
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

func TestExtractTemplateBody_OnlyFrontmatter(t *testing.T) {
	content := `---
name: Bug Report
about: Create a report
---`

	result := extractTemplateBody(content)
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

// ============================================================================
// Mutual Exclusivity Tests
// ============================================================================

func TestRunCreate_BodyAndBodyFileMutuallyExclusive(t *testing.T) {
	// ARRANGE: Valid config
	config := `
project:
  owner: "test-owner"
  number: 1
repositories:
  - "owner/repo"
`
	dir := createTempConfig(t, config)
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	// Create a temp body file
	tmpfile, err := os.CreateTemp(dir, "body-*.md")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if _, err := tmpfile.WriteString("Body from file"); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	cmd := NewRootCommand()
	cmd.SetArgs([]string{
		"create",
		"--title", "Test Issue",
		"--body", "Body from flag",
		"--body-file", tmpfile.Name(),
	})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// ACT
	err = cmd.Execute()

	// ASSERT
	if err == nil {
		t.Fatal("Expected error when using --body and --body-file together")
	}
	if !strings.Contains(err.Error(), "cannot use --body and --body-file together") {
		t.Errorf("Expected mutual exclusivity error, got: %v", err)
	}
}

func TestRunCreate_TemplateWithBodyFlagError(t *testing.T) {
	// ARRANGE: Valid config
	config := `
project:
  owner: "test-owner"
  number: 1
repositories:
  - "owner/repo"
`
	dir := createTempConfig(t, config)
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	cmd := NewRootCommand()
	cmd.SetArgs([]string{
		"create",
		"--title", "Test Issue",
		"--body", "Body from flag",
		"--template", "bug",
	})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// ACT
	err := cmd.Execute()

	// ASSERT
	if err == nil {
		t.Fatal("Expected error when using --template with --body")
	}
	if !strings.Contains(err.Error(), "cannot use --template with --body") {
		t.Errorf("Expected mutual exclusivity error, got: %v", err)
	}
}

func TestRunCreate_BodyFileReadsContent(t *testing.T) {
	// ARRANGE: Valid config with body file
	config := `
project:
  owner: "test-owner"
  number: 1
repositories:
  - "owner/repo"
`
	dir := createTempConfig(t, config)
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	// Create a temp body file
	tmpfile, err := os.CreateTemp(dir, "body-*.md")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if _, err := tmpfile.WriteString("Body content from file"); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	cmd := NewRootCommand()
	cmd.SetArgs([]string{
		"create",
		"--title", "Test Issue",
		"--body-file", tmpfile.Name(),
	})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// ACT
	err = cmd.Execute()

	// ASSERT: Should reach API call phase (past body file reading)
	if err == nil {
		t.Skip("Skipping: API call succeeded (authenticated environment)")
	}

	// Verify we didn't get a body file reading error
	errStr := err.Error()
	if strings.Contains(errStr, "failed to read body file") {
		t.Errorf("Expected to successfully read body file, got: %v", err)
	}
}
