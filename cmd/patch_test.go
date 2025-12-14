package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rubrical-studios/gh-pmu/internal/api"
	"github.com/rubrical-studios/gh-pmu/internal/config"
	"github.com/spf13/cobra"
)

// mockPatchClient implements patchClient for testing
type mockPatchClient struct {
	// Return values
	createdIssue          *api.Issue
	openIssues            []api.Issue
	closedIssues          []api.Issue
	project               *api.Project
	addedItemID           string
	issueByNumber         *api.Issue
	projectItemID         string
	projectItemFieldValue string
	patchIssues           []api.Issue

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
	getPatchIssuesErr      error
}

func (m *mockPatchClient) CreateIssue(owner, repo, title, body string, labels []string) (*api.Issue, error) {
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

func (m *mockPatchClient) GetOpenIssuesByLabel(owner, repo, label string) ([]api.Issue, error) {
	if m.getOpenIssuesErr != nil {
		return nil, m.getOpenIssuesErr
	}
	return m.openIssues, nil
}

func (m *mockPatchClient) GetClosedIssuesByLabel(owner, repo, label string) ([]api.Issue, error) {
	if m.getClosedIssuesErr != nil {
		return nil, m.getClosedIssuesErr
	}
	return m.closedIssues, nil
}

func (m *mockPatchClient) AddIssueToProject(projectID, issueID string) (string, error) {
	m.addToProjectCalls = append(m.addToProjectCalls, addToProjectCall{
		projectID: projectID,
		issueID:   issueID,
	})
	if m.addToProjectErr != nil {
		return "", m.addToProjectErr
	}
	return m.addedItemID, nil
}

func (m *mockPatchClient) SetProjectItemField(projectID, itemID, fieldID, value string) error {
	m.setFieldCalls = append(m.setFieldCalls, setFieldCall{
		projectID: projectID,
		itemID:    itemID,
		fieldID:   fieldID,
		value:     value,
	})
	return m.setFieldErr
}

func (m *mockPatchClient) GetProject(owner string, number int) (*api.Project, error) {
	if m.getProjectErr != nil {
		return nil, m.getProjectErr
	}
	return m.project, nil
}

func (m *mockPatchClient) GetIssueByNumber(owner, repo string, number int) (*api.Issue, error) {
	if m.getIssueErr != nil {
		return nil, m.getIssueErr
	}
	return m.issueByNumber, nil
}

func (m *mockPatchClient) GetProjectItemID(projectID, issueID string) (string, error) {
	if m.getProjectItemErr != nil {
		return "", m.getProjectItemErr
	}
	return m.projectItemID, nil
}

func (m *mockPatchClient) GetProjectItemFieldValue(projectID, itemID, fieldID string) (string, error) {
	if m.getProjectItemFieldErr != nil {
		return "", m.getProjectItemFieldErr
	}
	return m.projectItemFieldValue, nil
}

func (m *mockPatchClient) GetIssuesByPatch(owner, repo, patchVersion string) ([]api.Issue, error) {
	if m.getPatchIssuesErr != nil {
		return nil, m.getPatchIssuesErr
	}
	return m.patchIssues, nil
}

func (m *mockPatchClient) UpdateIssueBody(issueID, body string) error {
	m.updateIssueBodyCalls = append(m.updateIssueBodyCalls, updateIssueBodyCall{
		issueID: issueID,
		body:    body,
	})
	return nil
}

func (m *mockPatchClient) WriteFile(path, content string) error {
	m.writeFileCalls = append(m.writeFileCalls, writeFileCall{
		path:    path,
		content: content,
	})
	return nil
}

func (m *mockPatchClient) MkdirAll(path string) error {
	return nil
}

func (m *mockPatchClient) GitAdd(paths ...string) error {
	m.gitAddCalls = append(m.gitAddCalls, gitAddCall{
		paths: paths,
	})
	return nil
}

func (m *mockPatchClient) CloseIssue(issueID string) error {
	m.closeIssueCalls = append(m.closeIssueCalls, closeIssueCall{
		issueID: issueID,
	})
	return nil
}

func (m *mockPatchClient) GitTag(tag, message string) error {
	m.gitTagCalls = append(m.gitTagCalls, gitTagCall{
		tag:     tag,
		message: message,
	})
	return nil
}

// testPatchConfig returns a test configuration for patch tests
func testPatchConfig() *config.Config {
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

// setupMockForPatch creates a mock configured for patch tests
func setupMockForPatch() *mockPatchClient {
	return &mockPatchClient{
		openIssues: []api.Issue{}, // No active patches
		createdIssue: &api.Issue{
			ID:     "ISSUE_123",
			Number: 100,
			Title:  "Patch: v1.1.5",
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
func newTestPatchCmd() (*cobra.Command, *bytes.Buffer) {
	cmd := &cobra.Command{Use: "patch"}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	return cmd, buf
}

// =============================================================================
// REQ-023: Start Patch
// =============================================================================

// AC-023-1: Given `patch start --version 1.1.5`, Then tracker issue created: "Patch: v1.1.5"
func TestRunPatchStartWithDeps_CreatesTrackerIssue(t *testing.T) {
	// ARRANGE
	mock := setupMockForPatch()
	cfg := testPatchConfig()
	cmd, _ := newTestPatchCmd()
	opts := &patchStartOptions{
		version: "1.1.5",
	}

	expectedTitle := "Patch: v1.1.5"

	// ACT
	err := runPatchStartWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mock.createIssueCalls) != 1 {
		t.Fatalf("Expected 1 CreateIssue call, got %d", len(mock.createIssueCalls))
	}

	call := mock.createIssueCalls[0]
	if call.title != expectedTitle {
		t.Errorf("Expected title '%s', got '%s'", expectedTitle, call.title)
	}
}

// AC-023-2: Given tracker issue created, Then has `patch` label
func TestRunPatchStartWithDeps_HasPatchLabel(t *testing.T) {
	// ARRANGE
	mock := setupMockForPatch()
	cfg := testPatchConfig()
	cmd, _ := newTestPatchCmd()
	opts := &patchStartOptions{
		version: "1.1.5",
	}

	// ACT
	err := runPatchStartWithDeps(cmd, opts, cfg, mock)

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
		if label == "patch" {
			hasLabel = true
			break
		}
	}
	if !hasLabel {
		t.Errorf("Expected 'patch' label, got labels: %v", call.labels)
	}
}

// AC-023-3: Given active patch exists, When running `patch start`, Then error
func TestRunPatchStartWithDeps_ActivePatchExists_ReturnsError(t *testing.T) {
	// ARRANGE
	mock := setupMockForPatch()
	mock.openIssues = []api.Issue{
		{
			ID:     "EXISTING_PATCH",
			Number: 50,
			Title:  "Patch: v1.1.4",
			State:  "OPEN",
		},
	}
	cfg := testPatchConfig()
	cmd, _ := newTestPatchCmd()
	opts := &patchStartOptions{
		version: "1.1.5",
	}

	// ACT
	err := runPatchStartWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err == nil {
		t.Fatalf("Expected error for active patch exists, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(strings.ToLower(errMsg), "active patch") {
		t.Errorf("Expected error to mention 'active patch', got: %s", errMsg)
	}
}

// =============================================================================
// REQ-025: Add Issue to Patch
// =============================================================================

func TestRunPatchAddWithDeps_SetsPatchField(t *testing.T) {
	// ARRANGE
	mock := setupMockForPatch()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Patch: v1.1.5",
			State:  "OPEN",
		},
	}
	mock.issueByNumber = &api.Issue{
		ID:     "ISSUE_42",
		Number: 42,
		Title:  "Fix critical bug",
	}
	mock.projectItemID = "ITEM_42"

	cfg := testPatchConfig()
	cfg.Fields["patch"] = config.Field{
		Field: "Patch",
	}

	cmd, _ := newTestPatchCmd()
	opts := &patchAddOptions{
		issueNumber: 42,
	}

	// ACT
	err := runPatchAddWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mock.setFieldCalls) != 1 {
		t.Fatalf("Expected 1 SetProjectItemField call, got %d", len(mock.setFieldCalls))
	}

	call := mock.setFieldCalls[0]
	if call.value != "v1.1.5" {
		t.Errorf("Expected field value 'v1.1.5', got '%s'", call.value)
	}
}

func TestRunPatchAddWithDeps_NoActivePatch_ReturnsError(t *testing.T) {
	// ARRANGE
	mock := setupMockForPatch()
	mock.openIssues = []api.Issue{} // No active patch

	cfg := testPatchConfig()
	cfg.Fields["patch"] = config.Field{
		Field: "Patch",
	}

	cmd, _ := newTestPatchCmd()
	opts := &patchAddOptions{
		issueNumber: 42,
	}

	// ACT
	err := runPatchAddWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err == nil {
		t.Fatalf("Expected error for no active patch, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "no active patch") {
		t.Errorf("Expected error to mention 'no active patch', got: %s", errMsg)
	}
}

// =============================================================================
// REQ-024: LTS Constraints
// =============================================================================

// AC-024-1: Given issue without `bug` or `hotfix` label, When running `patch add`,
// Then warning: "Issue #42 is not labeled bug/hotfix"
func TestRunPatchAddWithDeps_NonBugIssue_ShowsWarning(t *testing.T) {
	// ARRANGE
	mock := setupMockForPatch()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Patch: v1.1.5",
			State:  "OPEN",
		},
	}
	mock.issueByNumber = &api.Issue{
		ID:     "ISSUE_42",
		Number: 42,
		Title:  "Add new feature",
		Labels: []api.Label{{Name: "enhancement"}}, // Not a bug
	}
	mock.projectItemID = "ITEM_42"

	cfg := testPatchConfig()
	cfg.Fields["patch"] = config.Field{
		Field: "Patch",
	}

	cmd, buf := newTestPatchCmd()
	opts := &patchAddOptions{
		issueNumber: 42,
	}

	// ACT
	err := runPatchAddWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error (warning only), got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "not labeled bug/hotfix") {
		t.Errorf("Expected output to contain warning about bug/hotfix label, got '%s'", output)
	}

	// Field should still be set
	if len(mock.setFieldCalls) != 1 {
		t.Fatalf("Expected field to still be set despite warning, got %d calls", len(mock.setFieldCalls))
	}
}

