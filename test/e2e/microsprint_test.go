//go:build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"
)

// TestMicrosprintLifecycle tests the complete microsprint workflow:
// start -> add issue -> current -> list -> close
func TestMicrosprintLifecycle(t *testing.T) {
	cfg := setupTestConfig(t)

	// Generate unique microsprint name with timestamp
	sprintName := fmt.Sprintf("e2e-test-%d", time.Now().UnixNano())

	// Track resources for cleanup
	var testIssueNum int

	// Cleanup at end of test
	defer func() {
		if testIssueNum > 0 {
			runCleanupAfterTest(t, testIssueNum)
		}
		// Ensure microsprint is closed even if test fails
		runPMU(t, cfg.Dir, "microsprint", "close", "--skip-retro")
	}()

	// Step 1: Start a new microsprint with --name flag
	t.Run("start microsprint", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "microsprint", "start", "--name", sprintName)
		assertExitCode(t, result, 0)
		assertContains(t, result.Stdout, sprintName)
	})

	// Step 2: Verify microsprint current shows the sprint
	t.Run("verify current microsprint", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "microsprint", "current")
		assertExitCode(t, result, 0)
		assertContains(t, result.Stdout, sprintName)
	})

	// Step 3: Create and add an issue to the microsprint
	t.Run("add issue to microsprint", func(t *testing.T) {
		// Create test issue
		testIssueNum = createTestIssue(t, cfg, "Microsprint Test Issue")

		// Add to microsprint
		result := runPMU(t, cfg.Dir, "microsprint", "add", fmt.Sprintf("%d", testIssueNum))
		assertExitCode(t, result, 0)
	})

	// Step 4: Verify microsprint list shows the active microsprint
	// Note: 'microsprint list' shows microsprints (with tracker), not issues within them
	t.Run("verify microsprint list", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "microsprint", "list")
		assertExitCode(t, result, 0)
		// Check that the microsprint name appears in the list
		assertContains(t, result.Stdout, sprintName)
		assertContains(t, result.Stdout, "Active")
	})

	// Step 5: Close microsprint with --skip-retro flag
	t.Run("close microsprint", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "microsprint", "close", "--skip-retro")
		assertExitCode(t, result, 0)
	})

	// Step 6: Verify no current microsprint
	t.Run("verify no current microsprint", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "microsprint", "current")
		// Should fail or show no active microsprint
		if result.ExitCode == 0 {
			assertNotContains(t, result.Stdout, sprintName)
		}
	})
}
