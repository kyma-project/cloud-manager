package awscertificate

import (
	"testing"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConvertTags(t *testing.T) {
	t.Run("Tags include cloud-manager name, Scope and Shoot", func(t *testing.T) {
		cert := &cloudresourcesv1beta1.AwsCertificate{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-cert",
				Labels: map[string]string{
					"custom-label": "custom-value",
				},
			},
		}

		scope := &cloudcontrolv1beta1.Scope{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-scope",
			},
			Spec: cloudcontrolv1beta1.ScopeSpec{
				ShootName: "test-shoot",
			},
		}

		tags := convertTags(cert, scope)

		// Verify required tags exist
		tagMap := make(map[string]string)
		for _, tag := range tags {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}

		assert.Equal(t, "test-cert", tagMap[common.TagCloudManagerName])
		assert.Equal(t, "cloud-manager", tagMap["kyma-project.io/managed-by"])
		assert.Equal(t, "test-scope", tagMap[common.TagScope])
		assert.Equal(t, "test-shoot", tagMap[common.TagShoot])
		assert.Equal(t, "custom-value", tagMap["custom-label"])
	})

	t.Run("Tags work without labels", func(t *testing.T) {
		cert := &cloudresourcesv1beta1.AwsCertificate{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-cert-no-labels",
			},
		}

		scope := &cloudcontrolv1beta1.Scope{
			ObjectMeta: metav1.ObjectMeta{
				Name: "scope-name",
			},
			Spec: cloudcontrolv1beta1.ScopeSpec{
				ShootName: "shoot-name",
			},
		}

		tags := convertTags(cert, scope)

		// Should have exactly 4 tags (cloud-manager-name, managed-by, Scope, Shoot)
		assert.Len(t, tags, 4)

		tagMap := make(map[string]string)
		for _, tag := range tags {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}

		assert.Equal(t, "test-cert-no-labels", tagMap[common.TagCloudManagerName])
		assert.Equal(t, "cloud-manager", tagMap["kyma-project.io/managed-by"])
		assert.Equal(t, "scope-name", tagMap[common.TagScope])
		assert.Equal(t, "shoot-name", tagMap[common.TagShoot])
	})

	t.Run("Tag pointers are properly set", func(t *testing.T) {
		cert := &cloudresourcesv1beta1.AwsCertificate{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-cert",
			},
		}

		scope := &cloudcontrolv1beta1.Scope{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-scope",
			},
			Spec: cloudcontrolv1beta1.ScopeSpec{
				ShootName: "test-shoot",
			},
		}

		tags := convertTags(cert, scope)

		// All tags should have non-nil Key and Value pointers
		for i, tag := range tags {
			assert.NotNil(t, tag.Key, "Tag %d Key should not be nil", i)
			assert.NotNil(t, tag.Value, "Tag %d Value should not be nil", i)
		}
	})
}
