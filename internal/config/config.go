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
