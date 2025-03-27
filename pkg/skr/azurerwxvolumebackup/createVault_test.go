package azurerwxvolumebackup

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	vaultClient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	spy "github.com/kyma-project/cloud-manager/pkg/testinfra/clientspy"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestCreateVault(t *testing.T) {

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

		ctx := context.Background()

		state.client, _ = vaultClient.NewMockClient()(ctx, "", "", "", "")

		t.Run("happy path", func(t *testing.T) {

			// Act
			err, res := createVault(ctx, state)

			// Assert
			assert.Equal(t, ctx, res, "should return same context")
			assert.Equal(t, composed.StopAndForget, err, "should expect stop and forget")

		})

		t.Run("unhappy path", func(t *testing.T) {

			// Arrange
			state.resourceGroupName = "http-error"

			// Act
			err, res := createVault(ctx, state)

			// Assert
			assert.Equal(t, ctx, res, "should return same context")
			assert.Equal(t, composed.StopWithRequeue, err, "should expect stop with requeue")

		})

	})

}
