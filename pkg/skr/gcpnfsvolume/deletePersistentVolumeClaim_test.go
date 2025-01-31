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
	"testing"
	"time"
)

func TestDeletePersistentVolumeClaim(t *testing.T) {

	t.Run("createPersistentVolumeClaim", func(t *testing.T) {

		var gcpNfsVolume *cloudresourcesv1beta1.GcpNfsVolume
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
					Finalizers: []string{api.CommonFinalizerDeletionHook},
				},
			}

			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-pvc",
					Namespace: "test-ns",
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

		t.Run("Should: invoke delete API call if PVC is found and not marked for deletion", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err, res := deletePersistentVolumeClaim(ctx, state)

			assert.Nil(t, res, "should return nil result")
			assert.Equal(t, err, composed.StopWithRequeue, "should return StopWithRequeue err")
			assert.EqualValues(t, 1, k8sClient.(spy.ClientSpy).DeleteCallCount(), "delete should be called")
		})

		t.Run("Should: do nothing if GcpNfsVolume is not marked for deletion", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			gcpNfsVolume.ObjectMeta.DeletionTimestamp = nil

			err, res := deletePersistentVolumeClaim(ctx, state)

			assert.Nil(t, res, "should return nil result")
			assert.Nil(t, err, "should return nil err")
			assert.EqualValues(t, 0, k8sClient.(spy.ClientSpy).DeleteCallCount(), "delete should not be called")
		})

		t.Run("Should: do nothing if PVC is not defined", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			state.PVC = nil

			err, res := deletePersistentVolumeClaim(ctx, state)

			assert.Nil(t, res, "should return nil result")
			assert.Nil(t, err, "should return nil err")
			assert.EqualValues(t, 0, k8sClient.(spy.ClientSpy).DeleteCallCount(), "delete should not be called")
		})

		t.Run("Should: do nothing if PVC is marked for deletion", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: v1.ObjectMeta{
					DeletionTimestamp: &v1.Time{
						Time: time.Now(),
					},
				},
			}
			state.PVC = pvc

			err, res := deletePersistentVolumeClaim(ctx, state)

			assert.Nil(t, res, "should return nil result")
			assert.Nil(t, err, "should return nil err")
			assert.EqualValues(t, 0, k8sClient.(spy.ClientSpy).DeleteCallCount(), "delete should not be called")
		})

	})
}
