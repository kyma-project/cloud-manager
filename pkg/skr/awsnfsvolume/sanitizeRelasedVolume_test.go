package awsnfsvolume

import (
	"testing"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"

	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSanitizeReleasedVolume(t *testing.T) {
	t.Run("sanitizeReleasedVolume", func(t *testing.T) {

		t.Run("should sanitize released PV and remove PVC ref", func(t *testing.T) {

		})

		t.Run("should do nothing if its marked for deletion", func(t *testing.T) {
		})

		t.Run("should do nothing if state is not defined", func(t *testing.T) {
		})

		t.Run("should do nothing if state.Volume is not in released phase", func(t *testing.T) {
		})

		t.Run("should do nothing if state.Volume has no defined ClaimRef", func(t *testing.T) {
		})

		t.Run("should log error and return if k8s api call fails", func(t *testing.T) {
		})

	})
}

func awsNfsInstanceTestFactory() cloudresourcesv1beta1.AwsNfsVolume {
	return cloudresourcesv1beta1.AwsNfsVolume{
		ObjectMeta: v1.ObjectMeta{
			DeletionTimestamp: &v1.Time{
				Time: time.Now(),
			},
		},
	}
}

func persistanteVolumeFactory() corev1.PersistentVolume {
	return corev1.PersistentVolume{
		Spec:   corev1.PersistentVolumeSpec{},
		Status: corev1.PersistentVolumeStatus{},
	}
}
