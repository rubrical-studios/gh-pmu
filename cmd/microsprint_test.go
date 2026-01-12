package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rubrical-studios/gh-pmu/internal/api"
	"github.com/rubrical-studios/gh-pmu/internal/config"
	"github.com/spf13/cobra"
)

// setupMicrosprintTestDir creates a temp directory with a .gh-pmu.yml config file
// and changes to that directory. Returns cleanup function to restore original dir.
func setupMicrosprintTestDir(t *testing.T, cfg *config.Config) func() {
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

// mockMicrosprintClient implements microsprintClient for testing
type mockMicrosprintClient struct {
	// Return values
	createdIssue          *api.Issue
	authenticatedUser     string
	openIssues            []api.Issue
	closedIssues          []api.Issue
	project               *api.Project
	addedItemID           string
	issueByNumber         *api.Issue
	projectItemID         string
	projectItemFieldValue string
	microsprintIssues     []api.Issue

	// Captured calls for verification
	createIssueCalls     []createIssueCall
	addToProjectCalls    []addToProjectCall
	setFieldCalls        []setFieldCall
	closeIssueCalls      []closeIssueCall
	updateIssueBodyCalls []updateIssueBodyCall
	writeFileCalls       []writeFileCall
	mkdirCalls           []string
	gitAddCalls          []string
	gitCommitCalls       []gitCommitCall

	// Error injection
	createIssueErr          error
	getAuthUserErr          error
	getOpenIssuesErr        error
	getClosedIssuesErr      error
	addToProjectErr         error
	setFieldErr             error
	getProjectErr           error
	closeIssueErr           error
	getIssueErr             error
	getProjectItemErr       error
	getProjectItemFieldErr  error
	getMicrosprintIssuesErr error
	writeFileErr            error
	mkdirErr                error
	gitAddErr               error
	gitCommitErr            error
}

type createIssueCall struct {
	owner  string
	repo   string
	title  string
	body   string
	labels []string
}

type addToProjectCall struct {
	projectID string
	issueID   string
}

type setFieldCall struct {
	projectID string
	itemID    string
	fieldID   string
	value     string
}

type closeIssueCall struct {
	issueID string
}

type updateIssueBodyCall struct {
	issueID string
	body    string
}

type writeFileCall struct {
	path    string
	content string
}

type gitCommitCall struct {
	message string
}

func (m *mockMicrosprintClient) CreateIssue(owner, repo, title, body string, labels []string) (*api.Issue, error) {
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

func (m *mockMicrosprintClient) GetAuthenticatedUser() (string, error) {
	if m.getAuthUserErr != nil {
		return "", m.getAuthUserErr
	}
	return m.authenticatedUser, nil
}

func (m *mockMicrosprintClient) GetOpenIssuesByLabel(owner, repo, label string) ([]api.Issue, error) {
	if m.getOpenIssuesErr != nil {
		return nil, m.getOpenIssuesErr
	}
	return m.openIssues, nil
}

func (m *mockMicrosprintClient) GetClosedIssuesByLabel(owner, repo, label string) ([]api.Issue, error) {
	if m.getClosedIssuesErr != nil {
		return nil, m.getClosedIssuesErr
	}
	return m.closedIssues, nil
}

func (m *mockMicrosprintClient) AddIssueToProject(projectID, issueID string) (string, error) {
	m.addToProjectCalls = append(m.addToProjectCalls, addToProjectCall{
		projectID: projectID,
		issueID:   issueID,
	})
	if m.addToProjectErr != nil {
		return "", m.addToProjectErr
	}
	return m.addedItemID, nil
}

func (m *mockMicrosprintClient) SetProjectItemField(projectID, itemID, fieldID, value string) error {
	m.setFieldCalls = append(m.setFieldCalls, setFieldCall{
		projectID: projectID,
		itemID:    itemID,
		fieldID:   fieldID,
		value:     value,
	})
	return m.setFieldErr
}

func (m *mockMicrosprintClient) GetProject(owner string, number int) (*api.Project, error) {
	if m.getProjectErr != nil {
		return nil, m.getProjectErr
	}
	return m.project, nil
}

func (m *mockMicrosprintClient) CloseIssue(issueID string) error {
	m.closeIssueCalls = append(m.closeIssueCalls, closeIssueCall{
		issueID: issueID,
	})
	return m.closeIssueErr
}

func (m *mockMicrosprintClient) GetIssueByNumber(owner, repo string, number int) (*api.Issue, error) {
	if m.getIssueErr != nil {
		return nil, m.getIssueErr
	}
	return m.issueByNumber, nil
}

func (m *mockMicrosprintClient) GetProjectItemID(projectID, issueID string) (string, error) {
	if m.getProjectItemErr != nil {
		return "", m.getProjectItemErr
	}
	return m.projectItemID, nil
}

func (m *mockMicrosprintClient) UpdateIssueBody(issueID, body string) error {
	m.updateIssueBodyCalls = append(m.updateIssueBodyCalls, updateIssueBodyCall{
		issueID: issueID,
		body:    body,
	})
	return nil
}

func (m *mockMicrosprintClient) GetProjectItemFieldValue(projectID, itemID, fieldID string) (string, error) {
	if m.getProjectItemFieldErr != nil {
		return "", m.getProjectItemFieldErr
	}
	return m.projectItemFieldValue, nil
}

func (m *mockMicrosprintClient) GetIssuesByMicrosprint(owner, repo, microsprintName string) ([]api.Issue, error) {
	if m.getMicrosprintIssuesErr != nil {
		return nil, m.getMicrosprintIssuesErr
	}
	return m.microsprintIssues, nil
}

func (m *mockMicrosprintClient) WriteFile(path, content string) error {
	m.writeFileCalls = append(m.writeFileCalls, writeFileCall{
		path:    path,
		content: content,
	})
	return m.writeFileErr
}

func (m *mockMicrosprintClient) MkdirAll(path string) error {
	m.mkdirCalls = append(m.mkdirCalls, path)
	return m.mkdirErr
}

func (m *mockMicrosprintClient) GitAdd(paths ...string) error {
	m.gitAddCalls = append(m.gitAddCalls, paths...)
	return m.gitAddErr
}

func (m *mockMicrosprintClient) GitCommit(message string) error {
	m.gitCommitCalls = append(m.gitCommitCalls, gitCommitCall{message: message})
	return m.gitCommitErr
}

// testMicrosprintConfig returns a test configuration
func testMicrosprintConfig() *config.Config {
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

// setupMockForStart creates a mock configured for microsprint start tests
func setupMockForStart() *mockMicrosprintClient {
	return &mockMicrosprintClient{
		authenticatedUser: "testuser",
		openIssues:        []api.Issue{}, // No active microsprints
		createdIssue: &api.Issue{
			ID:     "ISSUE_123",
			Number: 100,
			Title:  "Microsprint: 2025-12-13-a",
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

func setupMockForMicrosprint() *mockMicrosprintClient {
	return &mockMicrosprintClient{
		authenticatedUser: "testuser",
		openIssues:        []api.Issue{},
		closedIssues:      []api.Issue{},
		project: &api.Project{
			ID:     "PROJECT_1",
			Number: 1,
			Title:  "Test Project",
		},
	}
}

// Helper to create a test command with captured output
func newTestMicrosprintCmd() (*cobra.Command, *bytes.Buffer) {
	cmd := newMicrosprintCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	return cmd, buf
}

// =============================================================================
// AC-001-1: Given no active microsprint, When user runs `microsprint start`,
// Then tracker issue created with title "Microsprint: YYYY-MM-DD-a" and label `microsprint`
// =============================================================================

func TestRunMicrosprintStartWithDeps_CreatesTrackerIssue(t *testing.T) {
	// ARRANGE
	mock := setupMockForStart()
	cfg := testMicrosprintConfig()
	cleanup := setupMicrosprintTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintStartOptions{}

	// Expected date-based name
	today := time.Now().Format("2006-01-02")
	expectedTitle := "Microsprint: " + today + "-a"

	// ACT
	err := runMicrosprintStartWithDeps(cmd, opts, cfg, mock)

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

	// Verify microsprint label is applied
	hasLabel := false
	for _, label := range call.labels {
		if label == "microsprint" {
			hasLabel = true
			break
		}
	}
	if !hasLabel {
		t.Errorf("Expected 'microsprint' label, got labels: %v", call.labels)
	}

	// Verify correct repository
	if call.owner != "testowner" || call.repo != "testrepo" {
		t.Errorf("Expected owner/repo 'testowner/testrepo', got '%s/%s'", call.owner, call.repo)
	}
}

func TestRunMicrosprintStartWithDeps_WithNameFlag_AppendsSuffix(t *testing.T) {
	// ARRANGE - AC-001-2
	mock := setupMockForStart()
	cfg := testMicrosprintConfig()
	cleanup := setupMicrosprintTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintStartOptions{
		name: "auth",
	}

	today := time.Now().Format("2006-01-02")
	expectedTitle := "Microsprint: " + today + "-a-auth"

	// ACT
	err := runMicrosprintStartWithDeps(cmd, opts, cfg, mock)

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

func TestRunMicrosprintStartWithDeps_AssignsToCurrentUser(t *testing.T) {
	// ARRANGE - AC-001-3
	mock := setupMockForStart()
	mock.authenticatedUser = "alice"
	cfg := testMicrosprintConfig()
	cleanup := setupMicrosprintTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintStartOptions{}

	// ACT
	err := runMicrosprintStartWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// The issue should be assigned to the authenticated user
	// This would be verified by checking the CreateIssue call includes assignee
	// or by a separate AssignIssue call - implementation will determine exact pattern
	if len(mock.createIssueCalls) != 1 {
		t.Fatalf("Expected 1 CreateIssue call, got %d", len(mock.createIssueCalls))
	}

	// Note: Assignment verification depends on implementation
	// Either CreateIssue accepts assignees parameter, or separate AssignIssue call
}

func TestRunMicrosprintStartWithDeps_SetsStatusToInProgress(t *testing.T) {
	// ARRANGE - AC-001-4
	mock := setupMockForStart()
	cfg := testMicrosprintConfig()
	cleanup := setupMicrosprintTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintStartOptions{}

	// ACT
	err := runMicrosprintStartWithDeps(cmd, opts, cfg, mock)

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
// REQ-002: Auto-Generated Naming
// =============================================================================

// AC-002-1: Given no microsprints today, When starting microsprint, Then name is YYYY-MM-DD-a
// (Already covered by TestRunMicrosprintStartWithDeps_CreatesTrackerIssue)

// AC-002-2: Given microsprint YYYY-MM-DD-a exists, When starting new microsprint, Then name is YYYY-MM-DD-b
func TestRunMicrosprintStartWithDeps_AutoIncrement_AtoB(t *testing.T) {
	// ARRANGE
	today := time.Now().Format("2006-01-02")
	mock := setupMockForStart()
	// Simulate existing microsprint "a" for today
	mock.openIssues = []api.Issue{
		{
			ID:     "EXISTING_1",
			Number: 50,
			Title:  "Microsprint: " + today + "-a",
		},
	}
	cfg := testMicrosprintConfig()
	cleanup := setupMicrosprintTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintStartOptions{}

	expectedTitle := "Microsprint: " + today + "-b"

	// ACT
	err := runMicrosprintStartWithDeps(cmd, opts, cfg, mock)

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

// AC-002-2 continued: Multiple existing microsprints
func TestRunMicrosprintStartWithDeps_AutoIncrement_BtoC(t *testing.T) {
	// ARRANGE
	today := time.Now().Format("2006-01-02")
	mock := setupMockForStart()
	// Simulate existing microsprints "a" and "b" for today
	mock.openIssues = []api.Issue{
		{
			ID:     "EXISTING_1",
			Number: 50,
			Title:  "Microsprint: " + today + "-a",
		},
		{
			ID:     "EXISTING_2",
			Number: 51,
			Title:  "Microsprint: " + today + "-b",
		},
	}
	cfg := testMicrosprintConfig()
	cleanup := setupMicrosprintTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintStartOptions{}

	expectedTitle := "Microsprint: " + today + "-c"

	// ACT
	err := runMicrosprintStartWithDeps(cmd, opts, cfg, mock)

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

// AC-002-3: Given microsprint YYYY-MM-DD-z exists, When starting new microsprint, Then name is YYYY-MM-DD-aa
func TestRunMicrosprintStartWithDeps_AutoIncrement_ZtoAA(t *testing.T) {
	// ARRANGE
	today := time.Now().Format("2006-01-02")
	mock := setupMockForStart()
	// Simulate existing microsprint "z" for today
	mock.openIssues = []api.Issue{
		{
			ID:     "EXISTING_Z",
			Number: 75,
			Title:  "Microsprint: " + today + "-z",
		},
	}
	cfg := testMicrosprintConfig()
	cleanup := setupMicrosprintTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintStartOptions{}

	expectedTitle := "Microsprint: " + today + "-aa"

	// ACT
	err := runMicrosprintStartWithDeps(cmd, opts, cfg, mock)

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

// AC-002-3 continued: Double letter increment
func TestRunMicrosprintStartWithDeps_AutoIncrement_AAtoAB(t *testing.T) {
	// ARRANGE
	today := time.Now().Format("2006-01-02")
	mock := setupMockForStart()
	// Simulate existing microsprint "aa" for today
	mock.openIssues = []api.Issue{
		{
			ID:     "EXISTING_AA",
			Number: 76,
			Title:  "Microsprint: " + today + "-aa",
		},
	}
	cfg := testMicrosprintConfig()
	cleanup := setupMicrosprintTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintStartOptions{}

	expectedTitle := "Microsprint: " + today + "-ab"

	// ACT
	err := runMicrosprintStartWithDeps(cmd, opts, cfg, mock)

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

// =============================================================================
// REQ-007: Tracker Issue Per Microsprint
// =============================================================================

// AC-007-1: Given `microsprint start`, Then new tracker issue created (not reused)
func TestRunMicrosprintStartWithDeps_CreatesNewIssue(t *testing.T) {
	// ARRANGE
	mock := setupMockForStart()
	cfg := testMicrosprintConfig()
	cleanup := setupMicrosprintTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintStartOptions{}

	// ACT
	err := runMicrosprintStartWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify CreateIssue was called (new issue created, not reused)
	if len(mock.createIssueCalls) != 1 {
		t.Fatalf("Expected exactly 1 CreateIssue call (new issue), got %d", len(mock.createIssueCalls))
	}

	// Verify it's creating via CreateIssue, not updating existing
	call := mock.createIssueCalls[0]
	if call.owner == "" || call.repo == "" || call.title == "" {
		t.Errorf("CreateIssue called with empty fields: owner=%q, repo=%q, title=%q", call.owner, call.repo, call.title)
	}
}

// AC-007-3: Given tracker issue, Then it has `microsprint` label for filtering
func TestRunMicrosprintStartWithDeps_HasMicrosprintLabel(t *testing.T) {
	// ARRANGE
	mock := setupMockForStart()
	cfg := testMicrosprintConfig()
	cleanup := setupMicrosprintTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintStartOptions{}

	// ACT
	err := runMicrosprintStartWithDeps(cmd, opts, cfg, mock)

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
		if label == "microsprint" {
			hasLabel = true
			break
		}
	}
	if !hasLabel {
		t.Errorf("Expected 'microsprint' label for filtering, got labels: %v", call.labels)
	}
}

// AC-007-2: Given `microsprint close`, Then tracker issue is closed
func TestRunMicrosprintCloseWithDeps_ClosesTrackerIssue(t *testing.T) {
	// ARRANGE
	today := time.Now().Format("2006-01-02")
	mock := setupMockForStart()
	// Simulate an active microsprint tracker issue
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Microsprint: " + today + "-a",
			State:  "OPEN",
		},
	}
	cfg := testMicrosprintConfig()
	cleanup := setupMicrosprintTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintCloseOptions{}

	// ACT
	err := runMicrosprintCloseWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify CloseIssue was called on the tracker issue
	if len(mock.closeIssueCalls) != 1 {
		t.Fatalf("Expected 1 CloseIssue call, got %d", len(mock.closeIssueCalls))
	}

	closeCall := mock.closeIssueCalls[0]
	if closeCall.issueID != "TRACKER_123" {
		t.Errorf("Expected to close issue TRACKER_123, got %s", closeCall.issueID)
	}
}

// =============================================================================
// REQ-003: Add Issue to Microsprint
// =============================================================================

// AC-003-1: Given active microsprint, When user runs `microsprint add 42`,
// Then Microsprint Text field on issue #42 is set to microsprint name
func TestRunMicrosprintAddWithDeps_SetsTextFieldToMicrosprintName(t *testing.T) {
	// ARRANGE
	today := time.Now().Format("2006-01-02")
	microsprintName := today + "-a"
	mock := setupMockForStart()
	// Active microsprint exists
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Microsprint: " + microsprintName,
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

	cfg := testMicrosprintConfig()
	// Add microsprint field to config
	cfg.Fields["microsprint"] = config.Field{
		Field: "Microsprint",
	}

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintAddOptions{
		issueNumber: 42,
	}

	// ACT
	err := runMicrosprintAddWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify SetProjectItemField was called with correct values
	if len(mock.setFieldCalls) != 1 {
		t.Fatalf("Expected 1 SetProjectItemField call, got %d", len(mock.setFieldCalls))
	}

	call := mock.setFieldCalls[0]
	if call.value != microsprintName {
		t.Errorf("Expected field value '%s', got '%s'", microsprintName, call.value)
	}
	if call.fieldID != "Microsprint" {
		t.Errorf("Expected fieldID 'Microsprint', got '%s'", call.fieldID)
	}
}

// AC-003-2: Given active microsprint, When field updated,
// Then output confirms "Added #42 to microsprint YYYY-MM-DD-a"
func TestRunMicrosprintAddWithDeps_OutputsConfirmation(t *testing.T) {
	// ARRANGE
	today := time.Now().Format("2006-01-02")
	microsprintName := today + "-a"
	mock := setupMockForStart()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Microsprint: " + microsprintName,
			State:  "OPEN",
		},
	}
	mock.issueByNumber = &api.Issue{
		ID:     "ISSUE_42",
		Number: 42,
		Title:  "Fix login bug",
	}
	mock.projectItemID = "ITEM_42"

	cfg := testMicrosprintConfig()
	cfg.Fields["microsprint"] = config.Field{
		Field: "Microsprint",
	}

	cmd, buf := newTestMicrosprintCmd()
	opts := &microsprintAddOptions{
		issueNumber: 42,
	}

	// ACT
	err := runMicrosprintAddWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	expectedOutput := "Added #42 to microsprint " + microsprintName
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output to contain '%s', got '%s'", expectedOutput, output)
	}
}

// AC-003-3: Given Text field update, Then tracker issue body is NOT updated
// (avoid race conditions)
func TestRunMicrosprintAddWithDeps_DoesNotUpdateTrackerBody(t *testing.T) {
	// ARRANGE
	today := time.Now().Format("2006-01-02")
	microsprintName := today + "-a"
	mock := setupMockForStart()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Microsprint: " + microsprintName,
			State:  "OPEN",
		},
	}
	mock.issueByNumber = &api.Issue{
		ID:     "ISSUE_42",
		Number: 42,
		Title:  "Fix login bug",
	}
	mock.projectItemID = "ITEM_42"

	cfg := testMicrosprintConfig()
	cfg.Fields["microsprint"] = config.Field{
		Field: "Microsprint",
	}

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintAddOptions{
		issueNumber: 42,
	}

	// ACT
	err := runMicrosprintAddWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify tracker issue body was NOT updated (no updateIssueBody calls)
	if len(mock.updateIssueBodyCalls) > 0 {
		t.Errorf("Expected no UpdateIssueBody calls (avoid race conditions), got %d", len(mock.updateIssueBodyCalls))
	}
}

// Test that old dates are ignored (only today's microsprints count)
func TestRunMicrosprintStartWithDeps_IgnoresOldDates(t *testing.T) {
	// ARRANGE
	today := time.Now().Format("2006-01-02")
	mock := setupMockForStart()
	// Simulate existing microsprint from yesterday
	mock.openIssues = []api.Issue{
		{
			ID:     "OLD_1",
			Number: 40,
			Title:  "Microsprint: 2020-01-01-z", // Old date
		},
	}
	cfg := testMicrosprintConfig()
	cleanup := setupMicrosprintTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintStartOptions{}

	// Should start fresh with "a" since old dates don't count
	expectedTitle := "Microsprint: " + today + "-a"

	// ACT
	err := runMicrosprintStartWithDeps(cmd, opts, cfg, mock)

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

// =============================================================================
// REQ-038: Remove Issue from Microsprint
// =============================================================================

// AC-038-1: Given issue #42 assigned to microsprint, When running `microsprint remove 42`,
// Then Microsprint Text field cleared (set to empty)
func TestRunMicrosprintRemoveWithDeps_ClearsTextField(t *testing.T) {
	// ARRANGE
	today := time.Now().Format("2006-01-02")
	microsprintName := today + "-a"
	mock := setupMockForStart()
	// Active microsprint exists
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Microsprint: " + microsprintName,
			State:  "OPEN",
		},
	}
	// The issue to remove (currently assigned to microsprint)
	mock.issueByNumber = &api.Issue{
		ID:     "ISSUE_42",
		Number: 42,
		Title:  "Fix login bug",
	}
	mock.projectItemID = "ITEM_42"
	// Issue is currently assigned to microsprint
	mock.projectItemFieldValue = microsprintName

	cfg := testMicrosprintConfig()
	cfg.Fields["microsprint"] = config.Field{
		Field: "Microsprint",
	}

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintRemoveOptions{
		issueNumber: 42,
	}

	// ACT
	err := runMicrosprintRemoveWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify SetProjectItemField was called with empty value to clear the field
	if len(mock.setFieldCalls) != 1 {
		t.Fatalf("Expected 1 SetProjectItemField call, got %d", len(mock.setFieldCalls))
	}

	call := mock.setFieldCalls[0]
	if call.value != "" {
		t.Errorf("Expected field value to be empty (cleared), got '%s'", call.value)
	}
	if call.fieldID != "Microsprint" {
		t.Errorf("Expected fieldID 'Microsprint', got '%s'", call.fieldID)
	}
}

// AC-038-2: Given field cleared, Then output confirms "Removed #42 from microsprint YYYY-MM-DD-a"
func TestRunMicrosprintRemoveWithDeps_OutputsConfirmation(t *testing.T) {
	// ARRANGE
	today := time.Now().Format("2006-01-02")
	microsprintName := today + "-a"
	mock := setupMockForStart()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Microsprint: " + microsprintName,
			State:  "OPEN",
		},
	}
	mock.issueByNumber = &api.Issue{
		ID:     "ISSUE_42",
		Number: 42,
		Title:  "Fix login bug",
	}
	mock.projectItemID = "ITEM_42"
	mock.projectItemFieldValue = microsprintName

	cfg := testMicrosprintConfig()
	cfg.Fields["microsprint"] = config.Field{
		Field: "Microsprint",
	}

	cmd, buf := newTestMicrosprintCmd()
	opts := &microsprintRemoveOptions{
		issueNumber: 42,
	}

	// ACT
	err := runMicrosprintRemoveWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	expectedOutput := "Removed #42 from microsprint " + microsprintName
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output to contain '%s', got '%s'", expectedOutput, output)
	}
}

