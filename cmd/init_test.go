package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func TestWriteConfig_IncludesTriageAndLabels(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &InitConfig{
		ProjectName:   "Test Project",
		ProjectOwner:  "owner",
		ProjectNumber: 1,
		Repositories:  []string{"owner/repo"},
	}

	err := writeConfig(tmpDir, cfg)
	if err != nil {
		t.Fatalf("writeConfig failed: %v", err)
	}

	content, _ := readFile(tmpDir + "/.gh-pmu.yml")

	// Should have project name
	if !bytes.Contains(content, []byte("name: Test Project")) {
		t.Error("Config should have project name")
	}

	// Should have triage section
	if !bytes.Contains(content, []byte("triage:")) {
		t.Error("Config should have triage section")
	}

	// Should have estimate triage rule
	if !bytes.Contains(content, []byte("estimate:")) {
		t.Error("Config should have estimate triage rule")
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

	err := writeConfigWithMetadata(tmpDir, cfg, metadata, nil)
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

func TestSplitRepository(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedOwner string
		expectedName  string
	}{
		{
			name:          "valid owner/repo format",
			input:         "rubrical-studios/gh-pmu",
			expectedOwner: "rubrical-studios",
			expectedName:  "gh-pmu",
		},
		{
			name:          "simple owner/repo",
			input:         "owner/repo",
			expectedOwner: "owner",
			expectedName:  "repo",
		},
		{
			name:          "no slash - invalid input",
			input:         "noslash",
			expectedOwner: "",
			expectedName:  "",
		},
		{
			name:          "empty string",
			input:         "",
			expectedOwner: "",
			expectedName:  "",
		},
		{
			name:          "multiple slashes - takes first split",
			input:         "owner/repo/extra",
			expectedOwner: "owner",
			expectedName:  "repo/extra",
		},
		{
			name:          "only slash",
			input:         "/",
			expectedOwner: "",
			expectedName:  "",
		},
		{
			name:          "owner with trailing slash",
			input:         "owner/",
			expectedOwner: "owner",
			expectedName:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, name := splitRepository(tt.input)
			if owner != tt.expectedOwner {
				t.Errorf("splitRepository(%q) owner = %q, want %q", tt.input, owner, tt.expectedOwner)
			}
			if name != tt.expectedName {
				t.Errorf("splitRepository(%q) name = %q, want %q", tt.input, name, tt.expectedName)
			}
		})
	}
}

func TestWriteConfigWithMetadata_EmptyMetadata(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &InitConfig{
		ProjectName:   "Test",
		ProjectOwner:  "owner",
		ProjectNumber: 1,
		Repositories:  []string{"owner/repo"},
	}

	// Empty metadata with no fields
	metadata := &ProjectMetadata{
		ProjectID: "PVT_empty",
		Fields:    []FieldMetadata{},
	}

	err := writeConfigWithMetadata(tmpDir, cfg, metadata, nil)
	if err != nil {
		t.Fatalf("writeConfigWithMetadata failed with empty fields: %v", err)
	}

	content, err := readFile(tmpDir + "/.gh-pmu.yml")
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	// Should still have metadata section
	if !bytes.Contains(content, []byte("metadata:")) {
		t.Error("Config should have metadata section even with empty fields")
	}
	if !bytes.Contains(content, []byte("PVT_empty")) {
		t.Error("Config should contain project ID")
	}
}

func TestWriteConfigWithMetadata_FieldOptions(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &InitConfig{
		ProjectOwner:  "owner",
		ProjectNumber: 1,
		Repositories:  []string{"owner/repo"},
	}

	metadata := &ProjectMetadata{
		ProjectID: "PVT_test",
		Fields: []FieldMetadata{
			{
				ID:       "PVTF_size",
				Name:     "Size",
				DataType: "SINGLE_SELECT",
				Options: []OptionMetadata{
					{ID: "size_xs", Name: "XS"},
					{ID: "size_s", Name: "S"},
					{ID: "size_m", Name: "M"},
					{ID: "size_l", Name: "L"},
					{ID: "size_xl", Name: "XL"},
				},
			},
		},
	}

	err := writeConfigWithMetadata(tmpDir, cfg, metadata, nil)
	if err != nil {
		t.Fatalf("writeConfigWithMetadata failed: %v", err)
	}

	content, _ := readFile(tmpDir + "/.gh-pmu.yml")

	// Check all options are written
	options := []string{"XS", "S", "M", "L", "XL"}
	for _, opt := range options {
		if !bytes.Contains(content, []byte(opt)) {
			t.Errorf("Config should contain option %q", opt)
		}
	}
}

