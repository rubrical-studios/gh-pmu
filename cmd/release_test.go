package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rubrical-studios/gh-pmu/internal/api"
	"github.com/rubrical-studios/gh-pmu/internal/config"
	"github.com/spf13/cobra"
)

// setupReleaseTestDir creates a temp directory with a .gh-pmu.yml config file
// and changes to that directory. Returns cleanup function to restore original dir.
func setupReleaseTestDir(t *testing.T, cfg *config.Config) func() {
	t.Helper()

	// Save original directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Create temp directory
	tempDir := t.TempDir()

	// Save config to temp directory
	configPath := filepath.Join(tempDir, ".gh-pmu.yml")
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to chdir to temp dir: %v", err)
	}

	// Return cleanup function
	return func() {
		_ = os.Chdir(originalDir)
	}
}

// mockReleaseClient implements releaseClient for testing
type mockReleaseClient struct {
	// Return values
	createdIssue           *api.Issue
	openIssues             []api.Issue
	closedIssues           []api.Issue
	project                *api.Project
	addedItemID            string
	issueByNumber          *api.Issue
	projectItemID          string
	projectItemIDs         map[string]string // issueID -> itemID mapping for per-issue returns
	projectItemFieldValue  string
	projectItemFieldValues map[string]string // itemID -> fieldValue mapping for per-issue status
	releaseIssues          []api.Issue

	// Captured calls for verification
	createIssueCalls     []createIssueCall
	addToProjectCalls    []addToProjectCall
	setFieldCalls        []setFieldCall
	updateIssueBodyCalls []updateIssueBodyCall
	writeFileCalls       []writeFileCall
	gitAddCalls          []gitAddCall
	closeIssueCalls      []closeIssueCall
	gitTagCalls          []gitTagCall

	// Error injection
	createIssueErr         error
	getOpenIssuesErr       error
	getClosedIssuesErr     error
	addToProjectErr        error
	setFieldErr            error
	getProjectErr          error
	getIssueErr            error
	getProjectItemErr      error
	getProjectItemFieldErr error
	getReleaseIssuesErr    error
	reopenIssueErr         error
}

// Helper types for call tracking
type gitAddCall struct {
	paths []string
}

type gitTagCall struct {
	tag     string
	message string
}

func (m *mockReleaseClient) CreateIssue(owner, repo, title, body string, labels []string) (*api.Issue, error) {
	m.createIssueCalls = append(m.createIssueCalls, createIssueCall{
		owner:  owner,
		repo:   repo,
		title:  title,
		body:   body,
		labels: labels,
	})
	if m.createIssueErr != nil {
		return nil, m.createIssueErr
	}
	return m.createdIssue, nil
}

func (m *mockReleaseClient) GetOpenIssuesByLabel(owner, repo, label string) ([]api.Issue, error) {
	if m.getOpenIssuesErr != nil {
		return nil, m.getOpenIssuesErr
	}
	return m.openIssues, nil
}

func (m *mockReleaseClient) GetClosedIssuesByLabel(owner, repo, label string) ([]api.Issue, error) {
	if m.getClosedIssuesErr != nil {
		return nil, m.getClosedIssuesErr
	}
	return m.closedIssues, nil
}

func (m *mockReleaseClient) AddIssueToProject(projectID, issueID string) (string, error) {
	m.addToProjectCalls = append(m.addToProjectCalls, addToProjectCall{
		projectID: projectID,
		issueID:   issueID,
	})
	if m.addToProjectErr != nil {
		return "", m.addToProjectErr
	}
	return m.addedItemID, nil
}

func (m *mockReleaseClient) SetProjectItemField(projectID, itemID, fieldID, value string) error {
	m.setFieldCalls = append(m.setFieldCalls, setFieldCall{
		projectID: projectID,
		itemID:    itemID,
		fieldID:   fieldID,
		value:     value,
	})
	return m.setFieldErr
}

func (m *mockReleaseClient) GetProject(owner string, number int) (*api.Project, error) {
	if m.getProjectErr != nil {
		return nil, m.getProjectErr
	}
	return m.project, nil
}

func (m *mockReleaseClient) GetIssueByNumber(owner, repo string, number int) (*api.Issue, error) {
	if m.getIssueErr != nil {
		return nil, m.getIssueErr
	}
	return m.issueByNumber, nil
}

func (m *mockReleaseClient) GetProjectItemID(projectID, issueID string) (string, error) {
	if m.getProjectItemErr != nil {
		return "", m.getProjectItemErr
	}
	// Check per-issue mapping first
	if m.projectItemIDs != nil {
		if itemID, ok := m.projectItemIDs[issueID]; ok {
			return itemID, nil
		}
		// If map is set but issueID not found, return error (not found)
		return "", fmt.Errorf("project item not found for issue %s", issueID)
	}
	return m.projectItemID, nil
}

func (m *mockReleaseClient) GetProjectItemFieldValue(projectID, itemID, fieldID string) (string, error) {
	if m.getProjectItemFieldErr != nil {
		return "", m.getProjectItemFieldErr
	}
	// Check per-item mapping first
	if m.projectItemFieldValues != nil {
		if value, ok := m.projectItemFieldValues[itemID]; ok {
			return value, nil
		}
	}
	return m.projectItemFieldValue, nil
}

func (m *mockReleaseClient) GetIssuesByRelease(owner, repo, releaseVersion string) ([]api.Issue, error) {
	if m.getReleaseIssuesErr != nil {
		return nil, m.getReleaseIssuesErr
	}
	return m.releaseIssues, nil
}

func (m *mockReleaseClient) UpdateIssueBody(issueID, body string) error {
	m.updateIssueBodyCalls = append(m.updateIssueBodyCalls, updateIssueBodyCall{
		issueID: issueID,
		body:    body,
	})
	return nil
}

func (m *mockReleaseClient) WriteFile(path, content string) error {
	m.writeFileCalls = append(m.writeFileCalls, writeFileCall{
		path:    path,
		content: content,
	})
	return nil
}

func (m *mockReleaseClient) MkdirAll(path string) error {
	return nil
}

func (m *mockReleaseClient) GitAdd(paths ...string) error {
	m.gitAddCalls = append(m.gitAddCalls, gitAddCall{
		paths: paths,
	})
	return nil
}

func (m *mockReleaseClient) CloseIssue(issueID string) error {
	m.closeIssueCalls = append(m.closeIssueCalls, closeIssueCall{
		issueID: issueID,
	})
	return nil
}

func (m *mockReleaseClient) ReopenIssue(issueID string) error {
	if m.reopenIssueErr != nil {
		return m.reopenIssueErr
	}
	return nil
}

func (m *mockReleaseClient) GitTag(tag, message string) error {
	m.gitTagCalls = append(m.gitTagCalls, gitTagCall{
		tag:     tag,
		message: message,
	})
	return nil
}

func (m *mockReleaseClient) GitCheckoutNewBranch(branch string) error {
	return nil
}

// testReleaseConfig returns a test configuration for release tests
func testReleaseConfig() *config.Config {
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
					"in_progress": "In progress",
				},
			},
		},
	}
}

// setupMockForRelease creates a mock configured for release start tests
func setupMockForRelease() *mockReleaseClient {
	return &mockReleaseClient{
		openIssues: []api.Issue{}, // No active releases
		createdIssue: &api.Issue{
			ID:     "ISSUE_123",
			Number: 100,
			Title:  "Release: v1.2.0",
			URL:    "https://github.com/testowner/testrepo/issues/100",
		},
		project: &api.Project{
			ID:     "PROJECT_1",
			Number: 1,
			Title:  "Test Project",
		},
		addedItemID: "ITEM_456",
	}
}