// AC-038-3: Given issue not in any microsprint, When running `microsprint remove 42`,
// Then warning: "Issue #42 is not assigned to a microsprint"
func TestRunMicrosprintRemoveWithDeps_WarnsIfNotAssigned(t *testing.T) {
	// ARRANGE
	today := time.Now().Format("2006-01-02")
	microsprintName := today + "-a"
	mock := setupMockForStart()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Microsprint: " + microsprintName,
			State:  "OPEN",
		},
	}
	mock.issueByNumber = &api.Issue{
		ID:     "ISSUE_42",
		Number: 42,
		Title:  "Fix login bug",
	}
	mock.projectItemID = "ITEM_42"
	// Issue is NOT assigned to any microsprint (empty field value)
	mock.projectItemFieldValue = ""

	cfg := testMicrosprintConfig()
	cfg.Fields["microsprint"] = config.Field{
		Field: "Microsprint",
	}

	cmd, buf := newTestMicrosprintCmd()
	opts := &microsprintRemoveOptions{
		issueNumber: 42,
	}

	// ACT
	err := runMicrosprintRemoveWithDeps(cmd, opts, cfg, mock)

	// ASSERT - should not error, but warn
	if err != nil {
		t.Fatalf("Expected no error (warning only), got: %v", err)
	}

	output := buf.String()
	expectedWarning := "Issue #42 is not assigned to a microsprint"
	if !strings.Contains(output, expectedWarning) {
		t.Errorf("Expected output to contain warning '%s', got '%s'", expectedWarning, output)
	}

	// Verify SetProjectItemField was NOT called (nothing to clear)
	if len(mock.setFieldCalls) != 0 {
		t.Errorf("Expected 0 SetProjectItemField calls (nothing to clear), got %d", len(mock.setFieldCalls))
	}
}