// AC-024-2: Given issue with `breaking-change` label, When running `patch add`,
// Then error: "Breaking changes not allowed in patches"
func TestRunPatchAddWithDeps_BreakingChange_ReturnsError(t *testing.T) {
	// ARRANGE
	mock := setupMockForPatch()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Patch: v1.1.5",
			State:  "OPEN",
		},
	}
	mock.issueByNumber = &api.Issue{
		ID:     "ISSUE_42",
		Number: 42,
		Title:  "Breaking API change",
		Labels: []api.Label{{Name: "breaking-change"}},
	}
	mock.projectItemID = "ITEM_42"

	cfg := testPatchConfig()
	cfg.Fields["patch"] = config.Field{
		Field: "Patch",
	}

	cmd, _ := newTestPatchCmd()
	opts := &patchAddOptions{
		issueNumber: 42,
	}

	// ACT
	err := runPatchAddWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err == nil {
		t.Fatalf("Expected error for breaking-change label, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "Breaking changes not allowed") {
		t.Errorf("Expected error to mention 'Breaking changes not allowed', got: %s", errMsg)
	}

	// Field should NOT be set
	if len(mock.setFieldCalls) != 0 {
		t.Errorf("Expected field NOT to be set for breaking change, got %d calls", len(mock.setFieldCalls))
	}
}