func TestWriteConfig_FilePermissions(t *testing.T) {
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

	// Check file exists and is readable
	info, err := os.Stat(tmpDir + "/.gh-pmu.yml")
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}

	// File should not be a directory
	if info.IsDir() {
		t.Error("Config file should not be a directory")
	}

	// File should have some content
	if info.Size() == 0 {
		t.Error("Config file should not be empty")
	}
}

// ============================================================================
// writeConfig Error Path Tests (IT-3.4)
// ============================================================================

func TestWriteConfig_InvalidDirectory(t *testing.T) {
	// Try to write to a non-existent directory
	nonExistentDir := "/nonexistent/path/that/does/not/exist"

	cfg := &InitConfig{
		ProjectOwner:  "owner",
		ProjectNumber: 1,
		Repositories:  []string{"owner/repo"},
	}

	err := writeConfig(nonExistentDir, cfg)
	if err == nil {
		t.Error("Expected error when writing to non-existent directory")
	}

	// Check error message mentions file write failure
	if !strings.Contains(err.Error(), "failed to write config file") {
		t.Errorf("Expected 'failed to write config file' error, got: %v", err)
	}
}

func TestWriteConfig_ReadOnlyDirectory(t *testing.T) {
	// Skip on Windows as permission handling differs
	if os.Getenv("OS") == "Windows_NT" || strings.Contains(os.Getenv("OS"), "Windows") {
		t.Skip("Skipping permission test on Windows")
	}

	tmpDir := t.TempDir()

	// Make directory read-only
	if err := os.Chmod(tmpDir, 0444); err != nil {
		t.Fatalf("Failed to make directory read-only: %v", err)
	}
	// Restore permissions for cleanup
	defer func() { _ = os.Chmod(tmpDir, 0755) }()

	cfg := &InitConfig{
		ProjectOwner:  "owner",
		ProjectNumber: 1,
		Repositories:  []string{"owner/repo"},
	}

	err := writeConfig(tmpDir, cfg)
	if err == nil {
		t.Error("Expected error when writing to read-only directory")
	}
}

func TestWriteConfigWithMetadata_InvalidDirectory(t *testing.T) {
	// Try to write to a non-existent directory
	nonExistentDir := "/nonexistent/path/that/does/not/exist"

	cfg := &InitConfig{
		ProjectOwner:  "owner",
		ProjectNumber: 1,
		Repositories:  []string{"owner/repo"},
	}

	metadata := &ProjectMetadata{
		ProjectID: "test-id",
		Fields:    []FieldMetadata{},
	}

	err := writeConfigWithMetadata(nonExistentDir, cfg, metadata, nil)
	if err == nil {
		t.Error("Expected error when writing to non-existent directory")
	}

	// Check error message mentions file write failure
	if !strings.Contains(err.Error(), "failed to write config file") {
		t.Errorf("Expected 'failed to write config file' error, got: %v", err)
	}
}

func TestWriteConfig_EmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Empty config should still work (though with empty/default values)
	cfg := &InitConfig{}

	err := writeConfig(tmpDir, cfg)
	if err != nil {
		t.Fatalf("writeConfig with empty config failed: %v", err)
	}

	// Verify file was created
	configPath := filepath.Join(tmpDir, ".gh-pmu.yml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
}

