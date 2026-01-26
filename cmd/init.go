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
	"github.com/rubrical-studios/gh-pmu/internal/defaults"
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

// initOptions holds the command-line options for init
type initOptions struct {
	nonInteractive bool
	project        int
	repo           string
	owner          string
	framework      string
	yes            bool
}

func newInitCommand() *cobra.Command {
	opts := &initOptions{}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize gh-pmu configuration for the current project",
		Long: `Initialize gh-pmu configuration by creating a .gh-pmu.yml file.

This command will:
- Auto-detect the current repository from git remote
- Discover and list available projects for selection
- Fetch and cache project field metadata from GitHub
- Create a .gh-pmu.yml configuration file

Non-interactive mode (--non-interactive) disables all prompts and requires
--project and --repo flags. Use this for CI/CD pipelines and automation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, args, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.nonInteractive, "non-interactive", false, "Disable UI and prompts (requires --project and --repo)")
	cmd.Flags().IntVar(&opts.project, "project", 0, "Project number")
	cmd.Flags().StringVar(&opts.repo, "repo", "", "Repository (owner/repo format)")
	cmd.Flags().StringVar(&opts.owner, "owner", "", "Project owner (defaults to repo owner)")
	cmd.Flags().StringVar(&opts.framework, "framework", "IDPF", "Framework type (IDPF or none)")
	cmd.Flags().BoolVarP(&opts.yes, "yes", "y", false, "Auto-confirm prompts")

	return cmd
}

func runInit(cmd *cobra.Command, args []string, opts *initOptions) error {
	// Handle non-interactive mode
	if opts.nonInteractive {
		return runInitNonInteractive(cmd, opts)
	}

	u := ui.New(cmd.OutOrStdout())
	reader := bufio.NewReader(os.Stdin)

	// Print header
	u.Header("gh-pmu init", "Configure project management settings")
	fmt.Fprintln(cmd.OutOrStdout())

	// Check if config already exists and preserve framework
	var existingFramework string
	if _, err := os.Stat(".gh-pmu.yml"); err == nil {
		// Try to load existing config to preserve framework
		if existingCfg, err := loadExistingFramework("."); err == nil {
			existingFramework = existingCfg
		}
		if !opts.yes {
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
	}

	// Use flag values if provided, otherwise auto-detect/prompt
	var owner string
	var defaultRepo string
	var projectNumber int

	// Handle --repo flag
	if opts.repo != "" {
		defaultRepo = opts.repo
		o, _ := splitRepository(opts.repo)
		owner = o
		u.Success(fmt.Sprintf("Using repository: %s", opts.repo))
	} else {
		// Auto-detect repository
		detectedRepo := detectRepository()
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
	}

	// Handle --owner flag (overrides repo owner for project lookup)
	if opts.owner != "" {
		owner = opts.owner
	}

	// Handle --project flag
	if opts.project > 0 {
		projectNumber = opts.project
	}

	// Initialize API client
	client := api.NewClient()

	var selectedProject *api.Project
	var err error
	var spinner *ui.Spinner

	// If project number was provided via flag, validate it directly
	if projectNumber > 0 {
		spinner = ui.NewSpinner(cmd.OutOrStdout(), fmt.Sprintf("Validating project %s/%d...", owner, projectNumber))
		spinner.Start()
		selectedProject, err = client.GetProject(owner, projectNumber)
		spinner.Stop()

		if err != nil {
			return fmt.Errorf("failed to find project: %w", err)
		}
		u.Success(fmt.Sprintf("Using project: %s (#%d)", selectedProject.Title, selectedProject.Number))
	} else {
		// Fetch projects for owner
		fmt.Fprintln(cmd.OutOrStdout())
		spinner = ui.NewSpinner(cmd.OutOrStdout(), fmt.Sprintf("Fetching projects for %s...", owner))
		spinner.Start()

		projects, err := client.ListProjects(owner)
		spinner.Stop()

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
	}

	// Step 2: Confirm repository
	var repo string
	if opts.repo != "" {
		// Repo already provided via flag, no need to prompt
		repo = opts.repo
	} else {
		fmt.Fprintln(cmd.OutOrStdout())
		u.Step(2, 2, "Confirm Repository")

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
	}

	// Prompt for IDPF framework (new projects only, preserve existing on re-init)
	var framework string
	if existingFramework != "" {
		// Preserve existing framework value on re-init
		framework = existingFramework
		u.Info(fmt.Sprintf("Framework preserved: %s", framework))
	} else if opts.framework != "" && opts.framework != "IDPF" {
		// Framework explicitly set via flag to non-default value
		framework = opts.framework
		if framework == "none" {
			u.Info("IDPF validation disabled")
		} else {
			u.Success(fmt.Sprintf("Framework: %s", framework))
		}
	} else if opts.yes || opts.framework == "IDPF" {
		// --yes flag or explicit IDPF: use default
		framework = "IDPF"
		u.Success("IDPF validation enabled")
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

	// Load embedded defaults
	defs, err := defaults.Load()
	if err != nil {
		return fmt.Errorf("failed to load embedded defaults: %w", err)
	}

	// Fetch project fields
	fmt.Fprintln(cmd.OutOrStdout())
	spinner = ui.NewSpinner(cmd.OutOrStdout(), "Fetching project fields...")
	spinner.Start()
	projectFields, err := client.GetProjectFields(selectedProject.ID)
	spinner.Stop()

	if err != nil {
		return fmt.Errorf("could not fetch project fields: %w", err)
	}

	// Remove deprecated Microsprint field if it exists
	microsprintField := findFieldByName(projectFields, "Microsprint")
	if microsprintField != nil {
		fmt.Fprintln(cmd.OutOrStdout())
		u.Warning("Found deprecated Microsprint field")
		spinner = ui.NewSpinner(cmd.OutOrStdout(), "Removing Microsprint field...")
		spinner.Start()
		err := client.DeleteProjectField(microsprintField.ID)
		spinner.Stop()
		if err != nil {
			u.Warning(fmt.Sprintf("Could not remove Microsprint field: %v", err))
		} else {
			u.Success("Removed deprecated Microsprint field")
		}
	}

	// Check and create required fields (IDPF only)
	repoOwner, repoName := splitRepository(repo)
	if framework == "IDPF" {
		fmt.Fprintln(cmd.OutOrStdout())
		u.Info("Validating required fields...")

		// Validate required fields exist with expected options
		for _, reqField := range defs.Fields.Required {
			field := findFieldByName(projectFields, reqField.Name)
			if field == nil {
				return fmt.Errorf("required field %q not found in project. Please ensure you copied from a Kanban template that includes the %s field", reqField.Name, reqField.Name)
			}

			// Validate field type
			if field.DataType != reqField.Type {
				return fmt.Errorf("field %q has type %s, expected %s", reqField.Name, field.DataType, reqField.Type)
			}

			// Validate options for SINGLE_SELECT fields
			if reqField.Type == "SINGLE_SELECT" && len(reqField.Options) > 0 {
				for _, reqOpt := range reqField.Options {
					found := false
					for _, opt := range field.Options {
						if opt.Name == reqOpt {
							found = true
							break
						}
					}
					if !found {
						return fmt.Errorf("field %q missing required option %q", reqField.Name, reqOpt)
					}
				}
			}
			u.Success(fmt.Sprintf("%s field validated", reqField.Name))
		}

		// Create optional fields if missing
		fmt.Fprintln(cmd.OutOrStdout())
		u.Info("Checking optional fields...")
		for _, optField := range defs.Fields.CreateIfMissing {
			exists, err := client.FieldExists(selectedProject.ID, optField.Name)
			if err != nil {
				u.Warning(fmt.Sprintf("Could not check %s field: %v", optField.Name, err))
				continue
			}
			if exists {
				u.Success(fmt.Sprintf("%s field exists", optField.Name))
			} else {
				spinner = ui.NewSpinner(cmd.OutOrStdout(), fmt.Sprintf("Creating %s field...", optField.Name))
				spinner.Start()
				_, err := client.CreateProjectField(selectedProject.ID, optField.Name, optField.Type, optField.Options)
				spinner.Stop()
				if err != nil {
					u.Warning(fmt.Sprintf("Could not create %s field: %v", optField.Name, err))
				} else {
					u.Success(fmt.Sprintf("Created %s field", optField.Name))
				}
			}
		}

		// Migrate legacy "release" label to "branch" (if needed)
		fmt.Fprintln(cmd.OutOrStdout())
		u.Info("Checking for label migrations...")
		legacyReleaseExists, _ := client.LabelExists(repoOwner, repoName, "release")
		branchExists, _ := client.LabelExists(repoOwner, repoName, "branch")

		if legacyReleaseExists {
			if !branchExists {
				// Rename "release" to "branch"
				spinner = ui.NewSpinner(cmd.OutOrStdout(), "Migrating 'release' label to 'branch'...")
				spinner.Start()
				err := client.UpdateLabel(repoOwner, repoName, "release", "branch", "0e8a16", "Branch tracker issue")
				spinner.Stop()
				if err != nil {
					u.Warning(fmt.Sprintf("Could not migrate release label: %v", err))
				} else {
					u.Success("Migrated 'release' label to 'branch'")
				}
			} else {
				// Both exist - delete the legacy "release" label
				spinner = ui.NewSpinner(cmd.OutOrStdout(), "Removing legacy 'release' label...")
				spinner.Start()
				err := client.DeleteLabel(repoOwner, repoName, "release")
				spinner.Stop()
				if err != nil {
					u.Warning(fmt.Sprintf("Could not remove legacy release label: %v", err))
				} else {
					u.Success("Removed legacy 'release' label")
				}
			}
		}

		// Check and create required labels from defaults
		fmt.Fprintln(cmd.OutOrStdout())
		u.Info("Checking repository labels...")
		for _, labelDef := range defs.Labels {
			exists, err := client.LabelExists(repoOwner, repoName, labelDef.Name)
			if err != nil {
				u.Warning(fmt.Sprintf("Could not check %s label: %v", labelDef.Name, err))
				continue
			}
			if exists {
				u.Success(fmt.Sprintf("%s label exists", labelDef.Name))
			} else {
				spinner = ui.NewSpinner(cmd.OutOrStdout(), fmt.Sprintf("Creating %s label...", labelDef.Name))
				spinner.Start()
				err := client.CreateLabel(repoOwner, repoName, labelDef.Name, labelDef.Color, labelDef.Description)
				spinner.Stop()
				if err != nil {
					u.Warning(fmt.Sprintf("Could not create %s label: %v", labelDef.Name, err))
				} else {
					u.Success(fmt.Sprintf("Created %s label", labelDef.Name))
				}
			}
		}
	} else {
		u.Info("Skipping field and label setup (framework: none)")
	}

	// Refetch fields after potential creation
	fields, _ := client.GetProjectFields(selectedProject.ID)

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

	// Write config
	cwd, _ := os.Getwd()
	if err := writeConfigWithMetadata(cwd, cfg, metadata); err != nil {
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

// runInitNonInteractive handles init in non-interactive mode (for CI/CD).
// It requires --project and --repo flags and outputs errors to STDERR.
func runInitNonInteractive(cmd *cobra.Command, opts *initOptions) error {
	// Validate required flags
	var missingFlags []string
	if opts.project == 0 {
		missingFlags = append(missingFlags, "--project")
	}
	if opts.repo == "" {
		missingFlags = append(missingFlags, "--repo")
	}

	if len(missingFlags) > 0 {
		fmt.Fprintf(os.Stderr, "error: non-interactive mode requires flags: %s\n", strings.Join(missingFlags, ", "))
		return fmt.Errorf("missing required flags: %s", strings.Join(missingFlags, ", "))
	}

	// Validate repo format
	repoOwner, repoName := splitRepository(opts.repo)
	if repoOwner == "" || repoName == "" {
		fmt.Fprintf(os.Stderr, "error: --repo must be in owner/repo format\n")
		return fmt.Errorf("invalid repo format: %s", opts.repo)
	}

	// Determine owner (from --owner flag or infer from repo)
	owner := opts.owner
	if owner == "" {
		owner = repoOwner
	}

	// Determine framework (defaults to IDPF)
	framework := opts.framework
	if framework == "" {
		framework = "IDPF"
	}

	// Check if config already exists
	var existingFramework string
	if _, err := os.Stat(".gh-pmu.yml"); err == nil {
		if existingCfg, err := loadExistingFramework("."); err == nil {
			existingFramework = existingCfg
		}
		if !opts.yes {
			fmt.Fprintf(os.Stderr, "error: .gh-pmu.yml already exists (use --yes to overwrite)\n")
			return fmt.Errorf("config already exists")
		}
	}

	// Preserve existing framework on re-init
	if existingFramework != "" {
		framework = existingFramework
	}

	// Initialize API client
	client := api.NewClient()

	// Validate project exists
	selectedProject, err := client.GetProject(owner, opts.project)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to find project %s/%d: %v\n", owner, opts.project, err)
		return fmt.Errorf("failed to find project: %w", err)
	}

	// Load embedded defaults
	defs, err := defaults.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to load defaults: %v\n", err)
		return fmt.Errorf("failed to load embedded defaults: %w", err)
	}

	// Fetch project fields
	projectFields, err := client.GetProjectFields(selectedProject.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not fetch project fields: %v\n", err)
		return fmt.Errorf("could not fetch project fields: %w", err)
	}

	// Check and create required fields (IDPF only)
	if framework == "IDPF" {
		// Validate required fields exist with expected options
		for _, reqField := range defs.Fields.Required {
			field := findFieldByName(projectFields, reqField.Name)
			if field == nil {
				fmt.Fprintf(os.Stderr, "error: required field %q not found in project\n", reqField.Name)
				return fmt.Errorf("required field %q not found in project", reqField.Name)
			}

			// Validate field type
			if field.DataType != reqField.Type {
				fmt.Fprintf(os.Stderr, "error: field %q has type %s, expected %s\n", reqField.Name, field.DataType, reqField.Type)
				return fmt.Errorf("field %q has type %s, expected %s", reqField.Name, field.DataType, reqField.Type)
			}

			// Validate options for SINGLE_SELECT fields
			if reqField.Type == "SINGLE_SELECT" && len(reqField.Options) > 0 {
				for _, reqOpt := range reqField.Options {
					found := false
					for _, opt := range field.Options {
						if opt.Name == reqOpt {
							found = true
							break
						}
					}
					if !found {
						fmt.Fprintf(os.Stderr, "error: field %q missing required option %q\n", reqField.Name, reqOpt)
						return fmt.Errorf("field %q missing required option %q", reqField.Name, reqOpt)
					}
				}
			}
		}

		// Create optional fields if missing
		for _, optField := range defs.Fields.CreateIfMissing {
			exists, err := client.FieldExists(selectedProject.ID, optField.Name)
			if err != nil {
				continue // Skip on error in non-interactive mode
			}
			if !exists {
				_, _ = client.CreateProjectField(selectedProject.ID, optField.Name, optField.Type, optField.Options)
			}
		}

		// Check and create required labels
		for _, labelDef := range defs.Labels {
			exists, err := client.LabelExists(repoOwner, repoName, labelDef.Name)
			if err != nil {
				continue
			}
			if !exists {
				_ = client.CreateLabel(repoOwner, repoName, labelDef.Name, labelDef.Color, labelDef.Description)
			}
		}
	}

	// Refetch fields after potential creation
	fields, _ := client.GetProjectFields(selectedProject.ID)

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
		ProjectNumber: opts.project,
		Repositories:  []string{opts.repo},
		Framework:     framework,
	}

	// Write config
	cwd, _ := os.Getwd()
	if err := writeConfigWithMetadata(cwd, cfg, metadata); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to write config: %v\n", err)
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Output success to stdout (minimal for CI/CD parsing)
	fmt.Fprintf(cmd.OutOrStdout(), "Created .gh-pmu.yml for %s (#%d)\n", selectedProject.Title, opts.project)

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

// existingConfigRaw is used for YAML unmarshaling to get framework
type existingConfigRaw struct {
	Framework string `yaml:"framework"`
}

// loadExistingFramework loads framework from existing config
func loadExistingFramework(dir string) (string, error) {
	configPath := filepath.Join(dir, ".gh-pmu.yml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}
	var raw existingConfigRaw
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return "", err
	}
	return raw.Framework, nil
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
type ConfigFileWithMetadata struct {
	Project      ProjectConfig           `yaml:"project"`
	Repositories []string                `yaml:"repositories"`
	Framework    string                  `yaml:"framework,omitempty"`
	Defaults     DefaultsConfig          `yaml:"defaults"`
	Fields       map[string]FieldMapping `yaml:"fields"`
	Triage       map[string]TriageRule   `yaml:"triage,omitempty"`
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
func writeConfigWithMetadata(dir string, cfg *InitConfig, metadata *ProjectMetadata) error {
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

// findFieldByName searches for a field by name in a slice of ProjectFields.
// Returns nil if not found.
func findFieldByName(fields []api.ProjectField, name string) *api.ProjectField {
	for i := range fields {
		if fields[i].Name == name {
			return &fields[i]
		}
	}
	return nil
}