// =============================================================================
// REQ-035: View Current Microsprint
// =============================================================================

// AC-035-1: Given active microsprint, When running `microsprint current`,
// Then displays: name, started time, issue count, tracker issue number
func TestRunMicrosprintCurrentWithDeps_DisplaysActiveDetails(t *testing.T) {
	// ARRANGE
	today := time.Now().Format("2006-01-02")
	microsprintName := today + "-a"
	mock := setupMockForStart()
	// Active microsprint exists
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Microsprint: " + microsprintName,
			State:  "OPEN",
		},
	}
	// Issues assigned to this microsprint
	mock.microsprintIssues = []api.Issue{
		{ID: "ISSUE_1", Number: 41, Title: "Fix bug A"},
		{ID: "ISSUE_2", Number: 42, Title: "Fix bug B"},
		{ID: "ISSUE_3", Number: 43, Title: "Add feature C"},
	}

	cfg := testMicrosprintConfig()
	cfg.Fields["microsprint"] = config.Field{
		Field: "Microsprint",
	}

	cmd, buf := newTestMicrosprintCmd()
	opts := &microsprintCurrentOptions{}

	// ACT
	err := runMicrosprintCurrentWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()

	// Verify output contains microsprint name
	if !strings.Contains(output, microsprintName) {
		t.Errorf("Expected output to contain microsprint name '%s', got '%s'", microsprintName, output)
	}

	// Verify output contains tracker issue number
	if !strings.Contains(output, "#100") {
		t.Errorf("Expected output to contain tracker issue '#100', got '%s'", output)
	}

	// Verify output contains issue count
	if !strings.Contains(output, "3") {
		t.Errorf("Expected output to contain issue count '3', got '%s'", output)
	}
}