func TestWriteConfig_OverwriteExisting(t *testing.T) {
	tmpDir := t.TempDir()

	// Write initial config
	cfg1 := &InitConfig{
		ProjectOwner:  "owner1",
		ProjectNumber: 1,
		Repositories:  []string{"owner/repo1"},
	}
	if err := writeConfig(tmpDir, cfg1); err != nil {
		t.Fatalf("Initial writeConfig failed: %v", err)
	}

	// Write second config (should overwrite)
	cfg2 := &InitConfig{
		ProjectOwner:  "owner2",
		ProjectNumber: 2,
		Repositories:  []string{"owner/repo2"},
	}
	if err := writeConfig(tmpDir, cfg2); err != nil {
		t.Fatalf("Second writeConfig failed: %v", err)
	}

	// Read file and verify it has new content
	configPath := filepath.Join(tmpDir, ".gh-pmu.yml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "owner2") {
		t.Error("Expected config to contain 'owner2' (new value)")
	}
	if strings.Contains(content, "owner1") {
		t.Error("Expected old 'owner1' to be overwritten")
	}
}

func TestWriteConfigWithMetadata_NilMetadataPanics(t *testing.T) {
	// Document that nil metadata causes a panic
	// This test verifies the current behavior - the function does not handle nil metadata
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when metadata is nil, but function didn't panic")
		}
	}()

	tmpDir := t.TempDir()

	cfg := &InitConfig{
		ProjectOwner:  "owner",
		ProjectNumber: 1,
		Repositories:  []string{"owner/repo"},
	}

	// This should panic because metadata is nil
	// Note: In production, metadata is always provided by the caller
	_ = writeConfigWithMetadata(tmpDir, cfg, nil, nil)
}

func TestParseGitRemote_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		remote   string
		expected string
	}{
		{
			name:     "GitHub enterprise HTTPS - not supported",
			remote:   "https://github.example.com/owner/repo.git",
			expected: "",
		},
		{
			name:     "GitLab URL - not supported",
			remote:   "https://gitlab.com/owner/repo.git",
			expected: "",
		},
		{
			name:     "Bitbucket URL - not supported",
			remote:   "https://bitbucket.org/owner/repo.git",
			expected: "",
		},
		{
			name:     "SSH with port - not standard GitHub",
			remote:   "ssh://git@github.com:22/owner/repo.git",
			expected: "",
		},
		{
			name:     "file protocol",
			remote:   "file:///path/to/repo.git",
			expected: "",
		},
		{
			name:     "random string",
			remote:   "not-a-valid-url",
			expected: "",
		},
		{
			name:     "empty string",
			remote:   "",
			expected: "",
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

func TestParseReleaseTitleForInit(t *testing.T) {
	tests := []struct {
		title       string
		wantVersion string
		wantTrack   string
	}{
		{"Release: v1.2.0", "1.2.0", "stable"},
		{"Release: v1.2.0 (Phoenix)", "1.2.0", "stable"},
		{"Release: patch/1.1.1", "1.1.1", "patch"},
		{"Release: beta/2.0.0", "2.0.0", "beta"},
		{"Release: hotfix/1.0.1", "1.0.1", "hotfix"},
		{"Release: rc/3.0.0", "3.0.0", "rc"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			version, track := parseReleaseTitleForInit(tt.title)
			if version != tt.wantVersion {
				t.Errorf("version: got %s, want %s", version, tt.wantVersion)
			}
			if track != tt.wantTrack {
				t.Errorf("track: got %s, want %s", track, tt.wantTrack)
			}
		})
	}
}

// ============================================================================
// extractReleaseVersionForInit Tests
// ============================================================================

func TestExtractReleaseVersionForInit_SimpleVersion(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"Release: v1.0.0", "v1.0.0"},
		{"Release: v2.5.1", "v2.5.1"},
		{"Release: 1.0.0", "1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			result := extractReleaseVersionForInit(tt.title)
			if result != tt.expected {
				t.Errorf("extractReleaseVersionForInit(%q) = %q, want %q", tt.title, result, tt.expected)
			}
		})
	}
}