// Helper to create a test command with captured output
func newTestReleaseCmd() (*cobra.Command, *bytes.Buffer) {
	cmd := &cobra.Command{Use: "release"}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	return cmd, buf
}

// =============================================================================
// REQ-017: Start Release
// =============================================================================

// AC-017-1: Given `release start --branch release/v1.2.0`, Then tracker issue created: "Release: release/v1.2.0"
func TestRunReleaseStartWithDeps_CreatesTrackerIssue(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	cfg := testReleaseConfig()
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestReleaseCmd()
	opts := &releaseStartOptions{
		branch: "release/v1.2.0",
	}

	expectedTitle := "Release: release/v1.2.0"

	// ACT
	err := runReleaseStartWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify CreateIssue was called
	if len(mock.createIssueCalls) != 1 {
		t.Fatalf("Expected 1 CreateIssue call, got %d", len(mock.createIssueCalls))
	}

	call := mock.createIssueCalls[0]

	// Verify title matches expected pattern
	if call.title != expectedTitle {
		t.Errorf("Expected title '%s', got '%s'", expectedTitle, call.title)
	}
}

// AC-017-3: Given tracker issue created, Then has `release` label
func TestRunReleaseStartWithDeps_HasReleaseLabel(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	cfg := testReleaseConfig()
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestReleaseCmd()
	opts := &releaseStartOptions{
		branch: "release/v1.2.0",
	}

	// ACT
	err := runReleaseStartWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mock.createIssueCalls) != 1 {
		t.Fatalf("Expected 1 CreateIssue call, got %d", len(mock.createIssueCalls))
	}

	call := mock.createIssueCalls[0]
	hasLabel := false
	for _, label := range call.labels {
		if label == "release" {
			hasLabel = true
			break
		}
	}
	if !hasLabel {
		t.Errorf("Expected 'release' label, got labels: %v", call.labels)
	}
}

// AC-017-4: Given active release exists, When running `release start`, Then error: "Active release exists"
func TestRunReleaseStartWithDeps_ActiveReleaseExists_ReturnsError(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	// Simulate an existing active release
	mock.openIssues = []api.Issue{
		{
			ID:     "EXISTING_RELEASE",
			Number: 50,
			Title:  "Release: release/v1.1.0",
			State:  "OPEN",
		},
	}
	cfg := testReleaseConfig()
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestReleaseCmd()
	opts := &releaseStartOptions{
		branch: "release/v1.2.0",
	}

	// ACT
	err := runReleaseStartWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err == nil {
		t.Fatalf("Expected error for active release exists, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(strings.ToLower(errMsg), "active release") {
		t.Errorf("Expected error to mention 'active release', got: %s", errMsg)
	}
}

// Test that release is added to project and status set to In Progress
func TestRunReleaseStartWithDeps_AddsToProjectAndSetsStatus(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	cfg := testReleaseConfig()
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestReleaseCmd()
	opts := &releaseStartOptions{
		branch: "release/v1.2.0",
	}

	// ACT
	err := runReleaseStartWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify issue was added to project
	if len(mock.addToProjectCalls) != 1 {
		t.Fatalf("Expected 1 AddIssueToProject call, got %d", len(mock.addToProjectCalls))
	}

	// Verify status field was set to In Progress
	statusSet := false
	for _, call := range mock.setFieldCalls {
		if call.value == "In progress" {
			statusSet = true
			break
		}
	}
	if !statusSet {
		t.Errorf("Expected status to be set to 'In progress', got calls: %+v", mock.setFieldCalls)
	}
}

// =============================================================================
// REQ-018: Version Validation
// =============================================================================

// AC-018-1: Given `release start --version 1.2.0`, Then accepted (valid semver)
func TestValidateVersion_ValidSemver_Accepted(t *testing.T) {
	validVersions := []string{
		"1.2.0",
		"0.1.0",
		"10.20.30",
		"1.0.0",
	}

	for _, version := range validVersions {
		err := validateVersion(version)
		if err != nil {
			t.Errorf("Expected version '%s' to be valid, got error: %v", version, err)
		}
	}
}

// AC-018-2: Given `release start --version 1.2`, Then error: "Invalid version format. Use semver: X.Y.Z"
func TestValidateVersion_InvalidFormat_ReturnsError(t *testing.T) {
	invalidVersions := []string{
		"1.2",
		"1",
		"1.2.3.4",
		"abc",
		"1.2.x",
		"",
	}

	for _, version := range invalidVersions {
		err := validateVersion(version)
		if err == nil {
			t.Errorf("Expected version '%s' to be invalid, got no error", version)
			continue
		}

		errMsg := err.Error()
		if !strings.Contains(errMsg, "Invalid version format") {
			t.Errorf("Expected error message to contain 'Invalid version format', got: %s", errMsg)
		}
	}
}

// AC-018-3: Given `release start --version v1.2.0`, Then accepted (v prefix allowed)
func TestValidateVersion_VPrefixAllowed(t *testing.T) {
	versionsWithPrefix := []string{
		"v1.2.0",
		"v0.1.0",
		"v10.20.30",
	}

	for _, version := range versionsWithPrefix {
		err := validateVersion(version)
		if err != nil {
			t.Errorf("Expected version '%s' (with v prefix) to be valid, got error: %v", version, err)
		}
	}
}

// Test that branch names are used literally
func TestRunReleaseStartWithDeps_BranchNameUsedLiterally(t *testing.T) {
	testCases := []struct {
		branch        string
		expectedTitle string
	}{
		{"release/v1.2.0", "Release: release/v1.2.0"},
		{"patch/v1.1.1", "Release: patch/v1.1.1"},
		{"hotfix-auth-bypass", "Release: hotfix-auth-bypass"},
	}

	for _, tc := range testCases {
		t.Run(tc.branch, func(t *testing.T) {
			mock := setupMockForRelease()
			cfg := testReleaseConfig()
			cleanup := setupReleaseTestDir(t, cfg)
			defer cleanup()

			cmd, _ := newTestReleaseCmd()
			opts := &releaseStartOptions{
				branch: tc.branch,
			}

			err := runReleaseStartWithDeps(cmd, opts, cfg, mock)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if len(mock.createIssueCalls) != 1 {
				t.Fatalf("Expected 1 CreateIssue call, got %d", len(mock.createIssueCalls))
			}

			if mock.createIssueCalls[0].title != tc.expectedTitle {
				t.Errorf("Expected title '%s', got '%s'", tc.expectedTitle, mock.createIssueCalls[0].title)
			}
		})
	}
}

// =============================================================================
// REQ-019: Add Issue to Release
// =============================================================================

// AC-019-1: Given active release v1.2.0, When running `release add 42`,
// Then Release field on #42 set to "v1.2.0"
func TestRunReleaseAddWithDeps_SetsReleaseField(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	// Active release exists
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Release: v1.2.0",
			State:  "OPEN",
		},
	}
	// The issue to add
	mock.issueByNumber = &api.Issue{
		ID:     "ISSUE_42",
		Number: 42,
		Title:  "Fix login bug",
	}
	// Project item for issue 42
	mock.projectItemID = "ITEM_42"

	cfg := testReleaseConfig()
	// Add release field to config
	cfg.Fields["release"] = config.Field{
		Field: "Release",
	}
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestReleaseCmd()
	opts := &releaseAddOptions{
		issueNumber: 42,
	}

	// ACT
	err := runReleaseAddWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify SetProjectItemField was called with correct values
	if len(mock.setFieldCalls) != 1 {
		t.Fatalf("Expected 1 SetProjectItemField call, got %d", len(mock.setFieldCalls))
	}

	call := mock.setFieldCalls[0]
	if call.value != "v1.2.0" {
		t.Errorf("Expected field value 'v1.2.0', got '%s'", call.value)
	}
	if call.fieldID != "Release" {
		t.Errorf("Expected fieldID 'Release', got '%s'", call.fieldID)
	}
}

