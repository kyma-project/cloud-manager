package awsnfsvolume

import (
	"context"
	"testing"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/bootstrap"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	spy "github.com/kyma-project/cloud-manager/pkg/testinfra/clientspy"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSanitizeReleasedVolume(t *testing.T) {

	t.Run("sanitizeReleasedVolume", func(t *testing.T) {

		var awsNfsVolume *cloudresourcesv1beta1.AwsNfsVolume
		var pv *corev1.PersistentVolume
		var state *State
		var k8sClient client.WithWatch

		createEmptyAwsNfsVolumeState := func(k8sClient client.WithWatch, awsNfsVolume *cloudresourcesv1beta1.AwsNfsVolume) *State {
			cluster := composed.NewStateCluster(k8sClient, k8sClient, nil, k8sClient.Scheme())
			return &State{
				State: composed.NewStateFactory(cluster).NewState(types.NamespacedName{}, awsNfsVolume),
			}
		}

		setupTest := func() {
			awsNfsVolume = &cloudresourcesv1beta1.AwsNfsVolume{}

			pv = &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pv",
				},
				Spec: corev1.PersistentVolumeSpec{
					ClaimRef: &corev1.ObjectReference{
						UID: "013d3f5e-e780-4979-a5b9-a740aae7187c",
					},
				},
				Status: corev1.PersistentVolumeStatus{
					Phase: corev1.VolumeReleased,
				},
			}
			scheme := bootstrap.SkrScheme
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				Build()

			k8sClient = spy.NewClientSpy(fakeClient)

			state = createEmptyAwsNfsVolumeState(k8sClient, awsNfsVolume)
			state.Volume = pv
		}

		t.Run("Should: sanitize released PV and remove PVC ref", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			assert.NotNilf(t, pv.Spec.ClaimRef, "Claim ref not nil")

			err, _ := sanitizeReleasedVolume(ctx, state)

			assert.Nilf(t, pv.Spec.ClaimRef, "Claim ref is nil")
			assert.NotNil(t, err, "should return non-nilnil err") // not an actual err, but requeue
			assert.EqualValues(t, 1, k8sClient.(spy.ClientSpy).UpdateCallCount(), "update should be called")
		})

		t.Run("Should: do nothing if its marked for deletion", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			awsNfsVolume.ObjectMeta = metav1.ObjectMeta{
				DeletionTimestamp: &metav1.Time{
					Time: time.Now(),
				},
			}

			err, res := sanitizeReleasedVolume(ctx, state)

			assert.Nil(t, res, "should return nil result")
			assert.Nil(t, err, "should return nil err")
			assert.EqualValues(t, 0, k8sClient.(spy.ClientSpy).UpdateCallCount(), "update should not be called")
		})

		t.Run("Should: do nothing if Volume is notdefined in state", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			state.Volume = nil

			err, res := sanitizeReleasedVolume(ctx, state)

			assert.Nil(t, res, "should return nil result")
			assert.Nil(t, err, "should return nil err")
			assert.EqualValues(t, 0, k8sClient.(spy.ClientSpy).UpdateCallCount(), "update should not be called")
		})

		t.Run("Should: do nothing if state.Volume.status is not in released phase", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			state.Volume.Status.Phase = corev1.VolumeBound

			err, res := sanitizeReleasedVolume(ctx, state)

			assert.Nil(t, res, "should return nil result")
			assert.Nil(t, err, "should return nil err")
			assert.EqualValues(t, 0, k8sClient.(spy.ClientSpy).UpdateCallCount(), "update should not be called")
		})

		t.Run("Should: do nothing if state.Volume has no defined ClaimRef", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			state.Volume.Spec.ClaimRef = nil

			err, res := sanitizeReleasedVolume(ctx, state)

			assert.Nil(t, res, "should return nil result")
			assert.Nil(t, err, "should return nil err")
			assert.EqualValues(t, 0, k8sClient.(spy.ClientSpy).UpdateCallCount(), "update should not be called")
		})

	})
}
