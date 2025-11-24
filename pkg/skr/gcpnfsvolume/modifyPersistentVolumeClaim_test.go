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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestModifyPersistentVolumeClaim(t *testing.T) {

	t.Run("modifyPersistentVolumeClaim", func(t *testing.T) {

		var gcpNfsVolume *cloudresourcesv1beta1.GcpNfsVolume
		var actualPVC *corev1.PersistentVolumeClaim
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
					PersistentVolumeClaim: &cloudresourcesv1beta1.GcpNfsVolumePvcSpec{
						Labels: map[string]string{
							"foo": "bar",
						},
						Annotations: map[string]string{
							"baz": "qux",
						},
					},
				},
				Status: cloudresourcesv1beta1.GcpNfsVolumeStatus{
					Conditions: []metav1.Condition{
						{
							Type:    cloudresourcesv1beta1.ConditionTypeReady,
							Status:  metav1.ConditionTrue,
							Reason:  cloudresourcesv1beta1.ConditionReasonReady,
							Message: "Volume is ready",
						},
					},
				},
			}

			actualPVC = &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-gcpnfsvol",
					Namespace:   "test-ns",
					Labels:      getVolumeClaimLabels(gcpNfsVolume),
					Annotations: getVolumeClaimAnnotations(gcpNfsVolume),
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							"storage": *gcpNfsVolumeCapacityToResourceQuantity(gcpNfsVolume),
						},
					},
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(commonscheme.SkrScheme).
				WithObjects(actualPVC).
				WithStatusSubresource(actualPVC).
				Build()

			k8sClient = spy.NewClientSpy(fakeClient)

			state = createEmptyGcpNfsVolumeState(k8sClient, gcpNfsVolume)
			state.PVC = actualPVC
		}

		t.Run("Should: modify PVC when labels change", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			gcpNfsVolume.Spec.PersistentVolumeClaim = &cloudresourcesv1beta1.GcpNfsVolumePvcSpec{
				Labels: map[string]string{
					"foo": "bar-modified",
					"oof": "rab",
				},
			}

			err, res := modifyPersistentVolumeClaim(ctx, state)

			assert.NotNil(t, err, "should return not-nil err") // not an actual error, but StopWithRequeueDelay
			assert.Nil(t, res, "should return nil res")
			assert.EqualValues(t, 1, k8sClient.(spy.ClientSpy).UpdateCallCount(), "update should be called")
		})

		t.Run("Should: modify PVC when annotation change", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			gcpNfsVolume.Spec.PersistentVolumeClaim = &cloudresourcesv1beta1.GcpNfsVolumePvcSpec{
				Annotations: map[string]string{
					"baz": "qux-modified",
					"zab": "xuq",
				},
			}

			err, res := modifyPersistentVolumeClaim(ctx, state)

			assert.NotNil(t, err, "should return not-nil err") // not an actual error, but StopWithRequeueDelay
			assert.Nil(t, res, "should return nil res")
			assert.EqualValues(t, 1, k8sClient.(spy.ClientSpy).UpdateCallCount(), "update should be called")
		})

		t.Run("Should: modify only PVC label when capacity changes", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			originalCapacity := actualPVC.Spec.Resources.Requests["storage"].DeepCopy()
			gcpNfsVolume.Spec.CapacityGb = 2000

			err, res := modifyPersistentVolumeClaim(ctx, state)

			postModifyCapacity := actualPVC.Spec.Resources.Requests["storage"].DeepCopy()
			assert.NotNil(t, err, "should return not-nil err") // not an actual error, but StopWithRequeueDelay
			assert.Nil(t, res, "should return nil res")
			assert.EqualValues(t, 1, k8sClient.(spy.ClientSpy).UpdateCallCount(), "update should be called")
			assert.True(t, originalCapacity.Equal(postModifyCapacity), "capacity should not change")
			assert.Equal(t, "2000Gi", actualPVC.Labels[cloudresourcesv1beta1.LabelStorageCapacity], "label value should be adjusted to match desired value")
		})

		t.Run("Should: do nothing if GcpNfsVolume is marked for deletion", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			gcpNfsVolume.ObjectMeta = metav1.ObjectMeta{
				DeletionTimestamp: &metav1.Time{
					Time: time.Now(),
				},
			}

			err, res := modifyPersistentVolumeClaim(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
			assert.EqualValues(t, 0, k8sClient.(spy.ClientSpy).UpdateCallCount(), "update should not be called")
		})

		t.Run("Should: do nothing if NFS is not ready", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			state.PVC = nil
			gcpNfsVolume.Status.Conditions = []metav1.Condition{}

			err, res := modifyPersistentVolumeClaim(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
			assert.EqualValues(t, 0, k8sClient.(spy.ClientSpy).UpdateCallCount(), "update should not be called")
		})

		t.Run("Should: do nothing if PVC is not loaded", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			state.PVC = nil

			err, res := modifyPersistentVolumeClaim(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
			assert.EqualValues(t, 0, k8sClient.(spy.ClientSpy).UpdateCallCount(), "update should not be called")
		})

		t.Run("Should: do nothing if PVC actual and desired PVC state is same", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err, res := modifyPersistentVolumeClaim(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
			assert.EqualValues(t, 0, k8sClient.(spy.ClientSpy).UpdateCallCount(), "update should not be called")
		})

	})
}
