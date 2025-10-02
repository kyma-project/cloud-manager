package azurerwxvolumebackup

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/bootstrap"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azurerwxvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	spy "github.com/kyma-project/cloud-manager/pkg/testinfra/clientspy"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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

	scheme := bootstrap.SkrScheme
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	k8sClient := spy.NewClientSpy(fakeClient)
	cluster := composed.NewStateCluster(k8sClient, k8sClient, nil, k8sClient.Scheme())

	return cluster

}

func setupDefaultState(ctx context.Context, backup *cloudresourcesv1beta1.AzureRwxVolumeBackup) *State {

	cluster := setupDefaultCluster()

	kcpScheme := bootstrap.KcpScheme

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

	kymaRef := klog.ObjectRef{
		Name:      "skr",
		Namespace: "test",
	}

	kcpClient := fake.NewClientBuilder().
		WithScheme(kcpScheme).
		WithObjects(scope).
		Build()
	kcpCluster := composed.NewStateCluster(kcpClient, kcpClient, nil, kcpScheme)

	state := &State{
		State: commonscope.NewStateFactory(kcpCluster, kymaRef).NewState(composed.NewStateFactory(cluster).NewState(types.NamespacedName{}, backup)),
	}

	state.client, _ = azurerwxvolumebackupclient.NewMockClient()(ctx, "", "", "", "")
	state.scope = scope
	state.fileShareName = "matchingFileShareName"
	state.subscriptionId = "test-subscription-id"

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

			t.Run("ListProtectedItems - one matching name; fail backup", func(t *testing.T) {

				backup.Status.Id = "asdf"
				kvp := map[string]int{
					"ListProtectedItems match": 1,
				}
				kvp2 := map[string]string{
					"TriggerBackup": "fail",
				}
				newCtx := addValuesToContext(ctx, kvp)
				newCtx = addValuesToContext(newCtx, kvp2)

				err, _ := createBackup(newCtx, state)

				assert.Equal(t, composed.StopWithRequeue, err)
				assert.Equal(t, cloudresourcesv1beta1.AzureRwxBackupError, backup.Status.State)

			})

		})

		t.Run("happy path", func(t *testing.T) {

			t.Run("ListProtectedItems - one matching name; succeed backup", func(t *testing.T) {

				backup.Status.Id = "asdf"

				kvp := map[string]int{
					"ListProtectedItems match": 1,
				}
				newCtx := addValuesToContext(ctx, kvp)

				err, _ := createBackup(newCtx, state)

				assert.Equal(t, composed.StopWithRequeue, err)
				assert.Equal(t, cloudresourcesv1beta1.AzureRwxBackupDone, backup.Status.State)

			})

		})

	})

}
