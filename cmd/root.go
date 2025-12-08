package cmd

import (
	"github.com/spf13/cobra"
)

var version = "dev"

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gh-pmu",
		Short: "GitHub Project Management CLI (Unified)",
		Long: `gh-pmu is a unified GitHub CLI extension for project management, sub-issue hierarchy, and project templating.

This extension combines and replaces:
  - gh-pm (https://github.com/yahsan2/gh-pm) - Project management
  - gh-sub-issue (https://github.com/yahsan2/gh-sub-issue) - Sub-issue hierarchy

Use 'gh pmu <command> --help' for more information about a command.`,
		Version: version,
	}

	cmd.AddCommand(newInitCommand())
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newViewCommand())
	cmd.AddCommand(newCreateCommand())
	cmd.AddCommand(newMoveCommand())
	cmd.AddCommand(newCloseCommand())
	cmd.AddCommand(newSubCommand())
	cmd.AddCommand(newIntakeCommand())
	cmd.AddCommand(newTriageCommand())
	cmd.AddCommand(newSplitCommand())

	return cmd
}

func Execute() error {
	return NewRootCommand().Execute()
}
