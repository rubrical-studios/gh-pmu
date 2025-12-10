package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ValidConfig_ReturnsProjectDetails(t *testing.T) {
	// ARRANGE: Path to valid test config
	configPath := filepath.Join("..", "..", "testdata", "config", "valid.gh-pmu.yml")

	// ACT: Load the configuration
	cfg, err := Load(configPath)

	// ASSERT: No error and correct values
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg.Project.Owner != "rubrical-studios" {
		t.Errorf("Expected owner 'rubrical-studios', got '%s'", cfg.Project.Owner)
	}

	if cfg.Project.Number != 13 {
		t.Errorf("Expected project number 13, got %d", cfg.Project.Number)
	}
}

func TestLoad_MinimalConfig_ReturnsRequiredFields(t *testing.T) {
	// ARRANGE: Path to minimal test config
	configPath := filepath.Join("..", "..", "testdata", "config", "minimal.gh-pmu.yml")

	// ACT: Load the configuration
	cfg, err := Load(configPath)

	// ASSERT: No error and required fields present
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg.Project.Owner != "rubrical-studios" {
		t.Errorf("Expected owner 'rubrical-studios', got '%s'", cfg.Project.Owner)
	}

	if cfg.Project.Number != 13 {
		t.Errorf("Expected project number 13, got %d", cfg.Project.Number)
	}

	if len(cfg.Repositories) != 1 {
		t.Errorf("Expected 1 repository, got %d", len(cfg.Repositories))
	}
}

func TestLoad_MissingFile_ReturnsError(t *testing.T) {
	// ARRANGE: Path to non-existent file
	configPath := filepath.Join("..", "..", "testdata", "config", "does-not-exist.yml")

	// ACT: Load the configuration
	_, err := Load(configPath)

	// ASSERT: Error is returned
	if err == nil {
		t.Fatal("Expected error for missing file, got nil")
	}
}

func TestLoad_InvalidYAML_ReturnsError(t *testing.T) {
	// ARRANGE: Path to invalid YAML
	configPath := filepath.Join("..", "..", "testdata", "config", "invalid-yaml-syntax.gh-pmu.yml")

	// ACT: Load the configuration
	_, err := Load(configPath)

	// ASSERT: Error is returned
	if err == nil {
		t.Fatal("Expected error for invalid YAML, got nil")
	}
}

func TestValidate_MissingOwner_ReturnsError(t *testing.T) {
	// ARRANGE: Config with missing owner
	cfg := &Config{
		Project: Project{
			Number: 13,
			// Owner is missing
		},
		Repositories: []string{"rubrical-studios/gh-pm-test"},
	}

	// ACT: Validate the config
	err := cfg.Validate()

	// ASSERT: Error mentions owner
	if err == nil {
		t.Fatal("Expected validation error for missing owner, got nil")
	}
}

func TestValidate_MissingNumber_ReturnsError(t *testing.T) {
	// ARRANGE: Config with missing project number
	cfg := &Config{
		Project: Project{
			Owner: "rubrical-studios",
			// Number is missing (zero value)
		},
		Repositories: []string{"rubrical-studios/gh-pm-test"},
	}

	// ACT: Validate the config
	err := cfg.Validate()

	// ASSERT: Error mentions number
	if err == nil {
		t.Fatal("Expected validation error for missing project number, got nil")
	}
}

func TestValidate_MissingRepositories_ReturnsError(t *testing.T) {
	// ARRANGE: Config with no repositories
	cfg := &Config{
		Project: Project{
			Owner:  "rubrical-studios",
			Number: 13,
		},
		Repositories: []string{}, // Empty
	}

	// ACT: Validate the config
	err := cfg.Validate()

	// ASSERT: Error mentions repositories
	if err == nil {
		t.Fatal("Expected validation error for missing repositories, got nil")
	}
}

func TestValidate_ValidConfig_ReturnsNil(t *testing.T) {
	// ARRANGE: Valid config
	cfg := &Config{
		Project: Project{
			Owner:  "rubrical-studios",
			Number: 13,
		},
		Repositories: []string{"rubrical-studios/gh-pm-test"},
	}

	// ACT: Validate the config
	err := cfg.Validate()

	// ASSERT: No error
	if err != nil {
		t.Fatalf("Expected no error for valid config, got: %v", err)
	}
}

func TestResolveFieldValue_WithAlias_ReturnsActualValue(t *testing.T) {
	// ARRANGE: Config with field aliases
	cfg := &Config{
		Fields: map[string]Field{
			"priority": {
				Field: "Priority",
				Values: map[string]string{
					"p0": "P0",
					"p1": "P1",
					"p2": "P2",
				},
			},
		},
	}

	// ACT: Resolve alias
	value := cfg.ResolveFieldValue("priority", "p1")

	// ASSERT: Returns actual value
	if value != "P1" {
		t.Errorf("Expected 'P1', got '%s'", value)
	}
}