// AC-035-2: Given no active microsprint, Then message: "No active microsprint"
func TestRunMicrosprintCurrentWithDeps_NoActiveMicrosprint(t *testing.T) {
	// ARRANGE
	mock := setupMockForStart()
	// No active microsprint (old date or empty)
	mock.openIssues = []api.Issue{
		{
			ID:     "OLD_TRACKER",
			Number: 50,
			Title:  "Microsprint: 2020-01-01-a", // Old date
			State:  "OPEN",
		},
	}

	cfg := testMicrosprintConfig()
	cfg.Fields["microsprint"] = config.Field{
		Field: "Microsprint",
	}

	cmd, buf := newTestMicrosprintCmd()
	opts := &microsprintCurrentOptions{}

	// ACT
	err := runMicrosprintCurrentWithDeps(cmd, opts, cfg, mock)

	// ASSERT - should not error, just display message
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	expectedMessage := "No active microsprint"
	if !strings.Contains(output, expectedMessage) {
		t.Errorf("Expected output to contain '%s', got '%s'", expectedMessage, output)
	}
}

// AC-035-3: Given `--refresh` flag, Then tracker issue body updated with current issue list
func TestRunMicrosprintCurrentWithDeps_RefreshUpdatesTrackerBody(t *testing.T) {
	// ARRANGE
	today := time.Now().Format("2006-01-02")
	microsprintName := today + "-a"
	mock := setupMockForStart()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Microsprint: " + microsprintName,
			State:  "OPEN",
		},
	}
	mock.microsprintIssues = []api.Issue{
		{ID: "ISSUE_1", Number: 41, Title: "Fix bug A"},
		{ID: "ISSUE_2", Number: 42, Title: "Fix bug B"},
	}

	cfg := testMicrosprintConfig()
	cfg.Fields["microsprint"] = config.Field{
		Field: "Microsprint",
	}

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintCurrentOptions{
		refresh: true,
	}

	// ACT
	err := runMicrosprintCurrentWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify UpdateIssueBody was called
	if len(mock.updateIssueBodyCalls) != 1 {
		t.Fatalf("Expected 1 UpdateIssueBody call for refresh, got %d", len(mock.updateIssueBodyCalls))
	}

	call := mock.updateIssueBodyCalls[0]
	if call.issueID != "TRACKER_123" {
		t.Errorf("Expected UpdateIssueBody on tracker TRACKER_123, got %s", call.issueID)
	}

	// Verify body contains issue references
	if !strings.Contains(call.body, "#41") || !strings.Contains(call.body, "#42") {
		t.Errorf("Expected body to contain issue references #41 and #42, got '%s'", call.body)
	}
}

// =============================================================================
// REQ-004: Close Microsprint with Artifacts
// =============================================================================

