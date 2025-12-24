package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/rubrical-studios/gh-pmu/internal/api"
	"github.com/rubrical-studios/gh-pmu/internal/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ErrRepoRootProtected is returned when attempting to write config to repo root during tests
var ErrRepoRootProtected = errors.New("cannot write config to repository root during tests")

// protectRepoRoot enables protection against writing to repo root (set by tests)
var protectRepoRoot bool

// SetRepoRootProtection enables or disables repo root write protection.
// This should be called by test setup to prevent accidental config writes.
func SetRepoRootProtection(enabled bool) {
	protectRepoRoot = enabled
}

// isRepoRoot checks if the given directory is the repository root by looking for go.mod
func isRepoRoot(dir string) bool {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(absDir, "go.mod"))
	return err == nil
}

func newInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize gh-pmu configuration for the current project",
		Long: `Initialize gh-pmu configuration by creating a .gh-pmu.yml file.

This command will:
- Auto-detect the current repository from git remote
- Discover and list available projects for selection
- Fetch and cache project field metadata from GitHub
- Create a .gh-pmu.yml configuration file`,
		RunE: runInit,
	}

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	u := ui.New(cmd.OutOrStdout())
	reader := bufio.NewReader(os.Stdin)

	// Print header
	u.Header("gh-pmu init", "Configure project management settings")
	fmt.Fprintln(cmd.OutOrStdout())

	// Check if config already exists and preserve framework and active releases
	var existingFramework string
	var existingActiveReleases []ReleaseActiveEntry
	if _, err := os.Stat(".gh-pmu.yml"); err == nil {
		// Try to load existing config to preserve framework and active releases
		if existingCfg, err := loadExistingConfigFull("."); err == nil {
			if existingCfg.Framework != "" {
				existingFramework = existingCfg.Framework
			}
			existingActiveReleases = existingCfg.ActiveReleases
		}
		u.Warning("Configuration file .gh-pmu.yml already exists")
		fmt.Fprint(cmd.OutOrStdout(), u.Prompt("Overwrite?", "y/N"))
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			u.Info("Aborted")
			return nil
		}
		fmt.Fprintln(cmd.OutOrStdout())
	}

	// Auto-detect repository
	detectedRepo := detectRepository()
	var owner string
	var defaultRepo string

	if detectedRepo != "" {
		o, _ := splitRepository(detectedRepo)
		owner = o
		defaultRepo = detectedRepo
		u.Success(fmt.Sprintf("Detected repository: %s", detectedRepo))
	} else {
		u.Warning("Could not detect repository from git remote")
		fmt.Fprint(cmd.OutOrStdout(), u.Prompt("Repository owner", ""))
		ownerInput, _ := reader.ReadString('\n')
		owner = strings.TrimSpace(ownerInput)
		if owner == "" {
			return fmt.Errorf("repository owner is required")
		}
	}

	// Initialize API client
	client := api.NewClient()

	// Fetch projects for owner
	fmt.Fprintln(cmd.OutOrStdout())
	spinner := ui.NewSpinner(cmd.OutOrStdout(), fmt.Sprintf("Fetching projects for %s...", owner))
	spinner.Start()

	projects, err := client.ListProjects(owner)
	spinner.Stop()

	var selectedProject *api.Project
	var projectNumber int

	if err != nil || len(projects) == 0 {
		// No projects found or error - fall back to manual entry
		if err != nil {
			u.Warning(fmt.Sprintf("Could not fetch projects: %v", err))
		} else {
			u.Warning(fmt.Sprintf("No projects found for %s", owner))
		}
		fmt.Fprintln(cmd.OutOrStdout())

		// Manual project number entry
		fmt.Fprint(cmd.OutOrStdout(), u.Prompt("Project number", ""))
		numberInput, _ := reader.ReadString('\n')
		numberInput = strings.TrimSpace(numberInput)
		projectNumber, err = strconv.Atoi(numberInput)
		if err != nil {
			return fmt.Errorf("invalid project number: %s", numberInput)
		}

		// Validate project exists
		spinner = ui.NewSpinner(cmd.OutOrStdout(), fmt.Sprintf("Validating project %s/%d...", owner, projectNumber))
		spinner.Start()
		selectedProject, err = client.GetProject(owner, projectNumber)
		spinner.Stop()

		if err != nil {
			return fmt.Errorf("failed to find project: %w", err)
		}
		u.Success(fmt.Sprintf("Found project: %s", selectedProject.Title))
	} else {
		// Projects found - show selection menu
		u.Success(fmt.Sprintf("Found %d project(s)", len(projects)))
		fmt.Fprintln(cmd.OutOrStdout())

		u.Step(1, 2, "Select Project")

		// Build menu options
		var menuOptions []string
		for _, p := range projects {
			menuOptions = append(menuOptions, fmt.Sprintf("%s (#%d)", p.Title, p.Number))
		}
		u.PrintMenu(menuOptions, true)

		// Get selection
		defaultSelection := "1"
		fmt.Fprint(cmd.OutOrStdout(), u.Prompt("Select", defaultSelection))
		selectionInput, _ := reader.ReadString('\n')
		selectionInput = strings.TrimSpace(selectionInput)

		if selectionInput == "" {
			selectionInput = defaultSelection
		}

		selection, err := strconv.Atoi(selectionInput)
		if err != nil {
			return fmt.Errorf("invalid selection: %s", selectionInput)
		}

		if selection == 0 {
			// Manual entry
			fmt.Fprint(cmd.OutOrStdout(), u.Prompt("Project number", ""))
			numberInput, _ := reader.ReadString('\n')
			numberInput = strings.TrimSpace(numberInput)
			projectNumber, err = strconv.Atoi(numberInput)
			if err != nil {
				return fmt.Errorf("invalid project number: %s", numberInput)
			}

			// Validate project exists
			spinner = ui.NewSpinner(cmd.OutOrStdout(), fmt.Sprintf("Validating project %s/%d...", owner, projectNumber))
			spinner.Start()
			selectedProject, err = client.GetProject(owner, projectNumber)
			spinner.Stop()

			if err != nil {
				return fmt.Errorf("failed to find project: %w", err)
			}
		} else if selection < 1 || selection > len(projects) {
			return fmt.Errorf("invalid selection: must be between 0 and %d", len(projects))
		} else {
			selectedProject = &projects[selection-1]
			projectNumber = selectedProject.Number
		}

		u.Success(fmt.Sprintf("Project: %s (#%d)", selectedProject.Title, selectedProject.Number))
	}

	// Step 2: Confirm repository
	fmt.Fprintln(cmd.OutOrStdout())
	u.Step(2, 2, "Confirm Repository")

	var repo string
	if defaultRepo != "" {
		fmt.Fprint(cmd.OutOrStdout(), u.Prompt("Repository", defaultRepo))
		repoInput, _ := reader.ReadString('\n')
		repoInput = strings.TrimSpace(repoInput)
		if repoInput != "" {
			repo = repoInput
		} else {
			repo = defaultRepo
		}
	} else {
		fmt.Fprint(cmd.OutOrStdout(), u.Prompt("Repository (owner/repo)", ""))
		repoInput, _ := reader.ReadString('\n')
		repo = strings.TrimSpace(repoInput)
	}

	if repo == "" {
		return fmt.Errorf("repository is required")
	}

	u.Success(fmt.Sprintf("Repository: %s", repo))

	// Prompt for IDPF framework (new projects only, preserve existing on re-init)
	var framework string
	if existingFramework != "" {
		// Preserve existing framework value on re-init
		framework = existingFramework
		u.Info(fmt.Sprintf("Framework preserved: %s", framework))
	} else {
		// New project - prompt for framework
		framework = "IDPF" // Default to IDPF
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprint(cmd.OutOrStdout(), u.Prompt("Use IDPF framework validation?", "Y/n"))
		frameworkInput, _ := reader.ReadString('\n')
		frameworkInput = strings.TrimSpace(strings.ToLower(frameworkInput))
		if frameworkInput == "n" || frameworkInput == "no" {
			framework = "none"
			u.Info("IDPF validation disabled")
		} else {
			u.Success("IDPF validation enabled")
		}
	}

	// Fetch project fields
	fmt.Fprintln(cmd.OutOrStdout())
	spinner = ui.NewSpinner(cmd.OutOrStdout(), "Fetching project fields...")
	spinner.Start()
	_, err = client.GetProjectFields(selectedProject.ID)
	spinner.Stop()

	if err != nil {
		u.Warning(fmt.Sprintf("Could not fetch project fields: %v", err))
	}

	// Check and create required fields (IDPF only)
	repoOwner, repoName := splitRepository(repo)
	if framework == "IDPF" {
		fmt.Fprintln(cmd.OutOrStdout())
		u.Info("Checking project fields...")
		requiredFields := []struct {
			name     string
			dataType string
		}{
			{"Release", "TEXT"},
			{"Microsprint", "TEXT"},
		}

		for _, rf := range requiredFields {
			exists, err := client.FieldExists(selectedProject.ID, rf.name)
			if err != nil {
				u.Warning(fmt.Sprintf("Could not check %s field: %v", rf.name, err))
				continue
			}
			if exists {
				u.Success(fmt.Sprintf("%s field exists", rf.name))
			} else {
				spinner = ui.NewSpinner(cmd.OutOrStdout(), fmt.Sprintf("Creating %s field...", rf.name))
				spinner.Start()
				_, err := client.CreateProjectField(selectedProject.ID, rf.name, rf.dataType, nil)
				spinner.Stop()
				if err != nil {
					u.Warning(fmt.Sprintf("Could not create %s field: %v", rf.name, err))
				} else {
					u.Success(fmt.Sprintf("Created %s field", rf.name))
				}
			}
		}

		// Check and create required labels (IDPF only)
		fmt.Fprintln(cmd.OutOrStdout())
		u.Info("Checking repository labels...")
		requiredLabels := []struct {
			name        string
			color       string
			description string
		}{
			{"release", "0e8a16", "Release tracker issue"},
			{"microsprint", "1d76db", "Microsprint tracker issue"},
		}

		for _, rl := range requiredLabels {
			exists, err := client.LabelExists(repoOwner, repoName, rl.name)
			if err != nil {
				u.Warning(fmt.Sprintf("Could not check %s label: %v", rl.name, err))
				continue
			}
			if exists {
				u.Success(fmt.Sprintf("%s label exists", rl.name))
			} else {
				spinner = ui.NewSpinner(cmd.OutOrStdout(), fmt.Sprintf("Creating %s label...", rl.name))
				spinner.Start()
				err := client.CreateLabel(repoOwner, repoName, rl.name, rl.color, rl.description)
				spinner.Stop()
				if err != nil {
					u.Warning(fmt.Sprintf("Could not create %s label: %v", rl.name, err))
				} else {
					u.Success(fmt.Sprintf("Created %s label", rl.name))
				}
			}
		}
	} else {
		u.Info("Skipping Release/Microsprint field and label creation (framework: none)")
	}

	// Refetch fields after potential creation
	fields, _ := client.GetProjectFields(selectedProject.ID)

	// Sync active releases from tracker issues
	fmt.Fprintln(cmd.OutOrStdout())
	u.Info("Syncing active releases...")
	activeReleases, releaseResults, err := syncActiveReleases(client, repoOwner, repoName)
	if err != nil {
		u.Warning(fmt.Sprintf("Could not sync releases: %v", err))
	} else if len(releaseResults) > 0 {
		for _, r := range releaseResults {
			u.Success(fmt.Sprintf("Found active release: %s (track: %s, tracker: #%d, issues: %d)",
				r.Entry.Version, r.Entry.Track, r.Entry.TrackerIssue, r.IssueCount))
		}
	} else {
		u.Info("No active releases found")
	}

	// Convert to metadata
	metadata := &ProjectMetadata{
		ProjectID: selectedProject.ID,
	}
	for _, f := range fields {
		fm := FieldMetadata{
			ID:       f.ID,
			Name:     f.Name,
			DataType: f.DataType,
		}
		for _, opt := range f.Options {
			fm.Options = append(fm.Options, OptionMetadata{
				ID:   opt.ID,
				Name: opt.Name,
			})
		}
		metadata.Fields = append(metadata.Fields, fm)
	}

	// Create config
	cfg := &InitConfig{
		ProjectName:   selectedProject.Title,
		ProjectOwner:  owner,
		ProjectNumber: projectNumber,
		Repositories:  []string{repo},
		Framework:     framework,
	}

	// Merge existing active releases with newly discovered ones (additive, no duplicates)
	mergedReleases := mergeActiveReleases(existingActiveReleases, activeReleases)

	// Write config
	cwd, _ := os.Getwd()
	if err := writeConfigWithMetadata(cwd, cfg, metadata, mergedReleases); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Print summary
	u.SummaryBox("Configuration saved", map[string]string{
		"Project":    fmt.Sprintf("%s (#%d)", selectedProject.Title, selectedProject.Number),
		"Repository": repo,
		"Fields":     fmt.Sprintf("%d cached", len(fields)),
		"Config":     ".gh-pmu.yml",
	}, []string{"Project", "Repository", "Fields", "Config"})

	return nil
}

