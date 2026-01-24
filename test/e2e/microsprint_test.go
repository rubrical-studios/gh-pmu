//go:build e2e

package e2e

import (
	"testing"
)

// TestMicrosprintDeprecation tests that microsprint commands show deprecation message
// and exit successfully without performing operations.
func TestMicrosprintStartWithBranch(t *testing.T) {
	cfg := setupTestConfig(t)

	// Microsprint commands are deprecated - they should show deprecation message and exit
	t.Run("start microsprint with branch", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "microsprint", "start", "--name", "test", "--branch", "test-branch")
		assertExitCode(t, result, 0)
		assertContains(t, result.Stdout, "deprecated")
		assertContains(t, result.Stdout, "No operations have been performed")
	})
}

// TestMicrosprintLifecycle tests that all microsprint commands show deprecation message.
func TestMicrosprintLifecycle(t *testing.T) {
	cfg := setupTestConfig(t)

	// Step 1: microsprint start shows deprecation
	t.Run("start microsprint", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "microsprint", "start")
		assertExitCode(t, result, 0)
		assertContains(t, result.Stdout, "deprecated")
	})

	// Step 2: microsprint current shows deprecation
	t.Run("verify current microsprint", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "microsprint", "current")
		assertExitCode(t, result, 0)
		assertContains(t, result.Stdout, "deprecated")
	})

	// Step 3: microsprint add shows deprecation
	t.Run("add issue to microsprint", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "microsprint", "add", "123")
		assertExitCode(t, result, 0)
		assertContains(t, result.Stdout, "deprecated")
	})

	// Step 4: microsprint list shows deprecation
	t.Run("verify microsprint list", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "microsprint", "list")
		assertExitCode(t, result, 0)
		assertContains(t, result.Stdout, "deprecated")
	})

	// Step 5: microsprint close shows deprecation
	t.Run("close microsprint", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "microsprint", "close")
		assertExitCode(t, result, 0)
		assertContains(t, result.Stdout, "deprecated")
	})

	// Step 6: microsprint remove shows deprecation
	t.Run("remove from microsprint", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "microsprint", "remove", "123")
		assertExitCode(t, result, 0)
		assertContains(t, result.Stdout, "deprecated")
	})
}
