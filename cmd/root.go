package cmd

import (
	"fmt"
	"os"

	"github.com/rubrical-studios/gh-pmu/internal/config"
	pkgversion "github.com/rubrical-studios/gh-pmu/internal/version"
	"github.com/spf13/cobra"
)

// version is set by ldflags during goreleaser builds.
// When empty (default), falls back to the source constant in internal/version.
var version = ""

func getVersion() string {
	if version != "" {
		return version
	}
	return pkgversion.Version
}

// exemptCommands are commands that do not require terms acceptance.
var exemptCommands = map[string]bool{
	"init":   true,
	"accept": true,
	"help":   true,
}

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gh pmu",
		Short: "Streamline GitHub project workflows",
		Long: `gh pmu streamlines GitHub project workflows with unified issue tracking and sub-issue hierarchy.

Designed for Kanban-style GitHub Projects with status-based columns
(Backlog, In Progress, In Review, Done). Works seamlessly with the
IDPF-Praxis framework for structured development workflows.

This extension combines and replaces:
  - gh-pm (https://github.com/yahsan2/gh-pm) - Project management
  - gh-sub-issue (https://github.com/yahsan2/gh-sub-issue) - Sub-issue hierarchy

Use 'gh pmu <command> --help' for more information about a command.`,
		Version: getVersion(),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return checkAcceptance(cmd)
		},
	}

	cmd.AddCommand(newInitCommand())
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newViewCommand())
	cmd.AddCommand(newCreateCommand())
	cmd.AddCommand(newEditCommand())
	cmd.AddCommand(newCommentCommand())
	cmd.AddCommand(newMoveCommand())
	cmd.AddCommand(newCloseCommand())
	cmd.AddCommand(newBoardCommand())
	cmd.AddCommand(newSubCommand())
	cmd.AddCommand(newFieldCommand())
	cmd.AddCommand(newIntakeCommand())
	cmd.AddCommand(newTriageCommand())
	cmd.AddCommand(newSplitCommand())
	cmd.AddCommand(newHistoryCommand())
	cmd.AddCommand(newFilterCommand())
	cmd.AddCommand(newBranchCommand())
	cmd.AddCommand(newAcceptCommand())

	return cmd
}

func Execute() error {
	return NewRootCommand().Execute()
}

// checkAcceptance verifies terms have been accepted before running commands.
func checkAcceptance(cmd *cobra.Command) error {
	// Dev/source builds skip acceptance gate — only ldflags-injected builds enforce it
	if version == "" {
		return nil
	}

	// Check if this is an exempt command
	name := cmd.Name()
	if exemptCommands[name] {
		return nil
	}

	// --help flag on any command is always allowed
	if help, _ := cmd.Flags().GetBool("help"); help {
		return nil
	}

	// Try to load config — if no config exists, skip gate (init not run yet)
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}

	cfg, err := config.LoadFromDirectory(cwd)
	if err != nil {
		// No config file — not initialized yet, skip acceptance check
		return nil
	}

	// Check acceptance state
	if cfg.Acceptance == nil || !cfg.Acceptance.Accepted {
		return fmt.Errorf("terms not accepted — run 'gh pmu accept' first")
	}

	// Check version — re-acceptance needed on major/minor bump
	if config.RequiresReAcceptance(cfg.Acceptance.Version, getVersion()) {
		return fmt.Errorf("terms acceptance outdated (accepted v%s, current v%s) — run 'gh pmu accept' to re-accept",
			cfg.Acceptance.Version, getVersion())
	}

	return nil
}