// AC-019-2: Given issue added, Then output: "Added #42 to release v1.2.0"
func TestRunReleaseAddWithDeps_OutputsConfirmation(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Release: v1.2.0",
			State:  "OPEN",
		},
	}
	mock.issueByNumber = &api.Issue{
		ID:     "ISSUE_42",
		Number: 42,
		Title:  "Fix login bug",
	}
	mock.projectItemID = "ITEM_42"

	cfg := testReleaseConfig()
	cfg.Fields["release"] = config.Field{
		Field: "Release",
	}
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, buf := newTestReleaseCmd()
	opts := &releaseAddOptions{
		issueNumber: 42,
	}

	// ACT
	err := runReleaseAddWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	expectedOutput := "Added #42 to release v1.2.0"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output to contain '%s', got '%s'", expectedOutput, output)
	}
}

// Test error when no active release exists
func TestRunReleaseAddWithDeps_NoActiveRelease_ReturnsError(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{} // No active release

	cfg := testReleaseConfig()
	cfg.Fields["release"] = config.Field{
		Field: "Release",
	}

	cmd, _ := newTestReleaseCmd()
	opts := &releaseAddOptions{
		issueNumber: 42,
	}

	// ACT
	err := runReleaseAddWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err == nil {
		t.Fatalf("Expected error for no active release, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "no active release") {
		t.Errorf("Expected error to mention 'no active release', got: %s", errMsg)
	}
}

// =============================================================================
// REQ-039: Remove Issue from Release
// =============================================================================

// AC-039-1: Given issue #42 assigned to release, When running `release remove 42`,
// Then Release Text field cleared (set to empty)
func TestRunReleaseRemoveWithDeps_ClearsReleaseField(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Release: v1.2.0",
			State:  "OPEN",
		},
	}
	mock.issueByNumber = &api.Issue{
		ID:     "ISSUE_42",
		Number: 42,
		Title:  "Fix login bug",
	}
	mock.projectItemID = "ITEM_42"
	mock.projectItemFieldValue = "v1.2.0" // Currently assigned

	cfg := testReleaseConfig()
	cfg.Fields["release"] = config.Field{
		Field: "Release",
	}

	cmd, _ := newTestReleaseCmd()
	opts := &releaseRemoveOptions{
		issueNumber: 42,
	}

	// ACT
	err := runReleaseRemoveWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mock.setFieldCalls) != 1 {
		t.Fatalf("Expected 1 SetProjectItemField call, got %d", len(mock.setFieldCalls))
	}

	call := mock.setFieldCalls[0]
	if call.value != "" {
		t.Errorf("Expected field value to be empty (cleared), got '%s'", call.value)
	}
}

// AC-039-2: Given field cleared, Then output confirms "Removed #42 from release vX.Y.Z"
func TestRunReleaseRemoveWithDeps_OutputsConfirmation(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Release: v1.2.0",
			State:  "OPEN",
		},
	}
	mock.issueByNumber = &api.Issue{
		ID:     "ISSUE_42",
		Number: 42,
		Title:  "Fix login bug",
	}
	mock.projectItemID = "ITEM_42"
	mock.projectItemFieldValue = "v1.2.0"

	cfg := testReleaseConfig()
	cfg.Fields["release"] = config.Field{
		Field: "Release",
	}

	cmd, buf := newTestReleaseCmd()
	opts := &releaseRemoveOptions{
		issueNumber: 42,
	}

	// ACT
	err := runReleaseRemoveWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	expectedOutput := "Removed #42 from release v1.2.0"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output to contain '%s', got '%s'", expectedOutput, output)
	}
}

// AC-039-3: Given issue not in any release, When running `release remove 42`,
// Then warning: "Issue #42 is not assigned to a release"
func TestRunReleaseRemoveWithDeps_WarnsIfNotAssigned(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Release: v1.2.0",
			State:  "OPEN",
		},
	}
	mock.issueByNumber = &api.Issue{
		ID:     "ISSUE_42",
		Number: 42,
		Title:  "Fix login bug",
	}
	mock.projectItemID = "ITEM_42"
	mock.projectItemFieldValue = "" // Not assigned

	cfg := testReleaseConfig()
	cfg.Fields["release"] = config.Field{
		Field: "Release",
	}

	cmd, buf := newTestReleaseCmd()
	opts := &releaseRemoveOptions{
		issueNumber: 42,
	}

	// ACT
	err := runReleaseRemoveWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error (warning only), got: %v", err)
	}

	output := buf.String()
	expectedWarning := "Issue #42 is not assigned to a release"
	if !strings.Contains(output, expectedWarning) {
		t.Errorf("Expected output to contain warning '%s', got '%s'", expectedWarning, output)
	}

	if len(mock.setFieldCalls) != 0 {
		t.Errorf("Expected 0 SetProjectItemField calls (nothing to clear), got %d", len(mock.setFieldCalls))
	}
}

// =============================================================================
// REQ-036: View Current Release
// =============================================================================

// AC-036-1: Given active release, When running `release current`, Then displays details
func TestRunReleaseCurrentWithDeps_DisplaysActiveDetails(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Release: v1.2.0 (Phoenix)",
			State:  "OPEN",
		},
	}
	mock.releaseIssues = []api.Issue{
		{ID: "ISSUE_1", Number: 41, Title: "Fix bug A"},
		{ID: "ISSUE_2", Number: 42, Title: "Fix bug B"},
		{ID: "ISSUE_3", Number: 43, Title: "Add feature C"},
	}

	cfg := testReleaseConfig()
	cmd, buf := newTestReleaseCmd()
	opts := &releaseCurrentOptions{}

	// ACT
	err := runReleaseCurrentWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "v1.2.0") {
		t.Errorf("Expected output to contain version 'v1.2.0', got '%s'", output)
	}
	if !strings.Contains(output, "#100") {
		t.Errorf("Expected output to contain tracker issue '#100', got '%s'", output)
	}
	if !strings.Contains(output, "3") {
		t.Errorf("Expected output to contain issue count '3', got '%s'", output)
	}
}

// AC-036-2: Given no active release, Then message: "No active release"
func TestRunReleaseCurrentWithDeps_NoActiveRelease(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{} // No active release

	cfg := testReleaseConfig()
	cmd, buf := newTestReleaseCmd()
	opts := &releaseCurrentOptions{}

	// ACT
	err := runReleaseCurrentWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	expectedMessage := "No active release"
	if !strings.Contains(output, expectedMessage) {
		t.Errorf("Expected output to contain '%s', got '%s'", expectedMessage, output)
	}
}

