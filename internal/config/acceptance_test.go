package config

import "testing"

func TestRequiresReAcceptance_MajorBump_ReturnsTrue(t *testing.T) {
	// ARRANGE: Accepted on 0.15.0, current is 1.0.0
	accepted := "0.15.0"
	current := "1.0.0"

	// ACT
	result := RequiresReAcceptance(accepted, current)

	// ASSERT
	if !result {
		t.Error("Expected re-acceptance required for major version bump")
	}
}

func TestRequiresReAcceptance_MinorBump_ReturnsTrue(t *testing.T) {
	// ARRANGE: Accepted on 0.15.0, current is 0.16.0
	accepted := "0.15.0"
	current := "0.16.0"

	// ACT
	result := RequiresReAcceptance(accepted, current)

	// ASSERT
	if !result {
		t.Error("Expected re-acceptance required for minor version bump")
	}
}

func TestRequiresReAcceptance_PatchBump_ReturnsFalse(t *testing.T) {
	// ARRANGE: Accepted on 0.15.0, current is 0.15.1
	accepted := "0.15.0"
	current := "0.15.1"

	// ACT
	result := RequiresReAcceptance(accepted, current)

	// ASSERT
	if result {
		t.Error("Expected no re-acceptance required for patch-only bump")
	}
}

func TestRequiresReAcceptance_SameVersion_ReturnsFalse(t *testing.T) {
	// ARRANGE: Same version
	accepted := "0.15.2"
	current := "0.15.2"

	// ACT
	result := RequiresReAcceptance(accepted, current)

	// ASSERT
	if result {
		t.Error("Expected no re-acceptance required for same version")
	}
}

func TestRequiresReAcceptance_EmptyAccepted_ReturnsTrue(t *testing.T) {
	// ARRANGE: No prior acceptance
	accepted := ""
	current := "0.15.2"

	// ACT
	result := RequiresReAcceptance(accepted, current)

	// ASSERT
	if !result {
		t.Error("Expected re-acceptance required when no prior acceptance")
	}
}

func TestRequiresReAcceptance_DevVersion_ReturnsFalse(t *testing.T) {
	// ARRANGE: Current is dev build
	accepted := "0.15.0"
	current := "dev"

	// ACT
	result := RequiresReAcceptance(accepted, current)

	// ASSERT
	if result {
		t.Error("Expected no re-acceptance required for dev builds")
	}
}

func TestRequiresReAcceptance_AcceptedOnDev_ReturnsTrue(t *testing.T) {
	// ARRANGE: Accepted on dev, current is real version
	accepted := "dev"
	current := "0.15.0"

	// ACT
	result := RequiresReAcceptance(accepted, current)

	// ASSERT
	if !result {
		t.Error("Expected re-acceptance required when accepted on dev build")
	}
}

func TestRequiresReAcceptance_MinorBumpMultiple_ReturnsTrue(t *testing.T) {
	// ARRANGE: Accepted on 0.14.3, current is 0.16.0
	accepted := "0.14.3"
	current := "0.16.0"

	// ACT
	result := RequiresReAcceptance(accepted, current)

	// ASSERT
	if !result {
		t.Error("Expected re-acceptance required for multi-minor version bump")
	}
}
