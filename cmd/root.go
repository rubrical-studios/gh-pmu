package cmd

import (
	"github.com/spf13/cobra"
)

var version = "dev"

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gh-pmu",
		Short:   "GitHub Project Management CLI (Unified)",
		Long:    `gh-pmu is a unified GitHub CLI extension for project management, sub-issue hierarchy, and project templating.

Note: This is the development version. It will replace 'gh pm' and 'gh sub-issue' when complete.`,
		Version: version,
	}

	cmd.AddCommand(newInitCommand())
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newViewCommand())
	cmd.AddCommand(newCreateCommand())
	cmd.AddCommand(newMoveCommand())
	cmd.AddCommand(newSubCommand())

	return cmd
}

func Execute() error {
	return NewRootCommand().Execute()
}