// parseGitRemote extracts owner/repo from a GitHub remote URL.
// Supports both HTTPS and SSH formats.
// Returns empty string if not a valid GitHub remote.
func parseGitRemote(remote string) string {
	if remote == "" {
		return ""
	}

	// HTTPS format: https://github.com/owner/repo.git or https://github.com/owner/repo
	httpsRegex := regexp.MustCompile(`^https://github\.com/([^/]+)/([^/]+?)(?:\.git)?$`)
	if matches := httpsRegex.FindStringSubmatch(remote); matches != nil {
		return matches[1] + "/" + matches[2]
	}

	// SSH format: git@github.com:owner/repo.git or git@github.com:owner/repo
	sshRegex := regexp.MustCompile(`^git@github\.com:([^/]+)/([^/]+?)(?:\.git)?$`)
	if matches := sshRegex.FindStringSubmatch(remote); matches != nil {
		return matches[1] + "/" + matches[2]
	}

	return ""
}

// detectRepository attempts to get the repository from git remote.
func detectRepository() string {
	// Try to get the origin remote URL
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return parseGitRemote(strings.TrimSpace(string(output)))
}

// existingConfigFull holds framework and active releases for preservation during re-init
type existingConfigFull struct {
	Framework      string `yaml:"framework"`
	ActiveReleases []ReleaseActiveEntry
}

