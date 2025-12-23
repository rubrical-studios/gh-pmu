package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the .gh-pmu.yml configuration file
type Config struct {
	Project      Project           `yaml:"project"`
	Repositories []string          `yaml:"repositories"`
	Framework    string            `yaml:"framework,omitempty"`
	Defaults     Defaults          `yaml:"defaults,omitempty"`
	Fields       map[string]Field  `yaml:"fields,omitempty"`
	Triage       map[string]Triage `yaml:"triage,omitempty"`
	Release      Release           `yaml:"release,omitempty"`
	Metadata     *Metadata         `yaml:"metadata,omitempty"`
	Cache        *Cache            `yaml:"cache,omitempty"`
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

// LoadFromDirectoryAndNormalize loads the config and normalizes the framework field.
// If the framework field is empty, it sets it to "IDPF" and saves the config.
// This ensures the config file is self-documenting about which framework is in use.
func LoadFromDirectoryAndNormalize(dir string) (*Config, error) {
	configPath, err := FindConfigFile(dir)
	if err != nil {
		return nil, err
	}

	cfg, err := Load(configPath)
	if err != nil {
		return nil, err
	}

	// Normalize: missing framework defaults to IDPF
	if cfg.Framework == "" {
		cfg.Framework = "IDPF"
		if err := cfg.Save(configPath); err != nil {
			// Log warning but don't fail - config is still usable
			// The next save operation will include the framework
			return cfg, nil
		}
	}

	return cfg, nil
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

// IsIDPF returns true if the config uses IDPF framework validation.
// IDPF is the default framework when not specified.
func (c *Config) IsIDPF() bool {
	return c.Framework == "IDPF" || c.Framework == "idpf"
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
	Active    []ActiveRelease        `yaml:"active,omitempty"`
	Coverage  *CoverageConfig        `yaml:"coverage,omitempty"`
}

// CoverageConfig contains configuration for release coverage gates
type CoverageConfig struct {
	Enabled      *bool    `yaml:"enabled,omitempty"`       // Enable coverage gate (default: true)
	Threshold    int      `yaml:"threshold,omitempty"`     // Minimum patch coverage % (default: 80)
	SkipPatterns []string `yaml:"skip_patterns,omitempty"` // Patterns to exclude from analysis
}

// ActiveRelease represents an active release in the config
type ActiveRelease struct {
	Version      string `yaml:"version"`
	TrackerIssue int    `yaml:"tracker_issue"`
	Track        string `yaml:"track"`
}

// ArtifactConfig contains configuration for release artifacts
type ArtifactConfig struct {
	Directory    string `yaml:"directory,omitempty"`     // Base directory (default: "Releases")
	ReleaseNotes bool   `yaml:"release_notes,omitempty"` // Generate release-notes.md (default: true)
	Changelog    bool   `yaml:"changelog,omitempty"`     // Generate changelog.md (default: true)
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

// Cache contains cached tracker data for fast list operations
type Cache struct {
	Releases     []CachedTracker `yaml:"releases,omitempty"`
	Microsprints []CachedTracker `yaml:"microsprints,omitempty"`
}

// CachedTracker represents a cached release or microsprint tracker issue
type CachedTracker struct {
	Number int    `yaml:"number"`
	Title  string `yaml:"title"`
	State  string `yaml:"state"` // "OPEN" or "CLOSED"
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

// IsCoverageGateEnabled returns whether coverage gate is enabled (default: true)
func (c *Config) IsCoverageGateEnabled() bool {
	if c.Release.Coverage == nil || c.Release.Coverage.Enabled == nil {
		return true // Default to enabled
	}
	return *c.Release.Coverage.Enabled
}

// GetCoverageThreshold returns the minimum patch coverage percentage (default: 80)
func (c *Config) GetCoverageThreshold() int {
	if c.Release.Coverage == nil || c.Release.Coverage.Threshold == 0 {
		return 80 // Default threshold
	}
	return c.Release.Coverage.Threshold
}

// GetCoverageSkipPatterns returns patterns to exclude from coverage analysis
func (c *Config) GetCoverageSkipPatterns() []string {
	if c.Release.Coverage == nil {
		return []string{"*_test.go", "mock_*.go"}
	}
	if len(c.Release.Coverage.SkipPatterns) == 0 {
		return []string{"*_test.go", "mock_*.go"}
	}
	return c.Release.Coverage.SkipPatterns
}

// AddActiveRelease adds a release to the active list if not already present
func (c *Config) AddActiveRelease(release ActiveRelease) {
	// Check for duplicates by tracker issue number
	for _, r := range c.Release.Active {
		if r.TrackerIssue == release.TrackerIssue {
			return // Already exists
		}
	}
	c.Release.Active = append(c.Release.Active, release)
}

// RemoveActiveRelease removes a release from the active list by tracker issue number
func (c *Config) RemoveActiveRelease(trackerIssue int) {
	var filtered []ActiveRelease
	for _, r := range c.Release.Active {
		if r.TrackerIssue != trackerIssue {
			filtered = append(filtered, r)
		}
	}
	c.Release.Active = filtered
}

// GetActiveReleases returns the list of active releases
func (c *Config) GetActiveReleases() []ActiveRelease {
	return c.Release.Active
}

// MergeActiveReleases merges discovered releases into the config (additive, no duplicates)
func (c *Config) MergeActiveReleases(releases []ActiveRelease) {
	for _, r := range releases {
		c.AddActiveRelease(r)
	}
}

// HasCachedReleases returns true if cached release data exists
func (c *Config) HasCachedReleases() bool {
	return c.Cache != nil && len(c.Cache.Releases) > 0
}

// HasCachedMicrosprints returns true if cached microsprint data exists
func (c *Config) HasCachedMicrosprints() bool {
	return c.Cache != nil && len(c.Cache.Microsprints) > 0
}

// GetCachedReleases returns cached release trackers
func (c *Config) GetCachedReleases() []CachedTracker {
	if c.Cache == nil {
		return nil
	}
	return c.Cache.Releases
}

// GetCachedMicrosprints returns cached microsprint trackers
func (c *Config) GetCachedMicrosprints() []CachedTracker {
	if c.Cache == nil {
		return nil
	}
	return c.Cache.Microsprints
}

// SetCachedReleases updates the cached release tracker data
func (c *Config) SetCachedReleases(trackers []CachedTracker) {
	if c.Cache == nil {
		c.Cache = &Cache{}
	}
	c.Cache.Releases = trackers
}

// SetCachedMicrosprints updates the cached microsprint tracker data
func (c *Config) SetCachedMicrosprints(trackers []CachedTracker) {
	if c.Cache == nil {
		c.Cache = &Cache{}
	}
	c.Cache.Microsprints = trackers
}

// UpdateCachedTracker updates or adds a single tracker in the cache
func (c *Config) UpdateCachedTracker(trackerType string, tracker CachedTracker) {
	if c.Cache == nil {
		c.Cache = &Cache{}
	}

	var trackers *[]CachedTracker
	if trackerType == "release" {
		trackers = &c.Cache.Releases
	} else if trackerType == "microsprint" {
		trackers = &c.Cache.Microsprints
	} else {
		return
	}

	// Update existing or append new
	for i, t := range *trackers {
		if t.Number == tracker.Number {
			(*trackers)[i] = tracker
			return
		}
	}
	*trackers = append(*trackers, tracker)
}

// TempDirName is the name of the temporary directory within the project root
const TempDirName = "tmp"

// GetProjectRoot returns the directory containing .gh-pmu.yml.
// It searches from the current working directory up the directory tree.
func GetProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	configPath, err := FindConfigFile(cwd)
	if err != nil {
		return "", err
	}

	return filepath.Dir(configPath), nil
}

// GetTempDir returns the path to the project's tmp directory and creates it if needed.
// It also ensures tmp/ is in .gitignore.
func GetTempDir() (string, error) {
	projectRoot, err := GetProjectRoot()
	if err != nil {
		return "", err
	}

	tempDir := filepath.Join(projectRoot, TempDirName)

	// Create tmp directory if it doesn't exist
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Ensure tmp/ is in .gitignore
	if err := ensureGitignore(projectRoot); err != nil {
		// Log warning but don't fail - gitignore is nice-to-have
		fmt.Fprintf(os.Stderr, "Warning: could not update .gitignore: %v\n", err)
	}

	return tempDir, nil
}

// CreateTempFile creates a temporary file in the project's tmp directory.
// The pattern follows os.CreateTemp conventions (e.g., "prefix-*.suffix").
// The caller is responsible for closing and removing the file.
func CreateTempFile(pattern string) (*os.File, error) {
	tempDir, err := GetTempDir()
	if err != nil {
		return nil, err
	}

	return os.CreateTemp(tempDir, pattern)
}

// ensureGitignore adds tmp/ to .gitignore if not already present
func ensureGitignore(projectRoot string) error {
	gitignorePath := filepath.Join(projectRoot, ".gitignore")

	// Check if tmp/ is already in .gitignore
	if file, err := os.Open(gitignorePath); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == TempDirName || line == TempDirName+"/" {
				return nil // Already present
			}
		}
	}

	// Append tmp/ to .gitignore
	file, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open .gitignore: %w", err)
	}
	defer file.Close()

	// Check if file is empty or ends with newline
	info, _ := file.Stat()
	needsNewline := info.Size() > 0

	var content string
	if needsNewline {
		// Read last byte to check if it's a newline
		if f, err := os.Open(gitignorePath); err == nil {
			defer f.Close()
			buf := make([]byte, 1)
			if _, err := f.ReadAt(buf, info.Size()-1); err == nil && buf[0] != '\n' {
				content = "\n"
			}
		}
	}
	content += TempDirName + "/\n"

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write to .gitignore: %w", err)
	}

	return nil
}