// AC-024-3: Given bug-labeled issue, When running `patch add`, Then no warning
func TestRunPatchAddWithDeps_BugLabeledIssue_NoWarning(t *testing.T) {
	// ARRANGE
	mock := setupMockForPatch()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Patch: v1.1.5",
			State:  "OPEN",
		},
	}
	mock.issueByNumber = &api.Issue{
		ID:     "ISSUE_42",
		Number: 42,
		Title:  "Fix critical bug",
		Labels: []api.Label{{Name: "bug"}}, // Properly labeled
	}
	mock.projectItemID = "ITEM_42"

	cfg := testPatchConfig()
	cfg.Fields["patch"] = config.Field{
		Field: "Patch",
	}

	cmd, buf := newTestPatchCmd()
	opts := &patchAddOptions{
		issueNumber: 42,
	}

	// ACT
	err := runPatchAddWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "not labeled bug/hotfix") {
		t.Errorf("Expected no warning for bug-labeled issue, got '%s'", output)
	}
}

// Test hotfix label also passes without warning
func TestRunPatchAddWithDeps_HotfixLabeledIssue_NoWarning(t *testing.T) {
	// ARRANGE
	mock := setupMockForPatch()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Patch: v1.1.5",
			State:  "OPEN",
		},
	}
	mock.issueByNumber = &api.Issue{
		ID:     "ISSUE_42",
		Number: 42,
		Title:  "Emergency hotfix",
		Labels: []api.Label{{Name: "hotfix"}}, // Properly labeled
	}
	mock.projectItemID = "ITEM_42"

	cfg := testPatchConfig()
	cfg.Fields["patch"] = config.Field{
		Field: "Patch",
	}

	cmd, buf := newTestPatchCmd()
	opts := &patchAddOptions{
		issueNumber: 42,
	}

	// ACT
	err := runPatchAddWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "not labeled bug/hotfix") {
		t.Errorf("Expected no warning for hotfix-labeled issue, got '%s'", output)
	}
}