func TestExtractReleaseVersionForInit_WithCodename(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"Release: v1.0.0 (Phoenix)", "v1.0.0"},
		{"Release: v2.5.1 (Alpha)", "v2.5.1"},
		{"Release: 3.0.0 (Beta Release)", "3.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			result := extractReleaseVersionForInit(tt.title)
			if result != tt.expected {
				t.Errorf("extractReleaseVersionForInit(%q) = %q, want %q", tt.title, result, tt.expected)
			}
		})
	}
}

func TestExtractReleaseVersionForInit_TrackPrefix(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"Release: patch/v1.0.1", "patch/v1.0.1"},
		{"Release: beta/v2.0.0-rc1", "beta/v2.0.0-rc1"},
		{"Release: hotfix/1.0.2", "hotfix/1.0.2"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			result := extractReleaseVersionForInit(tt.title)
			if result != tt.expected {
				t.Errorf("extractReleaseVersionForInit(%q) = %q, want %q", tt.title, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// mergeActiveReleases Tests
// ============================================================================

func TestMergeActiveReleases_EmptyBoth(t *testing.T) {
	result := mergeActiveReleases(nil, nil)
	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d items", len(result))
	}
}

func TestMergeActiveReleases_EmptyExisting(t *testing.T) {
	discovered := []ReleaseActiveEntry{
		{Version: "1.0.0", TrackerIssue: 100, Track: "stable"},
	}

	result := mergeActiveReleases(nil, discovered)
	if len(result) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result))
	}
	if result[0].Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", result[0].Version)
	}
}

func TestMergeActiveReleases_EmptyDiscovered(t *testing.T) {
	existing := []ReleaseActiveEntry{
		{Version: "0.9.0", TrackerIssue: 50, Track: "stable"},
	}

	result := mergeActiveReleases(existing, nil)
	if len(result) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result))
	}
	if result[0].Version != "0.9.0" {
		t.Errorf("Expected version 0.9.0, got %s", result[0].Version)
	}
}

func TestMergeActiveReleases_NoDuplicates(t *testing.T) {
	existing := []ReleaseActiveEntry{
		{Version: "1.0.0", TrackerIssue: 100, Track: "stable"},
	}
	discovered := []ReleaseActiveEntry{
		{Version: "1.0.0", TrackerIssue: 100, Track: "stable"},
	}

	result := mergeActiveReleases(existing, discovered)
	if len(result) != 1 {
		t.Errorf("Expected 1 item (no duplicates), got %d", len(result))
	}
}

func TestMergeActiveReleases_MergesDifferentReleases(t *testing.T) {
	existing := []ReleaseActiveEntry{
		{Version: "0.9.0", TrackerIssue: 50, Track: "stable"},
	}
	discovered := []ReleaseActiveEntry{
		{Version: "1.0.0", TrackerIssue: 100, Track: "stable"},
	}

	result := mergeActiveReleases(existing, discovered)
	if len(result) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(result))
	}

	// Discovered should come first
	if result[0].Version != "1.0.0" {
		t.Errorf("Expected first item to be discovered (1.0.0), got %s", result[0].Version)
	}
	if result[1].Version != "0.9.0" {
		t.Errorf("Expected second item to be existing (0.9.0), got %s", result[1].Version)
	}
}

func TestMergeActiveReleases_DedupesByTrackerIssue(t *testing.T) {
	// Same tracker issue, different versions (edge case - shouldn't happen in practice)
	existing := []ReleaseActiveEntry{
		{Version: "old-version", TrackerIssue: 100, Track: "stable"},
	}
	discovered := []ReleaseActiveEntry{
		{Version: "new-version", TrackerIssue: 100, Track: "stable"},
	}

	result := mergeActiveReleases(existing, discovered)
	// Should have 1 item (discovered takes precedence for same tracker issue)
	if len(result) != 1 {
		t.Errorf("Expected 1 item (deduped by tracker issue), got %d", len(result))
	}
	if result[0].Version != "new-version" {
		t.Errorf("Expected discovered version to win, got %s", result[0].Version)
	}
}

// ============================================================================
// loadExistingConfigFull Tests
// ============================================================================

