package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/scooter-indie/gh-pmu/internal/api"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize gh-pmu configuration for the current project",
		Long: `Initialize gh-pmu configuration by creating a .gh-pmu.yml file.

This command will:
- Prompt for project owner and number
- Auto-detect the current repository (if in a git repo)
- Fetch and cache project field metadata from GitHub
- Validate the project exists before saving`,
		RunE: runInit,
	}

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	cmd.Println("Initializing gh-pm configuration...")

	reader := bufio.NewReader(os.Stdin)

	// Check if config already exists
	if _, err := os.Stat(".gh-pmu.yml"); err == nil {
		cmd.Print("Configuration file .gh-pmu.yml already exists. Overwrite? [y/N]: ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			cmd.Println("Aborted.")
			return nil
		}
	}

	// Auto-detect repository
	detectedRepo := detectRepository()
	var owner string
	var defaultRepo string

	if detectedRepo != "" {
		o, _ := splitRepository(detectedRepo)
		owner = o
		defaultRepo = detectedRepo
		cmd.Printf("Detected repository: %s\n", detectedRepo)
	}

	// Prompt for project owner
	if owner == "" {
		cmd.Print("Project owner (GitHub username or organization): ")
		ownerInput, _ := reader.ReadString('\n')
		owner = strings.TrimSpace(ownerInput)
	} else {
		cmd.Printf("Project owner [%s]: ", owner)
		ownerInput, _ := reader.ReadString('\n')
		ownerInput = strings.TrimSpace(ownerInput)
		if ownerInput != "" {
			owner = ownerInput
		}
	}

	if owner == "" {
		return fmt.Errorf("project owner is required")
	}

	// Prompt for project number
	cmd.Print("Project number: ")
	numberInput, _ := reader.ReadString('\n')
	numberInput = strings.TrimSpace(numberInput)
	projectNumber, err := strconv.Atoi(numberInput)
	if err != nil {
		return fmt.Errorf("invalid project number: %s", numberInput)
	}

	// Validate project exists
	cmd.Printf("Validating project %s/%d...\n", owner, projectNumber)
	client := api.NewClient()

	project, err := client.GetProject(owner, projectNumber)
	if err != nil {
		return fmt.Errorf("failed to find project: %w", err)
	}
	cmd.Printf("Found project: %s\n", project.Title)

	// Prompt for repository
	var repo string
	if defaultRepo != "" {
		cmd.Printf("Repository [%s]: ", defaultRepo)
		repoInput, _ := reader.ReadString('\n')
		repoInput = strings.TrimSpace(repoInput)
		if repoInput != "" {
			repo = repoInput
		} else {
			repo = defaultRepo
		}
	} else {
		cmd.Print("Repository (owner/repo): ")
		repoInput, _ := reader.ReadString('\n')
		repo = strings.TrimSpace(repoInput)
	}

	if repo == "" {
		return fmt.Errorf("repository is required")
	}

	// Fetch project fields
	cmd.Println("Fetching project fields...")
	fields, err := client.GetProjectFields(project.ID)
	if err != nil {
		cmd.Printf("Warning: could not fetch project fields: %v\n", err)
		fields = nil
	}

	// Convert to metadata
	metadata := &ProjectMetadata{
		ProjectID: project.ID,
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
		ProjectOwner:  owner,
		ProjectNumber: projectNumber,
		Repositories:  []string{repo},
	}

	// Write config
	cwd, _ := os.Getwd()
	if err := writeConfigWithMetadata(cwd, cfg, metadata); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	cmd.Println("Configuration saved to .gh-pmu.yml")
	cmd.Printf("Project: %s (#%d)\n", project.Title, project.Number)
	cmd.Printf("Repository: %s\n", repo)
	cmd.Printf("Fields cached: %d\n", len(fields))

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
	ProjectOwner  string
	ProjectNumber int
	Repositories  []string
}

// ConfigFile represents the .gh-pmu.yml file structure.
type ConfigFile struct {
	Project      ProjectConfig            `yaml:"project"`
	Repositories []string                 `yaml:"repositories"`
	Defaults     DefaultsConfig           `yaml:"defaults"`
	Fields       map[string]FieldMapping  `yaml:"fields"`
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

// ConfigFileWithMetadata extends ConfigFile with metadata section.
type ConfigFileWithMetadata struct {
	Project      ProjectConfig           `yaml:"project"`
	Repositories []string                `yaml:"repositories"`
	Defaults     DefaultsConfig          `yaml:"defaults"`
	Fields       map[string]FieldMapping `yaml:"fields"`
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
	configFile := &ConfigFile{
		Project: ProjectConfig{
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

	configFile := &ConfigFileWithMetadata{
		Project: ProjectConfig{
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
