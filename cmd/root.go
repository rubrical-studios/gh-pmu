package cmd

import (
	"github.com/spf13/cobra"
)

var version = "dev"

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
		Version: version,
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

	return cmd
}

func Execute() error {
	return NewRootCommand().Execute()
}