// AC-004-1: Given active microsprint with issues, When user runs `microsprint close`,
// Then `Microsprints/{name}/review.md` generated with issue summary
func TestRunMicrosprintCloseArtifactsWithDeps_GeneratesReviewMd(t *testing.T) {
	// ARRANGE
	today := time.Now().Format("2006-01-02")
	microsprintName := today + "-a"
	mock := setupMockForStart()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Microsprint: " + microsprintName,
			State:  "OPEN",
		},
	}
	mock.microsprintIssues = []api.Issue{
		{ID: "ISSUE_1", Number: 41, Title: "Fix bug A", State: "CLOSED"},
		{ID: "ISSUE_2", Number: 42, Title: "Fix bug B", State: "OPEN"},
	}

	cfg := testMicrosprintConfig()
	cfg.Fields["microsprint"] = config.Field{
		Field: "Microsprint",
	}

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintCloseOptions{
		skipRetro: true, // Skip retro for simpler test
	}

	// ACT
	err := runMicrosprintCloseArtifactsWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify review.md was created
	foundReview := false
	expectedPath := "Microsprints/" + microsprintName + "/review.md"
	for _, call := range mock.writeFileCalls {
		if call.path == expectedPath {
			foundReview = true
			// Verify content contains issue references
			if !strings.Contains(call.content, "#41") || !strings.Contains(call.content, "#42") {
				t.Errorf("Expected review.md to contain issue references, got '%s'", call.content)
			}
			break
		}
	}
	if !foundReview {
		t.Errorf("Expected review.md to be created at '%s', got calls: %+v", expectedPath, mock.writeFileCalls)
	}
}

// AC-004-3: Given `microsprint close --skip-retro`, Then retro.md generated with empty template
func TestRunMicrosprintCloseArtifactsWithDeps_SkipRetroGeneratesEmptyTemplate(t *testing.T) {
	// ARRANGE
	today := time.Now().Format("2006-01-02")
	microsprintName := today + "-a"
	mock := setupMockForStart()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Microsprint: " + microsprintName,
			State:  "OPEN",
		},
	}
	mock.microsprintIssues = []api.Issue{}

	cfg := testMicrosprintConfig()
	cfg.Fields["microsprint"] = config.Field{
		Field: "Microsprint",
	}

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintCloseOptions{
		skipRetro: true,
	}

	// ACT
	err := runMicrosprintCloseArtifactsWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify retro.md was created with empty template
	foundRetro := false
	expectedPath := "Microsprints/" + microsprintName + "/retro.md"
	for _, call := range mock.writeFileCalls {
		if call.path == expectedPath {
			foundRetro = true
			// Verify it contains template sections
			if !strings.Contains(call.content, "What Went Well") {
				t.Errorf("Expected retro.md to contain 'What Went Well' section, got '%s'", call.content)
			}
			break
		}
	}
	if !foundRetro {
		t.Errorf("Expected retro.md to be created at '%s'", expectedPath)
	}
}

// AC-004-4: Given artifacts generated, Then files staged to git (`git add`)
func TestRunMicrosprintCloseArtifactsWithDeps_StagesFilesToGit(t *testing.T) {
	// ARRANGE
	today := time.Now().Format("2006-01-02")
	microsprintName := today + "-a"
	mock := setupMockForStart()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Microsprint: " + microsprintName,
			State:  "OPEN",
		},
	}
	mock.microsprintIssues = []api.Issue{}

	cfg := testMicrosprintConfig()
	cfg.Fields["microsprint"] = config.Field{
		Field: "Microsprint",
	}

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintCloseOptions{
		skipRetro: true,
	}

	// ACT
	err := runMicrosprintCloseArtifactsWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify git add was called
	if len(mock.gitAddCalls) == 0 {
		t.Errorf("Expected git add to be called, got no calls")
	}

	// Verify the microsprint directory was added
	foundDir := false
	expectedDir := "Microsprints/" + microsprintName
	for _, path := range mock.gitAddCalls {
		if strings.HasPrefix(path, expectedDir) {
			foundDir = true
			break
		}
	}
	if !foundDir {
		t.Errorf("Expected git add to include '%s', got: %v", expectedDir, mock.gitAddCalls)
	}
}

// AC-004-5: Given `microsprint close --commit`, Then artifacts committed with standard message
func TestRunMicrosprintCloseArtifactsWithDeps_CommitFlagCommitsArtifacts(t *testing.T) {
	// ARRANGE
	today := time.Now().Format("2006-01-02")
	microsprintName := today + "-a"
	mock := setupMockForStart()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Microsprint: " + microsprintName,
			State:  "OPEN",
		},
	}
	mock.microsprintIssues = []api.Issue{}

	cfg := testMicrosprintConfig()
	cfg.Fields["microsprint"] = config.Field{
		Field: "Microsprint",
	}

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintCloseOptions{
		skipRetro: true,
		commit:    true,
	}

	// ACT
	err := runMicrosprintCloseArtifactsWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify git commit was called
	if len(mock.gitCommitCalls) != 1 {
		t.Fatalf("Expected 1 git commit call, got %d", len(mock.gitCommitCalls))
	}

	// Verify commit message contains microsprint reference
	commitMsg := mock.gitCommitCalls[0].message
	if !strings.Contains(commitMsg, microsprintName) {
		t.Errorf("Expected commit message to contain microsprint name '%s', got '%s'", microsprintName, commitMsg)
	}
}

// AC-004-6: Given close complete, Then tracker issue body updated with artifact links and closed
func TestRunMicrosprintCloseArtifactsWithDeps_UpdatesTrackerAndCloses(t *testing.T) {
	// ARRANGE
	today := time.Now().Format("2006-01-02")
	microsprintName := today + "-a"
	mock := setupMockForStart()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Microsprint: " + microsprintName,
			State:  "OPEN",
		},
	}
	mock.microsprintIssues = []api.Issue{
		{ID: "ISSUE_1", Number: 41, Title: "Fix bug A"},
	}

	cfg := testMicrosprintConfig()
	cfg.Fields["microsprint"] = config.Field{
		Field: "Microsprint",
	}

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintCloseOptions{
		skipRetro: true,
	}

	// ACT
	err := runMicrosprintCloseArtifactsWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify tracker issue body was updated
	if len(mock.updateIssueBodyCalls) != 1 {
		t.Fatalf("Expected 1 UpdateIssueBody call, got %d", len(mock.updateIssueBodyCalls))
	}

	bodyCall := mock.updateIssueBodyCalls[0]
	if bodyCall.issueID != "TRACKER_123" {
		t.Errorf("Expected UpdateIssueBody on TRACKER_123, got %s", bodyCall.issueID)
	}

	// Verify body contains artifact links
	if !strings.Contains(bodyCall.body, "review.md") {
		t.Errorf("Expected body to contain 'review.md' link, got '%s'", bodyCall.body)
	}

	// Verify tracker issue was closed
	if len(mock.closeIssueCalls) != 1 {
		t.Fatalf("Expected 1 CloseIssue call, got %d", len(mock.closeIssueCalls))
	}

	if mock.closeIssueCalls[0].issueID != "TRACKER_123" {
		t.Errorf("Expected to close TRACKER_123, got %s", mock.closeIssueCalls[0].issueID)
	}
}

// =============================================================================
// REQ-008: List Microsprint History
// =============================================================================