func TestLoadExistingConfigFull_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Write a config file
	configContent := `
framework: IDPF
project:
  owner: test
  number: 1
release:
  active:
    - version: "1.0.0"
      tracker_issue: 100
      track: stable
`
	configPath := filepath.Join(tmpDir, ".gh-pmu.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Load and verify
	result, err := loadExistingConfigFull(tmpDir)
	if err != nil {
		t.Fatalf("loadExistingConfigFull failed: %v", err)
	}

	if result.Framework != "IDPF" {
		t.Errorf("Expected framework 'IDPF', got %q", result.Framework)
	}
	if len(result.ActiveReleases) != 1 {
		t.Fatalf("Expected 1 active release, got %d", len(result.ActiveReleases))
	}
	if result.ActiveReleases[0].Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %q", result.ActiveReleases[0].Version)
	}
}

func TestLoadExistingConfigFull_NoActiveReleases(t *testing.T) {
	tmpDir := t.TempDir()

	configContent := `
framework: none
project:
  owner: test
  number: 1
`
	configPath := filepath.Join(tmpDir, ".gh-pmu.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	result, err := loadExistingConfigFull(tmpDir)
	if err != nil {
		t.Fatalf("loadExistingConfigFull failed: %v", err)
	}

	if result.Framework != "none" {
		t.Errorf("Expected framework 'none', got %q", result.Framework)
	}
	if len(result.ActiveReleases) != 0 {
		t.Errorf("Expected 0 active releases, got %d", len(result.ActiveReleases))
	}
}

func TestLoadExistingConfigFull_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	// Don't create any config file

	_, err := loadExistingConfigFull(tmpDir)
	if err == nil {
		t.Error("Expected error for missing config file")
	}
}

func TestLoadExistingConfigFull_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	configContent := `
not valid: yaml:
  - bad indent
    really bad
`
	configPath := filepath.Join(tmpDir, ".gh-pmu.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	_, err := loadExistingConfigFull(tmpDir)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

// ============================================================================
// isRepoRoot Tests
// ============================================================================

func TestIsRepoRoot_WithGoMod(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a go.mod file
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test"), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	result := isRepoRoot(tmpDir)
	if !result {
		t.Error("Expected isRepoRoot to return true when go.mod exists")
	}
}

func TestIsRepoRoot_WithoutGoMod(t *testing.T) {
	tmpDir := t.TempDir()

	result := isRepoRoot(tmpDir)
	if result {
		t.Error("Expected isRepoRoot to return false when go.mod doesn't exist")
	}
}

func TestIsRepoRoot_InvalidPath(t *testing.T) {
	result := isRepoRoot("/nonexistent/path/12345")
	if result {
		t.Error("Expected isRepoRoot to return false for invalid path")
	}
}

// ============================================================================
// SetRepoRootProtection Tests
// ============================================================================

func TestSetRepoRootProtection_EnablesProtection(t *testing.T) {
	// Reset protection state after test
	defer SetRepoRootProtection(false)

	SetRepoRootProtection(true)

	tmpDir := t.TempDir()
	// Create go.mod to simulate repo root
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test"), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	cfg := &InitConfig{
		ProjectOwner:  "owner",
		ProjectNumber: 1,
		Repositories:  []string{"owner/repo"},
	}

	err := writeConfig(tmpDir, cfg)
	if err != ErrRepoRootProtected {
		t.Errorf("Expected ErrRepoRootProtected, got: %v", err)
	}
}

func TestSetRepoRootProtection_DisabledAllowsWrite(t *testing.T) {
	SetRepoRootProtection(false)

	tmpDir := t.TempDir()
	// Create go.mod to simulate repo root
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test"), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	cfg := &InitConfig{
		ProjectOwner:  "owner",
		ProjectNumber: 1,
		Repositories:  []string{"owner/repo"},
	}

	err := writeConfig(tmpDir, cfg)
	if err != nil {
		t.Errorf("Expected write to succeed with protection disabled, got: %v", err)
	}
}

