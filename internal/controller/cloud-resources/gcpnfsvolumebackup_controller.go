/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloudresources

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpnfsbackupclientv1 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v1"
	gcpnfsbackupclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v2"
	gcpnfsvolumebackupv1 "github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolumebackup/v1"
	gcpnfsvolumebackupv2 "github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolumebackup/v2"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"

	reconcile2 "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	ctrl "sigs.k8s.io/controller-runtime"
)

// gcpNfsVolumeBackupRunner is a common interface for v1 and v2 reconcilers
type gcpNfsVolumeBackupRunner interface {
	Run(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
}

// GcpNfsVolumeBackupReconciler reconciles a GcpNfsVolumeBackup object
type GcpNfsVolumeBackupReconciler struct {
	reconciler gcpNfsVolumeBackupRunner
}

//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=gcpnfsvolumebackups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=gcpnfsvolumebackups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=gcpnfsvolumebackups/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GcpNfsVolumeBackup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *GcpNfsVolumeBackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.reconciler.Run(ctx, req)
}

type GcpNfsVolumeBackupReconcilerFactory struct {
	fileBackupClientProviderV1 gcpclient.ClientProvider[gcpnfsbackupclientv1.FileBackupClient]
	fileBackupClientProviderV2 gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient]
	env                        abstractions.Environment
}

func (f *GcpNfsVolumeBackupReconcilerFactory) New(args reconcile2.ReconcilerArguments) reconcile.Reconciler {
	// Check feature flag at reconciler creation time (after feature.Initialize has run)
	if feature.GcpBackupV2.Value(context.Background()) {
		reconciler := gcpnfsvolumebackupv2.NewReconciler(
			args.KymaRef,
			args.KcpCluster,
			args.SkrCluster,
			f.fileBackupClientProviderV2,
		)
		return &GcpNfsVolumeBackupReconciler{reconciler: &reconciler}
	}

	reconciler := gcpnfsvolumebackupv1.NewReconciler(
		args.KymaRef,
		args.KcpCluster,
		args.SkrCluster,
		f.fileBackupClientProviderV1,
		f.env,
	)
	return &GcpNfsVolumeBackupReconciler{reconciler: &reconciler}
}

func SetupGcpNfsVolumeBackupReconciler(
	reg skrruntime.SkrRegistry,
	fileBackupClientProviderV1 gcpclient.ClientProvider[gcpnfsbackupclientv1.FileBackupClient],
	fileBackupClientProviderV2 gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient],
	env abstractions.Environment,
) error {
	return reg.Register().
		WithFactory(&GcpNfsVolumeBackupReconcilerFactory{
			fileBackupClientProviderV1: fileBackupClientProviderV1,
			fileBackupClientProviderV2: fileBackupClientProviderV2,
			env:                        env,
		}).
		For(&cloudresourcesv1beta1.GcpNfsVolumeBackup{}).
		Complete()
}
