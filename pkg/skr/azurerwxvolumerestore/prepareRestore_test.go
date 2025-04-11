package azurerwxvolumerestore

import (
	"context"
	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	spy "github.com/kyma-project/cloud-manager/pkg/testinfra/clientspy"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestSetStartTime(t *testing.T) {
	t.Run("setStartTime", func(t *testing.T) {

		var azureRwxVolumeRestore *cloudresourcesv1beta1.AzureRwxVolumeRestore
		var state *State
		var k8sClient client.WithWatch

		kcpScheme := runtime.NewScheme()
		utilruntime.Must(clientgoscheme.AddToScheme(kcpScheme))
		utilruntime.Must(cloudcontrolv1beta1.AddToScheme(kcpScheme))

		scope := &cloudcontrolv1beta1.Scope{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-scope",
				Namespace: "test-ns",
			},
			Spec: cloudcontrolv1beta1.ScopeSpec{
				Scope: cloudcontrolv1beta1.ScopeInfo{
					Azure: &cloudcontrolv1beta1.AzureScope{
						SubscriptionId: "test-subscription-id",
					},
				},
			},
		}

		kcpClient := fake.NewClientBuilder().
			WithScheme(kcpScheme).
			WithObjects(scope).
			Build()
		kcpCluster := composed.NewStateCluster(kcpClient, kcpClient, nil, kcpScheme)
		createEmptyState := func(k8sClient client.WithWatch, azureRwxVolumeRestore *cloudresourcesv1beta1.AzureRwxVolumeRestore) *State {
			cluster := composed.NewStateCluster(k8sClient, k8sClient, nil, k8sClient.Scheme())
			return &State{
				State: commonscope.NewStateFactory(kcpCluster, kymaRef).NewState(composed.NewStateFactory(cluster).NewState(types.NamespacedName{}, azureRwxVolumeRestore)),
			}
		}

		setupTest := func(withObj bool) {
			azureRwxVolumeRestore = &cloudresourcesv1beta1.AzureRwxVolumeRestore{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-azure-restore",
					Namespace: "test-ns-2",
				},
				Spec: cloudresourcesv1beta1.AzureRwxVolumeRestoreSpec{

					Destination: cloudresourcesv1beta1.PvcSource{
						Pvc: cloudresourcesv1beta1.PvcRef{
							Name:      "test-azure-restore-pvc",
							Namespace: "test-ns",
						},
					},
					Source: cloudresourcesv1beta1.AzureRwxVolumeRestoreSource{
						Backup: cloudresourcesv1beta1.AzureRwxVolumeBackupRef{
							Name:      "test-azure-restore-backup",
							Namespace: "test-ns",
						},
					},
				},
			}

			scheme := runtime.NewScheme()
			utilruntime.Must(clientgoscheme.AddToScheme(scheme))
			utilruntime.Must(cloudresourcesv1beta1.AddToScheme(scheme))
			var fakeClient client.WithWatch
			if withObj {
				fakeClient = fake.NewClientBuilder().WithScheme(scheme).
					WithObjects(azureRwxVolumeRestore).
					WithStatusSubresource(azureRwxVolumeRestore).
					Build()
			} else {
				fakeClient = fake.NewClientBuilder().WithScheme(scheme).Build()
			}
			k8sClient = spy.NewClientSpy(fakeClient)
			state = createEmptyState(k8sClient, azureRwxVolumeRestore)
		}

		t.Run("Should: set start time", func(t *testing.T) {
			setupTest(true)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			err := k8sClient.Get(ctx, types.NamespacedName{Name: "test-azure-restore", Namespace: "test-ns-2"}, azureRwxVolumeRestore)
			assert.Nil(t, err, "should get azureRwxVolumeRestore")
			assert.Nil(t, azureRwxVolumeRestore.Status.StartTime, "should not have start time set")

			err, res := prepareRestore(ctx, state)

			assert.Equal(t, ctx, res, "should return same context")
			assert.Nil(t, err, "should return nil err")
			// reload the object and verify that the start time is set
			err = k8sClient.Get(ctx, types.NamespacedName{Name: "test-azure-restore", Namespace: "test-ns-2"}, azureRwxVolumeRestore)
			assert.Nil(t, err, "should get azureRwxVolumeRestore")
			assert.NotNil(t, azureRwxVolumeRestore.Status.StartTime, "should have start time set")
			assert.NotEmpty(t, azureRwxVolumeRestore.Status.RestoredDir, "should have restored dir set")
		})

		t.Run("Should: retry if setting start time fails", func(t *testing.T) {
			setupTest(false)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err, res := prepareRestore(ctx, state)

			assert.Equal(t, ctx, res, "should return same context")
			assert.Equal(t, err, composed.StopWithRequeue, "should stop and requeue")
		})

		t.Run("Should: skip if restoredDir is already set", func(t *testing.T) {
			setupTest(true)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			state.ObjAsAzureRwxVolumeRestore().Status.RestoredDir = uuid.NewString()

			err, res := prepareRestore(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
			// reload the object and verify that the start time hasn't changed
			err = k8sClient.Get(ctx, types.NamespacedName{Name: "test-azure-restore", Namespace: "test-ns-2"}, azureRwxVolumeRestore)
			assert.Nil(t, err, "should get azureRwxVolumeRestore")
			assert.Nil(t, azureRwxVolumeRestore.Status.StartTime, "should not have set start time")
		})

	})

}