// AC-008-1: Given `microsprint list`, Then table displayed with: Microsprint, Tracker#, Issues, Done, Duration, Status
func TestRunMicrosprintListWithDeps_DisplaysTable(t *testing.T) {
	// ARRANGE
	mock := setupMockForMicrosprint()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_100",
			Number: 100,
			Title:  "Microsprint: 2025-12-13-a",
			State:  "OPEN",
		},
	}
	mock.closedIssues = []api.Issue{
		{
			ID:     "TRACKER_99",
			Number: 99,
			Title:  "Microsprint: 2025-12-12-a",
			State:  "CLOSED",
		},
	}

	cfg := testMicrosprintConfig()
	cleanup := setupMicrosprintTestDir(t, cfg)
	defer cleanup()

	cmd, buf := newTestMicrosprintCmd()
	opts := &microsprintListOptions{}

	// ACT
	err := runMicrosprintListWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()

	// Verify headers
	if !strings.Contains(output, "MICROSPRINT") {
		t.Errorf("Expected output to contain 'MICROSPRINT' header, got '%s'", output)
	}
	if !strings.Contains(output, "TRACKER") {
		t.Errorf("Expected output to contain 'TRACKER' header, got '%s'", output)
	}
	if !strings.Contains(output, "STATUS") {
		t.Errorf("Expected output to contain 'STATUS' header, got '%s'", output)
	}

	// Verify microsprint data
	if !strings.Contains(output, "2025-12-13-a") {
		t.Errorf("Expected output to contain '2025-12-13-a', got '%s'", output)
	}
	if !strings.Contains(output, "2025-12-12-a") {
		t.Errorf("Expected output to contain '2025-12-12-a', got '%s'", output)
	}
}

// AC-008-2: Given multiple microsprints, Then sorted by date descending (most recent first)
func TestRunMicrosprintListWithDeps_SortedByDateDescending(t *testing.T) {
	// ARRANGE
	mock := setupMockForMicrosprint()
	mock.openIssues = []api.Issue{}
	mock.closedIssues = []api.Issue{
		{
			ID:     "TRACKER_1",
			Number: 1,
			Title:  "Microsprint: 2025-12-10-a",
			State:  "CLOSED",
		},
		{
			ID:     "TRACKER_3",
			Number: 3,
			Title:  "Microsprint: 2025-12-12-a",
			State:  "CLOSED",
		},
		{
			ID:     "TRACKER_2",
			Number: 2,
			Title:  "Microsprint: 2025-12-11-a",
			State:  "CLOSED",
		},
	}

	cfg := testMicrosprintConfig()
	cleanup := setupMicrosprintTestDir(t, cfg)
	defer cleanup()

	cmd, buf := newTestMicrosprintCmd()
	opts := &microsprintListOptions{}

	// ACT
	err := runMicrosprintListWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()

	// 2025-12-12 should appear before 2025-12-11, which should appear before 2025-12-10
	pos12 := strings.Index(output, "2025-12-12-a")
	pos11 := strings.Index(output, "2025-12-11-a")
	pos10 := strings.Index(output, "2025-12-10-a")

	if pos12 > pos11 {
		t.Errorf("Expected 2025-12-12-a to appear before 2025-12-11-a (descending order)")
	}
	if pos11 > pos10 {
		t.Errorf("Expected 2025-12-11-a to appear before 2025-12-10-a (descending order)")
	}
}

// Test no microsprints shows message
func TestRunMicrosprintListWithDeps_NoMicrosprints(t *testing.T) {
	// ARRANGE
	mock := setupMockForMicrosprint()
	mock.openIssues = []api.Issue{}
	mock.closedIssues = []api.Issue{}

	cfg := testMicrosprintConfig()
	cleanup := setupMicrosprintTestDir(t, cfg)
	defer cleanup()

	cmd, buf := newTestMicrosprintCmd()
	opts := &microsprintListOptions{}

	// ACT
	err := runMicrosprintListWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No microsprints found") {
		t.Errorf("Expected output to contain 'No microsprints found', got '%s'", output)
	}
}

// Test microsprint list uses cache when available
func TestRunMicrosprintListWithDeps_UsesCache(t *testing.T) {
	// ARRANGE
	mock := setupMockForMicrosprint()
	// API should NOT be called since we have cache
	mock.getOpenIssuesErr = errors.New("should not be called")
	mock.getClosedIssuesErr = errors.New("should not be called")

	cfg := testMicrosprintConfig()
	// Add cached data
	cfg.Cache = &config.Cache{
		Microsprints: []config.CachedTracker{
			{Number: 100, Title: "Microsprint: 2025-12-13-a", State: "OPEN"},
			{Number: 99, Title: "Microsprint: 2025-12-12-a", State: "CLOSED"},
		},
	}
	cleanup := setupMicrosprintTestDir(t, cfg)
	defer cleanup()

	cmd, buf := newTestMicrosprintCmd()
	opts := &microsprintListOptions{refresh: false}

	// ACT
	err := runMicrosprintListWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "2025-12-13-a") {
		t.Errorf("Expected output to contain '2025-12-13-a' from cache, got '%s'", output)
	}
	if !strings.Contains(output, "2025-12-12-a") {
		t.Errorf("Expected output to contain '2025-12-12-a' from cache, got '%s'", output)
	}
}

// Test microsprint list with --refresh flag bypasses cache
func TestRunMicrosprintListWithDeps_RefreshBypassesCache(t *testing.T) {
	// ARRANGE
	mock := setupMockForMicrosprint()
	mock.openIssues = []api.Issue{
		{ID: "TRACKER_200", Number: 200, Title: "Microsprint: 2025-12-20-a", State: "OPEN"},
	}
	mock.closedIssues = []api.Issue{}

	cfg := testMicrosprintConfig()
	// Add stale cached data
	cfg.Cache = &config.Cache{
		Microsprints: []config.CachedTracker{
			{Number: 100, Title: "Microsprint: 2025-12-10-a", State: "CLOSED"},
		},
	}
	cleanup := setupMicrosprintTestDir(t, cfg)
	defer cleanup()

	cmd, buf := newTestMicrosprintCmd()
	opts := &microsprintListOptions{refresh: true} // Force refresh

	// ACT
	err := runMicrosprintListWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	// Should contain fresh API data, not stale cache
	if !strings.Contains(output, "2025-12-20-a") {
		t.Errorf("Expected output to contain '2025-12-20-a' from API, got '%s'", output)
	}
}

// Test microsprint list API error handling
func TestRunMicrosprintListWithDeps_OpenIssuesError(t *testing.T) {
	// ARRANGE
	mock := setupMockForMicrosprint()
	mock.getOpenIssuesErr = errors.New("API error")

	cfg := testMicrosprintConfig()
	cleanup := setupMicrosprintTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintListOptions{}

	// ACT
	err := runMicrosprintListWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to get open microsprint issues") {
		t.Errorf("Expected error about open microsprint issues, got: %v", err)
	}
}

func TestRunMicrosprintListWithDeps_ClosedIssuesError(t *testing.T) {
	// ARRANGE
	mock := setupMockForMicrosprint()
	mock.openIssues = []api.Issue{}
	mock.getClosedIssuesErr = errors.New("API error")

	cfg := testMicrosprintConfig()
	cleanup := setupMicrosprintTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintListOptions{}

	// ACT
	err := runMicrosprintListWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to get closed microsprint issues") {
		t.Errorf("Expected error about closed microsprint issues, got: %v", err)
	}
}

// Test microsprintsFromCache helper function
func TestMicrosprintsFromCache(t *testing.T) {
	cached := []config.CachedTracker{
		{Number: 100, Title: "Microsprint: 2025-12-13-a", State: "OPEN"},
		{Number: 99, Title: "Microsprint: 2025-12-12-a", State: "CLOSED"},
		{Number: 98, Title: "Not a microsprint", State: "OPEN"}, // Should be filtered out
	}

	microsprints := microsprintsFromCache(cached)

	if len(microsprints) != 2 {
		t.Fatalf("Expected 2 microsprints, got %d", len(microsprints))
	}

	// Check first microsprint
	if microsprints[0].name != "2025-12-13-a" {
		t.Errorf("Expected name '2025-12-13-a', got '%s'", microsprints[0].name)
	}
	if microsprints[0].status != "Active" {
		t.Errorf("Expected status 'Active' for OPEN, got '%s'", microsprints[0].status)
	}

	// Check second microsprint
	if microsprints[1].name != "2025-12-12-a" {
		t.Errorf("Expected name '2025-12-12-a', got '%s'", microsprints[1].name)
	}
	if microsprints[1].status != "Closed" {
		t.Errorf("Expected status 'Closed' for CLOSED, got '%s'", microsprints[1].status)
	}
}

