//go:build e2e

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestInitNonInteractiveMode tests the init command in non-interactive mode.
// This uses the test project #41 configured in config_test.go.
func TestInitNonInteractiveMode(t *testing.T) {
	// Create a temp directory for the init test
	tmpDir := t.TempDir()

	// Initialize git repo (some init validation may need it)
	initGitRepo(t, tmpDir)

	t.Run("non-interactive_creates_config", func(t *testing.T) {
		// Run init in non-interactive mode
		result := runPMU(t, tmpDir, "init",
			"--non-interactive",
			"--project", "41",
			"--repo", "rubrical-studios/gh-pmu-e2e-test",
		)

		assertExitCode(t, result, 0)

		// Verify config file was created
		configPath := filepath.Join(tmpDir, ".gh-pmu.yml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Config file was not created")
		}

		// Read and validate config
		configData, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config: %v", err)
		}

		var config map[string]interface{}
		if err := yaml.Unmarshal(configData, &config); err != nil {
			t.Fatalf("Failed to parse config: %v", err)
		}

		// Verify project section
		project, ok := config["project"].(map[string]interface{})
		if !ok {
			t.Error("Config missing 'project' section")
		} else {
			if project["number"] != 41 {
				t.Errorf("Expected project number 41, got %v", project["number"])
			}
			if project["owner"] != "rubrical-studios" {
				t.Errorf("Expected owner 'rubrical-studios', got %v", project["owner"])
			}
		}

		// Verify repositories section
		repos, ok := config["repositories"].([]interface{})
		if !ok {
			t.Error("Config missing 'repositories' section")
		} else if len(repos) == 0 {
			t.Error("No repositories configured")
		} else if repos[0] != "rubrical-studios/gh-pmu-e2e-test" {
			t.Errorf("Expected repo 'rubrical-studios/gh-pmu-e2e-test', got %v", repos[0])
		}

		// Verify framework defaults to IDPF
		if framework, ok := config["framework"].(string); ok {
			if framework != "IDPF" {
				t.Errorf("Expected framework 'IDPF', got %q", framework)
			}
		}

		// Verify metadata section exists
		if _, ok := config["metadata"]; !ok {
			t.Error("Config missing 'metadata' section")
		}
	})
}

// TestInitNonInteractiveFrameworkNone tests init with --framework none.
func TestInitNonInteractiveFrameworkNone(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	result := runPMU(t, tmpDir, "init",
		"--non-interactive",
		"--project", "41",
		"--repo", "rubrical-studios/gh-pmu-e2e-test",
		"--framework", "none",
	)

	assertExitCode(t, result, 0)

	// Read and validate config
	configPath := filepath.Join(tmpDir, ".gh-pmu.yml")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(configData, &config); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Verify framework is none
	if framework, ok := config["framework"].(string); ok {
		if framework != "none" {
			t.Errorf("Expected framework 'none', got %q", framework)
		}
	} else {
		t.Error("Framework field not found in config")
	}
}

// TestInitNonInteractiveWithOwner tests init with explicit --owner flag.
func TestInitNonInteractiveWithOwner(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	// Using explicit owner (same as repo owner in this test)
	result := runPMU(t, tmpDir, "init",
		"--non-interactive",
		"--project", "41",
		"--repo", "rubrical-studios/gh-pmu-e2e-test",
		"--owner", "rubrical-studios",
	)

	assertExitCode(t, result, 0)

	// Verify config was created
	configPath := filepath.Join(tmpDir, ".gh-pmu.yml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
}

// TestInitNonInteractiveOverwrite tests the --yes flag for overwriting.
func TestInitNonInteractiveOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	// Create existing config
	existingConfig := "project:\n  name: existing\n  owner: test\n  number: 999\n"
	configPath := filepath.Join(tmpDir, ".gh-pmu.yml")
	if err := os.WriteFile(configPath, []byte(existingConfig), 0644); err != nil {
		t.Fatalf("Failed to write existing config: %v", err)
	}

	// Run init with --yes to overwrite
	result := runPMU(t, tmpDir, "init",
		"--non-interactive",
		"--project", "41",
		"--repo", "rubrical-studios/gh-pmu-e2e-test",
		"--yes",
	)

	assertExitCode(t, result, 0)

	// Verify config was overwritten
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	if strings.Contains(string(configData), "existing") {
		t.Error("Expected existing config to be overwritten")
	}
	if !strings.Contains(string(configData), "rubrical-studios") {
		t.Error("Expected new config to contain repo owner")
	}
}

// TestInitNonInteractiveMissingFlags tests error handling for missing flags.
func TestInitNonInteractiveMissingFlags(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("missing_both_flags", func(t *testing.T) {
		result := runPMU(t, tmpDir, "init", "--non-interactive")
		assertExitCode(t, result, 1)
		assertContains(t, result.Stderr, "--project")
		assertContains(t, result.Stderr, "--repo")
	})

	t.Run("missing_project", func(t *testing.T) {
		result := runPMU(t, tmpDir, "init",
			"--non-interactive",
			"--repo", "owner/repo",
		)
		assertExitCode(t, result, 1)
		assertContains(t, result.Stderr, "--project")
	})

	t.Run("missing_repo", func(t *testing.T) {
		result := runPMU(t, tmpDir, "init",
			"--non-interactive",
			"--project", "41",
		)
		assertExitCode(t, result, 1)
		assertContains(t, result.Stderr, "--repo")
	})
}

// TestInitNonInteractiveInvalidRepoFormat tests error handling for invalid repo format.
func TestInitNonInteractiveInvalidRepoFormat(t *testing.T) {
	tmpDir := t.TempDir()

	result := runPMU(t, tmpDir, "init",
		"--non-interactive",
		"--project", "41",
		"--repo", "invalid-repo-format",
	)

	assertExitCode(t, result, 1)
	// Should mention owner/repo format in error
	if !strings.Contains(result.Stderr, "owner/repo") && !strings.Contains(result.Stderr, "invalid") {
		t.Errorf("Expected error about repo format, got: %s", result.Stderr)
	}
}

// TestInitNonInteractiveExistingConfigNoYes tests error when config exists without --yes.
func TestInitNonInteractiveExistingConfigNoYes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create existing config
	configPath := filepath.Join(tmpDir, ".gh-pmu.yml")
	if err := os.WriteFile(configPath, []byte("project:\n  owner: test\n"), 0644); err != nil {
		t.Fatalf("Failed to write existing config: %v", err)
	}

	result := runPMU(t, tmpDir, "init",
		"--non-interactive",
		"--project", "41",
		"--repo", "rubrical-studios/gh-pmu-e2e-test",
	)

	assertExitCode(t, result, 1)
	// Should mention --yes in error
	if !strings.Contains(result.Stderr, "--yes") && !strings.Contains(result.Stderr, "already exists") {
		t.Errorf("Expected error about --yes or already exists, got: %s", result.Stderr)
	}
}