// AC-036-3: Given `--refresh` flag, Then tracker issue body updated
func TestRunReleaseCurrentWithDeps_RefreshUpdatesTrackerBody(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Release: v1.2.0",
			State:  "OPEN",
		},
	}
	mock.releaseIssues = []api.Issue{
		{ID: "ISSUE_1", Number: 41, Title: "Fix bug A"},
		{ID: "ISSUE_2", Number: 42, Title: "Fix bug B"},
	}

	cfg := testReleaseConfig()
	cmd, _ := newTestReleaseCmd()
	opts := &releaseCurrentOptions{
		refresh: true,
	}

	// ACT
	err := runReleaseCurrentWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mock.updateIssueBodyCalls) != 1 {
		t.Fatalf("Expected 1 UpdateIssueBody call for refresh, got %d", len(mock.updateIssueBodyCalls))
	}

	call := mock.updateIssueBodyCalls[0]
	if call.issueID != "TRACKER_123" {
		t.Errorf("Expected UpdateIssueBody on TRACKER_123, got %s", call.issueID)
	}
	if !strings.Contains(call.body, "#41") || !strings.Contains(call.body, "#42") {
		t.Errorf("Expected body to contain issue references, got '%s'", call.body)
	}
}

// =============================================================================
// REQ-020: Release Artifacts
// =============================================================================

// AC-020-1: Given `release close`, Then `Releases/v1.2.0/release-notes.md` generated
func TestRunReleaseCloseWithDeps_GeneratesReleaseNotes(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Release: v1.2.0",
			State:  "OPEN",
		},
	}
	mock.releaseIssues = []api.Issue{
		{ID: "ISSUE_1", Number: 41, Title: "Add new feature", State: "CLOSED", Labels: []api.Label{{Name: "enhancement"}}},
		{ID: "ISSUE_2", Number: 42, Title: "Fix bug", State: "CLOSED", Labels: []api.Label{{Name: "bug"}}},
	}

	cfg := testReleaseConfig()
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestReleaseCmd()
	opts := &releaseCloseOptions{releaseName: "v1.2.0", yes: true}

	// ACT
	err := runReleaseCloseWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify WriteFile was called for release-notes.md
	found := false
	for _, call := range mock.writeFileCalls {
		if strings.Contains(call.path, "Releases/v1.2.0/release-notes.md") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected WriteFile call for Releases/v1.2.0/release-notes.md, got calls: %+v", mock.writeFileCalls)
	}
}

// AC-020-2: Given release-notes.md, Then contains: date, codename (if set), tracker issue, issues grouped by label
func TestRunReleaseCloseWithDeps_ReleaseNotesContainsDetails(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Release: v1.2.0 (Phoenix)",
			State:  "OPEN",
		},
	}
	mock.releaseIssues = []api.Issue{
		{ID: "ISSUE_1", Number: 41, Title: "Add new feature", State: "CLOSED", Labels: []api.Label{{Name: "enhancement"}}},
		{ID: "ISSUE_2", Number: 42, Title: "Fix critical bug", State: "CLOSED", Labels: []api.Label{{Name: "bug"}}},
	}

	cfg := testReleaseConfig()
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestReleaseCmd()
	opts := &releaseCloseOptions{releaseName: "v1.2.0", yes: true}

	// ACT
	err := runReleaseCloseWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Find the release-notes.md content
	var releaseNotesContent string
	for _, call := range mock.writeFileCalls {
		if strings.Contains(call.path, "release-notes.md") {
			releaseNotesContent = call.content
			break
		}
	}

	if releaseNotesContent == "" {
		t.Fatalf("release-notes.md was not written")
	}

	// Verify content contains required elements
	if !strings.Contains(releaseNotesContent, "v1.2.0") {
		t.Errorf("Expected release-notes.md to contain version 'v1.2.0'")
	}
	if !strings.Contains(releaseNotesContent, "Phoenix") {
		t.Errorf("Expected release-notes.md to contain codename 'Phoenix'")
	}
	if !strings.Contains(releaseNotesContent, "#100") {
		t.Errorf("Expected release-notes.md to contain tracker issue '#100'")
	}
	// Check issue grouping
	if !strings.Contains(releaseNotesContent, "#41") || !strings.Contains(releaseNotesContent, "#42") {
		t.Errorf("Expected release-notes.md to contain issue references")
	}
}

// AC-020-3: Given `release close`, Then `Releases/v1.2.0/changelog.md` generated
func TestRunReleaseCloseWithDeps_GeneratesChangelog(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Release: v1.2.0",
			State:  "OPEN",
		},
	}
	mock.releaseIssues = []api.Issue{
		{ID: "ISSUE_1", Number: 41, Title: "Add feature", State: "CLOSED"},
	}

	cfg := testReleaseConfig()
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestReleaseCmd()
	opts := &releaseCloseOptions{releaseName: "v1.2.0", yes: true}

	// ACT
	err := runReleaseCloseWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify WriteFile was called for changelog.md
	found := false
	for _, call := range mock.writeFileCalls {
		if strings.Contains(call.path, "Releases/v1.2.0/changelog.md") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected WriteFile call for Releases/v1.2.0/changelog.md, got calls: %+v", mock.writeFileCalls)
	}
}

// AC-020-4: Given artifacts, Then staged to git
func TestRunReleaseCloseWithDeps_StagesArtifacts(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Release: v1.2.0",
			State:  "OPEN",
		},
	}
	mock.releaseIssues = []api.Issue{
		{ID: "ISSUE_1", Number: 41, Title: "Add feature", State: "CLOSED"},
	}

	cfg := testReleaseConfig()
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestReleaseCmd()
	opts := &releaseCloseOptions{releaseName: "v1.2.0", yes: true}

	// ACT
	err := runReleaseCloseWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify GitAdd was called
	if len(mock.gitAddCalls) == 0 {
		t.Errorf("Expected GitAdd to be called for staging artifacts")
	}

	// Verify both files were staged
	var stagedFiles []string
	for _, call := range mock.gitAddCalls {
		stagedFiles = append(stagedFiles, call.paths...)
	}

	hasReleaseNotes := false
	hasChangelog := false
	for _, path := range stagedFiles {
		if strings.Contains(path, "release-notes.md") {
			hasReleaseNotes = true
		}
		if strings.Contains(path, "changelog.md") {
			hasChangelog = true
		}
	}

	if !hasReleaseNotes {
		t.Errorf("Expected release-notes.md to be staged")
	}
	if !hasChangelog {
		t.Errorf("Expected changelog.md to be staged")
	}
}

// Test that release close closes the tracker issue
func TestRunReleaseCloseWithDeps_ClosesTrackerIssue(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Release: v1.2.0",
			State:  "OPEN",
		},
	}
	mock.releaseIssues = []api.Issue{}

	cfg := testReleaseConfig()
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestReleaseCmd()
	opts := &releaseCloseOptions{releaseName: "v1.2.0", yes: true}

	// ACT
	err := runReleaseCloseWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify CloseIssue was called
	if len(mock.closeIssueCalls) != 1 {
		t.Fatalf("Expected 1 CloseIssue call, got %d", len(mock.closeIssueCalls))
	}

	if mock.closeIssueCalls[0].issueID != "TRACKER_123" {
		t.Errorf("Expected to close TRACKER_123, got %s", mock.closeIssueCalls[0].issueID)
	}
}

