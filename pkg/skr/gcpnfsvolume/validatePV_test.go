package gcpnfsvolume

import (
	"context"
	"testing"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	spy "github.com/kyma-project/cloud-manager/pkg/testinfra/clientspy"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestValidatePV(t *testing.T) {

	t.Run("validatePV", func(t *testing.T) {

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
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gcpnfsvol",
					Namespace: "test-ns",
				},
				Spec: cloudresourcesv1beta1.GcpNfsVolumeSpec{
					CapacityGb: 1000,
				},
				Status: cloudresourcesv1beta1.GcpNfsVolumeStatus{
					Id: "26c33227-7abb-471d-b5fd-aeb125c50790",
				},
			}

			pv = &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name:   gcpNfsVolume.Status.Id,
					Labels: getVolumeLabels(gcpNfsVolume),
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(commonscheme.SkrScheme).
				WithObjects(pv).
				WithStatusSubresource(pv).
				Build()

			k8sClient = spy.NewClientSpy(fakeClient)

			state = createEmptyGcpNfsVolumeState(k8sClient, gcpNfsVolume)
			state.PV = &corev1.PersistentVolume{}
		}

		t.Run("Should: do nothing if GcpNfsVolume is marked for deletion", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			gcpNfsVolume.DeletionTimestamp = &metav1.Time{
				Time: time.Now(),
			}

			err, res := validatePV(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
		})

		t.Run("Should: do nothing if APIServer cant find requested PV", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			fakeClient := fake.NewClientBuilder().
				WithScheme(commonscheme.SkrScheme).
				Build()
			k8sClient.(spy.ClientSpy).SetClient(fakeClient)

			err, res := validatePV(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
		})

		t.Run("Should: do nothing if found PV has expected labels", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err, res := validatePV(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
		})

		t.Run("Should: set Status to Error and returns error when PV belongs to another GcpNfsVolume", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			pv.Labels[cloudresourcesv1beta1.LabelNfsVolName] = "another-owner-name"
			fakeClient := fake.NewClientBuilder().
				WithScheme(commonscheme.SkrScheme).
				WithObjects(pv).
				WithStatusSubresource(pv).
				Build()
			k8sClient.(spy.ClientSpy).SetClient(fakeClient)

			err, _ := validatePV(ctx, state)

			assert.NotNilf(t, err, "error should be returned")
			errorConditions := meta.FindStatusCondition(gcpNfsVolume.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
			assert.NotEmpty(t, errorConditions, "error condition should be added")
		})
	})
}
