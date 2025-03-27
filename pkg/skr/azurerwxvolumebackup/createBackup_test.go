package azurerwxvolumebackup

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureClient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	spy "github.com/kyma-project/cloud-manager/pkg/testinfra/clientspy"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestCreateBackup(t *testing.T) {

	t.Run("createBackup", func(t *testing.T) {
		// Arrange
		var backup *cloudresourcesv1beta1.AzureRwxVolumeBackup
		var k8sClient client.WithWatch
		var fakeClient client.WithWatch
		var state *State

		backup = &cloudresourcesv1beta1.AzureRwxVolumeBackup{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: "backupName",
			},
			Spec: cloudresourcesv1beta1.AzureRwxVolumeBackupSpec{
				Location: "uswest",
			},
			Status: cloudresourcesv1beta1.AzureRwxVolumeBackupStatus{},
		}

		scheme := runtime.NewScheme()
		fakeClient = fake.NewClientBuilder().WithScheme(scheme).Build()
		k8sClient = spy.NewClientSpy(fakeClient)
		cluster := composed.NewStateCluster(k8sClient, k8sClient, nil, k8sClient.Scheme())

		state = &State{
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

		ctx := context.Background()

		state.client, _ = azureClient.NewMockClient()(ctx, "", "", "", "")
		state.scope = scope
		state.fileShareName = "matchingFileShareName"

		t.Run("unhappy path - Id is empty", func(t *testing.T) {
			err, _ := createBackup(ctx, state)

			assert.Equal(t, composed.StopWithRequeue, err)

		})

		t.Run("unhappy path - Id is empty", func(t *testing.T) {
			backup.Status.Id = "asdf"

			state.vaultName = "fail ListProtectedItems"
			err, _ := createBackup(ctx, state)

			assert.Equal(t, composed.StopWithRequeue, err)

		})

		t.Run("unhappy path - more than one matching name", func(t *testing.T) {

			backup.Status.Id = "asdf"
			state.vaultName = "more than 1"
			err, _ := createBackup(ctx, state)

			assert.Equal(t, composed.StopAndForget, err)

		})

		t.Run("unhappy path - already protected and one matching name; fail backup", func(t *testing.T) {

			backup.Status.Id = "asdf"
			state.vaultName = "exactly 1 - fail"
			err, _ := createBackup(ctx, state)

			assert.Equal(t, composed.StopWithRequeue, err)

		})

		t.Run("happy path - already protected and one matching name; succeed backup", func(t *testing.T) {

			backup.Status.Id = "asdf"
			state.vaultName = "exactly 1 - succeed"
			err, _ := createBackup(ctx, state)

			assert.Equal(t, composed.StopWithRequeue, err)

		})

		// Not yet protected

		t.Run("unhappy path - CreateBackupPolicy fails", func(t *testing.T) {
			backup.Status.Id = "asdf"
			state.vaultName = "vaultName - fail CreateBackupPolicy"
			err, _ := createBackup(ctx, state)
			assert.Equal(t, composed.StopWithRequeue, err)

		})

		t.Run("unhappy path - ListBackupProtectableItems Fail", func(t *testing.T) {
			backup.Status.Id = "asdf"
			state.vaultName = "vaultName - fail ListBackupProtectableItems"
			err, _ := createBackup(ctx, state)
			assert.Equal(t, composed.StopWithRequeue, err)

		})

	})

}

func TestHandleNoBackupPolicy(t *testing.T) {

}
