//go:build e2e

package e2e

import (
	"fmt"
	"testing"
)

// TestCreateToCloseWorkflow tests the complete issue lifecycle:
// create -> backlog -> in_progress -> in_review -> done
func TestCreateToCloseWorkflow(t *testing.T) {
	cfg := setupTestConfig(t)

	var issueNum int

	// Cleanup at end of test
	defer func() {
		if issueNum > 0 {
			runCleanupAfterTest(t, issueNum)
		}
	}()

	// Step 1: Create new issue with --title, --status, --priority
	t.Run("create issue", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "create",
			"--title", "[E2E] Create-to-Close Workflow Test",
			"--status", "backlog",
			"--priority", "p2",
		)
		assertExitCode(t, result, 0)
		issueNum = extractIssueNumber(t, result.Stdout)
		t.Logf("Created issue #%d", issueNum)
	})

	// Step 2: Move through workflow sequentially (not parallel)
	// backlog -> in_progress
	t.Run("move to in_progress", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "move", fmt.Sprintf("%d", issueNum), "--status", "in_progress")
		assertExitCode(t, result, 0)
		assertContains(t, result.Stdout, "In progress")
	})

	// in_progress -> in_review
	t.Run("move to in_review", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "move", fmt.Sprintf("%d", issueNum), "--status", "in_review")
		assertExitCode(t, result, 0)
		assertContains(t, result.Stdout, "In review")
	})

	// in_review -> done
	t.Run("move to done", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "move", fmt.Sprintf("%d", issueNum), "--status", "done")
		assertExitCode(t, result, 0)
		assertContains(t, result.Stdout, "Done")
	})

	// Step 3: Verify final state shows "Done" status
	t.Run("verify final state", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "view", fmt.Sprintf("%d", issueNum), "--json")
		assertExitCode(t, result, 0)
		assertContains(t, result.Stdout, "Done")
	})
}

// TestSubIssueWorkflow tests sub-issue operations:
// create parent -> create sub -> list subs -> remove sub
func TestSubIssueWorkflow(t *testing.T) {
	cfg := setupTestConfig(t)

	var parentIssueNum, subIssueNum int

	// Cleanup at end of test
	defer func() {
		if subIssueNum > 0 {
			runCleanupAfterTest(t, subIssueNum)
		}
		if parentIssueNum > 0 {
			runCleanupAfterTest(t, parentIssueNum)
		}
	}()

	// Step 1: Create parent issue
	t.Run("create parent issue", func(t *testing.T) {
		parentIssueNum = createTestIssue(t, cfg, "Sub-Issue Workflow - Parent")
		t.Logf("Created parent issue #%d", parentIssueNum)
	})

	// Step 2: Create sub-issue via sub create --parent
	t.Run("create sub-issue", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "sub", "create",
			"--parent", fmt.Sprintf("%d", parentIssueNum),
			"--title", "[E2E] Sub-Issue Workflow - Child",
		)
		assertExitCode(t, result, 0)
		subIssueNum = extractIssueNumber(t, result.Stdout)
		t.Logf("Created sub-issue #%d", subIssueNum)
	})

	// Step 3: Verify sub list shows the sub-issue
	t.Run("verify sub list", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "sub", "list", fmt.Sprintf("%d", parentIssueNum))
		assertExitCode(t, result, 0)
		assertContains(t, result.Stdout, fmt.Sprintf("#%d", subIssueNum))
	})

	// Step 4: Remove sub-issue via sub remove
	t.Run("remove sub-issue", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "sub", "remove",
			fmt.Sprintf("%d", parentIssueNum),
			fmt.Sprintf("%d", subIssueNum),
		)
		assertExitCode(t, result, 0)
	})

	// Step 5: Verify removal succeeded (sub no longer in list)
	t.Run("verify removal", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "sub", "list", fmt.Sprintf("%d", parentIssueNum))
		// Either the list is empty or the sub-issue is not present
		if result.ExitCode == 0 {
			assertNotContains(t, result.Stdout, fmt.Sprintf("#%d", subIssueNum))
		}
	})
}