// =============================================================================
// REQ-040: Remove Issue from Patch
// =============================================================================

func TestRunPatchRemoveWithDeps_ClearsPatchField(t *testing.T) {
	// ARRANGE
	mock := setupMockForPatch()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Patch: v1.1.5",
			State:  "OPEN",
		},
	}
	mock.issueByNumber = &api.Issue{
		ID:     "ISSUE_42",
		Number: 42,
		Title:  "Fix bug",
	}
	mock.projectItemID = "ITEM_42"
	mock.projectItemFieldValue = "v1.1.5" // Currently assigned

	cfg := testPatchConfig()
	cfg.Fields["patch"] = config.Field{
		Field: "Patch",
	}

	cmd, _ := newTestPatchCmd()
	opts := &patchRemoveOptions{
		issueNumber: 42,
	}

	// ACT
	err := runPatchRemoveWithDeps(cmd, opts, cfg, mock)

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

func TestRunPatchRemoveWithDeps_WarnsIfNotAssigned(t *testing.T) {
	// ARRANGE
	mock := setupMockForPatch()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Patch: v1.1.5",
			State:  "OPEN",
		},
	}
	mock.issueByNumber = &api.Issue{
		ID:     "ISSUE_42",
		Number: 42,
		Title:  "Fix bug",
	}
	mock.projectItemID = "ITEM_42"
	mock.projectItemFieldValue = "" // Not assigned

	cfg := testPatchConfig()
	cfg.Fields["patch"] = config.Field{
		Field: "Patch",
	}

	cmd, buf := newTestPatchCmd()
	opts := &patchRemoveOptions{
		issueNumber: 42,
	}

	// ACT
	err := runPatchRemoveWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error (warning only), got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "not assigned to a patch") {
		t.Errorf("Expected output to contain warning, got '%s'", output)
	}

	if len(mock.setFieldCalls) != 0 {
		t.Errorf("Expected 0 SetProjectItemField calls, got %d", len(mock.setFieldCalls))
	}
}

// =============================================================================
// REQ-037: View Current Patch
// =============================================================================