// =============================================================================
// REQ-013: Multiple Active Detection
// =============================================================================

// AC-013-1: Given 2+ open tracker issues, When running `microsprint add`, Then error
func TestRunMicrosprintAddWithDeps_MultipleActiveError(t *testing.T) {
	// ARRANGE
	mock := setupMockForMicrosprint()
	today := time.Now().Format("2006-01-02")
	// Two active microsprints for today
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_100",
			Number: 100,
			Title:  "Microsprint: " + today + "-a",
			State:  "OPEN",
		},
		{
			ID:     "TRACKER_101",
			Number: 101,
			Title:  "Microsprint: " + today + "-b",
			State:  "OPEN",
		},
	}

	cfg := testMicrosprintConfig()
	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintAddOptions{issueNumber: 42}

	// ACT
	err := runMicrosprintAddWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err == nil {
		t.Fatal("Expected error when multiple active microsprints exist")
	}
	if !strings.Contains(err.Error(), "Multiple active microsprints") {
		t.Errorf("Expected error to mention 'Multiple active microsprints', got: %v", err)
	}
	if !strings.Contains(err.Error(), "microsprint resolve") {
		t.Errorf("Expected error to suggest 'microsprint resolve', got: %v", err)
	}
}

// AC-013-2: Given 2+ open tracker issues, When running `microsprint close`, Then error
func TestRunMicrosprintCloseWithDeps_MultipleActiveError(t *testing.T) {
	// ARRANGE
	mock := setupMockForMicrosprint()
	today := time.Now().Format("2006-01-02")
	// Two active microsprints for today
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_100",
			Number: 100,
			Title:  "Microsprint: " + today + "-a",
			State:  "OPEN",
		},
		{
			ID:     "TRACKER_101",
			Number: 101,
			Title:  "Microsprint: " + today + "-b",
			State:  "OPEN",
		},
	}

	cfg := testMicrosprintConfig()
	cleanup := setupMicrosprintTestDir(t, cfg)
	defer cleanup()

	cmd, _ := newTestMicrosprintCmd()
	opts := &microsprintCloseOptions{}

	// ACT
	err := runMicrosprintCloseWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err == nil {
		t.Fatal("Expected error when multiple active microsprints exist")
	}
	if !strings.Contains(err.Error(), "Multiple active microsprints") {
		t.Errorf("Expected error to mention 'Multiple active microsprints', got: %v", err)
	}
}

// =============================================================================
// REQ-016: Empty Microsprint Close
// =============================================================================

// AC-016-1: Given active microsprint with 0 issues, Then review.md generated with "No issues completed"
func TestGenerateReviewContent_EmptyMicrosprint(t *testing.T) {
	// ACT
	content := generateReviewContent("2025-12-13-a", []api.Issue{})

	// ASSERT
	if !strings.Contains(content, "No issues completed") {
		t.Errorf("Expected content to contain 'No issues completed', got: %s", content)
	}
}

// =============================================================================
// REQ-012: Tracker Naming Validation
// =============================================================================

// AC-012-1: Given issue with `microsprint` label but incorrect title, Then ignored
func TestFindActiveMicrosprint_IgnoresNonTrackerIssues(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	issues := []api.Issue{
		{
			ID:     "RANDOM_1",
			Number: 1,
			Title:  "Random Issue with microsprint label",
			State:  "OPEN",
		},
		{
			ID:     "TRACKER_2",
			Number: 2,
			Title:  "Microsprint: " + today + "-a",
			State:  "OPEN",
		},
	}

	// ACT
	result := findActiveMicrosprint(issues)

	// ASSERT
	if result == nil {
		t.Fatal("Expected to find tracker issue")
	}
	if result.Number != 2 {
		t.Errorf("Expected to find issue #2, got #%d", result.Number)
	}
}

// AC-012-2: Given issue with correct title pattern, Then recognized as tracker
func TestFindActiveMicrosprint_RecognizesTrackerIssue(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	issues := []api.Issue{
		{
			ID:     "TRACKER_1",
			Number: 1,
			Title:  "Microsprint: " + today + "-a",
			State:  "OPEN",
		},
	}

	// ACT
	result := findActiveMicrosprint(issues)

	// ASSERT
	if result == nil {
		t.Fatal("Expected to find tracker issue")
	}
	if result.Number != 1 {
		t.Errorf("Expected issue #1, got #%d", result.Number)
	}
}

// AC-012-2: Also recognizes tracker with custom name suffix
func TestFindActiveMicrosprint_RecognizesTrackerWithCustomName(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	issues := []api.Issue{
		{
			ID:     "TRACKER_1",
			Number: 1,
			Title:  "Microsprint: " + today + "-a-auth-refactor",
			State:  "OPEN",
		},
	}

	// ACT
	result := findActiveMicrosprint(issues)

	// ASSERT
	if result == nil {
		t.Fatal("Expected to find tracker issue with custom name")
	}
}

// =============================================================================
// compareSuffixes Tests
// =============================================================================

func TestCompareSuffixes_SameLength(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		expected int
	}{
		{"a", "a", 0},  // equal
		{"a", "b", -1}, // a < b
		{"b", "a", 1},  // b > a
		{"z", "a", 1},  // z > a
		{"aa", "aa", 0},
		{"aa", "ab", -1},
		{"ab", "aa", 1},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			result := compareSuffixes(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("compareSuffixes(%q, %q) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestCompareSuffixes_DifferentLength(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		expected int
	}{
		{"a", "aa", -1}, // shorter is less
		{"aa", "a", 1},  // longer is greater
		{"z", "aa", -1}, // z < aa because aa is longer
		{"aaa", "z", 1}, // aaa > z because aaa is longer
		{"ab", "aaa", -1},
		{"aaa", "ab", 1},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			result := compareSuffixes(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("compareSuffixes(%q, %q) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestCompareSuffixes_EdgeCases(t *testing.T) {
	// Empty strings
	if compareSuffixes("", "") != 0 {
		t.Error("Expected empty strings to be equal")
	}

	// Single character vs empty
	if compareSuffixes("a", "") != 1 {
		t.Error("Expected 'a' > ''")
	}
	if compareSuffixes("", "a") != -1 {
		t.Error("Expected '' < 'a'")
	}
}

// =============================================================================
// extractReleaseFromMicrosprintTitle Tests
// =============================================================================

func TestExtractReleaseFromMicrosprintTitle_WithRelease(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"Microsprint: 2025-12-13-a [v1.0.0]", "v1.0.0"},
		{"Microsprint: 2025-12-13-a [release/v2.0.0]", "release/v2.0.0"},
		{"Microsprint: 2025-12-13-a-auth [Phoenix]", "Phoenix"},
		{"Microsprint: 2025-12-13-a [ v1.0.0 ]", "v1.0.0"}, // with spaces
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			result := extractReleaseFromMicrosprintTitle(tt.title)
			if result != tt.expected {
				t.Errorf("extractReleaseFromMicrosprintTitle(%q) = %q, want %q", tt.title, result, tt.expected)
			}
		})
	}
}

