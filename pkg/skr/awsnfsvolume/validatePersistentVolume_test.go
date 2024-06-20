package awsnfsvolume

import (
	"context"
	"k8s.io/apimachinery/pkg/api/meta"
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

func TestValidatePersistentVolume(t *testing.T) {

	t.Run("validatePersistentVolume", func(t *testing.T) {

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
			awsNfsVolume = &cloudresourcesv1beta1.AwsNfsVolume{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-awsnfsvol",
					Namespace: "test-ns",
				},
				//Spec: cloudresourcesv1beta1.AwsNfsVolumeSpec{
				//
				//},
			}

			pv = &corev1.PersistentVolume{
				ObjectMeta: v1.ObjectMeta{
					Name:   "test-awsnfsvol",
					Labels: getVolumeLabels(awsNfsVolume),
				},
			}

			scheme := runtime.NewScheme()
			utilruntime.Must(clientgoscheme.AddToScheme(scheme))
			utilruntime.Must(cloudresourcesv1beta1.AddToScheme(scheme))
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(pv).
				WithStatusSubresource(pv).
				Build()

			k8sClient = spy.NewClientSpy(fakeClient)

			state = createEmptyAwsNfsVolumeState(k8sClient, awsNfsVolume)
			state.Volume = &corev1.PersistentVolume{}
		}

		t.Run("Should: do nothing if AwsNfsVolume is marked for deletion", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			awsNfsVolume.ObjectMeta.DeletionTimestamp = &v1.Time{
				Time: time.Now(),
			}

			err, res := validatePersistentVolume(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
		})

		t.Run("Should: do nothing if APIServer cant find requested PersistentVolume", func(t *testing.T) {
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

			err, res := validatePersistentVolume(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
		})

		t.Run("Should: do nothing if found PersistentVolume has expected labels", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err, res := validatePersistentVolume(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
		})

		t.Run("Should: set Status to Error and returns error when PersistentVolume belongs to another AwsNfsVolume", func(t *testing.T) {
			setupTest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			pv.Labels[cloudresourcesv1beta1.LabelNfsVolName] = "another-owner-name"
			scheme := runtime.NewScheme()
			utilruntime.Must(clientgoscheme.AddToScheme(scheme))
			utilruntime.Must(cloudresourcesv1beta1.AddToScheme(scheme))
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(pv).
				WithStatusSubresource(pv).
				Build()
			k8sClient.(spy.ClientSpy).SetClient(fakeClient)

			err, _ := validatePersistentVolume(ctx, state)

			assert.NotNilf(t, err, "error should be returned")
			errorConditions := meta.FindStatusCondition(awsNfsVolume.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
			assert.NotEmpty(t, errorConditions, "error condition should be added")
		})
	})
}
