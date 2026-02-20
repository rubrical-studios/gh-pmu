// Package defaults provides embedded default configuration for gh-pmu.
package defaults

import (
	_ "embed"

	"gopkg.in/yaml.v3"
)

//go:embed defaults.yml
var defaultsYAML []byte

//go:embed terms.txt
var termsText string

// Defaults holds the parsed default configuration.
type Defaults struct {
	Labels []LabelDef `yaml:"labels"`
	Fields FieldsDef  `yaml:"fields"`
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

// GetLabel returns the label definition for a given name, or nil if not found.
func (d *Defaults) GetLabel(name string) *LabelDef {
	for i := range d.Labels {
		if d.Labels[i].Name == name {
			return &d.Labels[i]
		}
	}
	return nil
}

// GetLabelNames returns a slice of all standard label names.
func (d *Defaults) GetLabelNames() []string {
	names := make([]string, len(d.Labels))
	for i := range d.Labels {
		names[i] = d.Labels[i].Name
	}
	return names
}

// IsStandardLabel returns true if the given label name is a standard label.
func (d *Defaults) IsStandardLabel(name string) bool {
	return d.GetLabel(name) != nil
}

// Terms returns the embedded terms and conditions text.
func Terms() string {
	return termsText
}
