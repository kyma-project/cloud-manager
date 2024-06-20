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
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestValidatePVC(t *testing.T) {

	t.Run("validatePVC", func(t *testing.T) {

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
				},
				Spec: cloudresourcesv1beta1.GcpNfsVolumeSpec{
					CapacityGb: 1000,
				},
			}

			pvc = &corev1.PersistentVolumeClaim{
				ObjectMeta: v1.ObjectMeta{
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

		t.Run("Should: do nothing if GcpNfsVolume is marked for deletion", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			gcpNfsVolume.ObjectMeta.DeletionTimestamp = &v1.Time{
				Time: time.Now(),
			}

			err, res := validatePVC(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
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

			err, res := validatePVC(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
		})

		t.Run("Should: do nothing if found PVC has expected labels", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err, res := validatePVC(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
		})

		t.Run("Should: set Status to Error and returns error when PVC belongs to another GcpNfsVolume", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			pvc.Labels[cloudresourcesv1beta1.LabelNfsVolName] = "another-owner-name"
			scheme := runtime.NewScheme()
			utilruntime.Must(clientgoscheme.AddToScheme(scheme))
			utilruntime.Must(cloudresourcesv1beta1.AddToScheme(scheme))
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(pvc).
				WithStatusSubresource(pvc).
				Build()
			k8sClient.(spy.ClientSpy).SetClient(fakeClient)

			err, _ := validatePVC(ctx, state)

			assert.NotNilf(t, err, "error should be returned")
			errorConditions := meta.FindStatusCondition(gcpNfsVolume.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
			assert.NotEmpty(t, errorConditions, "error condition should be added")
		})
	})
}
