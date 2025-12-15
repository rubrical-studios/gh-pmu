package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rubrical-studios/gh-pmu/internal/api"
	"github.com/rubrical-studios/gh-pmu/internal/config"
	"github.com/spf13/cobra"
)

// mockReleaseClient implements releaseClient for testing
type mockReleaseClient struct {
	// Return values
	createdIssue          *api.Issue
	openIssues            []api.Issue
	closedIssues          []api.Issue
	project               *api.Project
	addedItemID           string
	issueByNumber         *api.Issue
	projectItemID         string
	projectItemFieldValue string
	releaseIssues         []api.Issue

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
	return m.projectItemID, nil
}

func (m *mockReleaseClient) GetProjectItemFieldValue(projectID, itemID, fieldID string) (string, error) {
	if m.getProjectItemFieldErr != nil {
		return "", m.getProjectItemFieldErr
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

func (m *mockReleaseClient) GitTag(tag, message string) error {
	m.gitTagCalls = append(m.gitTagCalls, gitTagCall{
		tag:     tag,
		message: message,
	})
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

// AC-017-1: Given `release start --version 1.2.0`, Then tracker issue created: "Release: v1.2.0"
func TestRunReleaseStartWithDeps_CreatesTrackerIssue(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	cfg := testReleaseConfig()
	cmd, _ := newTestReleaseCmd()
	opts := &releaseStartOptions{
		version: "1.2.0",
	}

	expectedTitle := "Release: v1.2.0"

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
	cmd, _ := newTestReleaseCmd()
	opts := &releaseStartOptions{
		version: "1.2.0",
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
			Title:  "Release: v1.1.0",
			State:  "OPEN",
		},
	}
	cfg := testReleaseConfig()
	cmd, _ := newTestReleaseCmd()
	opts := &releaseStartOptions{
		version: "1.2.0",
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
	cmd, _ := newTestReleaseCmd()
	opts := &releaseStartOptions{
		version: "1.2.0",
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

// AC-018-4: Given closed release v1.2.0 exists, When starting v1.2.0, Then error: "Version v1.2.0 already released"
func TestRunReleaseStartWithDeps_DuplicateVersion_ReturnsError(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	// Add method to get closed issues
	mock.closedIssues = []api.Issue{
		{
			ID:     "CLOSED_RELEASE",
			Number: 40,
			Title:  "Release: v1.2.0",
			State:  "CLOSED",
		},
	}
	cfg := testReleaseConfig()
	cmd, _ := newTestReleaseCmd()
	opts := &releaseStartOptions{
		version: "1.2.0",
	}

	// ACT
	err := runReleaseStartWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err == nil {
		t.Fatalf("Expected error for duplicate version, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "already released") {
		t.Errorf("Expected error to mention 'already released', got: %s", errMsg)
	}
}

// Test that version validation is called during release start
func TestRunReleaseStartWithDeps_InvalidVersion_ReturnsError(t *testing.T) {
	// ARRANGE
	mock := setupMockForRelease()
	cfg := testReleaseConfig()
	cmd, _ := newTestReleaseCmd()
	opts := &releaseStartOptions{
		version: "1.2", // Invalid - missing patch version
	}

	// ACT
	err := runReleaseStartWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err == nil {
		t.Fatalf("Expected error for invalid version, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "Invalid version format") {
		t.Errorf("Expected error to mention 'Invalid version format', got: %s", errMsg)
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
		{ID: "ISSUE_1", Number: 41, Title: "Add new feature", Labels: []api.Label{{Name: "enhancement"}}},
		{ID: "ISSUE_2", Number: 42, Title: "Fix bug", Labels: []api.Label{{Name: "bug"}}},
	}

	cfg := testReleaseConfig()
	cmd, _ := newTestReleaseCmd()
	opts := &releaseCloseOptions{}

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
		{ID: "ISSUE_1", Number: 41, Title: "Add new feature", Labels: []api.Label{{Name: "enhancement"}}},
		{ID: "ISSUE_2", Number: 42, Title: "Fix critical bug", Labels: []api.Label{{Name: "bug"}}},
	}

	cfg := testReleaseConfig()
	cmd, _ := newTestReleaseCmd()
	opts := &releaseCloseOptions{}

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
		{ID: "ISSUE_1", Number: 41, Title: "Add feature"},
	}

	cfg := testReleaseConfig()
	cmd, _ := newTestReleaseCmd()
	opts := &releaseCloseOptions{}

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
		{ID: "ISSUE_1", Number: 41, Title: "Add feature"},
	}

	cfg := testReleaseConfig()
	cmd, _ := newTestReleaseCmd()
	opts := &releaseCloseOptions{}

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
	cmd, _ := newTestReleaseCmd()
	opts := &releaseCloseOptions{}

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
	cmd, _ := newTestReleaseCmd()
	opts := &releaseCloseOptions{}

	// ACT
	err := runReleaseCloseWithDeps(cmd, opts, cfg, mock)

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
	cmd, _ := newTestReleaseCmd()
	opts := &releaseCloseOptions{
		tag: true,
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
	cmd, _ := newTestReleaseCmd()
	opts := &releaseCloseOptions{
		tag: false, // No --tag flag
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
