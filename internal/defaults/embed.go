// Package defaults provides embedded default configuration for gh-pmu.
package defaults

import (
	_ "embed"

	"gopkg.in/yaml.v3"
)

//go:embed defaults.yml
var defaultsYAML []byte

// Defaults holds the parsed default configuration.
type Defaults struct {
	Labels []LabelDef   `yaml:"labels"`
	Fields FieldsDef    `yaml:"fields"`
}

// LabelDef represents a label definition.
type LabelDef struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Color       string `yaml:"color"`
}

// FieldsDef holds field definitions separated by requirement level.
type FieldsDef struct {
	Required        []FieldDef `yaml:"required"`
	CreateIfMissing []FieldDef `yaml:"create_if_missing"`
}

// FieldDef represents a project field definition.
type FieldDef struct {
	Name    string   `yaml:"name"`
	Type    string   `yaml:"type"`
	Options []string `yaml:"options,omitempty"`
}

// Load parses and returns the embedded defaults.
func Load() (*Defaults, error) {
	var d Defaults
	if err := yaml.Unmarshal(defaultsYAML, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// MustLoad parses and returns the embedded defaults, panicking on error.
func MustLoad() *Defaults {
	d, err := Load()
	if err != nil {
		panic("failed to load embedded defaults: " + err.Error())
	}
	return d
}
