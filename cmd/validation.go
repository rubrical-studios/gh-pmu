package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rubrical-studios/gh-pmu/internal/api"
	"github.com/rubrical-studios/gh-pmu/internal/config"
)

// Regex patterns for checkbox detection
var (
	checkedBoxRegex   = regexp.MustCompile(`- \[x\]`)
	uncheckedBoxRegex = regexp.MustCompile(`- \[ \]`)
)

// ValidationError represents a validation failure with actionable message
type ValidationError struct {
	IssueNumber int
	Message     string
	Suggestion  string
}

func (e *ValidationError) Error() string {
	if e.Suggestion != "" {
		return fmt.Sprintf("Issue #%d: %s\n\n%s", e.IssueNumber, e.Message, e.Suggestion)
	}
	return fmt.Sprintf("Issue #%d: %s", e.IssueNumber, e.Message)
}

// ValidationErrors collects multiple validation failures
type ValidationErrors struct {
	Errors []ValidationError
}

func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return ""
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Validation failed for %d issues:\n", len(e.Errors)))
	for _, err := range e.Errors {
		sb.WriteString(fmt.Sprintf("\n  - Issue #%d: %s", err.IssueNumber, err.Message))
	}
	return sb.String()
}

func (e *ValidationErrors) Add(err ValidationError) {
	e.Errors = append(e.Errors, err)
}

