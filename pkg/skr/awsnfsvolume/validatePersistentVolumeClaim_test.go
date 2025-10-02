package awsnfsvolume

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/common/bootstrap"
	"k8s.io/apimachinery/pkg/api/meta"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	spy "github.com/kyma-project/cloud-manager/pkg/testinfra/clientspy"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestValidatePersistentVolumeClaim(t *testing.T) {

	t.Run("validatePersistentVolumeClaim", func(t *testing.T) {

		var awsNfsVolume *cloudresourcesv1beta1.AwsNfsVolume
		var pvc *corev1.PersistentVolumeClaim
		var state *State
		var k8sClient client.WithWatch

		createEmptyAwsNfsVolumeState := func(k8sClient client.WithWatch, awsNfsVolume *cloudresourcesv1beta1.AwsNfsVolume) *State {
			cluster := composed.NewStateCluster(k8sClient, k8sClient, nil, k8sClient.Scheme())
			return &State{
				State: composed.NewStateFactory(cluster).NewState(types.NamespacedName{}, awsNfsVolume),
			}
		}

		setupTest := func() {
			awsNfsVolume = &cloudresourcesv1beta1.AwsNfsVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-awsnfsvol",
					Namespace: "test-ns",
				},
			}

			pvc = &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-awsnfsvol",
					Namespace: "test-ns",
					Labels:    getVolumeClaimLabels(awsNfsVolume),
				},
			}

			scheme := bootstrap.SkrScheme
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(pvc).
				WithStatusSubresource(pvc).
				Build()

			k8sClient = spy.NewClientSpy(fakeClient)

			state = createEmptyAwsNfsVolumeState(k8sClient, awsNfsVolume)
			state.PVC = &corev1.PersistentVolumeClaim{}
		}

		t.Run("Should: do nothing if AwsNfsVolume is marked for deletion", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			awsNfsVolume.DeletionTimestamp = &metav1.Time{
				Time: time.Now(),
			}

			err, res := validatePersistentVolumeClaim(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
		})

		t.Run("Should: do nothing if APIServer cant find requested PersistentVolumeClaim", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			scheme := bootstrap.SkrScheme
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				Build()
			k8sClient.(spy.ClientSpy).SetClient(fakeClient)

			err, res := validatePersistentVolumeClaim(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
		})

		t.Run("Should: do nothing if found PersistentVolumeClaim has expected labels", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err, res := validatePersistentVolumeClaim(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
		})

		t.Run("Should: set Status to Error and returns error when PersistentVolumeClaim belongs to another AwsNfsVolume", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			pvc.Labels[cloudresourcesv1beta1.LabelNfsVolName] = "another-owner-name"
			scheme := bootstrap.SkrScheme
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(pvc).
				WithStatusSubresource(pvc).
				Build()
			k8sClient.(spy.ClientSpy).SetClient(fakeClient)

			err, _ := validatePersistentVolumeClaim(ctx, state)

			assert.NotNilf(t, err, "error should be returned")
			errorConditions := meta.FindStatusCondition(awsNfsVolume.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
			assert.NotEmpty(t, errorConditions, "error condition should be added")
		})
	})
}