// Test error when no active release
func TestRunReleaseCloseWithDeps_NoActiveRelease_ReturnsError(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{} // No active release

	cfg := testReleaseConfig()
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestReleaseCmd()
	opts := &releaseCloseOptions{releaseName: "v1.2.0", yes: true}

	// ACT
	err := runReleaseCloseWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err == nil {
		t.Fatalf("Expected error for no active release, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "release not found") {
		t.Errorf("Expected error to mention 'release not found', got: %s", errMsg)
	}
}

// =============================================================================
// REQ-021: Release Git Tag
// =============================================================================

// AC-021-1: Given `release close --tag`, Then `git tag -a v1.2.0 -m "Release v1.2.0"` executed
func TestRunReleaseCloseWithDeps_WithTag_CreatesGitTag(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Release: v1.2.0",
			State:  "OPEN",
		},
	}
	mock.releaseIssues = []api.Issue{}

	cfg := testReleaseConfig()
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestReleaseCmd()
	opts := &releaseCloseOptions{
		releaseName: "v1.2.0",
		yes:         true,
		tag:         true,
	}

	// ACT
	err := runReleaseCloseWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify GitTag was called
	if len(mock.gitTagCalls) != 1 {
		t.Fatalf("Expected 1 GitTag call, got %d", len(mock.gitTagCalls))
	}

	call := mock.gitTagCalls[0]
	if call.tag != "v1.2.0" {
		t.Errorf("Expected tag 'v1.2.0', got '%s'", call.tag)
	}
	if !strings.Contains(call.message, "Release v1.2.0") {
		t.Errorf("Expected message to contain 'Release v1.2.0', got '%s'", call.message)
	}
}

// AC-021-2: Given tag created, Then NOT pushed (user controls push timing)
// This is verified by NOT having a GitPush call in the implementation

// AC-021-3: Given `release close` (no --tag), Then no tag created
func TestRunReleaseCloseWithDeps_NoTag_NoGitTagCreated(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Release: v1.2.0",
			State:  "OPEN",
		},
	}
	mock.releaseIssues = []api.Issue{}

	cfg := testReleaseConfig()
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestReleaseCmd()
	opts := &releaseCloseOptions{
		releaseName: "v1.2.0",
		yes:         true,
		tag:         false, // No --tag flag
	}

	// ACT
	err := runReleaseCloseWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify GitTag was NOT called
	if len(mock.gitTagCalls) != 0 {
		t.Errorf("Expected 0 GitTag calls (no --tag flag), got %d", len(mock.gitTagCalls))
	}
}

// =============================================================================
// REQ-022: List Releases
// =============================================================================

// AC-022-1: Given `release list`, Then table: Version, Codename, Tracker#, Issues, Date, Status
func TestRunReleaseListWithDeps_DisplaysReleaseTable(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_200",
			Number: 200,
			Title:  "Release: v2.0.0 (Phoenix)",
			State:  "OPEN",
		},
	}
	mock.closedIssues = []api.Issue{
		{
			ID:     "TRACKER_100",
			Number: 100,
			Title:  "Release: v1.0.0",
			State:  "CLOSED",
		},
	}

	cfg := testReleaseConfig()
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, buf := newTestReleaseCmd()
	opts := &releaseListOptions{}

	// ACT
	err := runReleaseListWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()

	// Verify headers
	if !strings.Contains(output, "VERSION") {
		t.Errorf("Expected output to contain 'VERSION' header, got '%s'", output)
	}
	if !strings.Contains(output, "STATUS") {
		t.Errorf("Expected output to contain 'STATUS' header, got '%s'", output)
	}
	if !strings.Contains(output, "TRACKER") {
		t.Errorf("Expected output to contain 'TRACKER' header, got '%s'", output)
	}

	// Verify release data
	if !strings.Contains(output, "v2.0.0") {
		t.Errorf("Expected output to contain 'v2.0.0', got '%s'", output)
	}
	if !strings.Contains(output, "v1.0.0") {
		t.Errorf("Expected output to contain 'v1.0.0', got '%s'", output)
	}
	if !strings.Contains(output, "Phoenix") {
		t.Errorf("Expected output to contain codename 'Phoenix', got '%s'", output)
	}
}

// AC-022-2: Given multiple releases, Then sorted by version descending
func TestRunReleaseListWithDeps_SortedByVersionDescending(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{}
	mock.closedIssues = []api.Issue{
		{
			ID:     "TRACKER_100",
			Number: 100,
			Title:  "Release: v1.0.0",
			State:  "CLOSED",
		},
		{
			ID:     "TRACKER_300",
			Number: 300,
			Title:  "Release: v3.0.0",
			State:  "CLOSED",
		},
		{
			ID:     "TRACKER_200",
			Number: 200,
			Title:  "Release: v2.0.0",
			State:  "CLOSED",
		},
	}

	cfg := testReleaseConfig()
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, buf := newTestReleaseCmd()
	opts := &releaseListOptions{}

	// ACT
	err := runReleaseListWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()

	// Find positions of versions in output - v3.0.0 should appear before v2.0.0, which should appear before v1.0.0
	pos3 := strings.Index(output, "v3.0.0")
	pos2 := strings.Index(output, "v2.0.0")
	pos1 := strings.Index(output, "v1.0.0")

	if pos3 > pos2 {
		t.Errorf("Expected v3.0.0 to appear before v2.0.0 (descending order)")
	}
	if pos2 > pos1 {
		t.Errorf("Expected v2.0.0 to appear before v1.0.0 (descending order)")
	}
}

// Test no releases shows message
func TestRunReleaseListWithDeps_NoReleases(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{}
	mock.closedIssues = []api.Issue{}

	cfg := testReleaseConfig()
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, buf := newTestReleaseCmd()
	opts := &releaseListOptions{}

	// ACT
	err := runReleaseListWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No releases found") {
		t.Errorf("Expected output to contain 'No releases found', got '%s'", output)
	}
}

// ============================================================================
// runReleaseReopenWithDeps Tests
// ============================================================================

func TestRunReleaseReopenWithDeps_Success(t *testing.T) {
	mock := setupMockForRelease()
	mock.closedIssues = []api.Issue{
		{ID: "closed-1", Number: 100, Title: "Release: v1.0.0"},
	}

	cfg := testReleaseConfig()
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, buf := newTestReleaseCmd()

	err := runReleaseReopenWithDeps(cmd, "v1.0.0", cfg, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Reopened release v1.0.0") {
		t.Errorf("expected 'Reopened release v1.0.0' in output, got: %s", output)
	}
}

func TestRunReleaseReopenWithDeps_WithCodename(t *testing.T) {
	mock := setupMockForRelease()
	mock.closedIssues = []api.Issue{
		{ID: "closed-1", Number: 100, Title: "Release: v1.0.0 (Phoenix)"},
	}

	cfg := testReleaseConfig()
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, buf := newTestReleaseCmd()

	err := runReleaseReopenWithDeps(cmd, "v1.0.0", cfg, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Reopened release v1.0.0") {
		t.Errorf("expected 'Reopened release v1.0.0' in output, got: %s", output)
	}
}

func TestRunReleaseReopenWithDeps_ReleaseNotFound(t *testing.T) {
	mock := setupMockForRelease()
	mock.closedIssues = []api.Issue{
		{ID: "closed-1", Number: 100, Title: "Release: v2.0.0"},
	}

	cfg := testReleaseConfig()
	cmd, _ := newTestReleaseCmd()

	err := runReleaseReopenWithDeps(cmd, "v1.0.0", cfg, mock)
	if err == nil {
		t.Fatal("expected error for release not found")
	}
	if !strings.Contains(err.Error(), "closed release not found") {
		t.Errorf("expected 'closed release not found' error, got: %v", err)
	}
}

func TestRunReleaseReopenWithDeps_GetClosedIssuesError(t *testing.T) {
	mock := setupMockForRelease()
	mock.getClosedIssuesErr = errors.New("API error")

	cfg := testReleaseConfig()
	cmd, _ := newTestReleaseCmd()

	err := runReleaseReopenWithDeps(cmd, "v1.0.0", cfg, mock)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "failed to get closed release issues") {
		t.Errorf("expected 'failed to get closed release issues' error, got: %v", err)
	}
}

