package api_tests

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

type testGcpNfsVolumeBackupBuilder struct {
	instance cloudresourcesv1beta1.GcpNfsVolumeBackup
}

func newTestGcpNfsVolumeBackupBuilder() *testGcpNfsVolumeBackupBuilder {
	return &testGcpNfsVolumeBackupBuilder{
		instance: cloudresourcesv1beta1.GcpNfsVolumeBackup{
			Spec: cloudresourcesv1beta1.GcpNfsVolumeBackupSpec{
				Source: cloudresourcesv1beta1.GcpNfsVolumeBackupSource{
					Volume: cloudresourcesv1beta1.GcpNfsVolumeRef{
						Name:      "test-volume",
						Namespace: "default",
					},
				},
			},
		},
	}
}

func (b *testGcpNfsVolumeBackupBuilder) Build() *cloudresourcesv1beta1.GcpNfsVolumeBackup {
	return &b.instance
}

func (b *testGcpNfsVolumeBackupBuilder) WithLocation(location string) *testGcpNfsVolumeBackupBuilder {
	b.instance.Spec.Location = location
	return b
}

var _ = Describe("Feature: SKR GcpNfsVolumeBackup", Ordered, func() {

	Context("Scenario: Location validation", func() {

		canCreateSkr(
			"GcpNfsVolumeBackup with empty location",
			newTestGcpNfsVolumeBackupBuilder().WithLocation(""),
		)

		// Test all valid GCP Filestore regions (43 regions)
		validRegions := []string{
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
			region := region // capture loop variable
			canCreateSkr(
				"GcpNfsVolumeBackup with location "+region,
				newTestGcpNfsVolumeBackupBuilder().WithLocation(region),
			)
		}

		// Test invalid regions
		invalidRegions := []string{
			"invalid-region",
			"us-west99",
			"europe-east1",
			"asia-west1",
			"us-west1-a", // zones are not allowed, only regions
			"us-central1-b",
			"UPPERCASE",
			"mixed-Case-1",
			"special@chars",
			"with spaces",
			"trailing-dash-",
			"-leading-dash",
			"double--dash",
		}

		for _, region := range invalidRegions {
			region := region // capture loop variable
			canNotCreateSkr(
				"GcpNfsVolumeBackup with invalid location "+region,
				newTestGcpNfsVolumeBackupBuilder().WithLocation(region),
				"location in body should match",
			)
		}

		canNotChangeSkr(
			"GcpNfsVolumeBackup Location cannot be changed",
			newTestGcpNfsVolumeBackupBuilder().WithLocation("us-central1"),
			func(b Builder[*cloudresourcesv1beta1.GcpNfsVolumeBackup]) {
				b.(*testGcpNfsVolumeBackupBuilder).WithLocation("us-west1")
			},
			"Location is immutable",
		)

		canNotChangeSkr(
			"GcpNfsVolumeBackup Location from empty to region",
			newTestGcpNfsVolumeBackupBuilder().WithLocation(""),
			func(b Builder[*cloudresourcesv1beta1.GcpNfsVolumeBackup]) {
				b.(*testGcpNfsVolumeBackupBuilder).WithLocation("us-central1")
			},
			"Location is immutable",
		)

		canNotChangeSkr(
			"GcpNfsVolumeBackup Location from region to empty",
			newTestGcpNfsVolumeBackupBuilder().WithLocation("us-central1"),
			func(b Builder[*cloudresourcesv1beta1.GcpNfsVolumeBackup]) {
				b.(*testGcpNfsVolumeBackupBuilder).WithLocation("")
			},
			"Location is immutable",
		)

		// Test sample of newly added regions specifically
		canCreateSkr(
			"GcpNfsVolumeBackup with africa-south1",
			newTestGcpNfsVolumeBackupBuilder().WithLocation("africa-south1"),
		)

		canCreateSkr(
			"GcpNfsVolumeBackup with asia-east2",
			newTestGcpNfsVolumeBackupBuilder().WithLocation("asia-east2"),
		)

		canCreateSkr(
			"GcpNfsVolumeBackup with europe-central2",
			newTestGcpNfsVolumeBackupBuilder().WithLocation("europe-central2"),
		)

		canCreateSkr(
			"GcpNfsVolumeBackup with me-central1",
			newTestGcpNfsVolumeBackupBuilder().WithLocation("me-central1"),
		)

		canCreateSkr(
			"GcpNfsVolumeBackup with us-south1",
			newTestGcpNfsVolumeBackupBuilder().WithLocation("us-south1"),
		)
	})
})
