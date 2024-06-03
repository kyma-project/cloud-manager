package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabels(t *testing.T) {
	t.Run("LabelsBuilder", func(t *testing.T) {

		t.Run("should define provided name label", func(t *testing.T) {
			builder := NewLabelBuilder()
			nameValue := "test-name-value"

			labels := builder.WithName(nameValue).Build()

			assert.Equal(t, labels[WellKnownK8sNameLabel], nameValue, "name label equals expected value")
		})

		t.Run("should define provided instance label", func(t *testing.T) {
			builder := NewLabelBuilder()
			instanceValue := "test-instance-value"

			labels := builder.WithInstance(instanceValue).Build()

			assert.Equal(t, labels[WellKnownK8sInstanceLabel], instanceValue, "instance label equals expected value")
		})

		t.Run("should define provided version label", func(t *testing.T) {
			builder := NewLabelBuilder()
			versionValue := "test-version-value"

			labels := builder.WithVersion(versionValue).Build()

			assert.Equal(t, labels[WellKnownK8sVersionLabel], versionValue, "version label equals expected value")
		})

		t.Run("should define provided component label", func(t *testing.T) {
			builder := NewLabelBuilder()
			componentValue := "test-component-value"

			labels := builder.WithComponent(componentValue).Build()

			assert.Equal(t, labels[WellKnownK8sComponentLabel], componentValue, "component label equals expected value")
		})

		t.Run("should define provided part-of label", func(t *testing.T) {
			builder := NewLabelBuilder()
			partOfValue := "test-part-of-value"

			labels := builder.WithPartOf(partOfValue).Build()

			assert.Equal(t, labels[WellKnownK8sPartOfLabel], partOfValue, "part-of label equals expected value")
		})

		t.Run("should define provided managed-by label", func(t *testing.T) {
			builder := NewLabelBuilder()
			managedBy := "test-managed-by-value"

			labels := builder.WithManagedBy(managedBy).Build()

			assert.Equal(t, labels[WellKnownK8sManagedByLabel], managedBy, "managed-by label equals expected value")
		})

		t.Run("should define default cloud manager labels", func(t *testing.T) {
			builder := NewLabelBuilder()

			labels := builder.WithCloudManagerDefaults().Build()

			assert.Equal(t, labels[WellKnownK8sComponentLabel], DefaultCloudManagerComponentLabelValue, "component label equals expected value")
			assert.Equal(t, labels[WellKnownK8sPartOfLabel], DefaultCloudManagerPartOfLabelValue, "part-of label equals expected value")
			assert.Equal(t, labels[WellKnownK8sManagedByLabel], DefaultCloudManagerManagedByLabelValue, "managed-by label equals expected value")
		})

		t.Run("should define custom label", func(t *testing.T) {
			builder := NewLabelBuilder()
			customLabelName := "foo.test.io/custom-label-for-test"
			customLabelValue := "this-is-a-test-label-value"

			labels := builder.WithCustomLabel(customLabelName, customLabelValue).Build()

			assert.Equal(t, labels[customLabelName], customLabelValue, "custom label is defined and has expecteed value")
		})

		t.Run("should define custom labels", func(t *testing.T) {
			builder := NewLabelBuilder()
			customLabelName := "foo.test.io/custom-label-for-test"
			customLabelValue := "this-is-a-test-label-value"
			otherCustomLabelName := "foo.test.io/other-custom-label-for-test"
			otherCustomLabelValue := "this-is-a-test-label-value-for-other-value"
			customLabels := map[string]string{
				customLabelName:      customLabelValue,
				otherCustomLabelName: otherCustomLabelValue,
			}

			labels := builder.WithCustomLabels(customLabels).Build()

			assert.Equal(t, labels[customLabelName], customLabelValue, "custom label is defined and has expecteed value")
			assert.Equal(t, labels[otherCustomLabelName], otherCustomLabelValue, "other custom label is defined and has expecteed value")
		})

	})
}
