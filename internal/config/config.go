package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config represents the .gh-pmu.yml configuration file
type Config struct {
	Project      Project           `yaml:"project"`
	Repositories []string          `yaml:"repositories"`
	Defaults     Defaults          `yaml:"defaults,omitempty"`
	Fields       map[string]Field  `yaml:"fields,omitempty"`
	Triage       map[string]Triage `yaml:"triage,omitempty"`
	Release      Release           `yaml:"release,omitempty"`
	Metadata     *Metadata         `yaml:"metadata,omitempty"`
}

// Project contains GitHub project configuration
type Project struct {
	Name   string `yaml:"name,omitempty"`
	Number int    `yaml:"number"`
	Owner  string `yaml:"owner"`
}

// Defaults contains default values for new issues
type Defaults struct {
	Priority string   `yaml:"priority,omitempty"`
	Status   string   `yaml:"status,omitempty"`
	Labels   []string `yaml:"labels,omitempty"`
}

// Field maps field aliases to GitHub project field names and values
type Field struct {
	Field  string            `yaml:"field"`
	Values map[string]string `yaml:"values,omitempty"`
}

// Triage contains configuration for triage rules
type Triage struct {
	Query       string            `yaml:"query"`
	Apply       TriageApply       `yaml:"apply,omitempty"`
	Interactive TriageInteractive `yaml:"interactive,omitempty"`
}

// TriageApply contains fields to apply during triage
type TriageApply struct {
	Labels []string          `yaml:"labels,omitempty"`
	Fields map[string]string `yaml:"fields,omitempty"`
}

// TriageInteractive contains interactive prompts for triage
type TriageInteractive struct {
	Status   bool `yaml:"status,omitempty"`
	Estimate bool `yaml:"estimate,omitempty"`
}

// Metadata contains cached project metadata from GitHub API
type Metadata struct {
	Project ProjectMetadata `yaml:"project,omitempty"`
	Fields  []FieldMetadata `yaml:"fields,omitempty"`
}

// ProjectMetadata contains cached project info
type ProjectMetadata struct {
	ID string `yaml:"id,omitempty"`
}

// FieldMetadata contains cached field info
type FieldMetadata struct {
	Name     string           `yaml:"name"`
	ID       string           `yaml:"id"`
	DataType string           `yaml:"data_type"`
	Options  []OptionMetadata `yaml:"options,omitempty"`
}

// OptionMetadata contains cached field option info
type OptionMetadata struct {
	Name string `yaml:"name"`
	ID   string `yaml:"id"`
}

// ConfigFileName is the default configuration file name
const ConfigFileName = ".gh-pmu.yml"

// Load reads and parses a configuration file from the given path
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// LoadFromDirectory finds and loads the config file from the given directory.
// It searches up the directory tree until it finds a .gh-pmu.yml file or
// reaches the filesystem root.
func LoadFromDirectory(dir string) (*Config, error) {
	configPath, err := FindConfigFile(dir)
	if err != nil {
		return nil, err
	}
	return Load(configPath)
}

// FindConfigFile searches for .gh-pmu.yml starting from dir and walking up
// the directory tree until found or filesystem root is reached.
func FindConfigFile(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	for {
		configPath := filepath.Join(dir, ConfigFileName)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return "", fmt.Errorf("no %s found in %s or any parent directory", ConfigFileName, startDir)
		}
		dir = parent
	}
}

// Validate checks that required configuration fields are present
func (c *Config) Validate() error {
	if c.Project.Owner == "" {
		return fmt.Errorf("project.owner is required")
	}

	if c.Project.Number == 0 {
		return fmt.Errorf("project.number is required")
	}

	if len(c.Repositories) == 0 {
		return fmt.Errorf("at least one repository is required")
	}

	return nil
}

// ResolveFieldValue maps an alias to its actual GitHub field value.
// If no alias is found, returns the original value unchanged.
func (c *Config) ResolveFieldValue(fieldKey, alias string) string {
	field, ok := c.Fields[fieldKey]
	if !ok {
		return alias
	}

	if actual, ok := field.Values[alias]; ok {
		return actual
	}

	return alias
}

// GetFieldName returns the actual GitHub field name for a given key.
// If no mapping exists, returns the original key unchanged.
func (c *Config) GetFieldName(fieldKey string) string {
	field, ok := c.Fields[fieldKey]
	if !ok {
		return fieldKey
	}

	if field.Field != "" {
		return field.Field
	}

	return fieldKey
}

