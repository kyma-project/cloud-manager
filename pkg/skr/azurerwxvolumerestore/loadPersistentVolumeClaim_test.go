package azurerwxvolumerestore

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	"testing"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	spy "github.com/kyma-project/cloud-manager/pkg/testinfra/clientspy"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestLoadPersistentVolumeClaim(t *testing.T) {

	t.Run("loadPersistentVolumeClaim", func(t *testing.T) {

		var pvc *corev1.PersistentVolumeClaim

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

		setupTest := func(withPvc bool, pvcStatus corev1.PersistentVolumeClaimPhase) {
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
			if withPvc {
				pvc = &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-azure-restore-pvc",
						Namespace: "test-ns",
						Annotations: map[string]string{
							"volume.kubernetes.io/storage-provisioner": "file.csi.azure.com",
						},
					},
					Status: corev1.PersistentVolumeClaimStatus{
						Phase: pvcStatus,
					},
				}
				fakeClient = fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(azureRwxVolumeRestore).
					WithStatusSubresource(azureRwxVolumeRestore).
					WithObjects(pvc).
					WithStatusSubresource(pvc).
					Build()
			} else {
				fakeClient = fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(azureRwxVolumeRestore).
					WithStatusSubresource(azureRwxVolumeRestore).
					Build()
			}
			k8sClient = spy.NewClientSpy(fakeClient)

			state = createEmptyState(k8sClient, azureRwxVolumeRestore)
		}

		t.Run("Should: load Bound pvc", func(t *testing.T) {
			setupTest(true, corev1.ClaimBound)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err, res := loadPersistentVolumeClaim(ctx, state)

			assert.Nil(t, res, "should return nil res")
			assert.Nil(t, err, "should return nil err")
			assert.NotNilf(t, state.pvc, "pvc should be loaded into state")
			assert.Equal(t, state.pvc, pvc, "loaded pvc should be the same as the one in the test")
		})

		t.Run("Should: fail pvc that is not Bound", func(t *testing.T) {
			setupTest(true, corev1.ClaimPending)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err, res := loadPersistentVolumeClaim(ctx, state)

			assert.Equal(t, ctx, res, "should return same context")
			assert.Equal(t, err, composed.StopAndForget, "should stop and forget")
			assert.Nil(t, state.pvc, "pvc should remain nil in state")
		})

		t.Run("Should: fail pvc with invalid provisioner", func(t *testing.T) {
			setupTest(true, corev1.ClaimPending)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			// patch the annotations to have invalid provisioner
			pvc.Annotations["volume.kubernetes.io/storage-provisioner"] = "invalid.provisioner"
			err := k8sClient.Patch(ctx, pvc, client.MergeFrom(pvc), client.FieldOwner("test"))
			assert.Nil(t, err, "should patch pvc with invalid provisioner")
			err, res := loadPersistentVolumeClaim(ctx, state)

			assert.Equal(t, ctx, res, "should return same context")
			assert.Equal(t, err, composed.StopAndForget, "should stop and forget")
			assert.Nil(t, state.pvc, "pvc should remain nil in state")
		})

		t.Run("Should: error out if APIServer cant find requested pvc", func(t *testing.T) {
			setupTest(false, corev1.ClaimPending)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err, res := loadPersistentVolumeClaim(ctx, state)

			assert.Equal(t, ctx, res, "should return same context")
			assert.Equal(t, err, composed.StopAndForget, "should stop and forget")
			assert.Nil(t, state.pvc, "pvc should remain nil in state as it is not found in APIServer")
		})

	})
}
