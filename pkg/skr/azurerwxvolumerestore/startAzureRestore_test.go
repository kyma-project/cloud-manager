package azurerwxvolumerestore

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azurerwxvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	spy "github.com/kyma-project/cloud-manager/pkg/testinfra/clientspy"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
	"time"
)

func TestStartAzureRestore(t *testing.T) {
	t.Run("startAzureRestore", func(t *testing.T) {

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
		setupTest := func(withObj bool, backupRecoveryPointId string, backupStorageAccountPath string) {
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
			azureRwxVolumeBackup := &cloudresourcesv1beta1.AzureRwxVolumeBackup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-azure-restore-backup",
					Namespace: "test-ns",
				},
				Status: cloudresourcesv1beta1.AzureRwxVolumeBackupStatus{},
			}
			azureRwxVolumeBackup.Status.RecoveryPointId = backupRecoveryPointId
			azureRwxVolumeBackup.Status.StorageAccountPath = backupStorageAccountPath
			startTime, _ := time.Parse(time.RFC3339, "2025-03-01T00:43:35.6367215Z")
			k8sStartTime := metav1.Time{
				Time: startTime,
			}
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
				Status: cloudresourcesv1beta1.AzureRwxVolumeRestoreStatus{
					RestoredDir: "test-restore-dir",
					StartTime:   &k8sStartTime,
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
			state.azureRwxVolumeBackup = azureRwxVolumeBackup
			state.storageClient, _ = azurerwxvolumebackupclient.NewMockClient()(nil, "", "", "", "")
			state.SetScope(scope)
		}

		t.Run("Should: start azure restore ", func(t *testing.T) {
			backupRecoveryPointId := "/subscriptions/3f1d2fbd-117a-4742-8bde-6edbcdee6a04/resourceGroups/rg-test/providers/Microsoft.RecoveryServices/vaults/v-test/backupFabrics/Azure/protectionContainers/StorageContainer;Storage;test;testsa/protectedItems/AzureFileShare;2DAC3CBDBBD863B2292F25490DC0794F35AAA4C27890D5DCA82B0A33E9596217/recoveryPoints/5639661428710522320"
			backupStorageAccountPath := "test-storage-account-path"
			setupTest(true, backupRecoveryPointId, backupStorageAccountPath)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			err := k8sClient.Get(ctx, types.NamespacedName{Name: "test-azure-restore", Namespace: "test-ns-2"}, azureRwxVolumeRestore)
			assert.Nil(t, err, "should get azureRwxVolumeRestore")
			assert.Empty(t, azureRwxVolumeRestore.Status.OpIdentifier, "should not have opIdentifier set")
			restore := state.ObjAsAzureRwxVolumeRestore()
			restore.Status.RestoredDir = "test-restore-dir"

			err, res := startAzureRestore(ctx, state)
			// wants to retry
			assert.Equal(t, ctx, res, "should return same context")
			assert.Equal(t, composed.StopWithRequeueDelay(util.Timing.T1000ms()), err, "should stop with requeue")
			// reload the object and verify that the opIdentifider is not set yet
			err = k8sClient.Get(ctx, types.NamespacedName{Name: "test-azure-restore", Namespace: "test-ns-2"}, azureRwxVolumeRestore)
			assert.Nil(t, err, "should get azureRwxVolumeRestore")
			assert.Equal(t, "v-test", azureRwxVolumeRestore.Status.OpIdentifier, "should have opIdentifier set")
			assert.Equal(t, cloudresourcesv1beta1.JobStateInProgress, azureRwxVolumeRestore.Status.State, "should be in progress")
		})

		t.Run("Should: fail if recoveryPointId is invalid in backup status ", func(t *testing.T) {
			setupTest(true, "invalid-recoveryPointId", "test-storage-account-path")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			err := k8sClient.Get(ctx, types.NamespacedName{Name: "test-azure-restore", Namespace: "test-ns-2"}, azureRwxVolumeRestore)
			assert.Nil(t, err, "should get azureRwxVolumeRestore")
			assert.Empty(t, azureRwxVolumeRestore.Status.OpIdentifier, "should not have opIdentifier set")
			restore := state.ObjAsAzureRwxVolumeRestore()
			restore.Status.RestoredDir = "test-restore-dir"

			err, res := startAzureRestore(ctx, state)
			// wants to retry
			assert.Equal(t, ctx, res, "should return same context")
			assert.Equal(t, composed.StopAndForget, err, "should stop and forget")
			// reload the object and verify that the opIdentifider is not set and state is failed
			err = k8sClient.Get(ctx, types.NamespacedName{Name: "test-azure-restore", Namespace: "test-ns-2"}, azureRwxVolumeRestore)
			assert.Nil(t, err, "should get azureRwxVolumeRestore")
			assert.Empty(t, azureRwxVolumeRestore.Status.OpIdentifier, "should still be empty")
			assert.Equal(t, cloudresourcesv1beta1.JobStateFailed, azureRwxVolumeRestore.Status.State, "should be in failed state")
		})
		t.Run("Should: fail if storageAccountPath is missing in backup status ", func(t *testing.T) {
			backupRecoverPointId := "/subscriptions/3f1d2fbd-117a-4742-8bde-6edbcdee6a04/resourceGroups/rg-test/providers/Microsoft.RecoveryServices/vaults/v-test/backupFabrics/Azure/protectionContainers/StorageContainer;Storage;test;testsa/protectedItems/AzureFileShare;2DAC3CBDBBD863B2292F25490DC0794F35AAA4C27890D5DCA82B0A33E9596217/recoveryPoints/5639661428710522320"
			setupTest(true, backupRecoverPointId, "")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			err := k8sClient.Get(ctx, types.NamespacedName{Name: "test-azure-restore", Namespace: "test-ns-2"}, azureRwxVolumeRestore)
			assert.Nil(t, err, "should get azureRwxVolumeRestore")
			assert.Empty(t, azureRwxVolumeRestore.Status.OpIdentifier, "should not have opIdentifier set")
			restore := state.ObjAsAzureRwxVolumeRestore()
			restore.Status.RestoredDir = "test-restore-dir"

			err, res := startAzureRestore(ctx, state)
			// wants to retry
			assert.Equal(t, ctx, res, "should return same context")
			assert.Equal(t, composed.StopAndForget, err, "should stop and forget")
			// reload the object and verify that the opIdentifider is not set and state is failed
			err = k8sClient.Get(ctx, types.NamespacedName{Name: "test-azure-restore", Namespace: "test-ns-2"}, azureRwxVolumeRestore)
			assert.Nil(t, err, "should get azureRwxVolumeRestore")
			assert.Empty(t, azureRwxVolumeRestore.Status.OpIdentifier, "should still be empty")
			assert.Equal(t, cloudresourcesv1beta1.JobStateFailed, azureRwxVolumeRestore.Status.State, "should be in failed state")
		})

	})

}
