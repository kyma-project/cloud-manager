package gcpnfsvolume

import (
	"context"
	"testing"

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

func TestLoadPersistentVolumeClaim(t *testing.T) {

	t.Run("loadPersistentVolumeClaim", func(t *testing.T) {

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
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gcpnfsvol",
					Namespace: "test-ns",
				},
				Spec: cloudresourcesv1beta1.GcpNfsVolumeSpec{
					CapacityGb: 1000,
				},
			}

			pvc = &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gcpnfsvol",
					Namespace: "test-ns",
					Labels:    getVolumeClaimLabels(gcpNfsVolume),
				},
			}

			scheme := runtime.NewScheme()
			utilruntime.Must(clientgoscheme.AddToScheme(scheme))
			utilruntime.Must(cloudresourcesv1beta1.AddToScheme(scheme))
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(pvc).
				WithStatusSubresource(pvc).
				Build()

			k8sClient = spy.NewClientSpy(fakeClient)

			state = createEmptyGcpNfsVolumeState(k8sClient, gcpNfsVolume)
			state.PV = &corev1.PersistentVolume{}
		}

		t.Run("Should: load existing PVC", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err, res := loadPersistentVolumeClaim(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
			assert.NotNilf(t, state.PVC, "pvc should be loaded into state")
			assert.Equal(t, gcpNfsVolume.Name, state.PVC.Name, "loaded pvc should have expected name")
			assert.Equal(t, gcpNfsVolume.Namespace, state.PVC.Namespace, "loaded pvc should have expected namespace")
		})

		t.Run("Should: do nothing if APIServer cant find requested PVC", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			scheme := runtime.NewScheme()
			utilruntime.Must(clientgoscheme.AddToScheme(scheme))
			utilruntime.Must(cloudresourcesv1beta1.AddToScheme(scheme))
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				Build()
			k8sClient.(spy.ClientSpy).SetClient(fakeClient)

			err, res := loadPersistentVolumeClaim(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
			assert.Nil(t, state.PVC, "pvc should remain nil in state as it is not found in APIServer")
		})

	})
}