// TestMultiIssueMoveWorkflow tests batch issue moves:
// create multiple issues -> move all in single command -> verify all updated
func TestMultiIssueMoveWorkflow(t *testing.T) {
	cfg := setupTestConfig(t)

	var issueNums []int

	// Cleanup at end of test
	defer func() {
		for _, num := range issueNums {
			runCleanupAfterTest(t, num)
		}
	}()

	// Step 1: Create multiple issues
	t.Run("create multiple issues", func(t *testing.T) {
		for i := 1; i <= 3; i++ {
			issueNum := createTestIssue(t, cfg, fmt.Sprintf("Multi-Move Test #%d", i))
			issueNums = append(issueNums, issueNum)
			t.Logf("Created issue #%d", issueNum)
		}
	})

	// Step 2: Move multiple issues in single command
	t.Run("move multiple issues", func(t *testing.T) {
		args := []string{"move"}
		for _, num := range issueNums {
			args = append(args, fmt.Sprintf("%d", num))
		}
		// --yes flag required for multi-issue moves (skips confirmation prompt)
		args = append(args, "--status", "in_progress", "--yes")

		result := runPMU(t, cfg.Dir, args...)
		t.Logf("Move command output:\nStdout: %s\nStderr: %s", result.Stdout, result.Stderr)
		assertExitCode(t, result, 0)

		// Verify all issues mentioned in output
		for _, num := range issueNums {
			assertContains(t, result.Stdout, fmt.Sprintf("#%d", num))
		}
	})

	// Step 3: Verify all issues have updated status (with retry for eventual consistency)
	// Use 10 retries with 2-second intervals for GitHub's eventual consistency
	t.Run("verify all issues updated", func(t *testing.T) {
		for _, num := range issueNums {
			// Use retry logic for eventual consistency (10 retries, 2s each = 20s max)
			result := waitForProjectSync(t, cfg, 10,
				[]string{"view", fmt.Sprintf("%d", num), "--json"},
				"In progress",
			)
			assertExitCode(t, result, 0)
			assertContains(t, result.Stdout, "In progress")
		}
	})
}

// TestCreateWithLabel verifies the --label flag applies labels correctly
func TestCreateWithLabel(t *testing.T) {
	cfg := setupTestConfig(t)

	var issueNum int

	// Cleanup at end of test
	defer func() {
		if issueNum > 0 {
			runCleanupAfterTest(t, issueNum)
		}
	}()

	// Step 1: Create issue with --label bug
	t.Run("create issue with label", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "create",
			"--title", "[E2E] Create With Label Test",
			"--status", "backlog",
			"--label", "bug",
		)
		assertExitCode(t, result, 0)
		issueNum = extractIssueNumber(t, result.Stdout)
		t.Logf("Created issue #%d with bug label", issueNum)
	})

	// Step 2: Verify label was applied
	t.Run("verify label applied", func(t *testing.T) {
		if issueNum == 0 {
			t.Skip("No issue number available")
		}
		assertHasLabel(t, issueNum, "bug")
	})
}

// TestCreateWithWorkflowLabels verifies workflow labels (epic, story, enhancement) can be applied
func TestCreateWithWorkflowLabels(t *testing.T) {
	cfg := setupTestConfig(t)

	var epicNum, storyNum, enhancementNum int

	// Cleanup at end of test
	defer func() {
		if enhancementNum > 0 {
			runCleanupAfterTest(t, enhancementNum)
		}
		if storyNum > 0 {
			runCleanupAfterTest(t, storyNum)
		}
		if epicNum > 0 {
			runCleanupAfterTest(t, epicNum)
		}
	}()

	// Test epic label
	t.Run("create issue with epic label", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "create",
			"--title", "[E2E] Epic Label Test",
			"--status", "backlog",
			"--label", "epic",
		)
		assertExitCode(t, result, 0)
		epicNum = extractIssueNumber(t, result.Stdout)
		t.Logf("Created issue #%d with epic label", epicNum)
	})

	t.Run("verify epic label applied", func(t *testing.T) {
		if epicNum == 0 {
			t.Skip("No epic issue number available")
		}
		assertHasLabel(t, epicNum, "epic")
	})

	// Test story label
	t.Run("create issue with story label", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "create",
			"--title", "[E2E] Story Label Test",
			"--status", "backlog",
			"--label", "story",
		)
		assertExitCode(t, result, 0)
		storyNum = extractIssueNumber(t, result.Stdout)
		t.Logf("Created issue #%d with story label", storyNum)
	})

	t.Run("verify story label applied", func(t *testing.T) {
		if storyNum == 0 {
			t.Skip("No story issue number available")
		}
		assertHasLabel(t, storyNum, "story")
	})

	// Test enhancement label
	t.Run("create issue with enhancement label", func(t *testing.T) {
		result := runPMU(t, cfg.Dir, "create",
			"--title", "[E2E] Enhancement Label Test",
			"--status", "backlog",
			"--label", "enhancement",
		)
		assertExitCode(t, result, 0)
		enhancementNum = extractIssueNumber(t, result.Stdout)
		t.Logf("Created issue #%d with enhancement label", enhancementNum)
	})

	t.Run("verify enhancement label applied", func(t *testing.T) {
		if enhancementNum == 0 {
			t.Skip("No enhancement issue number available")
		}
		assertHasLabel(t, enhancementNum, "enhancement")
	})
}