func TestResolveFieldValue_NoAlias_ReturnsOriginal(t *testing.T) {
	// ARRANGE: Config with field aliases
	cfg := &Config{
		Fields: map[string]Field{
			"priority": {
				Field: "Priority",
				Values: map[string]string{
					"p0": "P0",
					"p1": "P1",
				},
			},
		},
	}

	// ACT: Try to resolve value that has no alias
	value := cfg.ResolveFieldValue("priority", "Unknown")

	// ASSERT: Returns original value unchanged
	if value != "Unknown" {
		t.Errorf("Expected 'Unknown', got '%s'", value)
	}
}

func TestResolveFieldValue_UnknownField_ReturnsOriginal(t *testing.T) {
	// ARRANGE: Config with no fields configured
	cfg := &Config{
		Fields: map[string]Field{},
	}

	// ACT: Try to resolve unknown field
	value := cfg.ResolveFieldValue("unknown", "some-value")

	// ASSERT: Returns original value unchanged
	if value != "some-value" {
		t.Errorf("Expected 'some-value', got '%s'", value)
	}
}

func TestGetFieldName_WithMapping_ReturnsActualName(t *testing.T) {
	// ARRANGE: Config with field mapping
	cfg := &Config{
		Fields: map[string]Field{
			"priority": {
				Field: "Priority",
			},
			"status": {
				Field: "Status",
			},
		},
	}

	// ACT: Get actual field name
	name := cfg.GetFieldName("priority")

	// ASSERT: Returns mapped name
	if name != "Priority" {
		t.Errorf("Expected 'Priority', got '%s'", name)
	}
}

func TestGetFieldName_NoMapping_ReturnsOriginal(t *testing.T) {
	// ARRANGE: Config with no field mapping
	cfg := &Config{
		Fields: map[string]Field{},
	}

	// ACT: Get field name for unmapped field
	name := cfg.GetFieldName("SomeField")

	// ASSERT: Returns original name
	if name != "SomeField" {
		t.Errorf("Expected 'SomeField', got '%s'", name)
	}
}

func TestLoadFromDirectory_FindsConfigFile(t *testing.T) {
	// ARRANGE: Directory containing valid config
	dir := filepath.Join("..", "..", "testdata", "config")

	// Create a temporary .gh-pmu.yml in testdata/config for this test
	// (We'll use the valid.gh-pmu.yml by copying it)
	testDir := t.TempDir()
	srcPath := filepath.Join(dir, "valid.gh-pmu.yml")
	dstPath := filepath.Join(testDir, ".gh-pmu.yml")

	// Copy the file
	data, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("Failed to read source file: %v", err)
	}
	if err := os.WriteFile(dstPath, data, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// ACT: Load from directory
	cfg, err := LoadFromDirectory(testDir)

	// ASSERT: Config loaded successfully
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg.Project.Owner != "rubrical-studios" {
		t.Errorf("Expected owner 'rubrical-studios', got '%s'", cfg.Project.Owner)
	}
}

func TestLoadFromDirectory_NoConfigFile_ReturnsError(t *testing.T) {
	// ARRANGE: Empty directory
	testDir := t.TempDir()

	// ACT: Try to load from directory with no config
	_, err := LoadFromDirectory(testDir)

	// ASSERT: Error is returned
	if err == nil {
		t.Fatal("Expected error for missing config file, got nil")
	}
}

func TestApplyEnvOverrides_OverridesOwner(t *testing.T) {
	// ARRANGE: Config and env var
	cfg := &Config{
		Project: Project{
			Owner:  "original-owner",
			Number: 13,
		},
	}
	t.Setenv("GH_PM_PROJECT_OWNER", "env-owner")

	// ACT: Apply overrides
	cfg.ApplyEnvOverrides()

	// ASSERT: Owner is overridden
	if cfg.Project.Owner != "env-owner" {
		t.Errorf("Expected owner 'env-owner', got '%s'", cfg.Project.Owner)
	}
}

func TestApplyEnvOverrides_OverridesNumber(t *testing.T) {
	// ARRANGE: Config and env var
	cfg := &Config{
		Project: Project{
			Owner:  "rubrical-studios",
			Number: 13,
		},
	}
	t.Setenv("GH_PM_PROJECT_NUMBER", "99")

	// ACT: Apply overrides
	cfg.ApplyEnvOverrides()

	// ASSERT: Number is overridden
	if cfg.Project.Number != 99 {
		t.Errorf("Expected project number 99, got %d", cfg.Project.Number)
	}
}

