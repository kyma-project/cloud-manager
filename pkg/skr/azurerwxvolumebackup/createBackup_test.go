package azurerwxvolumebackup

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureClient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	spy "github.com/kyma-project/cloud-manager/pkg/testinfra/clientspy"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func setupDefaultBackup() *cloudresourcesv1beta1.AzureRwxVolumeBackup {

	backup := &cloudresourcesv1beta1.AzureRwxVolumeBackup{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: "backupName",
		},
		Spec: cloudresourcesv1beta1.AzureRwxVolumeBackupSpec{
			Location: "uswest",
		},
		Status: cloudresourcesv1beta1.AzureRwxVolumeBackupStatus{},
	}

	return backup

}

func setupDefaultCluster() composed.StateCluster {

	scheme := runtime.NewScheme()
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	k8sClient := spy.NewClientSpy(fakeClient)
	cluster := composed.NewStateCluster(k8sClient, k8sClient, nil, k8sClient.Scheme())

	return cluster

}

func setupDefaultState(ctx context.Context, backup *cloudresourcesv1beta1.AzureRwxVolumeBackup) *State {

	cluster := setupDefaultCluster()

	state := &State{
		State: composed.NewStateFactory(cluster).NewState(types.NamespacedName{}, backup),
	}

	scope := &cloudcontrolv1beta1.Scope{
		ObjectMeta: v1.ObjectMeta{
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

	state.client, _ = azureClient.NewMockClient()(ctx, "", "", "", "")
	state.scope = scope
	state.fileShareName = "matchingFileShareName"

	return state

}

func TestCreateBackup(t *testing.T) {

	t.Run("createBackup - fileshare already protected", func(t *testing.T) {

		// Arrange
		ctx := context.Background()
		backup := setupDefaultBackup()
		state := setupDefaultState(ctx, backup)

		t.Run("unhappy paths", func(t *testing.T) {

			t.Run("Id is empty", func(t *testing.T) {
				err, _ := createBackup(ctx, state)

				assert.Equal(t, composed.StopWithRequeue, err)

			})

			t.Run("ListProtectedItems - fails", func(t *testing.T) {
				backup.Status.Id = "asdf"

				state.vaultName = "fail ListProtectedItems"
				err, _ := createBackup(ctx, state)

				assert.Equal(t, composed.StopWithRequeue, err)
				assert.Equal(t, cloudresourcesv1beta1.AzureRwxBackupError, backup.Status.State)

			})

			t.Run("ListProtectedItems - more than one matching name", func(t *testing.T) {

				backup.Status.Id = "asdf"
				state.vaultName = "more than 1"
				err, _ := createBackup(ctx, state)

				assert.Equal(t, composed.StopAndForget, err)
				assert.Equal(t, cloudresourcesv1beta1.AzureRwxBackupFailed, backup.Status.State)

			})

			t.Run("ListProtectedItems - one matching name; fail backup", func(t *testing.T) {

				backup.Status.Id = "asdf"
				state.vaultName = "exactly 1 - fail"
				err, _ := createBackup(ctx, state)

				assert.Equal(t, composed.StopWithRequeue, err)
				assert.Equal(t, cloudresourcesv1beta1.AzureRwxBackupError, backup.Status.State)

			})

		})

		t.Run("happy path", func(t *testing.T) {

			t.Run("ListProtectedItems - one matching name; succeed backup", func(t *testing.T) {

				backup.Status.Id = "asdf"
				state.vaultName = "exactly 1 - succeed"
				err, _ := createBackup(ctx, state)

				assert.Equal(t, composed.StopWithRequeue, err)
				assert.Equal(t, cloudresourcesv1beta1.AzureRwxBackupDone, backup.Status.State)

			})

		})

	})

	t.Run("createBackup - fileshare not yet protected", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		backup := setupDefaultBackup()
		state := setupDefaultState(ctx, backup)

		// Not yet protected

		t.Run("unhappy paths", func(t *testing.T) {

			t.Run("CreateBackupPolicy fails", func(t *testing.T) {
				backup.Status.Id = "asdf"
				state.vaultName = "vaultName - fail CreateBackupPolicy"
				err, _ := createBackup(ctx, state)
				assert.Equal(t, composed.StopWithRequeue, err)
				assert.Equal(t, cloudresourcesv1beta1.AzureRwxBackupError, backup.Status.State)

			})

			t.Run("ListBackupProtectableItems - Fail", func(t *testing.T) {
				backup.Status.Id = "asdf"
				state.vaultName = "vaultName - fail ListBackupProtectableItems"
				err, _ := createBackup(ctx, state)
				assert.Equal(t, composed.StopWithRequeue, err)
				assert.Equal(t, cloudresourcesv1beta1.AzureRwxBackupError, backup.Status.State)
			})

			t.Run("ListBackupProtectableItems - zero matching protectable items", func(t *testing.T) {
				backup.Status.Id = "asdf"
				state.vaultName = "vaultName - zero"
				err, _ := createBackup(ctx, state)
				assert.Equal(t, composed.StopWithRequeue, err)
				assert.Equal(t, cloudresourcesv1beta1.AzureRwxBackupError, backup.Status.State)
			})

			t.Run("ListBackupProtectableItems -  more than one matching protectable items", func(t *testing.T) {
				backup.Status.Id = "asdf"
				state.vaultName = "vaultName - more than 1"
				err, _ := createBackup(ctx, state)
				assert.Equal(t, composed.StopAndForget, err)
				assert.Equal(t, cloudresourcesv1beta1.AzureRwxBackupFailed, backup.Status.State)
			})

			t.Run("ListBackupProtectableItems -  no matching protectable items", func(t *testing.T) {
				backup.Status.Id = "asdf"
				state.vaultName = "vaultName - zero"
				err, _ := createBackup(ctx, state)
				assert.Equal(t, composed.StopWithRequeue, err)
				assert.Equal(t, cloudresourcesv1beta1.AzureRwxBackupError, backup.Status.State)
			})

			t.Run("ListBackupProtectableItems - exactly one protectable item with nil name", func(t *testing.T) {
				backup.Status.Id = "asdf"
				state.vaultName = "vaultName - one nil"
				err, _ := createBackup(ctx, state)
				assert.Equal(t, composed.StopAndForget, err)
				assert.Equal(t, cloudresourcesv1beta1.AzureRwxBackupFailed, backup.Status.State)
			})

			t.Run("CreateOrUpdateProtectedItem - Fails", func(t *testing.T) {
				backup.Status.Id = "asdf"
				state.vaultName = "vaultName - one fail CreateOrUpdateProtectedItem"
				err, _ := createBackup(ctx, state)
				assert.Equal(t, composed.StopWithRequeue, err)
				assert.Equal(t, cloudresourcesv1beta1.AzureRwxBackupError, backup.Status.State)

			})

			t.Run("CreateOrUpdateProtectedItem - succeeds but TriggerBackup fails", func(t *testing.T) {
				backup.Status.Id = "asdf"
				state.vaultName = "vaultName - one pass CreateOrUpdateProtectedItem - fail TriggerBackup"
				err, _ := createBackup(ctx, state)
				assert.Equal(t, composed.StopWithRequeue, err)
				assert.Equal(t, cloudresourcesv1beta1.AzureRwxBackupError, backup.Status.State)

			})

		})

		t.Run("happy path", func(t *testing.T) {

			t.Run("CreateOrUpdateProtectedItem succeeds and TriggerBackup succeeds", func(t *testing.T) {
				backup.Status.Id = "asdf"
				state.vaultName = "vaultName - one pass CreateOrUpdateProtectedItem - succeed TriggerBackup"
				err, _ := createBackup(ctx, state)
				assert.Equal(t, composed.StopWithRequeueDelay(util.Timing.T60000ms()), err)
				assert.Equal(t, cloudresourcesv1beta1.AzureRwxBackupDone, backup.Status.State)

			})

		})

	})

}