func TestRunReleaseReopenWithDeps_ReopenError(t *testing.T) {
	mock := setupMockForRelease()
	mock.closedIssues = []api.Issue{
		{ID: "closed-1", Number: 100, Title: "Release: v1.0.0"},
	}
	mock.reopenIssueErr = errors.New("reopen failed")

	cfg := testReleaseConfig()
	cmd, _ := newTestReleaseCmd()

	err := runReleaseReopenWithDeps(cmd, "v1.0.0", cfg, mock)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "failed to reopen tracker issue") {
		t.Errorf("expected 'failed to reopen tracker issue' error, got: %v", err)
	}
}

func TestRunReleaseReopenWithDeps_NoRepositories(t *testing.T) {
	mock := setupMockForRelease()
	cfg := testReleaseConfig()
	cfg.Repositories = []string{}

	cmd, _ := newTestReleaseCmd()

	err := runReleaseReopenWithDeps(cmd, "v1.0.0", cfg, mock)
	if err == nil {
		t.Fatal("expected error for no repositories")
	}
	if !strings.Contains(err.Error(), "no repositories") {
		t.Errorf("expected 'no repositories' error, got: %v", err)
	}
}

// =============================================================================
// generateReleaseTrackerTemplate Tests
// =============================================================================

func TestGenerateReleaseTrackerTemplate_ContainsBranchName(t *testing.T) {
	branch := "release/v1.2.0"
	result := generateReleaseTrackerTemplate(branch)

	if !strings.Contains(result, "`"+branch+"`") {
		t.Errorf("Template should contain branch name in backticks, got: %s", result)
	}
}

func TestGenerateReleaseTrackerTemplate_ContainsWarnings(t *testing.T) {
	result := generateReleaseTrackerTemplate("release/v1.0.0")

	warnings := []string{
		"**Release Tracker Issue**",
		"**Do not manually:**",
		"Close or reopen this issue",
		"Change the title",
		"Remove the `release` label",
	}

	for _, warning := range warnings {
		if !strings.Contains(result, warning) {
			t.Errorf("Template should contain warning %q", warning)
		}
	}
}

func TestGenerateReleaseTrackerTemplate_ContainsCommands(t *testing.T) {
	branch := "release/v1.0.0"
	result := generateReleaseTrackerTemplate(branch)

	commands := []string{
		"`gh pmu release add <issue>`",
		"`gh pmu release remove <issue>`",
		"`gh pmu release close " + branch + "`",
	}

	for _, cmd := range commands {
		if !strings.Contains(result, cmd) {
			t.Errorf("Template should contain command %q", cmd)
		}
	}
}

func TestGenerateReleaseTrackerTemplate_ContainsIssuesSection(t *testing.T) {
	result := generateReleaseTrackerTemplate("release/v1.0.0")

	if !strings.Contains(result, "## Issues in this release") {
		t.Error("Template should contain 'Issues in this release' section")
	}
	if !strings.Contains(result, "Release field in the project") {
		t.Error("Template should explain issues are tracked via the Release field")
	}
}

func TestGenerateReleaseTrackerTemplate_DifferentBranchFormats(t *testing.T) {
	tests := []struct {
		branch string
	}{
		{"release/v1.0.0"},
		{"patch/v1.0.1"},
		{"hotfix-auth-bypass"},
		{"v2.0.0-beta"},
	}

	for _, tt := range tests {
		t.Run(tt.branch, func(t *testing.T) {
			result := generateReleaseTrackerTemplate(tt.branch)
			if !strings.Contains(result, "`"+tt.branch+"`") {
				t.Errorf("Template should contain branch name %q in backticks", tt.branch)
			}
			if !strings.Contains(result, "gh pmu release close "+tt.branch) {
				t.Errorf("Template should contain close command with branch name %q", tt.branch)
			}
		})
	}
}

func TestCalculateNextVersions(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		wantPatch      string
		wantMinor      string
		wantMajor      string
		wantErr        bool
	}{
		{
			name:           "standard version",
			currentVersion: "v1.2.3",
			wantPatch:      "v1.2.4",
			wantMinor:      "v1.3.0",
			wantMajor:      "v2.0.0",
		},
		{
			name:           "without v prefix",
			currentVersion: "1.2.3",
			wantPatch:      "v1.2.4",
			wantMinor:      "v1.3.0",
			wantMajor:      "v2.0.0",
		},
		{
			name:           "zero version",
			currentVersion: "v0.0.0",
			wantPatch:      "v0.0.1",
			wantMinor:      "v0.1.0",
			wantMajor:      "v1.0.0",
		},
		{
			name:           "invalid format",
			currentVersion: "invalid",
			wantErr:        true,
		},
		{
			name:           "incomplete version",
			currentVersion: "v1.2",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			versions, err := calculateNextVersions(tt.currentVersion)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if versions.patch != tt.wantPatch {
				t.Errorf("patch: got %s, want %s", versions.patch, tt.wantPatch)
			}
			if versions.minor != tt.wantMinor {
				t.Errorf("minor: got %s, want %s", versions.minor, tt.wantMinor)
			}
			if versions.major != tt.wantMajor {
				t.Errorf("major: got %s, want %s", versions.major, tt.wantMajor)
			}
		})
	}
}

// =============================================================================
// Command Flag Existence Tests
// =============================================================================

func TestReleaseStartCommand_HasBranchFlag(t *testing.T) {
	cmd := NewRootCommand()
	startCmd, _, err := cmd.Find([]string{"release", "start"})
	if err != nil {
		t.Fatalf("release start command not found: %v", err)
	}

	flag := startCmd.Flags().Lookup("branch")
	if flag == nil {
		t.Fatal("Expected --branch flag to exist")
	}
}

func TestReleaseCloseCommand_Flags(t *testing.T) {
	cmd := NewRootCommand()
	closeCmd, _, err := cmd.Find([]string{"release", "close"})
	if err != nil {
		t.Fatalf("release close command not found: %v", err)
	}

	tests := []struct {
		flag      string
		shorthand string
	}{
		{"yes", "y"},
		{"tag", ""},
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			flag := closeCmd.Flags().Lookup(tt.flag)
			if flag == nil {
				t.Fatalf("Expected --%s flag to exist", tt.flag)
			}
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("Expected --%s shorthand to be '%s', got '%s'", tt.flag, tt.shorthand, flag.Shorthand)
			}
		})
	}
}

func TestReleaseCurrentCommand_HasRefreshFlag(t *testing.T) {
	cmd := NewRootCommand()
	currentCmd, _, err := cmd.Find([]string{"release", "current"})
	if err != nil {
		t.Fatalf("release current command not found: %v", err)
	}

	flag := currentCmd.Flags().Lookup("refresh")
	if flag == nil {
		t.Fatal("Expected --refresh flag to exist")
	}
}

func TestReleaseAddCommand_Structure(t *testing.T) {
	cmd := NewRootCommand()
	addCmd, _, err := cmd.Find([]string{"release", "add"})
	if err != nil {
		t.Fatalf("release add command not found: %v", err)
	}

	if addCmd.Use != "add <issue-number>" {
		t.Errorf("Expected Use 'add <issue-number>', got %s", addCmd.Use)
	}

	// Requires exactly 1 argument
	if err := addCmd.Args(addCmd, []string{}); err == nil {
		t.Error("Expected error when no arguments provided")
	}
	if err := addCmd.Args(addCmd, []string{"123"}); err != nil {
		t.Errorf("Unexpected error with one argument: %v", err)
	}
}

