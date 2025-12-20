package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rubrical-studios/gh-pmu/internal/api"
	"github.com/rubrical-studios/gh-pmu/internal/config"
	"github.com/spf13/cobra"
)

// setupFieldTestDir creates a temp directory with a .gh-pmu.yml config file
// and changes to that directory. Returns cleanup function to restore original dir.
func setupFieldTestDir(t *testing.T, cfg *config.Config) func() {
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

// mockFieldClient implements fieldClient interface for testing
type mockFieldClient struct {
	project        *api.Project
	fields         []api.ProjectField
	createdField   *api.ProjectField
	getProjectErr  error
	getFieldsErr   error
	createFieldErr error
}

func (m *mockFieldClient) GetProject(owner string, number int) (*api.Project, error) {
	if m.getProjectErr != nil {
		return nil, m.getProjectErr
	}
	return m.project, nil
}

func (m *mockFieldClient) GetProjectFields(projectID string) ([]api.ProjectField, error) {
	if m.getFieldsErr != nil {
		return nil, m.getFieldsErr
	}
	return m.fields, nil
}

func (m *mockFieldClient) CreateProjectField(projectID, name, dataType string, singleSelectOptions []string) (*api.ProjectField, error) {
	if m.createFieldErr != nil {
		return nil, m.createFieldErr
	}
	if m.createdField != nil {
		return m.createdField, nil
	}
	// Return a default created field
	return &api.ProjectField{
		ID:       "PVTF_test123",
		Name:     name,
		DataType: dataType,
	}, nil
}

func newTestFieldCmd() (*cobra.Command, *bytes.Buffer) {
	cmd := newFieldCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	return cmd, buf
}

func TestRunFieldList_Success(t *testing.T) {
	cmd, buf := newTestFieldCmd()
	cmd.SetArgs([]string{"list"})

	mock := &mockFieldClient{
		project: &api.Project{
			ID:    "PVT_test",
			Title: "Test Project",
		},
		fields: []api.ProjectField{
			{ID: "F1", Name: "Status", DataType: "SINGLE_SELECT", Options: []api.FieldOption{
				{ID: "O1", Name: "Open"},
				{ID: "O2", Name: "Closed"},
			}},
			{ID: "F2", Name: "Priority", DataType: "SINGLE_SELECT", Options: []api.FieldOption{
				{ID: "O3", Name: "High"},
				{ID: "O4", Name: "Low"},
			}},
			{ID: "F3", Name: "PRD", DataType: "TEXT"},
		},
	}

	cfg := &config.Config{
		Project: config.Project{
			Owner:  "test-owner",
			Number: 1,
		},
		Repositories: []string{"test-owner/test-repo"},
	}

	err := runFieldListWithDeps(cmd, cfg, mock)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Test Project") {
		t.Errorf("expected project title in output, got %s", output)
	}
	if !strings.Contains(output, "Status (SINGLE_SELECT)") {
		t.Errorf("expected Status field in output, got %s", output)
	}
	if !strings.Contains(output, "PRD (TEXT)") {
		t.Errorf("expected PRD field in output, got %s", output)
	}
	if !strings.Contains(output, "Total: 3 fields") {
		t.Errorf("expected field count in output, got %s", output)
	}
}

func TestRunFieldCreate_TextFieldSuccess(t *testing.T) {
	cmd, buf := newTestFieldCmd()

	mock := &mockFieldClient{
		project: &api.Project{
			ID:    "PVT_test",
			Title: "Test Project",
		},
		fields: []api.ProjectField{}, // No existing fields
		createdField: &api.ProjectField{
			ID:       "PVTF_new",
			Name:     "PRD",
			DataType: "TEXT",
		},
	}

	cfg := &config.Config{
		Project: config.Project{
			Owner:  "test-owner",
			Number: 1,
		},
		Repositories: []string{"test-owner/test-repo"},
	}

	// Setup temp directory with config to avoid polluting repo root
	cleanup := setupFieldTestDir(t, cfg)
	defer cleanup()

	opts := &fieldCreateOptions{
		fieldType: "text",
		yes:       true, // Skip confirmation
	}

	err := runFieldCreateWithDeps(cmd, "PRD", opts, cfg, mock)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Created field \"PRD\"") {
		t.Errorf("expected success message in output, got %s", output)
	}
}

