package framework

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// =============================================================================
// REQ-029: Framework Detection
// =============================================================================

// AC-029-1: Given `.gh-pmu.yml` with `framework: IDPF-Agile`, Then framework detected as Agile
func TestDetectFramework_FromGhPmuYml(t *testing.T) {
	// ARRANGE
	tempDir := t.TempDir()

	// Create .gh-pmu.yml with framework field
	ghPmuContent := `project:
  owner: testowner
  number: 1
repositories:
  - testowner/testrepo
framework: IDPF-Agile
`
	err := os.WriteFile(filepath.Join(tempDir, ".gh-pmu.yml"), []byte(ghPmuContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// ACT
	framework, err := DetectFramework(tempDir)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if framework != "IDPF-Agile" {
		t.Errorf("Expected framework 'IDPF-Agile', got '%s'", framework)
	}
}

// AC-029-2: Given `framework-config.json` with `processFramework: "IDPF-Structured"`, Then framework detected
func TestDetectFramework_FromFrameworkConfigJson(t *testing.T) {
	// ARRANGE
	tempDir := t.TempDir()

	// Create framework-config.json with processFramework
	frameworkConfigContent := `{
  "projectType": {
    "processFramework": "IDPF-Structured"
  }
}`
	err := os.WriteFile(filepath.Join(tempDir, "framework-config.json"), []byte(frameworkConfigContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// ACT
	framework, err := DetectFramework(tempDir)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if framework != "IDPF-Structured" {
		t.Errorf("Expected framework 'IDPF-Structured', got '%s'", framework)
	}
}

// AC-029-3: Given no framework config, Then no framework restriction applied (empty string)
func TestDetectFramework_NoConfig_ReturnsEmpty(t *testing.T) {
	// ARRANGE
	tempDir := t.TempDir()

	// No config files created - empty directory

	// ACT
	framework, err := DetectFramework(tempDir)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if framework != "" {
		t.Errorf("Expected empty framework (no restriction), got '%s'", framework)
	}
}

// Test priority: .gh-pmu.yml takes precedence over framework-config.json
func TestDetectFramework_GhPmuYmlTakesPrecedence(t *testing.T) {
	// ARRANGE
	tempDir := t.TempDir()

	// Create both config files with different values
	ghPmuContent := `project:
  owner: testowner
  number: 1
repositories:
  - testowner/testrepo
framework: IDPF-Agile
`
	err := os.WriteFile(filepath.Join(tempDir, ".gh-pmu.yml"), []byte(ghPmuContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .gh-pmu.yml: %v", err)
	}

	frameworkConfigContent := `{
  "projectType": {
    "processFramework": "IDPF-Structured"
  }
}`
	err = os.WriteFile(filepath.Join(tempDir, "framework-config.json"), []byte(frameworkConfigContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create framework-config.json: %v", err)
	}

	// ACT
	framework, err := DetectFramework(tempDir)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// .gh-pmu.yml should take precedence
	if framework != "IDPF-Agile" {
		t.Errorf("Expected framework 'IDPF-Agile' (from .gh-pmu.yml), got '%s'", framework)
	}
}

// Test .gh-pmu.yml without framework field falls back to framework-config.json
func TestDetectFramework_GhPmuYmlWithoutFramework_FallsBackToFrameworkConfig(t *testing.T) {
	// ARRANGE
	tempDir := t.TempDir()

	// Create .gh-pmu.yml WITHOUT framework field
	ghPmuContent := `project:
  owner: testowner
  number: 1
repositories:
  - testowner/testrepo
`
	err := os.WriteFile(filepath.Join(tempDir, ".gh-pmu.yml"), []byte(ghPmuContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .gh-pmu.yml: %v", err)
	}

	frameworkConfigContent := `{
  "projectType": {
    "processFramework": "IDPF-Structured"
  }
}`
	err = os.WriteFile(filepath.Join(tempDir, "framework-config.json"), []byte(frameworkConfigContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create framework-config.json: %v", err)
	}

	// ACT
	framework, err := DetectFramework(tempDir)

	// ASSERT
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should fall back to framework-config.json
	if framework != "IDPF-Structured" {
		t.Errorf("Expected framework 'IDPF-Structured' (from framework-config.json), got '%s'", framework)
	}
}

// =============================================================================
// REQ-030: Command Restrictions
// =============================================================================

// AC-030-1: Given IDPF-Structured framework, When running `microsprint start`,
// Then error: "Microsprint commands not applicable for IDPF-Structured. Use `gh pmu release start --version X.Y.Z`"
func TestValidateCommand_StructuredFramework_MicrosprintBlocked(t *testing.T) {
	// ARRANGE
	framework := "IDPF-Structured"
	command := "microsprint"

	// ACT
	err := ValidateCommand(framework, command)

	// ASSERT
	if err == nil {
		t.Fatalf("Expected error for microsprint on IDPF-Structured, got nil")
	}

	// Verify error message contains helpful suggestion
	errMsg := err.Error()
	if !strings.Contains(errMsg, "not applicable") {
		t.Errorf("Expected error to mention 'not applicable', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "release") {
		t.Errorf("Expected error to suggest 'release' command, got: %s", errMsg)
	}
}

// AC-030-2: Given IDPF-LTS framework, When running `microsprint start`,
// Then error suggesting `patch` command
func TestValidateCommand_LTSFramework_MicrosprintBlocked(t *testing.T) {
	// ARRANGE
	framework := "IDPF-LTS"
	command := "microsprint"

	// ACT
	err := ValidateCommand(framework, command)

	// ASSERT
	if err == nil {
		t.Fatalf("Expected error for microsprint on IDPF-LTS, got nil")
	}

	// Verify error message contains helpful suggestion
	errMsg := err.Error()
	if !strings.Contains(errMsg, "not applicable") {
		t.Errorf("Expected error to mention 'not applicable', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "patch") {
		t.Errorf("Expected error to suggest 'patch' command, got: %s", errMsg)
	}
}

// AC-030-3: Given IDPF-Agile framework, When running `patch start`,
// Then error suggesting `microsprint` command
func TestValidateCommand_AgileFramework_PatchBlocked(t *testing.T) {
	// ARRANGE
	framework := "IDPF-Agile"
	command := "patch"

	// ACT
	err := ValidateCommand(framework, command)

	// ASSERT
	if err == nil {
		t.Fatalf("Expected error for patch on IDPF-Agile, got nil")
	}

	// Verify error message contains helpful suggestion
	errMsg := err.Error()
	if !strings.Contains(errMsg, "not applicable") {
		t.Errorf("Expected error to mention 'not applicable', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "microsprint") {
		t.Errorf("Expected error to suggest 'microsprint' command, got: %s", errMsg)
	}
}

// Test that valid commands pass validation
func TestValidateCommand_AgileFramework_MicrosprintAllowed(t *testing.T) {
	// ARRANGE
	framework := "IDPF-Agile"
	command := "microsprint"

	// ACT
	err := ValidateCommand(framework, command)

	// ASSERT
	if err != nil {
		t.Errorf("Expected microsprint to be allowed on IDPF-Agile, got error: %v", err)
	}
}

func TestValidateCommand_StructuredFramework_ReleaseAllowed(t *testing.T) {
	// ARRANGE
	framework := "IDPF-Structured"
	command := "release"

	// ACT
	err := ValidateCommand(framework, command)

	// ASSERT
	if err != nil {
		t.Errorf("Expected release to be allowed on IDPF-Structured, got error: %v", err)
	}
}

func TestValidateCommand_LTSFramework_PatchAllowed(t *testing.T) {
	// ARRANGE
	framework := "IDPF-LTS"
	command := "patch"

	// ACT
	err := ValidateCommand(framework, command)

	// ASSERT
	if err != nil {
		t.Errorf("Expected patch to be allowed on IDPF-LTS, got error: %v", err)
	}
}

// =============================================================================
// REQ-031: No Framework Fallback
// =============================================================================

// AC-031-1: Given no framework config, When running any workflow command,
// Then command executes without restriction
func TestValidateCommand_NoFramework_AllCommandsAllowed(t *testing.T) {
	// ARRANGE
	framework := "" // No framework detected
	commands := []string{"microsprint", "release", "patch"}

	for _, command := range commands {
		// ACT
		err := ValidateCommand(framework, command)

		// ASSERT
		if err != nil {
			t.Errorf("Expected %s to be allowed with no framework, got error: %v", command, err)
		}
	}
}

// AC-031-2: Given no framework, Then microsprint, release, and patch commands all available
func TestValidateCommand_NoFramework_MicrosprintAvailable(t *testing.T) {
	// ARRANGE
	framework := ""
	command := "microsprint"

	// ACT
	err := ValidateCommand(framework, command)

	// ASSERT
	if err != nil {
		t.Errorf("Expected microsprint to be available with no framework, got error: %v", err)
	}
}

func TestValidateCommand_NoFramework_ReleaseAvailable(t *testing.T) {
	// ARRANGE
	framework := ""
	command := "release"

	// ACT
	err := ValidateCommand(framework, command)

	// ASSERT
	if err != nil {
		t.Errorf("Expected release to be available with no framework, got error: %v", err)
	}
}

func TestValidateCommand_NoFramework_PatchAvailable(t *testing.T) {
	// ARRANGE
	framework := ""
	command := "patch"

	// ACT
	err := ValidateCommand(framework, command)

	// ASSERT
	if err != nil {
		t.Errorf("Expected patch to be available with no framework, got error: %v", err)
	}
}
