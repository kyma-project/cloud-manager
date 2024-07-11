package gcpnfsvolume

import (
	"context"
	"testing"
	"time"

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
)

func TestCreatePersistentVolumeClaim(t *testing.T) {

	t.Run("createPersistentVolumeClaim", func(t *testing.T) {

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
			gcpNfsVolume = &cloudresourcesv1beta1.GcpNfsVolume{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-gcpnfsvol",
					Namespace: "test-ns",
				},
				Spec: cloudresourcesv1beta1.GcpNfsVolumeSpec{
					CapacityGb: 1000,
				},
			}

			pv = &corev1.PersistentVolume{
				ObjectMeta: v1.ObjectMeta{
					Name: "test-pv",
				},
				Spec: corev1.PersistentVolumeSpec{
					ClaimRef: nil,
					Capacity: corev1.ResourceList{
						"storage": *gcpNfsVolumeCapacityToResourceQuantity(gcpNfsVolume),
					},
					StorageClassName: "",
				},
				Status: corev1.PersistentVolumeStatus{
					Phase: corev1.VolumeAvailable,
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

		t.Run("Should: create PVC", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err, res := createPersistentVolumeClaim(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
			assert.EqualValues(t, 1, k8sClient.(spy.ClientSpy).CreateCallCount(), "should call create")
			createdPvc := &corev1.PersistentVolumeClaim{}
			_ = k8sClient.(spy.ClientSpy).Client().Get(
				ctx,
				types.NamespacedName{
					Name:      gcpNfsVolume.Name,
					Namespace: gcpNfsVolume.Namespace,
				},
				createdPvc)
			assert.Equal(t, gcpNfsVolume.Name, createdPvc.Name, "created PVC should have expected name")
			assert.Equal(t, gcpNfsVolume.Namespace, createdPvc.Namespace, "created PVC should have expected namespace")
			assert.Equal(t, []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany}, createdPvc.Spec.AccessModes, "created PVC should have expected access mode")
			assert.Equal(t, corev1.PersistentVolumeFilesystem, *createdPvc.Spec.VolumeMode, "created PVC should have expected VolumeMode")
			assert.Equal(t, pv.Name, createdPvc.Spec.VolumeName, "created PVC should have PV name as VolumeName")
			assert.Equal(t, pv.Spec.StorageClassName, *createdPvc.Spec.StorageClassName, "created PVC should have same StorageClassName as PV")
			createdPVCCapacity := pv.Spec.Capacity["storage"]
			createdPVCCapacityInt64, _ := createdPVCCapacity.AsInt64()
			gcpNfsVolCapacity := gcpNfsVolumeCapacityToResourceQuantity(gcpNfsVolume)
			gcpNfsVolCapacityInt64, _ := gcpNfsVolCapacity.AsInt64()
			assert.Equal(t, createdPVCCapacityInt64, gcpNfsVolCapacityInt64, "created PVC have same storage as parent GcpNfsVolume")
		})

		t.Run("Should: do nothing if GcpNfsVolume is marked for deletion", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			gcpNfsVolume.ObjectMeta = v1.ObjectMeta{
				DeletionTimestamp: &v1.Time{
					Time: time.Now(),
				},
			}

			err, res := sanitizeReleasedVolume(ctx, state)

			assert.Nil(t, res, "should return nil result")
			assert.Nil(t, err, "should return nil err")
			assert.EqualValues(t, 0, k8sClient.(spy.ClientSpy).CreateCallCount(), "create should not be called")
		})

		t.Run("Should: do nothing if PVC already exists (already defined in state)", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			state.PVC = &corev1.PersistentVolumeClaim{}

			err, res := sanitizeReleasedVolume(ctx, state)

			assert.Nil(t, res, "should return nil result")
			assert.Nil(t, err, "should return nil err")
			assert.EqualValues(t, 0, k8sClient.(spy.ClientSpy).CreateCallCount(), "create should not be called")
		})

		t.Run("Should: do nothing if PV for binding does not exist", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			state.PV = nil

			err, res := sanitizeReleasedVolume(ctx, state)

			assert.Nil(t, res, "should return nil result")
			assert.Nil(t, err, "should return nil err")
			assert.EqualValues(t, 0, k8sClient.(spy.ClientSpy).CreateCallCount(), "create should not be called")
		})

	})
}
