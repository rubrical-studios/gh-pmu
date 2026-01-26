package cmd

import "github.com/rubrical-studios/gh-pmu/internal/api"

// Field name constants for project fields.
// The Branch field was previously named Release. For backward compatibility,
// we check for Branch first and fall back to Release for existing projects.
const (
	// BranchFieldName is the current name for the branch tracking field.
	BranchFieldName = "Branch"

	// LegacyReleaseFieldName is the legacy name for the branch tracking field.
	// Used for backward compatibility with existing projects.
	LegacyReleaseFieldName = "Release"
)

// ResolveBranchFieldName returns the appropriate field name for branch tracking
// based on available project fields. It checks for "Branch" first and falls back
// to "Release" for backward compatibility with existing projects.
func ResolveBranchFieldName(fields []api.ProjectField) string {
	hasBranch := false
	hasRelease := false

	for _, f := range fields {
		switch f.Name {
		case BranchFieldName:
			hasBranch = true
		case LegacyReleaseFieldName:
			hasRelease = true
		}
	}

	// Prefer Branch field if it exists
	if hasBranch {
		return BranchFieldName
	}

	// Fall back to Release for existing projects
	if hasRelease {
		return LegacyReleaseFieldName
	}

	// Default to Branch for new projects (field may not exist yet)
	return BranchFieldName
}

// ResolveBranchFieldNameFromFieldValues returns the appropriate field name for branch tracking
// based on field values. This is useful when you have FieldValue slices instead of ProjectField slices.
func ResolveBranchFieldNameFromFieldValues(fieldValues []api.FieldValue) string {
	for _, fv := range fieldValues {
		if fv.Field == BranchFieldName {
			return BranchFieldName
		}
	}
	// Default to legacy name if Branch not found in values
	return LegacyReleaseFieldName
}
