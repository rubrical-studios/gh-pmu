package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// semverRegex matches valid semver versions with optional v prefix
var semverRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)$`)

// validateVersion validates that a version string is valid semver format
// Accepts X.Y.Z or vX.Y.Z format
func validateVersion(version string) error {
	if !semverRegex.MatchString(version) {
		return fmt.Errorf("Invalid version format. Use semver: X.Y.Z")
	}
	return nil
}

// compareVersions compares two semver versions
// Returns: positive if v1 > v2, negative if v1 < v2, zero if equal
func compareVersions(v1, v2 string) int {
	// Strip 'v' prefix
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < 3; i++ {
		var n1, n2 int
		if i < len(parts1) {
			_, _ = fmt.Sscanf(parts1[i], "%d", &n1)
		}
		if i < len(parts2) {
			_, _ = fmt.Sscanf(parts2[i], "%d", &n2)
		}
		if n1 != n2 {
			return n1 - n2
		}
	}
	return 0
}

// nextVersions contains calculated next version options
type nextVersions struct {
	patch string
	minor string
	major string
}

// calculateNextVersions computes the next patch, minor, and major versions
func calculateNextVersions(currentVersion string) (*nextVersions, error) {
	// Strip 'v' prefix for parsing
	version := strings.TrimPrefix(currentVersion, "v")
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid version format: %s", currentVersion)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", parts[0])
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", parts[1])
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version: %s", parts[2])
	}

	return &nextVersions{
		patch: fmt.Sprintf("v%d.%d.%d", major, minor, patch+1),
		minor: fmt.Sprintf("v%d.%d.0", major, minor+1),
		major: fmt.Sprintf("v%d.0.0", major+1),
	}, nil
}

// deprecationWarning prints a deprecation warning to stderr
func deprecationWarning(cmd *cobra.Command, newCmd string) {
	fmt.Fprintf(os.Stderr, "WARNING: 'gh pmu release %s' is deprecated. Use 'gh pmu branch %s' instead.\n\n", cmd.Name(), newCmd)
}

// newReleaseCommand creates the release command group (deprecated wrapper for branch)
func newReleaseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "release",
		Short:      "[DEPRECATED] Use 'gh pmu branch' instead",
		Long:       `DEPRECATED: The release command has been renamed to branch. Please use 'gh pmu branch' instead.`,
		Deprecated: "use 'gh pmu branch' instead",
	}

	// Add deprecated subcommands that wrap branch subcommands
	cmd.AddCommand(newDeprecatedReleaseStartCommand())
	cmd.AddCommand(newDeprecatedReleaseAddCommand())
	cmd.AddCommand(newDeprecatedReleaseRemoveCommand())
	cmd.AddCommand(newDeprecatedReleaseCurrentCommand())
	cmd.AddCommand(newDeprecatedReleaseCloseCommand())
	cmd.AddCommand(newDeprecatedReleaseReopenCommand())
	cmd.AddCommand(newDeprecatedReleaseListCommand())

	return cmd
}

// newDeprecatedReleaseStartCommand wraps branch start
func newDeprecatedReleaseStartCommand() *cobra.Command {
	var branchName string

	cmd := &cobra.Command{
		Use:        "start",
		Short:      "[DEPRECATED] Use 'gh pmu branch start' instead",
		Deprecated: "use 'gh pmu branch start' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			deprecationWarning(cmd, "start --name "+branchName)
			// Get the branch command and run it
			branchCmd := newBranchCommand()
			branchCmd.SetArgs([]string{"start", "--name", branchName})
			return branchCmd.Execute()
		},
	}

	cmd.Flags().StringVar(&branchName, "branch", "", "Branch name (required)")
	_ = cmd.MarkFlagRequired("branch")

	return cmd
}

// newDeprecatedReleaseAddCommand wraps branch add
func newDeprecatedReleaseAddCommand() *cobra.Command {
	var issueNumber int

	cmd := &cobra.Command{
		Use:        "add",
		Short:      "[DEPRECATED] Use 'gh pmu branch add' instead",
		Deprecated: "use 'gh pmu branch add' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			deprecationWarning(cmd, fmt.Sprintf("add %d", issueNumber))
			branchCmd := newBranchCommand()
			branchCmd.SetArgs([]string{"add", fmt.Sprintf("%d", issueNumber)})
			return branchCmd.Execute()
		},
	}

	cmd.Flags().IntVar(&issueNumber, "issue", 0, "Issue number to add (required)")
	_ = cmd.MarkFlagRequired("issue")

	return cmd
}

// newDeprecatedReleaseRemoveCommand wraps branch remove
func newDeprecatedReleaseRemoveCommand() *cobra.Command {
	var issueNumber int

	cmd := &cobra.Command{
		Use:        "remove",
		Short:      "[DEPRECATED] Use 'gh pmu branch remove' instead",
		Deprecated: "use 'gh pmu branch remove' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			deprecationWarning(cmd, fmt.Sprintf("remove %d", issueNumber))
			branchCmd := newBranchCommand()
			branchCmd.SetArgs([]string{"remove", fmt.Sprintf("%d", issueNumber)})
			return branchCmd.Execute()
		},
	}

	cmd.Flags().IntVar(&issueNumber, "issue", 0, "Issue number to remove (required)")
	_ = cmd.MarkFlagRequired("issue")

	return cmd
}

// newDeprecatedReleaseCurrentCommand wraps branch current
func newDeprecatedReleaseCurrentCommand() *cobra.Command {
	var refresh bool

	cmd := &cobra.Command{
		Use:        "current",
		Short:      "[DEPRECATED] Use 'gh pmu branch current' instead",
		Deprecated: "use 'gh pmu branch current' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			deprecationWarning(cmd, "current")
			branchCmd := newBranchCommand()
			cmdArgs := []string{"current"}
			if refresh {
				cmdArgs = append(cmdArgs, "--refresh")
			}
			branchCmd.SetArgs(cmdArgs)
			return branchCmd.Execute()
		},
	}

	cmd.Flags().BoolVar(&refresh, "refresh", false, "Refresh from remote")

	return cmd
}

// newDeprecatedReleaseCloseCommand wraps branch close
func newDeprecatedReleaseCloseCommand() *cobra.Command {
	var tag, yes bool
	var branchName string

	cmd := &cobra.Command{
		Use:        "close [branch]",
		Short:      "[DEPRECATED] Use 'gh pmu branch close' instead",
		Deprecated: "use 'gh pmu branch close' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			deprecationWarning(cmd, "close")
			branchCmd := newBranchCommand()
			cmdArgs := []string{"close"}
			if len(args) > 0 {
				cmdArgs = append(cmdArgs, args[0])
			}
			if tag {
				cmdArgs = append(cmdArgs, "--tag")
			}
			if yes {
				cmdArgs = append(cmdArgs, "--yes")
			}
			branchCmd.SetArgs(cmdArgs)
			return branchCmd.Execute()
		},
	}

	cmd.Flags().BoolVar(&tag, "tag", false, "Create git tag")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation")
	cmd.Flags().StringVar(&branchName, "name", "", "Branch name to close")

	return cmd
}

// newDeprecatedReleaseReopenCommand wraps branch reopen
func newDeprecatedReleaseReopenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "reopen <branch>",
		Short:      "[DEPRECATED] Use 'gh pmu branch reopen' instead",
		Deprecated: "use 'gh pmu branch reopen' instead",
		Args:       cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deprecationWarning(cmd, "reopen "+args[0])
			branchCmd := newBranchCommand()
			branchCmd.SetArgs([]string{"reopen", args[0]})
			return branchCmd.Execute()
		},
	}

	return cmd
}

// newDeprecatedReleaseListCommand wraps branch list
func newDeprecatedReleaseListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "list",
		Short:      "[DEPRECATED] Use 'gh pmu branch list' instead",
		Deprecated: "use 'gh pmu branch list' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			deprecationWarning(cmd, "list")
			branchCmd := newBranchCommand()
			branchCmd.SetArgs([]string{"list"})
			return branchCmd.Execute()
		},
	}

	return cmd
}
