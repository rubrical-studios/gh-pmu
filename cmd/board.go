package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rubrical-studios/gh-pmu/internal/api"
	"github.com/rubrical-studios/gh-pmu/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// boardClient defines the interface for API methods used by board functions.
// This allows for easier testing with mock implementations.
type boardClient interface {
	GetProject(owner string, number int) (*api.Project, error)
	GetProjectItems(projectID string, filter *api.ProjectItemsFilter) ([]api.ProjectItem, error)
}

type boardOptions struct {
	status   string
	priority string
	limit    int
	noBorder bool
	json     bool
}

// Box drawing characters
const (
	boardTopLeft     = "┌"
	boardTopRight    = "┐"
	boardBottomLeft  = "└"
	boardBottomRight = "┘"
	boardHorizontal  = "─"
	boardVertical    = "│"
	boardTopTee      = "┬"
	boardBottomTee   = "┴"
	boardCross       = "┼"
	boardLeftTee     = "├"
	boardRightTee    = "┤"
)

func newBoardCommand() *cobra.Command {
	opts := &boardOptions{
		limit: 10,
	}

	cmd := &cobra.Command{
		Use:   "board",
		Short: "Display issues in a columnar board view",
		Long: `Display issues grouped by status in a terminal board view.

Shows issues organized in columns by their status field, similar to
a Kanban board. Each column displays issue numbers and truncated titles.

Examples:
  # Show full board with all status columns
  gh pmu board

  # Show only a single status column
  gh pmu board --status in_progress

  # Filter by priority across all columns
  gh pmu board --priority p0

  # Limit items per column
  gh pmu board --limit 5

  # Output without borders (simpler display)
  gh pmu board --no-border

  # Output as JSON grouped by status
  gh pmu board --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBoard(cmd, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.status, "status", "s", "", "Show only specified status column")
	cmd.Flags().StringVarP(&opts.priority, "priority", "p", "", "Filter by priority")
	cmd.Flags().IntVarP(&opts.limit, "limit", "n", 10, "Limit issues per column")
	cmd.Flags().BoolVar(&opts.noBorder, "no-border", false, "Display without box borders")
	cmd.Flags().BoolVar(&opts.json, "json", false, "Output as JSON grouped by status")

	return cmd
}

func runBoard(cmd *cobra.Command, opts *boardOptions) error {
	// Load configuration
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	cfg, err := config.LoadFromDirectory(cwd)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w\nRun 'gh pmu init' to create a configuration file", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create API client
	client := api.NewClient()

	return runBoardWithDeps(cmd, opts, cfg, client)
}

// runBoardWithDeps is the testable implementation of runBoard
func runBoardWithDeps(cmd *cobra.Command, opts *boardOptions, cfg *config.Config, client boardClient) error {
	// Get project
	project, err := client.GetProject(cfg.Project.Owner, cfg.Project.Number)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Build filter
	var filter *api.ProjectItemsFilter
	if len(cfg.Repositories) > 0 {
		filter = &api.ProjectItemsFilter{
			Repository: cfg.Repositories[0],
		}
	}

	// Fetch project items
	items, err := client.GetProjectItems(project.ID, filter)
	if err != nil {
		return fmt.Errorf("failed to get project items: %w", err)
	}

	// Apply priority filter if specified
	if opts.priority != "" {
		targetPriority := cfg.ResolveFieldValue("priority", opts.priority)
		items = filterByFieldValue(items, "Priority", targetPriority)
	}

	// Get status columns from config
	columns := getStatusColumns(cfg)

	// If --status is specified, filter to single column
	if opts.status != "" {
		targetStatus := cfg.ResolveFieldValue("status", opts.status)
		filtered := []statusColumn{}
		for _, col := range columns {
			if strings.EqualFold(col.value, targetStatus) {
				filtered = append(filtered, col)
				break
			}
		}
		if len(filtered) == 0 {
			// If not found in config, create a column for the raw value
			filtered = append(filtered, statusColumn{alias: opts.status, value: targetStatus})
		}
		columns = filtered
	}

	// Group items by status
	grouped := groupByStatus(items, columns)

	// Apply limit per column
	for status, columnItems := range grouped {
		if opts.limit > 0 && len(columnItems) > opts.limit {
			grouped[status] = columnItems[:opts.limit]
		}
	}

	// Output
	if opts.json {
		return outputBoardJSON(cmd, grouped, columns)
	}

	if opts.noBorder {
		return outputBoardSimple(cmd, grouped, columns)
	}

	return outputBoardBox(cmd, grouped, columns, opts.limit)
}

// statusColumn represents a status column for the board
type statusColumn struct {
	alias string
	value string
}

// getStatusColumns extracts status columns from config in order
func getStatusColumns(cfg *config.Config) []statusColumn {
	var columns []statusColumn

	// Try to get from config fields
	if statusField, ok := cfg.Fields["status"]; ok && len(statusField.Values) > 0 {
		// Note: Go maps are unordered, so we'll use a predefined order
		// that matches common workflow patterns
		preferredOrder := []string{"backlog", "ready", "in_progress", "in_review", "done"}

		for _, alias := range preferredOrder {
			if value, ok := statusField.Values[alias]; ok {
				columns = append(columns, statusColumn{alias: alias, value: value})
			}
		}

		// Add any remaining statuses not in preferred order
		for alias, value := range statusField.Values {
			found := false
			for _, col := range columns {
				if col.alias == alias {
					found = true
					break
				}
			}
			if !found {
				columns = append(columns, statusColumn{alias: alias, value: value})
			}
		}
	}

	// Fallback to defaults if no config
	if len(columns) == 0 {
		columns = []statusColumn{
			{alias: "backlog", value: "Backlog"},
			{alias: "in_progress", value: "In progress"},
			{alias: "in_review", value: "In review"},
			{alias: "done", value: "Done"},
		}
	}

	return columns
}

// groupByStatus groups items by their status field value
func groupByStatus(items []api.ProjectItem, columns []statusColumn) map[string][]api.ProjectItem {
	grouped := make(map[string][]api.ProjectItem)

	// Initialize all columns
	for _, col := range columns {
		grouped[col.value] = []api.ProjectItem{}
	}

	// Group items
	for _, item := range items {
		if item.Issue == nil {
			continue
		}
		status := getFieldValue(item, "Status")
		if status == "" {
			status = "(none)"
		}
		grouped[status] = append(grouped[status], item)
	}

	return grouped
}

// getTerminalWidth returns the terminal width or a default
func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 120 // default width
	}
	return width
}

// truncateString truncates a string to maxLen with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// outputBoardBox outputs the board with box drawing characters
func outputBoardBox(cmd *cobra.Command, grouped map[string][]api.ProjectItem, columns []statusColumn, limit int) error {
	termWidth := getTerminalWidth()
	numCols := len(columns)
	if numCols == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No status columns configured")
		return nil
	}

	// Calculate column width
	colWidth := (termWidth - numCols - 1) / numCols
	if colWidth < 15 {
		colWidth = 15
	}
	if colWidth > 30 {
		colWidth = 30
	}

	// Find max rows needed
	maxRows := 0
	for _, col := range columns {
		items := grouped[col.value]
		if len(items) > maxRows {
			maxRows = len(items)
		}
	}
	if limit > 0 && maxRows > limit {
		maxRows = limit
	}

	out := cmd.OutOrStdout()

	// Top border
	fmt.Fprint(out, boardTopLeft)
	for i := range columns {
		fmt.Fprint(out, strings.Repeat(boardHorizontal, colWidth))
		if i < numCols-1 {
			fmt.Fprint(out, boardTopTee)
		}
	}
	fmt.Fprintln(out, boardTopRight)

	// Header row
	fmt.Fprint(out, boardVertical)
	for _, col := range columns {
		items := grouped[col.value]
		header := fmt.Sprintf("%s (%d)", col.value, len(items))
		header = truncateString(header, colWidth-2)
		padding := colWidth - len(header) - 1
		if padding < 0 {
			padding = 0
		}
		fmt.Fprintf(out, " %s%s", header, strings.Repeat(" ", padding))
		fmt.Fprint(out, boardVertical)
	}
	fmt.Fprintln(out)

	// Header separator
	fmt.Fprint(out, boardLeftTee)
	for i := range columns {
		fmt.Fprint(out, strings.Repeat(boardHorizontal, colWidth))
		if i < numCols-1 {
			fmt.Fprint(out, boardCross)
		}
	}
	fmt.Fprintln(out, boardRightTee)

	// Data rows
	for row := 0; row < maxRows; row++ {
		fmt.Fprint(out, boardVertical)
		for _, col := range columns {
			items := grouped[col.value]
			var cell string
			if row < len(items) {
				item := items[row]
				if item.Issue != nil {
					cell = fmt.Sprintf("#%d %s", item.Issue.Number, item.Issue.Title)
				}
			}
			cell = truncateString(cell, colWidth-2)
			padding := colWidth - len(cell) - 1
			if padding < 0 {
				padding = 0
			}
			fmt.Fprintf(out, " %s%s", cell, strings.Repeat(" ", padding))
			fmt.Fprint(out, boardVertical)
		}
		fmt.Fprintln(out)
	}

	// Bottom border
	fmt.Fprint(out, boardBottomLeft)
	for i := range columns {
		fmt.Fprint(out, strings.Repeat(boardHorizontal, colWidth))
		if i < numCols-1 {
			fmt.Fprint(out, boardBottomTee)
		}
	}
	fmt.Fprintln(out, boardBottomRight)

	return nil
}

// outputBoardSimple outputs the board without borders
func outputBoardSimple(cmd *cobra.Command, grouped map[string][]api.ProjectItem, columns []statusColumn) error {
	out := cmd.OutOrStdout()

	for _, col := range columns {
		items := grouped[col.value]
		fmt.Fprintf(out, "\n## %s (%d)\n", col.value, len(items))
		if len(items) == 0 {
			fmt.Fprintln(out, "  (empty)")
			continue
		}
		for _, item := range items {
			if item.Issue != nil {
				fmt.Fprintf(out, "  #%d %s\n", item.Issue.Number, item.Issue.Title)
			}
		}
	}
	fmt.Fprintln(out)

	return nil
}

// outputBoardJSON outputs the board as JSON
func outputBoardJSON(cmd *cobra.Command, grouped map[string][]api.ProjectItem, columns []statusColumn) error {
	type jsonIssue struct {
		Number   int    `json:"number"`
		Title    string `json:"title"`
		Priority string `json:"priority,omitempty"`
	}
	type jsonColumn struct {
		Status string      `json:"status"`
		Count  int         `json:"count"`
		Issues []jsonIssue `json:"issues"`
	}

	var output []jsonColumn
	for _, col := range columns {
		items := grouped[col.value]
		jc := jsonColumn{
			Status: col.value,
			Count:  len(items),
			Issues: []jsonIssue{},
		}
		for _, item := range items {
			if item.Issue != nil {
				jc.Issues = append(jc.Issues, jsonIssue{
					Number:   item.Issue.Number,
					Title:    item.Issue.Title,
					Priority: getFieldValue(item, "Priority"),
				})
			}
		}
		output = append(output, jc)
	}

	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}
