package config

import (
	"fmt"
	"strconv"
	"strings"
)

// Acceptance holds the terms acceptance state for a repository.
type Acceptance struct {
	Accepted bool   `yaml:"accepted" json:"accepted"`
	User     string `yaml:"user,omitempty" json:"user,omitempty"`
	Date     string `yaml:"date,omitempty" json:"date,omitempty"`
	Version  string `yaml:"version,omitempty" json:"version,omitempty"`
}

// RequiresReAcceptance returns true if the user needs to (re-)accept terms.
// Re-acceptance is required when:
//   - No prior acceptance (empty acceptedVersion)
//   - Major or minor version changed
//   - Prior acceptance was on a dev build
//
// Re-acceptance is NOT required when:
//   - Only patch version changed
//   - Current build is "dev"
//   - Versions are identical
func RequiresReAcceptance(acceptedVersion, currentVersion string) bool {
	if acceptedVersion == "" {
		return true
	}

	// Dev builds never require re-acceptance
	if currentVersion == "dev" {
		return false
	}

	// Accepted on dev always requires re-acceptance for real versions
	if acceptedVersion == "dev" {
		return true
	}

	acceptedMajor, acceptedMinor, err := parseMajorMinor(acceptedVersion)
	if err != nil {
		return true
	}

	currentMajor, currentMinor, err := parseMajorMinor(currentVersion)
	if err != nil {
		return true
	}

	return acceptedMajor != currentMajor || acceptedMinor != currentMinor
}

// parseMajorMinor extracts major and minor version numbers from a semver string.
func parseMajorMinor(version string) (int, int, error) {
	parts := strings.SplitN(version, ".", 3)
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("invalid version: %s", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	return major, minor, nil
}