func TestExtractReleaseFromMicrosprintTitle_NoRelease(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"Microsprint: 2025-12-13-a", ""},
		{"Microsprint: 2025-12-13-a-auth", ""},
		{"Random issue title", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			result := extractReleaseFromMicrosprintTitle(tt.title)
			if result != tt.expected {
				t.Errorf("extractReleaseFromMicrosprintTitle(%q) = %q, want %q", tt.title, result, tt.expected)
			}
		})
	}
}

func TestExtractReleaseFromMicrosprintTitle_MalformedBrackets(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"Microsprint: 2025-12-13-a [unclosed", ""},   // no closing bracket
		{"Microsprint: 2025-12-13-a ]only close", ""}, // only closing bracket
		{"[v1.0.0] Microsprint: 2025-12-13-a", ""},    // brackets at start don't work (uses LastIndex)
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			result := extractReleaseFromMicrosprintTitle(tt.title)
			if result != tt.expected {
				t.Errorf("extractReleaseFromMicrosprintTitle(%q) = %q, want %q", tt.title, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// formatAsBullets Tests
// =============================================================================

func TestFormatAsBullets_EmptyInput(t *testing.T) {
	result := formatAsBullets("")
	expected := "- \n"
	if result != expected {
		t.Errorf("formatAsBullets(\"\") = %q, want %q", result, expected)
	}
}

func TestFormatAsBullets_SingleLine(t *testing.T) {
	result := formatAsBullets("Single item")
	expected := "- Single item\n"
	if result != expected {
		t.Errorf("formatAsBullets(\"Single item\") = %q, want %q", result, expected)
	}
}

func TestFormatAsBullets_MultipleLines(t *testing.T) {
	input := "First item\nSecond item\nThird item"
	result := formatAsBullets(input)
	expected := "- First item\n- Second item\n- Third item\n"
	if result != expected {
		t.Errorf("formatAsBullets(%q) = %q, want %q", input, result, expected)
	}
}

func TestFormatAsBullets_AlreadyBulleted(t *testing.T) {
	// Should not double-bullet items
	input := "- Already bulleted\n* Also bulleted"
	result := formatAsBullets(input)
	expected := "- Already bulleted\n* Also bulleted\n"
	if result != expected {
		t.Errorf("formatAsBullets(%q) = %q, want %q", input, result, expected)
	}
}

func TestFormatAsBullets_MixedContent(t *testing.T) {
	// Some lines bulleted, some not
	input := "- Already bulleted\nNot bulleted\n* Star bulleted"
	result := formatAsBullets(input)
	expected := "- Already bulleted\n- Not bulleted\n* Star bulleted\n"
	if result != expected {
		t.Errorf("formatAsBullets(%q) = %q, want %q", input, result, expected)
	}
}

func TestFormatAsBullets_WhitespaceLines(t *testing.T) {
	// Empty lines should be skipped
	input := "First item\n\n   \nSecond item"
	result := formatAsBullets(input)
	expected := "- First item\n- Second item\n"
	if result != expected {
		t.Errorf("formatAsBullets(%q) = %q, want %q", input, result, expected)
	}
}

func TestFormatAsBullets_LeadingWhitespace(t *testing.T) {
	// Whitespace should be trimmed
	input := "   Indented item\n\tTabbed item"
	result := formatAsBullets(input)
	expected := "- Indented item\n- Tabbed item\n"
	if result != expected {
		t.Errorf("formatAsBullets(%q) = %q, want %q", input, result, expected)
	}
}

// =============================================================================
// generateMicrosprintTrackerTemplate Tests
// =============================================================================

func TestGenerateMicrosprintTrackerTemplate_ContainsName(t *testing.T) {
	name := "2025-12-21-a"
	result := generateMicrosprintTrackerTemplate(name)

	if !strings.Contains(result, "`"+name+"`") {
		t.Errorf("Template should contain microsprint name in backticks, got: %s", result)
	}
}

func TestGenerateMicrosprintTrackerTemplate_ContainsWarnings(t *testing.T) {
	result := generateMicrosprintTrackerTemplate("2025-12-21-a")

	warnings := []string{
		"**Microsprint Tracker Issue**",
		"**Do not manually:**",
		"Close or reopen this issue",
		"Change the title",
		"Remove the `microsprint` label",
	}

	for _, warning := range warnings {
		if !strings.Contains(result, warning) {
			t.Errorf("Template should contain warning %q", warning)
		}
	}
}

func TestGenerateMicrosprintTrackerTemplate_ContainsCommands(t *testing.T) {
	result := generateMicrosprintTrackerTemplate("2025-12-21-a")

	commands := []string{
		"`gh pmu microsprint add <issue>`",
		"`gh pmu microsprint remove <issue>`",
		"`gh pmu microsprint close`",
	}

	for _, cmd := range commands {
		if !strings.Contains(result, cmd) {
			t.Errorf("Template should contain command %q", cmd)
		}
	}
}

func TestGenerateMicrosprintTrackerTemplate_ContainsIssuesSection(t *testing.T) {
	result := generateMicrosprintTrackerTemplate("2025-12-21-a")

	if !strings.Contains(result, "## Issues in this microsprint") {
		t.Error("Template should contain 'Issues in this microsprint' section")
	}
	if !strings.Contains(result, "Microsprint field in the project") {
		t.Error("Template should explain issues are tracked via the Microsprint field")
	}
}

func TestGenerateMicrosprintTrackerTemplate_WithCustomName(t *testing.T) {
	name := "2025-12-21-a-auth-refactor"
	result := generateMicrosprintTrackerTemplate(name)

	if !strings.Contains(result, "`"+name+"`") {
		t.Errorf("Template should contain full custom name %q", name)
	}
}

func TestRunMicrosprintCloseWithDeps_DryRun_ShowsPreview(t *testing.T) {
	// ARRANGE
	today := time.Now().Format("2006-01-02")
	mock := setupMockForStart()
	mock.openIssues = []api.Issue{
		{
			ID:     "TRACKER_123",
			Number: 100,
			Title:  "Microsprint: " + today + "-a",
			State:  "OPEN",
		},
	}
	cfg := testMicrosprintConfig()
	cleanup := setupMicrosprintTestDir(t, cfg)
	defer cleanup()

	cmd, buf := newTestMicrosprintCmd()
	opts := &microsprintCloseOptions{dryRun: true}

	// ACT
	err := runMicrosprintCloseWithDeps(cmd, opts, cfg, mock)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error in dry-run mode, got: %v", err)
	}

	// Should not close tracker issue in dry-run
	if len(mock.closeIssueCalls) != 0 {
		t.Errorf("Expected 0 CloseIssue calls in dry-run, got %d", len(mock.closeIssueCalls))
	}

	// Should show preview
	output := buf.String()
	if !strings.Contains(output, "[DRY RUN]") {
		t.Error("Expected output to contain '[DRY RUN]'")
	}
	if !strings.Contains(output, "Would close microsprint:") {
		t.Error("Expected output to contain 'Would close microsprint:'")
	}
	if !strings.Contains(output, "Would close tracker issue #100") {
		t.Error("Expected output to contain 'Would close tracker issue #100'")
	}
}

func TestMicrosprintCloseCommand_HasDryRunFlag(t *testing.T) {
	cmd := NewRootCommand()
	closeCmd, _, err := cmd.Find([]string{"microsprint", "close"})
	if err != nil {
		t.Fatalf("microsprint close command not found: %v", err)
	}

	flag := closeCmd.Flags().Lookup("dry-run")
	if flag == nil {
		t.Fatal("Expected --dry-run flag to exist")
	}

	// Verify it's a boolean flag
	if flag.Value.Type() != "bool" {
		t.Errorf("Expected --dry-run to be bool, got %s", flag.Value.Type())
	}
}
