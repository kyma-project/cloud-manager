package azurerwxvolumerestore

import (
	"context"
	"testing"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	spy "github.com/kyma-project/cloud-manager/pkg/testinfra/clientspy"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestLoadPersistentVolume(t *testing.T) {

	t.Run("loadPersistentVolume", func(t *testing.T) {

		var azureRwxVolumeRestore *cloudresourcesv1beta1.AzureRwxVolumeRestore
		var state *State
		var k8sClient client.WithWatch
		var pv *corev1.PersistentVolume

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

		setupTest := func(withPv bool, pvStatus corev1.PersistentVolumePhase) {
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

			var fakeClient client.WithWatch
			if withPv {
				pv = &corev1.PersistentVolume{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-azure-restore-pv",
					},
					Status: corev1.PersistentVolumeStatus{
						Phase: pvStatus,
					},
					Spec: corev1.PersistentVolumeSpec{
						PersistentVolumeSource: corev1.PersistentVolumeSource{
							CSI: &corev1.CSIPersistentVolumeSource{
								VolumeHandle: "shoot--kyma-dev--c-6ea9b9b#f21d936aa5673444a95852a#pv-shoot-kyma-dev-c-6ea9b9b-8aa269ae-f581-427b-b05c-a2a2bbfca###default",
							},
						},
					},
				}
				fakeClient = fake.NewClientBuilder().
					WithScheme(commonscheme.SkrScheme).
					WithObjects(azureRwxVolumeRestore).
					WithStatusSubresource(azureRwxVolumeRestore).
					WithObjects(pv).
					WithStatusSubresource(pv).
					Build()
			} else {
				fakeClient = fake.NewClientBuilder().
					WithScheme(commonscheme.SkrScheme).
					WithObjects(azureRwxVolumeRestore).
					WithStatusSubresource(azureRwxVolumeRestore).
					Build()
			}
			k8sClient = spy.NewClientSpy(fakeClient)

			state = createEmptyState(k8sClient, azureRwxVolumeRestore)
			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-azure-restore-pvc",
					Namespace: "test-ns",
					Annotations: map[string]string{
						"volume.kubernetes.io/storage-provisioner": "file.csi.azure.com",
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "test-azure-restore-pv",
				},
			}
			state.pvc = pvc
		}

		t.Run("Should: load Bound PV", func(t *testing.T) {
			setupTest(true, corev1.VolumeBound)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err, res := loadPersistentVolume(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
			// shoot--kyma-dev--c-6ea9b9b#f21d936aa5673444a95852a#pv-shoot-kyma-dev-c-6ea9b9b-8aa269ae-f581-427b-b05c-a2a2bbfca###default
			assert.Equal(t, "shoot--kyma-dev--c-6ea9b9b", state.resourceGroupName, "resource group name should be set in state")
			assert.Equal(t, "f21d936aa5673444a95852a", state.storageAccountName, "storage account name should be set in state")
			assert.Equal(t, "pv-shoot-kyma-dev-c-6ea9b9b-8aa269ae-f581-427b-b05c-a2a2bbfca", state.fileShareName, "file share name should be set in state")
		})

		t.Run("Should: fail PV that is not Bound", func(t *testing.T) {
			setupTest(true, corev1.VolumePending)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err, res := loadPersistentVolume(ctx, state)

			assert.Equal(t, ctx, res, "should return same context")
			assert.Equal(t, err, composed.StopAndForget, "should stop and forget")
			assert.Empty(t, state.resourceGroupName, "resource group name should remain empty in state")
			assert.Empty(t, state.storageAccountName, "storage account name should remain empty in state")
			assert.Empty(t, state.fileShareName, "file share name should remain empty in state")
		})

		t.Run("Should: error out if APIServer cant find requested PV", func(t *testing.T) {
			setupTest(false, corev1.VolumePending)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err, res := loadPersistentVolume(ctx, state)

			assert.Equal(t, ctx, res, "should return same context")
			assert.Equal(t, err, composed.StopAndForget, "should stop and forget")
			assert.Empty(t, state.resourceGroupName, "resource group name should remain empty in state")
			assert.Empty(t, state.storageAccountName, "storage account name should remain empty in state")
			assert.Empty(t, state.fileShareName, "file share name should remain empty in state")
		})

	})
}