func TestRunPatchCurrentWithDeps_DisplaysActiveDetails(t *testing.T) {
	// ARRANGE
	mock := setupMockForPatch()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Patch: v1.1.5",
			State:  "OPEN",
		},
	}
	mock.patchIssues = []api.Issue{
		{ID: "ISSUE_1", Number: 41, Title: "Fix bug A"},
		{ID: "ISSUE_2", Number: 42, Title: "Fix bug B"},
	}

	cfg := testPatchConfig()
	cmd, buf := newTestPatchCmd()
	opts := &patchCurrentOptions{}

	// ACT
	err := runPatchCurrentWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "v1.1.5") {
		t.Errorf("Expected output to contain version 'v1.1.5', got '%s'", output)
	}
	if !strings.Contains(output, "#100") {
		t.Errorf("Expected output to contain tracker issue '#100', got '%s'", output)
	}
	if !strings.Contains(output, "2") {
		t.Errorf("Expected output to contain issue count '2', got '%s'", output)
	}
}

func TestRunPatchCurrentWithDeps_NoActivePatch(t *testing.T) {
	// ARRANGE
	mock := setupMockForPatch()
	mock.openIssues = []api.Issue{} // No active patch

	cfg := testPatchConfig()
	cmd, buf := newTestPatchCmd()
	opts := &patchCurrentOptions{}

	// ACT
	err := runPatchCurrentWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No active patch") {
		t.Errorf("Expected output to contain 'No active patch', got '%s'", output)
	}
}

// =============================================================================
// REQ-026: Patch Artifacts
// =============================================================================

func TestRunPatchCloseWithDeps_GeneratesPatchNotes(t *testing.T) {
	// ARRANGE
	mock := setupMockForPatch()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Patch: v1.1.5",
			State:  "OPEN",
		},
	}
	mock.patchIssues = []api.Issue{
		{ID: "ISSUE_1", Number: 41, Title: "Fix bug A"},
	}

	cfg := testPatchConfig()
	cmd, _ := newTestPatchCmd()
	opts := &patchCloseOptions{}

	// ACT
	err := runPatchCloseWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	found := false
	for _, call := range mock.writeFileCalls {
		if strings.Contains(call.path, "Patches/v1.1.5/patch-notes.md") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected WriteFile call for Patches/v1.1.5/patch-notes.md, got calls: %+v", mock.writeFileCalls)
	}
}

func TestRunPatchCloseWithDeps_StagesArtifacts(t *testing.T) {
	// ARRANGE
	mock := setupMockForPatch()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Patch: v1.1.5",
			State:  "OPEN",
		},
	}
	mock.patchIssues = []api.Issue{}

	cfg := testPatchConfig()
	cmd, _ := newTestPatchCmd()
	opts := &patchCloseOptions{}

	// ACT
	err := runPatchCloseWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mock.gitAddCalls) == 0 {
		t.Errorf("Expected GitAdd to be called for staging artifacts")
	}
}

func TestRunPatchCloseWithDeps_ClosesTrackerIssue(t *testing.T) {
	// ARRANGE
	mock := setupMockForPatch()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Patch: v1.1.5",
			State:  "OPEN",
		},
	}
	mock.patchIssues = []api.Issue{}

	cfg := testPatchConfig()
	cmd, _ := newTestPatchCmd()
	opts := &patchCloseOptions{}

	// ACT
	err := runPatchCloseWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mock.closeIssueCalls) != 1 {
		t.Fatalf("Expected 1 CloseIssue call, got %d", len(mock.closeIssueCalls))
	}

	if mock.closeIssueCalls[0].issueID != "TRACKER_123" {
		t.Errorf("Expected to close TRACKER_123, got %s", mock.closeIssueCalls[0].issueID)
	}
}

// =============================================================================
// REQ-027: Patch Git Tag
// =============================================================================