func TestWriteConfigWithMetadata_RepoRootProtection(t *testing.T) {
	// Reset protection state after test
	defer SetRepoRootProtection(false)

	SetRepoRootProtection(true)

	tmpDir := t.TempDir()
	// Create go.mod to simulate repo root
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test"), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	cfg := &InitConfig{
		ProjectOwner:  "owner",
		ProjectNumber: 1,
		Repositories:  []string{"owner/repo"},
	}

	metadata := &ProjectMetadata{
		ProjectID: "PVT_test",
		Fields:    []FieldMetadata{},
	}

	err := writeConfigWithMetadata(tmpDir, cfg, metadata, nil)
	if err != ErrRepoRootProtected {
		t.Errorf("Expected ErrRepoRootProtected, got: %v", err)
	}
}

// ============================================================================
// optionNameToAlias Tests (Issue #442)
// ============================================================================

func TestOptionNameToAlias(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple lowercase",
			input:    "Backlog",
			expected: "backlog",
		},
		{
			name:     "space to underscore",
			input:    "In progress",
			expected: "in_progress",
		},
		{
			name:     "multiple spaces",
			input:    "In Review",
			expected: "in_review",
		},
		{
			name:     "emoji prefix with space",
			input:    "🅿️ Parking Lot",
			expected: "parking_lot",
		},
		{
			name:     "emoji prefix no space",
			input:    "🚀Ready",
			expected: "ready",
		},
		{
			name:     "multiple emojis",
			input:    "✅ ✓ Done",
			expected: "done",
		},
		{
			name:     "emoji only",
			input:    "🔥",
			expected: "",
		},
		{
			name:     "already lowercase underscore",
			input:    "in_progress",
			expected: "in_progress",
		},
		{
			name:     "uppercase with underscore",
			input:    "IN_PROGRESS",
			expected: "in_progress",
		},
		{
			name:     "leading and trailing spaces",
			input:    "  Backlog  ",
			expected: "backlog",
		},
		{
			name:     "P0 priority",
			input:    "P0",
			expected: "p0",
		},
		{
			name:     "complex emoji",
			input:    "🏃‍♂️ Running",
			expected: "running",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := optionNameToAlias(tt.input)
			if result != tt.expected {
				t.Errorf("optionNameToAlias(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// buildFieldMappingsFromMetadata Tests (Issue #442)
// ============================================================================

func TestBuildFieldMappingsFromMetadata_StatusField(t *testing.T) {
	metadata := &ProjectMetadata{
		ProjectID: "PVT_test",
		Fields: []FieldMetadata{
			{
				ID:       "PVTF_status",
				Name:     "Status",
				DataType: "SINGLE_SELECT",
				Options: []OptionMetadata{
					{ID: "opt1", Name: "Backlog"},
					{ID: "opt2", Name: "In progress"},
					{ID: "opt3", Name: "🅿️ Parking Lot"},
					{ID: "opt4", Name: "Done"},
				},
			},
		},
	}

	mappings := buildFieldMappingsFromMetadata(metadata)

	// Check status field exists
	status, ok := mappings["status"]
	if !ok {
		t.Fatal("Expected 'status' field mapping")
	}

	if status.Field != "Status" {
		t.Errorf("Expected field name 'Status', got %q", status.Field)
	}

	// Check all values are mapped
	expectedMappings := map[string]string{
		"backlog":     "Backlog",
		"in_progress": "In progress",
		"parking_lot": "🅿️ Parking Lot",
		"done":        "Done",
	}

	for alias, expected := range expectedMappings {
		if actual, ok := status.Values[alias]; !ok {
			t.Errorf("Missing alias %q in status values", alias)
		} else if actual != expected {
			t.Errorf("status.Values[%q] = %q, want %q", alias, actual, expected)
		}
	}
}

func TestBuildFieldMappingsFromMetadata_PriorityField(t *testing.T) {
	metadata := &ProjectMetadata{
		ProjectID: "PVT_test",
		Fields: []FieldMetadata{
			{
				ID:       "PVTF_priority",
				Name:     "Priority",
				DataType: "SINGLE_SELECT",
				Options: []OptionMetadata{
					{ID: "opt1", Name: "P0"},
					{ID: "opt2", Name: "P1"},
					{ID: "opt3", Name: "P2"},
					{ID: "opt4", Name: "P3"},
				},
			},
		},
	}

	mappings := buildFieldMappingsFromMetadata(metadata)

	priority, ok := mappings["priority"]
	if !ok {
		t.Fatal("Expected 'priority' field mapping")
	}

	// Check P3 is included (not hardcoded)
	if _, ok := priority.Values["p3"]; !ok {
		t.Error("Expected 'p3' to be included in priority values")
	}
}

func TestBuildFieldMappingsFromMetadata_FallbackDefaults(t *testing.T) {
	// Empty metadata - should use fallback defaults
	metadata := &ProjectMetadata{
		ProjectID: "PVT_test",
		Fields:    []FieldMetadata{},
	}

	mappings := buildFieldMappingsFromMetadata(metadata)

	// Should have default status
	status, ok := mappings["status"]
	if !ok {
		t.Fatal("Expected default 'status' field mapping")
	}
	if _, ok := status.Values["backlog"]; !ok {
		t.Error("Expected default 'backlog' in status values")
	}

	// Should have default priority
	priority, ok := mappings["priority"]
	if !ok {
		t.Fatal("Expected default 'priority' field mapping")
	}
	if _, ok := priority.Values["p0"]; !ok {
		t.Error("Expected default 'p0' in priority values")
	}
}

func TestBuildFieldMappingsFromMetadata_NoOptions(t *testing.T) {
	// Fields exist but have no options - should fall back to defaults
	metadata := &ProjectMetadata{
		ProjectID: "PVT_test",
		Fields: []FieldMetadata{
			{
				ID:       "PVTF_status",
				Name:     "Status",
				DataType: "SINGLE_SELECT",
				Options:  []OptionMetadata{}, // Empty options
			},
		},
	}

	mappings := buildFieldMappingsFromMetadata(metadata)

	// Should fall back to default status values
	status := mappings["status"]
	if _, ok := status.Values["backlog"]; !ok {
		t.Error("Expected fallback to default 'backlog' when no options")
	}
}

func TestBuildFieldMappingsFromMetadata_CaseInsensitiveFieldName(t *testing.T) {
	metadata := &ProjectMetadata{
		ProjectID: "PVT_test",
		Fields: []FieldMetadata{
			{
				ID:       "PVTF_status",
				Name:     "STATUS", // Uppercase
				DataType: "SINGLE_SELECT",
				Options: []OptionMetadata{
					{ID: "opt1", Name: "Active"},
				},
			},
		},
	}

	mappings := buildFieldMappingsFromMetadata(metadata)

	status, ok := mappings["status"]
	if !ok {
		t.Fatal("Expected 'status' field mapping for uppercase 'STATUS'")
	}
	if status.Field != "STATUS" {
		t.Errorf("Expected field name to preserve case 'STATUS', got %q", status.Field)
	}
}

func TestWriteConfigWithMetadata_IncludesParkingLot(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &InitConfig{
		ProjectOwner:  "owner",
		ProjectNumber: 1,
		Repositories:  []string{"owner/repo"},
	}

	metadata := &ProjectMetadata{
		ProjectID: "PVT_test",
		Fields: []FieldMetadata{
			{
				ID:       "PVTF_status",
				Name:     "Status",
				DataType: "SINGLE_SELECT",
				Options: []OptionMetadata{
					{ID: "opt1", Name: "Backlog"},
					{ID: "opt2", Name: "🅿️ Parking Lot"},
					{ID: "opt3", Name: "Done"},
				},
			},
		},
	}

	err := writeConfigWithMetadata(tmpDir, cfg, metadata, nil)
	if err != nil {
		t.Fatalf("writeConfigWithMetadata failed: %v", err)
	}

	content, _ := readFile(tmpDir + "/.gh-pmu.yml")

	// Should contain parking_lot alias
	if !bytes.Contains(content, []byte("parking_lot:")) {
		t.Error("Config should contain 'parking_lot:' alias")
	}

	// Should contain the original name with emoji
	if !bytes.Contains(content, []byte("Parking Lot")) {
		t.Error("Config should contain 'Parking Lot' value")
	}
}