func (e *ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// issueValidationContext holds all info needed to validate an issue
type issueValidationContext struct {
	Number         int
	CurrentStatus  string
	CurrentRelease string
	Body           string
	ActiveReleases []string // Discovered from GitHub release tracker issues
}

// validateStatusTransition checks IDPF rules for a status transition
// Set force=true to bypass checkbox validation (but NOT body or release requirements)
func validateStatusTransition(cfg *config.Config, ctx *issueValidationContext, targetStatus, targetRelease string, force bool) *ValidationError {
	// Skip validation if not using IDPF
	if !cfg.IsIDPF() {
		return nil
	}

	// Normalize status values for comparison
	currentStatus := strings.ToLower(ctx.CurrentStatus)
	targetStatusLower := strings.ToLower(targetStatus)

	// Rule 1: Body required for in_review/done (NOT bypassed by --force)
	if targetStatusLower == "in_review" || targetStatusLower == "in review" || targetStatusLower == "done" {
		if isBodyEmpty(ctx.Body) {
			return &ValidationError{
				IssueNumber: ctx.Number,
				Message:     fmt.Sprintf("Empty body. Cannot move to '%s' without issue content.", targetStatus),
				Suggestion:  fmt.Sprintf("Use: gh issue edit %d --body \"<description>\"", ctx.Number),
			}
		}
	}

	// Rule 2: All checkboxes must be checked for in_review/done (bypassed by --force)
	if targetStatusLower == "in_review" || targetStatusLower == "in review" || targetStatusLower == "done" {
		unchecked := countUncheckedBoxes(ctx.Body)
		if unchecked > 0 && !force {
			uncheckedItems := getUncheckedItems(ctx.Body)
			itemList := ""
			if len(uncheckedItems) > 0 {
				itemList = "\n" + strings.Join(uncheckedItems, "\n")
			}
			return &ValidationError{
				IssueNumber: ctx.Number,
				Message:     fmt.Sprintf("Has %d unchecked checkbox(es):%s", unchecked, itemList),
				Suggestion:  fmt.Sprintf("Complete these items before moving to %s, or use --force to bypass.", targetStatus),
			}
		}
	}

	// Rule 3: Release required for backlog → ready/in_progress
	if currentStatus == "backlog" && (targetStatusLower == "ready" || targetStatusLower == "in progress" || targetStatusLower == "in_progress") {
		// Check if release is being set or already set
		releaseValue := targetRelease
		if releaseValue == "" {
			releaseValue = ctx.CurrentRelease
		}

		if releaseValue == "" {
			return &ValidationError{
				IssueNumber: ctx.Number,
				Message:     fmt.Sprintf("No release assignment. Cannot move from 'backlog' to '%s' without a release.", targetStatus),
				Suggestion:  fmt.Sprintf("Use: gh pmu move %d --release \"release/vX.Y.Z\"", ctx.Number),
			}
		}

		// Validate release exists in active releases (if we have discovered releases)
		if !isReleaseActiveInContext(ctx.ActiveReleases, releaseValue) {
			suggestion := "Use 'gh pmu release start' to create a new release."
			if len(ctx.ActiveReleases) > 0 {
				suggestion = fmt.Sprintf("Available releases: %s\n\n%s", strings.Join(ctx.ActiveReleases, ", "), suggestion)
			}
			return &ValidationError{
				IssueNumber: ctx.Number,
				Message:     fmt.Sprintf("Release \"%s\" not found in active releases.", releaseValue),
				Suggestion:  suggestion,
			}
		}
	}

	return nil
}

// isReleaseActiveInContext checks if a release name exists in the discovered active releases
func isReleaseActiveInContext(activeReleases []string, releaseName string) bool {
	// If no active releases discovered, allow any release (backwards compatibility)
	if len(activeReleases) == 0 {
		return true
	}

	for _, active := range activeReleases {
		if strings.EqualFold(active, releaseName) {
			return true
		}
	}

	return false
}

// discoverActiveReleases fetches active release names from GitHub issues with "release" label
// Returns the extracted release names (e.g., "v1.2.0") from issue titles like "Release: v1.2.0"
func discoverActiveReleases(issues []api.Issue) []string {
	var releases []string
	for _, issue := range issues {
		if strings.HasPrefix(issue.Title, "Release: ") {
			// Extract version from title (e.g., "Release: v1.2.0" or "Release: v1.2.0 (Phoenix)")
			version := strings.TrimPrefix(issue.Title, "Release: ")
			// Remove codename in parentheses if present
			if idx := strings.Index(version, " ("); idx > 0 {
				version = version[:idx]
			}
			releases = append(releases, strings.TrimSpace(version))
		}
	}
	return releases
}

// isBodyEmpty checks if the body is empty (empty string or whitespace only)
func isBodyEmpty(body string) bool {
	return strings.TrimSpace(body) == ""
}

// countUncheckedBoxes counts the number of unchecked checkboxes in the body
func countUncheckedBoxes(body string) int {
	return len(uncheckedBoxRegex.FindAllString(body, -1))
}

// countCheckedBoxes counts the number of checked checkboxes in the body
func countCheckedBoxes(body string) int {
	return len(checkedBoxRegex.FindAllString(body, -1))
}

// getUncheckedItems extracts the text of all unchecked checkbox items
func getUncheckedItems(body string) []string {
	// Match unchecked checkboxes with their text content
	uncheckedItemRegex := regexp.MustCompile(`- \[ \] (.+)`)
	matches := uncheckedItemRegex.FindAllStringSubmatch(body, -1)

	var items []string
	for _, match := range matches {
		if len(match) > 1 {
			items = append(items, "  [ ] "+strings.TrimSpace(match[1]))
		}
	}
	return items
}

// getFieldValueFromSlice extracts a field value from a slice of field values
func getFieldValueFromSlice(fieldValues []api.FieldValue, fieldName string) string {
	for _, fv := range fieldValues {
		if strings.EqualFold(fv.Field, fieldName) {
			return fv.Value
		}
	}
	return ""
}

// buildValidationContext creates a validation context from project item data
func buildValidationContext(number int, body string, fieldValues []api.FieldValue, activeReleases []string) *issueValidationContext {
	return &issueValidationContext{
		Number:         number,
		CurrentStatus:  getFieldValueFromSlice(fieldValues, "Status"),
		CurrentRelease: getFieldValueFromSlice(fieldValues, "Release"),
		Body:           body,
		ActiveReleases: activeReleases,
	}
}