func TestRunPatchCloseWithDeps_WithTag_CreatesGitTag(t *testing.T) {
	// ARRANGE
	mock := setupMockForPatch()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Patch: v1.1.5",
			State:  "OPEN",
		},
	}
	mock.patchIssues = []api.Issue{}

	cfg := testPatchConfig()
	cmd, _ := newTestPatchCmd()
	opts := &patchCloseOptions{
		tag: true,
	}

	// ACT
	err := runPatchCloseWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mock.gitTagCalls) != 1 {
		t.Fatalf("Expected 1 GitTag call, got %d", len(mock.gitTagCalls))
	}

	call := mock.gitTagCalls[0]
	if call.tag != "v1.1.5" {
		t.Errorf("Expected tag 'v1.1.5', got '%s'", call.tag)
	}
}

func TestRunPatchCloseWithDeps_NoTag_NoGitTagCreated(t *testing.T) {
	// ARRANGE
	mock := setupMockForPatch()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Patch: v1.1.5",
			State:  "OPEN",
		},
	}
	mock.patchIssues = []api.Issue{}

	cfg := testPatchConfig()
	cmd, _ := newTestPatchCmd()
	opts := &patchCloseOptions{
		tag: false,
	}

	// ACT
	err := runPatchCloseWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mock.gitTagCalls) != 0 {
		t.Errorf("Expected 0 GitTag calls, got %d", len(mock.gitTagCalls))
	}
}

// =============================================================================
// REQ-028: List Patches
// =============================================================================

func TestRunPatchListWithDeps_DisplaysPatchTable(t *testing.T) {
	// ARRANGE
	mock := setupMockForPatch()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_200",
			Number: 200,
			Title:  "Patch: v2.0.1",
			State:  "OPEN",
		},
	}
	mock.closedIssues = []api.Issue{
		{
			ID:     "TRACKER_100",
			Number: 100,
			Title:  "Patch: v1.0.1",
			State:  "CLOSED",
		},
	}

	cfg := testPatchConfig()
	cmd, buf := newTestPatchCmd()
	opts := &patchListOptions{}

	// ACT
	err := runPatchListWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "VERSION") {
		t.Errorf("Expected output to contain 'VERSION' header, got '%s'", output)
	}
	if !strings.Contains(output, "v2.0.1") {
		t.Errorf("Expected output to contain 'v2.0.1', got '%s'", output)
	}
	if !strings.Contains(output, "v1.0.1") {
		t.Errorf("Expected output to contain 'v1.0.1', got '%s'", output)
	}
}

func TestRunPatchListWithDeps_SortedByVersionDescending(t *testing.T) {
	// ARRANGE
	mock := setupMockForPatch()
	mock.openIssues = []api.Issue{}
	mock.closedIssues = []api.Issue{
		{
			ID:     "TRACKER_100",
			Number: 100,
			Title:  "Patch: v1.0.1",
			State:  "CLOSED",
		},
		{
			ID:     "TRACKER_300",
			Number: 300,
			Title:  "Patch: v3.0.1",
			State:  "CLOSED",
		},
		{
			ID:     "TRACKER_200",
			Number: 200,
			Title:  "Patch: v2.0.1",
			State:  "CLOSED",
		},
	}

	cfg := testPatchConfig()
	cmd, buf := newTestPatchCmd()
	opts := &patchListOptions{}

	// ACT
	err := runPatchListWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()

	pos3 := strings.Index(output, "v3.0.1")
	pos2 := strings.Index(output, "v2.0.1")
	pos1 := strings.Index(output, "v1.0.1")

	if pos3 > pos2 {
		t.Errorf("Expected v3.0.1 to appear before v2.0.1 (descending order)")
	}
	if pos2 > pos1 {
		t.Errorf("Expected v2.0.1 to appear before v1.0.1 (descending order)")
	}
}

func TestRunPatchListWithDeps_NoPatches(t *testing.T) {
	// ARRANGE
	mock := setupMockForPatch()
	mock.openIssues = []api.Issue{}
	mock.closedIssues = []api.Issue{}

	cfg := testPatchConfig()
	cmd, buf := newTestPatchCmd()
	opts := &patchListOptions{}

	// ACT
	err := runPatchListWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No patches found") {
		t.Errorf("Expected output to contain 'No patches found', got '%s'", output)
	}
}
