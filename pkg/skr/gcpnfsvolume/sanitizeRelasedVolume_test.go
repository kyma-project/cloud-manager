package gcpnfsvolume

import (
	"context"
	"testing"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	spy "github.com/kyma-project/cloud-manager/pkg/testinfra/clientspy"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSanitizeReleasedVolume(t *testing.T) {

	t.Run("sanitizeReleasedVolume", func(t *testing.T) {

		var gcpNfsVolume *cloudresourcesv1beta1.GcpNfsVolume
		var pv *corev1.PersistentVolume
		var state *State
		var k8sClient client.WithWatch

		createEmptyGcpNfsVolumeState := func(k8sClient client.WithWatch, gcpNfsVolume *cloudresourcesv1beta1.GcpNfsVolume) *State {
			cluster := composed.NewStateCluster(k8sClient, k8sClient, nil, k8sClient.Scheme())
			return &State{
				State: composed.NewStateFactory(cluster).NewState(types.NamespacedName{}, gcpNfsVolume),
			}
		}

		setupTest := func() {
			gcpNfsVolume = &cloudresourcesv1beta1.GcpNfsVolume{}

			pv = &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pv",
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
			scheme := runtime.NewScheme()
			utilruntime.Must(clientgoscheme.AddToScheme(scheme))
			utilruntime.Must(cloudresourcesv1beta1.AddToScheme(scheme))
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				Build()

			k8sClient = spy.NewClientSpy(fakeClient)

			state = createEmptyGcpNfsVolumeState(k8sClient, gcpNfsVolume)
			state.PV = pv
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
			gcpNfsVolume.ObjectMeta = metav1.ObjectMeta{
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
			state.PV = nil

			err, res := sanitizeReleasedVolume(ctx, state)

			assert.Nil(t, res, "should return nil result")
			assert.Nil(t, err, "should return nil err")
			assert.EqualValues(t, 0, k8sClient.(spy.ClientSpy).UpdateCallCount(), "update should not be called")
		})

		t.Run("Should: do nothing if state.Volume.status is not in released phase", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			state.PV.Status.Phase = corev1.VolumeBound

			err, res := sanitizeReleasedVolume(ctx, state)

			assert.Nil(t, res, "should return nil result")
			assert.Nil(t, err, "should return nil err")
			assert.EqualValues(t, 0, k8sClient.(spy.ClientSpy).UpdateCallCount(), "update should not be called")
		})

		t.Run("Should: do nothing if state.Volume has no defined ClaimRef", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			state.PV.Spec.ClaimRef = nil

			err, res := sanitizeReleasedVolume(ctx, state)

			assert.Nil(t, res, "should return nil result")
			assert.Nil(t, err, "should return nil err")
			assert.EqualValues(t, 0, k8sClient.(spy.ClientSpy).UpdateCallCount(), "update should not be called")
		})

	})
}
