package gcpnfsvolume

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-manager/pkg/composed"
	spy "github.com/kyma-project/cloud-manager/pkg/testinfra/clientspy"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"testing"
	"time"
)

func TestRemovePersistentVolumeClaimFinalizer(t *testing.T) {

	t.Run("removePersistentVolumeClaimFinalizer", func(t *testing.T) {

		var gcpNfsVolume *cloudresourcesv1beta1.GcpNfsVolume
		var pvc *corev1.PersistentVolumeClaim
		var state *State
		var k8sClient client.WithWatch

		createEmptyGcpNfsVolumeState := func(k8sClient client.WithWatch, gcpNfsVolume *cloudresourcesv1beta1.GcpNfsVolume) *State {
			cluster := composed.NewStateCluster(k8sClient, k8sClient, nil, k8sClient.Scheme())
			return &State{
				State: composed.NewStateFactory(cluster).NewState(types.NamespacedName{}, gcpNfsVolume),
			}
		}

		setupTest := func() {
			gcpNfsVolume = &cloudresourcesv1beta1.GcpNfsVolume{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-gcpnfsvol",
					Namespace: "test-ns",
					DeletionTimestamp: &v1.Time{
						Time: time.Now(),
					},
					Finalizers: []string{cloudresourcesv1beta1.Finalizer},
				},
			}

			pvc = &corev1.PersistentVolumeClaim{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-pvc",
					Namespace: "test-ns",
					DeletionTimestamp: &v1.Time{
						Time: time.Now(),
					},
					Finalizers: []string{cloudresourcesv1beta1.Finalizer},
				},
			}

			scheme := runtime.NewScheme()
			utilruntime.Must(clientgoscheme.AddToScheme(scheme))
			utilruntime.Must(cloudresourcesv1beta1.AddToScheme(scheme))
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(gcpNfsVolume).
				WithStatusSubresource(gcpNfsVolume).
				WithObjects(pvc).
				WithStatusSubresource(pvc).
				Build()

			k8sClient = spy.NewClientSpy(fakeClient)

			state = createEmptyGcpNfsVolumeState(k8sClient, gcpNfsVolume)
			state.PVC = pvc
		}

		t.Run("Should: delete finalizer", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err, res := removePersistenceVolumeClaimFinalizer(ctx, state)

			assert.Nil(t, res, "should return nil result")
			assert.Nil(t, err, "should return nil err")
			assert.EqualValues(t, 1, k8sClient.(spy.ClientSpy).UpdateCallCount(), "update should be called")
			assert.False(t, controllerutil.ContainsFinalizer(state.PVC, cloudresourcesv1beta1.Finalizer), "finalizer is removed")
		})

		t.Run("Should: do nothing if PVC is not marked for deletion", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			pvc.ObjectMeta.DeletionTimestamp = nil

			err, res := removePersistenceVolumeClaimFinalizer(ctx, state)

			assert.Nil(t, res, "should return nil result")
			assert.Nil(t, err, "should return nil err")
			assert.EqualValues(t, 0, k8sClient.(spy.ClientSpy).UpdateCallCount(), "update should not be called")
			assert.True(t, controllerutil.ContainsFinalizer(state.PVC, cloudresourcesv1beta1.Finalizer), "finalizer is not removed")
		})

		t.Run("Should: do nothing if PVC is not defined", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			state.PVC = nil

			err, res := removePersistenceVolumeClaimFinalizer(ctx, state)

			assert.Nil(t, res, "should return nil result")
			assert.Nil(t, err, "should return nil err")
			assert.EqualValues(t, 0, k8sClient.(spy.ClientSpy).UpdateCallCount(), "update should not be called")
			assert.True(t, controllerutil.ContainsFinalizer(pvc, cloudresourcesv1beta1.Finalizer), "finalizer is not removed")
		})

		t.Run("Should: do nothing if PVC does not contain Finalizer", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			pvc.ObjectMeta.Finalizers = nil
			state.PVC = pvc

			err, res := removePersistenceVolumeClaimFinalizer(ctx, state)

			assert.Nil(t, res, "should return nil result")
			assert.Nil(t, err, "should return nil err")
			assert.EqualValues(t, 0, k8sClient.(spy.ClientSpy).UpdateCallCount(), "update should not be called")
		})

	})
}