// ApplyEnvOverrides applies environment variable overrides to the config.
// Supported environment variables:
//   - GH_PM_PROJECT_OWNER: overrides project.owner
//   - GH_PM_PROJECT_NUMBER: overrides project.number
func (c *Config) ApplyEnvOverrides() {
	if owner := os.Getenv("GH_PM_PROJECT_OWNER"); owner != "" {
		c.Project.Owner = owner
	}

	if numberStr := os.Getenv("GH_PM_PROJECT_NUMBER"); numberStr != "" {
		if number, err := strconv.Atoi(numberStr); err == nil {
			c.Project.Number = number
		}
	}
}

// Save writes the configuration back to the given path
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// AddFieldMetadata adds or updates field metadata in the config
func (c *Config) AddFieldMetadata(field FieldMetadata) {
	if c.Metadata == nil {
		c.Metadata = &Metadata{}
	}

	// Check if field already exists, update if so
	for i, f := range c.Metadata.Fields {
		if f.Name == field.Name {
			c.Metadata.Fields[i] = field
			return
		}
	}

	// Add new field
	c.Metadata.Fields = append(c.Metadata.Fields, field)
}

// Release contains release management configuration
type Release struct {
	Tracks    map[string]TrackConfig `yaml:"tracks,omitempty"`
	Artifacts *ArtifactConfig        `yaml:"artifacts,omitempty"`
}

// ArtifactConfig contains configuration for release artifacts
type ArtifactConfig struct {
	Directory     string          `yaml:"directory,omitempty"`     // Base directory (default: "Releases")
	ReleaseNotes  bool            `yaml:"release_notes,omitempty"` // Generate release-notes.md (default: true)
	Changelog     bool            `yaml:"changelog,omitempty"`     // Generate changelog.md (default: true)
}

// TrackConfig contains configuration for a release track
type TrackConfig struct {
	Prefix      string            `yaml:"prefix"`
	Default     bool              `yaml:"default,omitempty"`
	Constraints *TrackConstraints `yaml:"constraints,omitempty"`
}

// TrackConstraints contains constraints for a release track
type TrackConstraints struct {
	Version string            `yaml:"version,omitempty"` // e.g., "patch_only"
	Labels  *LabelConstraints `yaml:"labels,omitempty"`
}

// LabelConstraints contains label requirements for a track
type LabelConstraints struct {
	Required  []string `yaml:"required,omitempty"`
	Forbidden []string `yaml:"forbidden,omitempty"`
}

// GetTrackPrefix returns the prefix for a given track name
// Returns "v" for stable track if not configured
func (c *Config) GetTrackPrefix(track string) string {
	if c.Release.Tracks == nil {
		// Default prefixes when not configured
		switch track {
		case "stable", "":
			return "v"
		default:
			return track + "/"
		}
	}

	if cfg, ok := c.Release.Tracks[track]; ok {
		return cfg.Prefix
	}

	// Default for unconfigured tracks
	if track == "stable" || track == "" {
		return "v"
	}
	return track + "/"
}

// GetDefaultTrack returns the default track name
func (c *Config) GetDefaultTrack() string {
	if c.Release.Tracks != nil {
		for name, cfg := range c.Release.Tracks {
			if cfg.Default {
				return name
			}
		}
	}
	return "stable"
}

// GetTrackConstraints returns constraints for a track, or nil if none
func (c *Config) GetTrackConstraints(track string) *TrackConstraints {
	if c.Release.Tracks == nil {
		return nil
	}
	if cfg, ok := c.Release.Tracks[track]; ok {
		return cfg.Constraints
	}
	return nil
}

// FormatReleaseFieldValue formats a version with the track prefix
func (c *Config) FormatReleaseFieldValue(version, track string) string {
	prefix := c.GetTrackPrefix(track)
	return prefix + version
}

// GetArtifactDirectory returns the base artifact directory
func (c *Config) GetArtifactDirectory() string {
	if c.Release.Artifacts != nil && c.Release.Artifacts.Directory != "" {
		return c.Release.Artifacts.Directory
	}
	return "Releases"
}

// GetArtifactPath returns the full artifact path for a release
// For stable: Releases/v1.0.0
// For other tracks: Releases/patch/v1.1.1
func (c *Config) GetArtifactPath(version, track string) string {
	baseDir := c.GetArtifactDirectory()
	if track == "stable" || track == "" {
		return fmt.Sprintf("%s/%s", baseDir, version)
	}
	return fmt.Sprintf("%s/%s/%s", baseDir, track, version)
}

// ShouldGenerateReleaseNotes returns whether release notes should be generated
func (c *Config) ShouldGenerateReleaseNotes() bool {
	if c.Release.Artifacts == nil {
		return true // Default to true
	}
	return c.Release.Artifacts.ReleaseNotes
}

// ShouldGenerateChangelog returns whether changelog should be generated
func (c *Config) ShouldGenerateChangelog() bool {
	if c.Release.Artifacts == nil {
		return true // Default to true
	}
	return c.Release.Artifacts.Changelog
}