func TestReleaseRemoveCommand_Structure(t *testing.T) {
	cmd := NewRootCommand()
	removeCmd, _, err := cmd.Find([]string{"release", "remove"})
	if err != nil {
		t.Fatalf("release remove command not found: %v", err)
	}

	if removeCmd.Use != "remove <issue-number>" {
		t.Errorf("Expected Use 'remove <issue-number>', got %s", removeCmd.Use)
	}

	// Requires exactly 1 argument
	if err := removeCmd.Args(removeCmd, []string{}); err == nil {
		t.Error("Expected error when no arguments provided")
	}
	if err := removeCmd.Args(removeCmd, []string{"123"}); err != nil {
		t.Errorf("Unexpected error with one argument: %v", err)
	}
}

func TestReleaseListCommand_Structure(t *testing.T) {
	cmd := NewRootCommand()
	listCmd, _, err := cmd.Find([]string{"release", "list"})
	if err != nil {
		t.Fatalf("release list command not found: %v", err)
	}

	if listCmd.Use != "list" {
		t.Errorf("Expected Use 'list', got %s", listCmd.Use)
	}

	if listCmd.Short == "" {
		t.Error("Expected Short description to be set")
	}
}

func TestReleaseReopenCommand_Structure(t *testing.T) {
	cmd := NewRootCommand()
	reopenCmd, _, err := cmd.Find([]string{"release", "reopen"})
	if err != nil {
		t.Fatalf("release reopen command not found: %v", err)
	}

	if reopenCmd.Use != "reopen <release-name>" {
		t.Errorf("Expected Use 'reopen <release-name>', got %s", reopenCmd.Use)
	}

	// Requires exactly 1 argument
	if err := reopenCmd.Args(reopenCmd, []string{}); err == nil {
		t.Error("Expected error when no arguments provided")
	}
	if err := reopenCmd.Args(reopenCmd, []string{"release/v1.0.0"}); err != nil {
		t.Errorf("Unexpected error with one argument: %v", err)
	}
}

func TestCompareVersions_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		{"equal", "v1.0.0", "v1.0.0", 0},
		{"a greater major", "v2.0.0", "v1.0.0", 1},
		{"b greater major", "v1.0.0", "v2.0.0", -1},
		{"a greater minor", "v1.2.0", "v1.1.0", 1},
		{"b greater minor", "v1.1.0", "v1.2.0", -1},
		{"a greater patch", "v1.0.2", "v1.0.1", 1},
		{"b greater patch", "v1.0.1", "v1.0.2", -1},
		{"without prefix", "1.0.0", "1.0.0", 0},
		{"mixed prefix", "v1.0.0", "1.0.0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareVersions(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// Parking Lot Exclusion Tests
// ============================================================================

func TestRunReleaseCloseWithDeps_SkipsParkingLotIssues(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Release: v1.2.0",
			State:  "OPEN",
		},
	}
	// 3 incomplete issues: 1 parking lot, 2 regular
	mock.releaseIssues = []api.Issue{
		{ID: "ISSUE_1", Number: 41, Title: "Parked feature idea", State: "OPEN"},
		{ID: "ISSUE_2", Number: 42, Title: "Incomplete work", State: "OPEN"},
		{ID: "ISSUE_3", Number: 43, Title: "Another incomplete", State: "OPEN"},
	}
	// Map issue IDs to item IDs
	mock.projectItemIDs = map[string]string{
		"ISSUE_1": "ITEM_1",
		"ISSUE_2": "ITEM_2",
		"ISSUE_3": "ITEM_3",
	}
	// ISSUE_1 is in Parking Lot status
	mock.projectItemFieldValues = map[string]string{
		"ITEM_1": "Parking Lot",
		"ITEM_2": "In Progress",
		"ITEM_3": "Ready",
	}

	cfg := testReleaseConfig()
	cfg.Fields["status"] = config.Field{
		Field: "Status",
		Values: map[string]string{
			"backlog":     "Backlog",
			"parking_lot": "Parking Lot",
		},
	}
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, output := newTestReleaseCmd()
	opts := &releaseCloseOptions{releaseName: "v1.2.0", yes: true}

	// ACT
	err := runReleaseCloseWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify output shows skipped parking lot issues
	outputStr := output.String()
	if !strings.Contains(outputStr, "Skipping 1 Parking Lot issue") {
		t.Errorf("Expected output to mention skipped parking lot issues, got: %s", outputStr)
	}

	// Verify only 2 issues were moved to backlog (not the parking lot one)
	statusSetCount := 0
	for _, call := range mock.setFieldCalls {
		if call.fieldID == "Status" && call.value == "Backlog" {
			statusSetCount++
		}
	}
	if statusSetCount != 2 {
		t.Errorf("Expected 2 issues moved to backlog, got %d. Calls: %+v", statusSetCount, mock.setFieldCalls)
	}
}

func TestRunReleaseCloseWithDeps_AllParkingLotNoMoves(t *testing.T) {
	// ARRANGE: All incomplete issues are in Parking Lot
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Release: v1.2.0",
			State:  "OPEN",
		},
	}
	mock.releaseIssues = []api.Issue{
		{ID: "ISSUE_1", Number: 41, Title: "Parked idea 1", State: "OPEN"},
		{ID: "ISSUE_2", Number: 42, Title: "Parked idea 2", State: "OPEN"},
	}
	mock.projectItemIDs = map[string]string{
		"ISSUE_1": "ITEM_1",
		"ISSUE_2": "ITEM_2",
	}
	mock.projectItemFieldValues = map[string]string{
		"ITEM_1": "Parking Lot",
		"ITEM_2": "Parking Lot",
	}

	cfg := testReleaseConfig()
	cfg.Fields["status"] = config.Field{
		Field: "Status",
		Values: map[string]string{
			"backlog":     "Backlog",
			"parking_lot": "Parking Lot",
		},
	}
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, output := newTestReleaseCmd()
	opts := &releaseCloseOptions{releaseName: "v1.2.0", yes: true}

	// ACT
	err := runReleaseCloseWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify output shows 2 parking lot issues skipped
	outputStr := output.String()
	if !strings.Contains(outputStr, "Skipping 2 Parking Lot issue") {
		t.Errorf("Expected output to mention skipped parking lot issues, got: %s", outputStr)
	}

	// Verify no status changes to Backlog
	for _, call := range mock.setFieldCalls {
		if call.fieldID == "Status" && call.value == "Backlog" {
			t.Errorf("No issues should be moved to backlog, but found call: %+v", call)
		}
	}

	// Verify "Moving incomplete issues" message is NOT shown
	if strings.Contains(outputStr, "Moving incomplete issues") {
		t.Errorf("Should not show 'Moving incomplete issues' when all are parking lot")
	}
}

func TestRunReleaseCloseWithDeps_NoParkingLotConfig(t *testing.T) {
	// ARRANGE: No parking_lot value configured, should use default "Parking Lot"
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Release: v1.2.0",
			State:  "OPEN",
		},
	}
	mock.releaseIssues = []api.Issue{
		{ID: "ISSUE_1", Number: 41, Title: "Parked idea", State: "OPEN"},
		{ID: "ISSUE_2", Number: 42, Title: "Regular issue", State: "OPEN"},
	}
	mock.projectItemIDs = map[string]string{
		"ISSUE_1": "ITEM_1",
		"ISSUE_2": "ITEM_2",
	}
	mock.projectItemFieldValues = map[string]string{
		"ITEM_1": "Parking Lot", // Uses default value
		"ITEM_2": "In Progress",
	}

	cfg := testReleaseConfig()
	// Status field configured but no parking_lot alias
	cfg.Fields["status"] = config.Field{
		Field: "Status",
		Values: map[string]string{
			"backlog": "Backlog",
		},
	}
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, output := newTestReleaseCmd()
	opts := &releaseCloseOptions{releaseName: "v1.2.0", yes: true}

	// ACT
	err := runReleaseCloseWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should still skip parking lot with default value
	outputStr := output.String()
	if !strings.Contains(outputStr, "Skipping 1 Parking Lot issue") {
		t.Errorf("Expected parking lot issue to be skipped even without config, got: %s", outputStr)
	}
}