func TestApplyEnvOverrides_InvalidNumber_Ignored(t *testing.T) {
	// ARRANGE: Config and invalid env var
	cfg := &Config{
		Project: Project{
			Owner:  "rubrical-studios",
			Number: 13,
		},
	}
	t.Setenv("GH_PM_PROJECT_NUMBER", "not-a-number")

	// ACT: Apply overrides
	cfg.ApplyEnvOverrides()

	// ASSERT: Number unchanged
	if cfg.Project.Number != 13 {
		t.Errorf("Expected project number 13 (unchanged), got %d", cfg.Project.Number)
	}
}

func TestApplyEnvOverrides_NoEnvVars_Unchanged(t *testing.T) {
	// ARRANGE: Config with no env vars set
	cfg := &Config{
		Project: Project{
			Owner:  "original-owner",
			Number: 13,
		},
	}
	// Ensure env vars are not set
	os.Unsetenv("GH_PM_PROJECT_OWNER")
	os.Unsetenv("GH_PM_PROJECT_NUMBER")

	// ACT: Apply overrides
	cfg.ApplyEnvOverrides()

	// ASSERT: Values unchanged
	if cfg.Project.Owner != "original-owner" {
		t.Errorf("Expected owner 'original-owner', got '%s'", cfg.Project.Owner)
	}
	if cfg.Project.Number != 13 {
		t.Errorf("Expected project number 13, got %d", cfg.Project.Number)
	}
}

func TestFindConfigFile_InCurrentDir_ReturnsPath(t *testing.T) {
	// ARRANGE: Create temp dir with config file
	testDir := t.TempDir()
	configPath := filepath.Join(testDir, ConfigFileName)
	if err := os.WriteFile(configPath, []byte("project:\n  owner: test\n  number: 1\n"), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// ACT: Find config starting from same dir
	found, err := FindConfigFile(testDir)

	// ASSERT: Found in current dir
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if found != configPath {
		t.Errorf("Expected %s, got %s", configPath, found)
	}
}

func TestFindConfigFile_InParentDir_ReturnsPath(t *testing.T) {
	// ARRANGE: Create nested dirs, config in parent
	parentDir := t.TempDir()
	childDir := filepath.Join(parentDir, "subdir")
	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}
	configPath := filepath.Join(parentDir, ConfigFileName)
	if err := os.WriteFile(configPath, []byte("project:\n  owner: test\n  number: 1\n"), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// ACT: Find config starting from child dir
	found, err := FindConfigFile(childDir)

	// ASSERT: Found in parent dir
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if found != configPath {
		t.Errorf("Expected %s, got %s", configPath, found)
	}
}

func TestFindConfigFile_InGrandparentDir_ReturnsPath(t *testing.T) {
	// ARRANGE: Create deeply nested dirs, config in grandparent
	grandparentDir := t.TempDir()
	parentDir := filepath.Join(grandparentDir, "parent")
	childDir := filepath.Join(parentDir, "child")
	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatalf("Failed to create nested dirs: %v", err)
	}
	configPath := filepath.Join(grandparentDir, ConfigFileName)
	if err := os.WriteFile(configPath, []byte("project:\n  owner: test\n  number: 1\n"), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// ACT: Find config starting from grandchild dir
	found, err := FindConfigFile(childDir)

	// ASSERT: Found in grandparent dir
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if found != configPath {
		t.Errorf("Expected %s, got %s", configPath, found)
	}
}

func TestFindConfigFile_NotFound_ReturnsError(t *testing.T) {
	// ARRANGE: Empty temp dir (no config anywhere in tree)
	testDir := t.TempDir()
	childDir := filepath.Join(testDir, "subdir")
	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// ACT: Try to find config
	_, err := FindConfigFile(childDir)

	// ASSERT: Error returned
	if err == nil {
		t.Fatal("Expected error when no config file exists, got nil")
	}
}

func TestLoadFromDirectory_FromSubdir_FindsParentConfig(t *testing.T) {
	// ARRANGE: Create nested structure with config in parent
	parentDir := t.TempDir()
	childDir := filepath.Join(parentDir, "subdir", "nested")
	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatalf("Failed to create nested dirs: %v", err)
	}

	configContent := `project:
  owner: rubrical-studios
  number: 13
repositories:
  - rubrical-studios/gh-pmu
`
	configPath := filepath.Join(parentDir, ConfigFileName)
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// ACT: Load from nested child dir
	cfg, err := LoadFromDirectory(childDir)

	// ASSERT: Config loaded from parent
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if cfg.Project.Owner != "rubrical-studios" {
		t.Errorf("Expected owner 'rubrical-studios', got '%s'", cfg.Project.Owner)
	}
	if cfg.Project.Number != 13 {
		t.Errorf("Expected number 13, got %d", cfg.Project.Number)
	}
}
