package azurerwxvolumerestore

import (
	"context"
	"testing"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azurerwxvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	spy "github.com/kyma-project/cloud-manager/pkg/testinfra/clientspy"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestFindAzureRestoreJob(t *testing.T) {
	t.Run("findAzureRestoreJob", func(t *testing.T) {

		var azureRwxVolumeRestore *cloudresourcesv1beta1.AzureRwxVolumeRestore
		var state *State
		var k8sClient client.WithWatch

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
			WithScheme(commonscheme.KcpScheme).
			WithObjects(scope).
			Build()
		kcpCluster := composed.NewStateCluster(kcpClient, kcpClient, nil, commonscheme.KcpScheme)

		createEmptyState := func(k8sClient client.WithWatch, azureRwxVolumeRestore *cloudresourcesv1beta1.AzureRwxVolumeRestore) *State {
			cluster := composed.NewStateCluster(k8sClient, k8sClient, nil, k8sClient.Scheme())
			return &State{
				State: commonscope.NewStateFactory(kcpCluster, kymaRef).NewState(composed.NewStateFactory(cluster).NewState(types.NamespacedName{}, azureRwxVolumeRestore)),
			}
		}

		setupTest := func(withObj bool, backupWithRecoveryPointId bool) {
			azureRwxVolumeBackup := &cloudresourcesv1beta1.AzureRwxVolumeBackup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-azure-restore-backup",
					Namespace: "test-ns",
				},
				Status: cloudresourcesv1beta1.AzureRwxVolumeBackupStatus{},
			}
			if backupWithRecoveryPointId {
				azureRwxVolumeBackup.Status.RecoveryPointId = "/subscriptions/3f1d2fbd-117a-4742-8bde-6edbcdee6a04/resourceGroups/rg-test/providers/Microsoft.RecoveryServices/vaults/v-test/backupFabrics/Azure/protectionContainers/StorageContainer;Storage;test;testsa/protectedItems/AzureFileShare;2DAC3CBDBBD863B2292F25490DC0794F35AAA4C27890D5DCA82B0A33E9596217/recoveryPoints/5639661428710522320"
			}
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

			var fakeClient client.WithWatch
			if withObj {
				fakeClient = fake.NewClientBuilder().WithScheme(commonscheme.SkrScheme).
					WithObjects(azureRwxVolumeRestore).
					WithStatusSubresource(azureRwxVolumeRestore).
					Build()
			} else {
				fakeClient = fake.NewClientBuilder().WithScheme(commonscheme.SkrScheme).Build()
			}
			k8sClient = spy.NewClientSpy(fakeClient)
			state = createEmptyState(k8sClient, azureRwxVolumeRestore)
			state.azureRwxVolumeBackup = azureRwxVolumeBackup
			state.storageClient, _ = azurerwxvolumebackupclient.NewMockClient()(nil, "", "", "", "")
		}

		t.Run("Should: find azure restore job", func(t *testing.T) {
			setupTest(true, true)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			err := k8sClient.Get(ctx, types.NamespacedName{Name: "test-azure-restore", Namespace: "test-ns-2"}, azureRwxVolumeRestore)
			assert.Nil(t, err, "should get azureRwxVolumeRestore")
			assert.Empty(t, azureRwxVolumeRestore.Status.OpIdentifier, "should not have opIdentifier set")
			restore := state.ObjAsAzureRwxVolumeRestore()
			restore.Status.RestoredDir = "test-restore-dir"
			jobId := "test-job-id"
			// This will only add a jobId to the array with inprogress status
			request := azurerwxvolumebackupclient.RestoreRequest{
				VaultName:                jobId,
				ResourceGroupName:        "",
				FabricName:               "",
				ContainerName:            "",
				ProtectedItemName:        "",
				RecoveryPointId:          "",
				SourceStorageAccountPath: "",
				TargetStorageAccountPath: "",
				TargetFileShareName:      "",
				TargetFolderName:         restore.Status.RestoredDir,
			}
			_, _ = state.storageClient.TriggerRestore(ctx, request)
			err, res := findAzureRestoreJob(ctx, state)
			// wants to retry
			assert.Equal(t, ctx, res, "should return same context")
			assert.Equal(t, composed.StopWithRequeueDelay(util.Timing.T10000ms()), err, "should stop with requeue")
			// reload the object and verify that the opIdentifider is not set yet
			err = k8sClient.Get(ctx, types.NamespacedName{Name: "test-azure-restore", Namespace: "test-ns-2"}, azureRwxVolumeRestore)
			assert.Nil(t, err, "should get azureRwxVolumeRestore")
			assert.Empty(t, azureRwxVolumeRestore.Status.OpIdentifier, "should still be empty")

			// 2nd time it should find the jobId
			err, res = findAzureRestoreJob(ctx, state)
			// wants to retry
			assert.Equal(t, ctx, res, "should return same context")
			assert.Nil(t, err, "should continue")
			// reload the object and verify that the opIdentifider is  set
			err = k8sClient.Get(ctx, types.NamespacedName{Name: "test-azure-restore", Namespace: "test-ns-2"}, azureRwxVolumeRestore)
			assert.Nil(t, err, "should get azureRwxVolumeRestore")
			assert.Equal(t, jobId, azureRwxVolumeRestore.Status.OpIdentifier, "should have opIdentifier set")
		})

		t.Run("Should: fail if recoveryPointId is missing", func(t *testing.T) {
			setupTest(true, false)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			err := k8sClient.Get(ctx, types.NamespacedName{Name: "test-azure-restore", Namespace: "test-ns-2"}, azureRwxVolumeRestore)
			assert.Nil(t, err, "should get azureRwxVolumeRestore")
			assert.Empty(t, azureRwxVolumeRestore.Status.OpIdentifier, "should not have opIdentifier set")
			restore := state.ObjAsAzureRwxVolumeRestore()
			restore.Status.RestoredDir = "test-restore-dir"

			jobId := "test-job-id"
			// This will only add a jobId to the array with inprogress status
			request := azurerwxvolumebackupclient.RestoreRequest{
				VaultName:                jobId,
				ResourceGroupName:        "",
				FabricName:               "",
				ContainerName:            "",
				ProtectedItemName:        "",
				RecoveryPointId:          "",
				SourceStorageAccountPath: "",
				TargetStorageAccountPath: "",
				TargetFileShareName:      "",
				TargetFolderName:         restore.Status.RestoredDir,
			}
			_, _ = state.storageClient.TriggerRestore(ctx, request)
			err, res := findAzureRestoreJob(ctx, state)
			// wants to retry
			assert.Equal(t, ctx, res, "should return same context")
			assert.Equal(t, composed.StopAndForget, err, "should stop and forget")
			// reload the object and verify that restore has failed
			err = k8sClient.Get(ctx, types.NamespacedName{Name: "test-azure-restore", Namespace: "test-ns-2"}, azureRwxVolumeRestore)
			assert.Nil(t, err, "should get azureRwxVolumeRestore")
			assert.Equal(t, cloudresourcesv1beta1.JobStateFailed, azureRwxVolumeRestore.Status.State, "should have a failed state")
		})

		t.Run("Should: continue if job is not found", func(t *testing.T) {
			setupTest(true, true)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			err := k8sClient.Get(ctx, types.NamespacedName{Name: "test-azure-restore", Namespace: "test-ns-2"}, azureRwxVolumeRestore)
			assert.Nil(t, err, "should get azureRwxVolumeRestore")
			assert.Empty(t, azureRwxVolumeRestore.Status.OpIdentifier, "should not have opIdentifier set")
			restore := state.ObjAsAzureRwxVolumeRestore()
			restore.Status.RestoredDir = "test-restore-dir"

			err, res := findAzureRestoreJob(ctx, state)
			// wants to retry
			assert.Equal(t, ctx, res, "should return same context")
			assert.Nil(t, err, "should proceed")
			// reload the object and verify that jobId is not set
			err = k8sClient.Get(ctx, types.NamespacedName{Name: "test-azure-restore", Namespace: "test-ns-2"}, azureRwxVolumeRestore)
			assert.Nil(t, err, "should get azureRwxVolumeRestore")
			assert.Empty(t, azureRwxVolumeRestore.Status.OpIdentifier, "should not have opIdentifier set")
		})

	})

}
