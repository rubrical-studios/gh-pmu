//go:build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"
)

// TestMicrosprintStartWithBranch tests starting a microsprint with --branch flag.
// Verifies the branch context is included in the tracker title and field.
func TestMicrosprintStartWithBranch(t *testing.T) {
	cfg := setupTestConfig(t)

	// Generate unique names
	branchName := fmt.Sprintf("release/e2e-ms-branch-%d", time.Now().UnixNano())
	sprintSuffix := fmt.Sprintf("branch-test-%d", time.Now().UnixNano()%10000)

	// Track resources for cleanup
	var trackerIssueNum int
	var branchTrackerNum int

	// Cleanup at end of test
	defer func() {
		if trackerIssueNum > 0 {
			runCleanupAfterTest(t, trackerIssueNum)
		}
		if branchTrackerNum > 0 {
			runCleanupAfterTest(t, branchTrackerNum)
		}
		// Ensure microsprint and branch are closed
		runPMU(t, cfg.Dir, "microsprint", "close", "--skip-retro")
		runPMU(t, cfg.Dir, "branch", "close", "--yes")
	}()

	// Step 1: Start a branch first (required context for --branch flag)
	t.Run("start branch", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "branch", "start", "--name", branchName)
		assertExitCode(t, result, 0)
		branchTrackerNum = extractIssueNumber(t, result.Stdout)
	})

	// Step 2: Start microsprint with --branch flag
	t.Run("start microsprint with branch", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "microsprint", "start", "--name", sprintSuffix, "--branch", branchName)
		assertExitCode(t, result, 0)

		// Verify branch context is mentioned in output
		assertContains(t, result.Stdout, branchName)

		// Extract tracker issue number
		trackerIssueNum = extractIssueNumber(t, result.Stdout)
	})

	// Step 3: Verify microsprint tracker has branch context in title
	t.Run("verify branch in tracker title", func(t *testing.T) {
		if trackerIssueNum == 0 {
			t.Skip("No tracker issue number available")
		}

		// View the microsprint current to check the title includes branch
		result := runPMU(t, cfg.Dir, "microsprint", "current")
		assertExitCode(t, result, 0)

		// Title should contain branch name in brackets: [release/e2e-ms-branch-xxx]
		assertContains(t, result.Stdout, branchName)
	})

	// Step 4: Clean up - close microsprint
	t.Run("close microsprint", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "microsprint", "close", "--skip-retro")
		assertExitCode(t, result, 0)
	})
}

// TestMicrosprintLifecycle tests the complete microsprint workflow:
// start -> add issue -> current -> list -> close
func TestMicrosprintLifecycle(t *testing.T) {
	cfg := setupTestConfig(t)

	// Generate unique microsprint name with timestamp
	sprintName := fmt.Sprintf("e2e-test-%d", time.Now().UnixNano())

	// Track resources for cleanup
	var testIssueNum int
	var trackerIssueNum int

	// Cleanup at end of test
	defer func() {
		if testIssueNum > 0 {
			runCleanupAfterTest(t, testIssueNum)
		}
		if trackerIssueNum > 0 {
			runCleanupAfterTest(t, trackerIssueNum)
		}
		// Ensure microsprint is closed even if test fails
		runPMU(t, cfg.Dir, "microsprint", "close", "--skip-retro")
	}()

	// Step 1: Start a new microsprint with --name flag
	t.Run("start microsprint", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "microsprint", "start", "--name", sprintName)
		assertExitCode(t, result, 0)
		assertContains(t, result.Stdout, sprintName)
		// Extract tracker issue number for label verification and cleanup
		trackerIssueNum = extractIssueNumber(t, result.Stdout)
	})

	// Step 1b: Verify tracker issue has 'microsprint' label
	t.Run("verify microsprint label", func(t *testing.T) {
		if trackerIssueNum == 0 {
			t.Skip("No tracker issue number available")
		}
		assertHasLabel(t, trackerIssueNum, "microsprint")
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

		// Add to microsprint (with retry for GitHub eventual consistency)
		// The issue may not appear in project queries immediately after creation
		var result *CommandResult
		for i := 0; i < 5; i++ {
			result = runPMU(t, cfg.Dir, "microsprint", "add", fmt.Sprintf("%d", testIssueNum))
			if result.ExitCode == 0 {
				break
			}
			if i < 4 {
				t.Logf("Retry %d/5: waiting for issue to appear in project...", i+1)
				time.Sleep(2 * time.Second)
			}
		}
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