func TestRunReleaseCloseWithDeps_ClearsReleaseAndMicrosprintFields(t *testing.T) {
	// ARRANGE: Incomplete issues that need to be moved to backlog
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Release: v1.2.0",
			State:  "OPEN",
		},
	}
	mock.releaseIssues = []api.Issue{
		{ID: "ISSUE_1", Number: 41, Title: "Incomplete work", State: "OPEN"},
	}
	mock.projectItemIDs = map[string]string{
		"ISSUE_1": "ITEM_1",
	}
	mock.projectItemFieldValues = map[string]string{
		"ITEM_1": "In Progress",
	}

	cfg := testReleaseConfig()
	cfg.Fields["status"] = config.Field{
		Field: "Status",
		Values: map[string]string{
			"backlog": "Backlog",
		},
	}
	cfg.Fields["release"] = config.Field{
		Field: "Release",
	}
	cfg.Fields["microsprint"] = config.Field{
		Field: "Microsprint",
	}
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestReleaseCmd()
	opts := &releaseCloseOptions{releaseName: "v1.2.0", yes: true}

	// ACT
	err := runReleaseCloseWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify Release field was cleared (set to "")
	releaseCleared := false
	for _, call := range mock.setFieldCalls {
		if call.fieldID == "Release" && call.value == "" {
			releaseCleared = true
			break
		}
	}
	if !releaseCleared {
		t.Errorf("Expected Release field to be cleared, calls: %+v", mock.setFieldCalls)
	}

	// Verify Microsprint field was cleared (set to "")
	microsprintCleared := false
	for _, call := range mock.setFieldCalls {
		if call.fieldID == "Microsprint" && call.value == "" {
			microsprintCleared = true
			break
		}
	}
	if !microsprintCleared {
		t.Errorf("Expected Microsprint field to be cleared, calls: %+v", mock.setFieldCalls)
	}

	// Verify Status was set to Backlog
	statusSet := false
	for _, call := range mock.setFieldCalls {
		if call.fieldID == "Status" && call.value == "Backlog" {
			statusSet = true
			break
		}
	}
	if !statusSet {
		t.Errorf("Expected Status field to be set to Backlog, calls: %+v", mock.setFieldCalls)
	}
}

func TestRunReleaseCloseWithDeps_GetProjectItemIDError_ContinuesWithWarning(t *testing.T) {
	// ARRANGE: GetProjectItemID fails for one issue but succeeds for another
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Release: v1.2.0",
			State:  "OPEN",
		},
	}
	mock.releaseIssues = []api.Issue{
		{ID: "ISSUE_1", Number: 41, Title: "Issue without project item", State: "OPEN"},
		{ID: "ISSUE_2", Number: 42, Title: "Normal issue", State: "OPEN"},
	}
	// Only ISSUE_2 has a project item ID - ISSUE_1 will fail lookup
	mock.projectItemIDs = map[string]string{
		"ISSUE_2": "ITEM_2",
	}
	mock.projectItemFieldValues = map[string]string{
		"ITEM_2": "In Progress",
	}

	cfg := testReleaseConfig()
	cfg.Fields["status"] = config.Field{
		Field: "Status",
		Values: map[string]string{
			"backlog": "Backlog",
		},
	}
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestReleaseCmd()
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	opts := &releaseCloseOptions{releaseName: "v1.2.0", yes: true}

	// ACT
	err := runReleaseCloseWithDeps(cmd, opts, cfg, mock)

	// ASSERT: Should succeed overall
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify warning was shown for issue 41
	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "Warning") || !strings.Contains(stderrStr, "#41") {
		t.Errorf("Expected warning about issue #41, got stderr: %s", stderrStr)
	}

	// Verify only 1 issue had status set (the one that succeeded)
	statusSetCount := 0
	for _, call := range mock.setFieldCalls {
		if call.fieldID == "Status" && call.value == "Backlog" {
			statusSetCount++
		}
	}
	if statusSetCount != 1 {
		t.Errorf("Expected 1 issue moved to backlog, got %d", statusSetCount)
	}
}

func TestRunReleaseCloseWithDeps_AllIssuesDone_NoMoveToBacklog(t *testing.T) {
	// ARRANGE: All release issues are closed (done)
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Release: v1.2.0",
			State:  "OPEN",
		},
	}
	// All issues are closed (done)
	mock.releaseIssues = []api.Issue{
		{ID: "ISSUE_1", Number: 41, Title: "Completed work", State: "CLOSED"},
		{ID: "ISSUE_2", Number: 42, Title: "Also done", State: "CLOSED"},
	}

	cfg := testReleaseConfig()
	cfg.Fields["status"] = config.Field{
		Field: "Status",
		Values: map[string]string{
			"backlog": "Backlog",
		},
	}
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, output := newTestReleaseCmd()
	opts := &releaseCloseOptions{releaseName: "v1.2.0", yes: true}

	// ACT
	err := runReleaseCloseWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify output shows 2 done, 0 incomplete
	outputStr := output.String()
	if !strings.Contains(outputStr, "2 done") || !strings.Contains(outputStr, "0 incomplete") {
		t.Errorf("Expected '2 done, 0 incomplete' in output, got: %s", outputStr)
	}

	// Verify no backlog status updates (all issues already done)
	backlogCount := 0
	for _, call := range mock.setFieldCalls {
		if call.fieldID == "Status" && call.value == "Backlog" {
			backlogCount++
		}
	}
	if backlogCount != 0 {
		t.Errorf("Expected no backlog status updates (all done), got %d: %+v", backlogCount, mock.setFieldCalls)
	}
}

func TestRunReleaseCloseWithDeps_DefaultBacklogValue(t *testing.T) {
	// ARRANGE: Status field has no backlog alias defined
	mock := setupMockForRelease()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Release: v1.2.0",
			State:  "OPEN",
		},
	}
	mock.releaseIssues = []api.Issue{
		{ID: "ISSUE_1", Number: 41, Title: "Incomplete work", State: "OPEN"},
	}
	mock.projectItemIDs = map[string]string{
		"ISSUE_1": "ITEM_1",
	}
	mock.projectItemFieldValues = map[string]string{
		"ITEM_1": "In Progress",
	}

	cfg := testReleaseConfig()
	// Status field with empty values map (no backlog alias)
	cfg.Fields["status"] = config.Field{
		Field:  "Status",
		Values: map[string]string{}, // No backlog alias
	}
	cleanup := setupReleaseTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestReleaseCmd()
	opts := &releaseCloseOptions{releaseName: "v1.2.0", yes: true}

	// ACT
	err := runReleaseCloseWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify Status was set to default "Backlog" (not an alias)
	statusSet := false
	for _, call := range mock.setFieldCalls {
		if call.fieldID == "Status" && call.value == "Backlog" {
			statusSet = true
			break
		}
	}
	if !statusSet {
		t.Errorf("Expected Status field to be set to default 'Backlog', calls: %+v", mock.setFieldCalls)
	}
}
