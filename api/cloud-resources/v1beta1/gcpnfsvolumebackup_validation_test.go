package v1beta1

import (
	"regexp"
	"testing"
)

// TestGcpNfsVolumeBackupLocationValidation tests the Location field validation pattern
func TestGcpNfsVolumeBackupLocationValidation(t *testing.T) {
	// This is the exact pattern from the kubebuilder validation
	pattern := `^$|^(africa-south1|asia-east1|asia-east2|asia-northeast1|asia-northeast2|asia-northeast3|asia-south1|asia-south2|asia-southeast1|asia-southeast2|asia-southeast3|australia-southeast1|australia-southeast2|europe-central2|europe-north1|europe-southwest1|europe-west1|europe-west10|europe-west12|europe-west2|europe-west3|europe-west4|europe-west6|europe-west8|europe-west9|me-central1|me-central2|me-west1|northamerica-northeast1|northamerica-northeast2|southamerica-east1|southamerica-west1|us-central1|us-east1|us-east4|us-east5|us-east7|us-south1|us-west1|us-west2|us-west3|us-west4|us-west8)$`
	re := regexp.MustCompile(pattern)

	// Test valid cases - all GCP Filestore regions
	validRegions := []string{
		"", // empty is valid
		// Africa
		"africa-south1",
		// Asia
		"asia-east1",
		"asia-east2",
		"asia-northeast1",
		"asia-northeast2",
		"asia-northeast3",
		"asia-south1",
		"asia-south2",
		"asia-southeast1",
		"asia-southeast2",
		"asia-southeast3",
		// Australia
		"australia-southeast1",
		"australia-southeast2",
		// Europe
		"europe-central2",
		"europe-north1",
		"europe-southwest1",
		"europe-west1",
		"europe-west10",
		"europe-west12",
		"europe-west2",
		"europe-west3",
		"europe-west4",
		"europe-west6",
		"europe-west8",
		"europe-west9",
		// Middle East
		"me-central1",
		"me-central2",
		"me-west1",
		// North America
		"northamerica-northeast1",
		"northamerica-northeast2",
		// South America
		"southamerica-east1",
		"southamerica-west1",
		// US
		"us-central1",
		"us-east1",
		"us-east4",
		"us-east5",
		"us-east7",
		"us-south1",
		"us-west1",
		"us-west2",
		"us-west3",
		"us-west4",
		"us-west8",
	}

	for _, region := range validRegions {
		if !re.MatchString(region) {
			t.Errorf("Valid region %q should match pattern but doesn't", region)
		}
	}

	// Test invalid cases
	invalidRegions := []string{
		"invalid-region",
		"us-west99",
		"europe-east1",
		"asia-west1",
		"us-west1-a",    // zones are not allowed
		"us-central1-b", // zones are not allowed
		"UPPERCASE",
		"mixed-Case-1",
		"special@chars",
		"with spaces",
		"trailing-dash-",
		"-leading-dash",
		"double--dash",
	}

	for _, region := range invalidRegions {
		if re.MatchString(region) {
			t.Errorf("Invalid region %q should NOT match pattern but does", region)
		}
	}
}

// TestGcpNfsVolumeBackupNewRegions specifically tests newly added regions
func TestGcpNfsVolumeBackupNewRegions(t *testing.T) {
	pattern := `^$|^(africa-south1|asia-east1|asia-east2|asia-northeast1|asia-northeast2|asia-northeast3|asia-south1|asia-south2|asia-southeast1|asia-southeast2|asia-southeast3|australia-southeast1|australia-southeast2|europe-central2|europe-north1|europe-southwest1|europe-west1|europe-west10|europe-west12|europe-west2|europe-west3|europe-west4|europe-west6|europe-west8|europe-west9|me-central1|me-central2|me-west1|northamerica-northeast1|northamerica-northeast2|southamerica-east1|southamerica-west1|us-central1|us-east1|us-east4|us-east5|us-east7|us-south1|us-west1|us-west2|us-west3|us-west4|us-west8)$`
	re := regexp.MustCompile(pattern)

	// These are regions that were previously missing from the validation
	newlyAddedRegions := []string{
		"africa-south1",
		"asia-east2",
		"asia-south2",
		"asia-southeast3",
		"australia-southeast2",
		"europe-central2",
		"europe-southwest1",
		"europe-west8",
		"europe-west9",
		"europe-west10",
		"europe-west12",
		"me-central1",
		"me-central2",
		"me-west1",
		"northamerica-northeast2",
		"southamerica-west1",
		"us-east5",
		"us-east7",
		"us-south1",
		"us-west8",
	}

	for _, region := range newlyAddedRegions {
		if !re.MatchString(region) {
			t.Errorf("Newly added region %q should match pattern but doesn't", region)
		}
	}
}

// TestGcpNfsVolumeBackupRegionCount verifies all 43 GCP Filestore regions are included
func TestGcpNfsVolumeBackupRegionCount(t *testing.T) {
	pattern := `^$|^(africa-south1|asia-east1|asia-east2|asia-northeast1|asia-northeast2|asia-northeast3|asia-south1|asia-south2|asia-southeast1|asia-southeast2|asia-southeast3|australia-southeast1|australia-southeast2|europe-central2|europe-north1|europe-southwest1|europe-west1|europe-west10|europe-west12|europe-west2|europe-west3|europe-west4|europe-west6|europe-west8|europe-west9|me-central1|me-central2|me-west1|northamerica-northeast1|northamerica-northeast2|southamerica-east1|southamerica-west1|us-central1|us-east1|us-east4|us-east5|us-east7|us-south1|us-west1|us-west2|us-west3|us-west4|us-west8)$`
	re := regexp.MustCompile(pattern)

	allRegions := []string{
		"africa-south1",
		"asia-east1", "asia-east2", "asia-northeast1", "asia-northeast2", "asia-northeast3",
		"asia-south1", "asia-south2", "asia-southeast1", "asia-southeast2", "asia-southeast3",
		"australia-southeast1", "australia-southeast2",
		"europe-central2", "europe-north1", "europe-southwest1",
		"europe-west1", "europe-west10", "europe-west12", "europe-west2", "europe-west3",
		"europe-west4", "europe-west6", "europe-west8", "europe-west9",
		"me-central1", "me-central2", "me-west1",
		"northamerica-northeast1", "northamerica-northeast2",
		"southamerica-east1", "southamerica-west1",
		"us-central1", "us-east1", "us-east4", "us-east5", "us-east7", "us-south1",
		"us-west1", "us-west2", "us-west3", "us-west4", "us-west8",
	}

	expectedCount := 43
	if len(allRegions) != expectedCount {
		t.Errorf("Expected %d regions in test list, got %d", expectedCount, len(allRegions))
	}

	// Verify all regions match
	for _, region := range allRegions {
		if !re.MatchString(region) {
			t.Errorf("Region %q from comprehensive list should match pattern but doesn't", region)
		}
	}
}
