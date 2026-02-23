package version

import "testing"

func TestVersion_IsNotEmpty(t *testing.T) {
	if Version == "" {
		t.Error("Version constant should not be empty")
	}
}

func TestVersion_IsValidSemver(t *testing.T) {
	// Version should look like a semver string (e.g., "0.15.1")
	// At minimum, it should contain at least one dot
	hasDot := false
	for _, c := range Version {
		if c == '.' {
			hasDot = true
			break
		}
	}
	if !hasDot {
		t.Errorf("Version %q does not look like a semver string", Version)
	}
}
