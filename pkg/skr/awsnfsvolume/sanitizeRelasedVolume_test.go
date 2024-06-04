package awsnfsvolume

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSanitizeReleasedVolume(t *testing.T) {

	t.Run("sanitizeReleasedVolume", func(t *testing.T) {

		var awsNfsVolume *cloudresourcesv1beta1.AwsNfsVolume
		var pv *corev1.PersistentVolume
		var state *State

		setupTest := func() {
			awsNfsVolume = &cloudresourcesv1beta1.AwsNfsVolume{}

			pv = &corev1.PersistentVolume{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-pv",
					Namespace: "test-ns",
				},
				Spec: corev1.PersistentVolumeSpec{
					ClaimRef: &corev1.ObjectReference{
						UID: "d004f326-ca6e-4d22-9a02-a9493a13b3f6",
					},
				},
				Status: corev1.PersistentVolumeStatus{
					Phase: corev1.VolumeReleased,
				},
			}
			state = testStateFactory(awsNfsVolume)
			state.Volume = pv

		}

		t.Run("should sanitize released PV and remove PVC ref", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err, _ := sanitizeReleasedVolume(ctx, state)

			assert.Nilf(t, pv.Spec.ClaimRef, "Claim ref is nil")
			assert.NotNil(t, err, "should return non-nilnil err") // not an actual err, but requeue
		})

		t.Run("should do nothing if its marked for deletion", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			awsNfsVolume.ObjectMeta = v1.ObjectMeta{
				DeletionTimestamp: &v1.Time{
					Time: time.Now(),
				},
			}

			err, res := sanitizeReleasedVolume(ctx, state)

			assert.Nil(t, res, "should return nil result")
			assert.Nil(t, err, "should return nil err")
		})

		t.Run("should do nothing if Volume is notdefined in state", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			state.Volume = nil

			err, res := sanitizeReleasedVolume(ctx, state)

			assert.Nil(t, res, "should return nil result")
			assert.Nil(t, err, "should return nil err")
		})

		t.Run("should do nothing if state.Volume.status is not in released phase", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			state.Volume.Status.Phase = corev1.VolumeBound

			err, res := sanitizeReleasedVolume(ctx, state)

			assert.Nil(t, res, "should return nil result")
			assert.Nil(t, err, "should return nil err")
		})

		t.Run("should do nothing if state.Volume has no defined ClaimRef", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			state.Volume.Spec.ClaimRef = nil

			err, res := sanitizeReleasedVolume(ctx, state)

			assert.Nil(t, res, "should return nil result")
			assert.Nil(t, err, "should return nil err")
		})

		t.Run("should log error and return if k8s api call fails", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err, res := sanitizeReleasedVolume(ctx, state)

			assert.NotNil(t, res, "should return non-nil result")
			assert.NotNil(t, err, "should return non-nilnil err")
		})

	})
}