// existingConfigRaw is used for YAML unmarshaling to get release.active
type existingConfigRaw struct {
	Framework string `yaml:"framework"`
	Release   struct {
		Active []ReleaseActiveEntry `yaml:"active"`
	} `yaml:"release"`
}

// loadExistingConfigFull loads framework and active releases from existing config
func loadExistingConfigFull(dir string) (*existingConfigFull, error) {
	configPath := filepath.Join(dir, ".gh-pmu.yml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var raw existingConfigRaw
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return &existingConfigFull{
		Framework:      raw.Framework,
		ActiveReleases: raw.Release.Active,
	}, nil
}

// splitRepository splits "owner/repo" into owner and repo parts.
func splitRepository(repo string) (owner, name string) {
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", ""
}

// InitConfig holds the configuration gathered during init.
type InitConfig struct {
	ProjectName   string
	ProjectOwner  string
	ProjectNumber int
	Repositories  []string
	Framework     string
}

// ConfigFile represents the .gh-pmu.yml file structure.
type ConfigFile struct {
	Project      ProjectConfig           `yaml:"project"`
	Repositories []string                `yaml:"repositories"`
	Defaults     DefaultsConfig          `yaml:"defaults"`
	Fields       map[string]FieldMapping `yaml:"fields"`
	Triage       map[string]TriageRule   `yaml:"triage,omitempty"`
}

// ProjectConfig represents the project section of config.
type ProjectConfig struct {
	Name   string `yaml:"name,omitempty"`
	Number int    `yaml:"number"`
	Owner  string `yaml:"owner"`
}

// DefaultsConfig represents default values for new items.
type DefaultsConfig struct {
	Priority string   `yaml:"priority"`
	Status   string   `yaml:"status"`
	Labels   []string `yaml:"labels,omitempty"`
}

// FieldMapping represents a field alias mapping.
type FieldMapping struct {
	Field  string            `yaml:"field"`
	Values map[string]string `yaml:"values"`
}

// ProjectMetadata holds cached project information from GitHub API.
type ProjectMetadata struct {
	ProjectID string
	Fields    []FieldMetadata
}

// FieldMetadata holds cached field information.
type FieldMetadata struct {
	ID       string
	Name     string
	DataType string
	Options  []OptionMetadata
}

// OptionMetadata holds option information for single-select fields.
type OptionMetadata struct {
	ID   string
	Name string
}

// MetadataSection represents the metadata section in config file.
type MetadataSection struct {
	Project MetadataProject `yaml:"project"`
	Fields  []MetadataField `yaml:"fields"`
}

// MetadataProject holds the project ID.
type MetadataProject struct {
	ID string `yaml:"id"`
}

// MetadataField represents a field in the metadata section.
type MetadataField struct {
	Name     string                `yaml:"name"`
	ID       string                `yaml:"id"`
	DataType string                `yaml:"data_type"`
	Options  []MetadataFieldOption `yaml:"options,omitempty"`
}

// MetadataFieldOption represents a field option.
type MetadataFieldOption struct {
	Name string `yaml:"name"`
	ID   string `yaml:"id"`
}

// TriageRule represents a single triage rule configuration.
type TriageRule struct {
	Query       string          `yaml:"query"`
	Apply       TriageApply     `yaml:"apply"`
	Interactive map[string]bool `yaml:"interactive,omitempty"`
}

// TriageApply represents what to apply when a triage rule matches.
type TriageApply struct {
	Labels []string          `yaml:"labels,omitempty"`
	Fields map[string]string `yaml:"fields,omitempty"`
}

// ConfigFileWithMetadata extends ConfigFile with metadata section.
// ReleaseConfig holds release-related configuration
type ReleaseConfig struct {
	Active []ReleaseActiveEntry `yaml:"active,omitempty"`
}

// ReleaseActiveEntry represents an active release
type ReleaseActiveEntry struct {
	Version      string `yaml:"version"`
	TrackerIssue int    `yaml:"tracker_issue"`
	Track        string `yaml:"track"`
}

type ConfigFileWithMetadata struct {
	Project      ProjectConfig           `yaml:"project"`
	Repositories []string                `yaml:"repositories"`
	Framework    string                  `yaml:"framework,omitempty"`
	Defaults     DefaultsConfig          `yaml:"defaults"`
	Fields       map[string]FieldMapping `yaml:"fields"`
	Triage       map[string]TriageRule   `yaml:"triage,omitempty"`
	Release      ReleaseConfig           `yaml:"release,omitempty"`
	Metadata     MetadataSection         `yaml:"metadata"`
}

// ProjectValidator is the interface for validating projects.
type ProjectValidator interface {
	GetProject(owner string, number int) (interface{}, error)
}

// validateProject checks if the project exists.
func validateProject(client ProjectValidator, owner string, number int) error {
	_, err := client.GetProject(owner, number)
	return err
}

// writeConfig writes the configuration to a .gh-pmu.yml file.
func writeConfig(dir string, cfg *InitConfig) error {
	// Safety check: prevent accidental writes to repo root during tests
	if protectRepoRoot && isRepoRoot(dir) {
		return ErrRepoRootProtected
	}

	configFile := &ConfigFile{
		Project: ProjectConfig{
			Name:   cfg.ProjectName,
			Owner:  cfg.ProjectOwner,
			Number: cfg.ProjectNumber,
		},
		Repositories: cfg.Repositories,
		Defaults: DefaultsConfig{
			Priority: "p2",
			Status:   "backlog",
		},
		Fields: map[string]FieldMapping{
			"priority": {
				Field: "Priority",
				Values: map[string]string{
					"p0": "P0",
					"p1": "P1",
					"p2": "P2",
				},
			},
			"status": {
				Field: "Status",
				Values: map[string]string{
					"backlog":     "Backlog",
					"ready":       "Ready",
					"in_progress": "In progress",
					"in_review":   "In review",
					"done":        "Done",
				},
			},
		},
		Triage: map[string]TriageRule{
			"estimate": {
				Query: "is:issue is:open -has:estimate",
				Apply: TriageApply{},
				Interactive: map[string]bool{
					"estimate": true,
				},
			},
		},
	}

	data, err := yaml.Marshal(configFile)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := filepath.Join(dir, ".gh-pmu.yml")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// writeConfigWithMetadata writes the configuration with project metadata.
func writeConfigWithMetadata(dir string, cfg *InitConfig, metadata *ProjectMetadata, activeReleases []ReleaseActiveEntry) error {
	// Safety check: prevent accidental writes to repo root during tests
	if protectRepoRoot && isRepoRoot(dir) {
		return ErrRepoRootProtected
	}

	// Convert metadata to YAML format
	var metadataFields []MetadataField
	for _, f := range metadata.Fields {
		mf := MetadataField{
			Name:     f.Name,
			ID:       f.ID,
			DataType: f.DataType,
		}
		for _, opt := range f.Options {
			mf.Options = append(mf.Options, MetadataFieldOption{
				Name: opt.Name,
				ID:   opt.ID,
			})
		}
		metadataFields = append(metadataFields, mf)
	}

	// Build field mappings dynamically from metadata
	fieldMappings := buildFieldMappingsFromMetadata(metadata)

	configFile := &ConfigFileWithMetadata{
		Project: ProjectConfig{
			Name:   cfg.ProjectName,
			Owner:  cfg.ProjectOwner,
			Number: cfg.ProjectNumber,
		},
		Repositories: cfg.Repositories,
		Framework:    cfg.Framework,
		Defaults: DefaultsConfig{
			Priority: "p2",
			Status:   "backlog",
		},
		Fields: fieldMappings,
		Triage: map[string]TriageRule{
			"estimate": {
				Query: "is:issue is:open -has:estimate",
				Apply: TriageApply{},
				Interactive: map[string]bool{
					"estimate": true,
				},
			},
		},
		Release: ReleaseConfig{
			Active: activeReleases,
		},
		Metadata: MetadataSection{
			Project: MetadataProject{
				ID: metadata.ProjectID,
			},
			Fields: metadataFields,
		},
	}

	data, err := yaml.Marshal(configFile)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := filepath.Join(dir, ".gh-pmu.yml")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ReleaseDiscoveryResult holds release info with issue count for display
type ReleaseDiscoveryResult struct {
	Entry      ReleaseActiveEntry
	IssueCount int
}

// syncActiveReleases queries open release issues and returns active release entries with issue counts
func syncActiveReleases(client *api.Client, owner, repo string) ([]ReleaseActiveEntry, []ReleaseDiscoveryResult, error) {
	issues, err := client.GetOpenIssuesByLabel(owner, repo, "release")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get release issues: %w", err)
	}

	var entries []ReleaseActiveEntry
	var results []ReleaseDiscoveryResult
	for _, issue := range issues {
		if !strings.HasPrefix(issue.Title, "Release: ") {
			continue
		}

		version, track := parseReleaseTitleForInit(issue.Title)

		// Get the release field value to count issues
		releaseFieldValue := extractReleaseVersionForInit(issue.Title)
		releaseIssues, _ := client.GetIssuesByRelease(owner, repo, releaseFieldValue)
		issueCount := len(releaseIssues)

		entry := ReleaseActiveEntry{
			Version:      version,
			TrackerIssue: issue.Number,
			Track:        track,
		}
		entries = append(entries, entry)
		results = append(results, ReleaseDiscoveryResult{
			Entry:      entry,
			IssueCount: issueCount,
		})
	}

	return entries, results, nil
}

// extractReleaseVersionForInit extracts the release field value from a title
func extractReleaseVersionForInit(title string) string {
	// Remove "Release: " prefix
	version := strings.TrimPrefix(title, "Release: ")
	// Remove codename if present
	if idx := strings.Index(version, " ("); idx > 0 {
		version = version[:idx]
	}
	return version
}

// parseReleaseTitleForInit parses a release title into version and track
func parseReleaseTitleForInit(title string) (version, track string) {
	// Remove "Release: " prefix
	remainder := strings.TrimPrefix(title, "Release: ")

	// Remove codename suffix if present (e.g., " (Phoenix)")
	if idx := strings.Index(remainder, " ("); idx != -1 {
		remainder = remainder[:idx]
	}

	// Check for track prefix (e.g., "patch/", "beta/")
	if strings.Contains(remainder, "/") {
		parts := strings.SplitN(remainder, "/", 2)
		track = parts[0]
		version = strings.TrimPrefix(parts[1], "v")
	} else {
		// Default track is "stable", version starts with v
		track = "stable"
		version = strings.TrimPrefix(remainder, "v")
	}

	return version, track
}

// mergeActiveReleases merges existing releases with newly discovered ones (additive, no duplicates)
func mergeActiveReleases(existing, discovered []ReleaseActiveEntry) []ReleaseActiveEntry {
	// Start with discovered releases (fresh from GitHub)
	result := make([]ReleaseActiveEntry, len(discovered))
	copy(result, discovered)

	// Create a map of tracker issue numbers for quick duplicate detection
	seen := make(map[int]bool)
	for _, r := range discovered {
		seen[r.TrackerIssue] = true
	}

	// Add existing releases that aren't already in discovered (handles releases that may have been added manually)
	for _, r := range existing {
		if !seen[r.TrackerIssue] {
			result = append(result, r)
			seen[r.TrackerIssue] = true
		}
	}

	return result
}

// buildFieldMappingsFromMetadata builds field mappings dynamically from project metadata.
// This ensures all field options (including "Parking Lot") are included in the config.
func buildFieldMappingsFromMetadata(metadata *ProjectMetadata) map[string]FieldMapping {
	mappings := make(map[string]FieldMapping)

	// Find Status and Priority fields in metadata
	for _, field := range metadata.Fields {
		fieldNameLower := strings.ToLower(field.Name)

		if fieldNameLower == "status" && len(field.Options) > 0 {
			values := make(map[string]string)
			for _, opt := range field.Options {
				alias := optionNameToAlias(opt.Name)
				values[alias] = opt.Name
			}
			mappings["status"] = FieldMapping{
				Field:  field.Name,
				Values: values,
			}
		}

		if fieldNameLower == "priority" && len(field.Options) > 0 {
			values := make(map[string]string)
			for _, opt := range field.Options {
				alias := optionNameToAlias(opt.Name)
				values[alias] = opt.Name
			}
			mappings["priority"] = FieldMapping{
				Field:  field.Name,
				Values: values,
			}
		}
	}

	// Fallback to defaults if fields not found in metadata
	if _, ok := mappings["status"]; !ok {
		mappings["status"] = FieldMapping{
			Field: "Status",
			Values: map[string]string{
				"backlog":     "Backlog",
				"ready":       "Ready",
				"in_progress": "In progress",
				"in_review":   "In review",
				"done":        "Done",
			},
		}
	}

	if _, ok := mappings["priority"]; !ok {
		mappings["priority"] = FieldMapping{
			Field: "Priority",
			Values: map[string]string{
				"p0": "P0",
				"p1": "P1",
				"p2": "P2",
			},
		}
	}

	return mappings
}

// optionNameToAlias converts a field option name to a CLI-friendly alias.
// Examples: "In progress" -> "in_progress", "🅿️ Parking Lot" -> "parking_lot"
func optionNameToAlias(name string) string {
	// Remove common emoji prefixes (strip all non-ASCII characters)
	var cleaned strings.Builder
	for _, r := range name {
		if r < 128 { // ASCII only
			cleaned.WriteRune(r)
		}
	}
	result := strings.TrimSpace(cleaned.String())

	// Convert to lowercase and replace spaces with underscores
	result = strings.ToLower(result)
	result = strings.ReplaceAll(result, " ", "_")

	// Remove any double underscores
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}

	// Trim leading/trailing underscores
	result = strings.Trim(result, "_")

	return result
}