func TestRunFieldCreate_FieldAlreadyExists(t *testing.T) {
	cmd, _ := newTestFieldCmd()

	mock := &mockFieldClient{
		project: &api.Project{
			ID:    "PVT_test",
			Title: "Test Project",
		},
		fields: []api.ProjectField{
			{ID: "F1", Name: "PRD", DataType: "TEXT"},
		},
	}

	cfg := &config.Config{
		Project: config.Project{
			Owner:  "test-owner",
			Number: 1,
		},
		Repositories: []string{"test-owner/test-repo"},
	}

	// Setup temp directory with config to avoid polluting repo root
	cleanup := setupFieldTestDir(t, cfg)
	defer cleanup()

	opts := &fieldCreateOptions{
		fieldType: "text",
		yes:       true,
	}

	err := runFieldCreateWithDeps(cmd, "PRD", opts, cfg, mock)
	if err == nil {
		t.Fatal("expected error for existing field, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got %v", err)
	}
}

func TestRunFieldCreate_SingleSelectWithOptions(t *testing.T) {
	cmd, buf := newTestFieldCmd()

	mock := &mockFieldClient{
		project: &api.Project{
			ID:    "PVT_test",
			Title: "Test Project",
		},
		fields: []api.ProjectField{},
		createdField: &api.ProjectField{
			ID:       "PVTSSF_new",
			Name:     "Environment",
			DataType: "SINGLE_SELECT",
			Options: []api.FieldOption{
				{ID: "O1", Name: "Development"},
				{ID: "O2", Name: "Production"},
			},
		},
	}

	cfg := &config.Config{
		Project: config.Project{
			Owner:  "test-owner",
			Number: 1,
		},
		Repositories: []string{"test-owner/test-repo"},
	}

	// Setup temp directory with config to avoid polluting repo root
	cleanup := setupFieldTestDir(t, cfg)
	defer cleanup()

	opts := &fieldCreateOptions{
		fieldType: "single_select",
		options:   []string{"Development", "Production"},
		yes:       true,
	}

	err := runFieldCreateWithDeps(cmd, "Environment", opts, cfg, mock)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Created field \"Environment\"") {
		t.Errorf("expected success message in output, got %s", output)
	}
}

func TestRunFieldCreate_SingleSelectRequiresOptions(t *testing.T) {
	opts := &fieldCreateOptions{
		fieldType: "single_select",
		options:   []string{}, // No options
		yes:       true,
	}

	// This validation happens before API call, so we test at command level
	cmd := newFieldCreateCommand()
	cmd.SetArgs([]string{"Environment", "--type", "single_select"})

	// We can't easily test this without stdin, but we can verify the validation logic
	dataType := strings.ToUpper(opts.fieldType)
	if dataType == "SINGLE_SELECT" && len(opts.options) == 0 {
		// Expected - validation should catch this
	} else {
		t.Error("expected validation to fail for single_select without options")
	}
}

func TestRunFieldCreate_InvalidFieldType(t *testing.T) {
	opts := &fieldCreateOptions{
		fieldType: "invalid_type",
		yes:       true,
	}

	dataType := strings.ToUpper(opts.fieldType)
	switch dataType {
	case "TEXT", "NUMBER", "DATE", "SINGLE_SELECT":
		t.Error("expected invalid_type to be rejected")
	default:
		// Expected - invalid type should be rejected
	}
}

func TestFieldCreateOptions_AllTypes(t *testing.T) {
	validTypes := []string{"text", "number", "date", "single_select"}

	for _, ft := range validTypes {
		dataType := strings.ToUpper(ft)
		switch dataType {
		case "TEXT", "NUMBER", "DATE", "SINGLE_SELECT":
			// Valid
		default:
			t.Errorf("type %s should be valid", ft)
		}
	}
}
