package cmd

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func TestInitCommand_Exists(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"init", "--help"})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("init command should exist: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("init")) {
		t.Error("Expected help output to mention 'init'")
	}
}

func TestDetectRepository_FromGitRemote(t *testing.T) {
	// Test with a known git remote URL
	tests := []struct {
		name     string
		remote   string
		expected string
	}{
		{
			name:     "HTTPS URL",
			remote:   "https://github.com/owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "HTTPS URL without .git",
			remote:   "https://github.com/owner/repo",
			expected: "owner/repo",
		},
		{
			name:     "SSH URL",
			remote:   "git@github.com:owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "SSH URL without .git",
			remote:   "git@github.com:owner/repo",
			expected: "owner/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseGitRemote(tt.remote)
			if result != tt.expected {
				t.Errorf("parseGitRemote(%q) = %q, want %q", tt.remote, result, tt.expected)
			}
		})
	}
}

func TestDetectRepository_InvalidRemote(t *testing.T) {
	tests := []string{
		"",
		"not-a-url",
		"https://gitlab.com/owner/repo",
	}

	for _, remote := range tests {
		t.Run(remote, func(t *testing.T) {
			result := parseGitRemote(remote)
			if result != "" {
				t.Errorf("parseGitRemote(%q) = %q, want empty string", remote, result)
			}
		})
	}
}

func TestWriteConfig_CreatesValidYAML(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	cfg := &InitConfig{
		ProjectOwner:  "test-owner",
		ProjectNumber: 5,
		Repositories:  []string{"test-owner/test-repo"},
	}

	err := writeConfig(tmpDir, cfg)
	if err != nil {
		t.Fatalf("writeConfig failed: %v", err)
	}

	// Verify file was created
	configPath := tmpDir + "/.gh-pmu.yml"
	content, err := readFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	// Check content contains expected values
	if !bytes.Contains(content, []byte("owner: test-owner")) {
		t.Error("Config should contain owner")
	}
	if !bytes.Contains(content, []byte("number: 5")) {
		t.Error("Config should contain project number")
	}
	if !bytes.Contains(content, []byte("test-owner/test-repo")) {
		t.Error("Config should contain repository")
	}
}

func TestWriteConfig_WithDefaults(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &InitConfig{
		ProjectOwner:  "owner",
		ProjectNumber: 1,
		Repositories:  []string{"owner/repo"},
	}

	err := writeConfig(tmpDir, cfg)
	if err != nil {
		t.Fatalf("writeConfig failed: %v", err)
	}

	content, _ := readFile(tmpDir + "/.gh-pmu.yml")

	// Should have default status field mapping
	if !bytes.Contains(content, []byte("status:")) {
		t.Error("Config should have default status field")
	}
}

// Helper to read file for tests
func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func TestValidateProject_Success(t *testing.T) {
	// Mock client that returns a valid project
	mockClient := &MockAPIClient{
		project: &MockProject{
			ID:    "PVT_test123",
			Title: "Test Project",
		},
	}

	err := validateProject(mockClient, "owner", 1)
	if err != nil {
		t.Errorf("validateProject should succeed for valid project: %v", err)
	}
}

func TestValidateProject_NotFound(t *testing.T) {
	// Mock client that returns not found error
	mockClient := &MockAPIClient{
		err: ErrProjectNotFound,
	}

	err := validateProject(mockClient, "owner", 999)
	if err == nil {
		t.Error("validateProject should fail for non-existent project")
	}
}

// MockProject represents a mock project for testing
type MockProject struct {
	ID    string
	Title string
}

// MockAPIClient is a mock implementation for testing
type MockAPIClient struct {
	project *MockProject
	err     error
}

// GetProject implements ProjectValidator interface
func (m *MockAPIClient) GetProject(owner string, number int) (interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.project, nil
}

// ErrProjectNotFound is returned when project doesn't exist
var ErrProjectNotFound = fmt.Errorf("project not found")

func TestWriteConfigWithMetadata_IncludesFields(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &InitConfig{
		ProjectOwner:  "owner",
		ProjectNumber: 1,
		Repositories:  []string{"owner/repo"},
	}

	metadata := &ProjectMetadata{
		ProjectID: "PVT_test123",
		Fields: []FieldMetadata{
			{
				ID:       "PVTF_status",
				Name:     "Status",
				DataType: "SINGLE_SELECT",
				Options: []OptionMetadata{
					{ID: "opt1", Name: "Backlog"},
					{ID: "opt2", Name: "Done"},
				},
			},
			{
				ID:       "PVTF_priority",
				Name:     "Priority",
				DataType: "SINGLE_SELECT",
				Options: []OptionMetadata{
					{ID: "opt3", Name: "High"},
					{ID: "opt4", Name: "Low"},
				},
			},
		},
	}

	err := writeConfigWithMetadata(tmpDir, cfg, metadata)
	if err != nil {
		t.Fatalf("writeConfigWithMetadata failed: %v", err)
	}

	content, _ := readFile(tmpDir + "/.gh-pmu.yml")

	// Should contain metadata section with project ID
	if !bytes.Contains(content, []byte("metadata:")) {
		t.Error("Config should have metadata section")
	}
	if !bytes.Contains(content, []byte("PVT_test123")) {
		t.Error("Config should contain project ID")
	}
	// Should contain field IDs
	if !bytes.Contains(content, []byte("PVTF_status")) {
		t.Error("Config should contain field IDs")
	}
}
