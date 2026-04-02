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
	gcpnfsrestoreclientv1 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client/v1"
	gcpnfsrestoreclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client/v2"
	gcpnfsvolumerestorev1 "github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolumerestore/v1"
	gcpnfsvolumerestorev2 "github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolumerestore/v2"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	reconcile2 "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	ctrl "sigs.k8s.io/controller-runtime"
)

// gcpNfsVolumeRestoreRunner is a common interface for v1 and v2 reconcilers
type gcpNfsVolumeRestoreRunner interface {
	Run(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
}

// GcpNfsVolumeRestoreReconciler reconciles a GcpNfsVolumeRestore object
type GcpNfsVolumeRestoreReconciler struct {
	reconciler gcpNfsVolumeRestoreRunner
}

//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=gcpnfsvolumerestores,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=gcpnfsvolumerestores/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=gcpnfsvolumerestores/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GcpNfsVolumeRestore object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *GcpNfsVolumeRestoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.reconciler.Run(ctx, req)
}

type GcpNfsVolumeRestoreReconcilerFactory struct {
	// v1 providers (old pattern: func(ctx, credFile) (T, error))
	fileRestoreClientProviderV1 gcpclient.ClientProvider[gcpnfsrestoreclientv1.FileRestoreClient]
	fileBackupClientProviderV1  gcpclient.ClientProvider[gcpnfsbackupclientv1.FileBackupClient]
	env                         abstractions.Environment
	// v2 providers (new pattern: func() T)
	fileRestoreClientProviderV2 gcpclient.GcpClientProvider[gcpnfsrestoreclientv2.FileRestoreClient]
	fileBackupClientProviderV2  gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient]
}

func (f *GcpNfsVolumeRestoreReconcilerFactory) New(args reconcile2.ReconcilerArguments) reconcile.Reconciler {
	// Check feature flag at reconciler creation time (after feature.Initialize has run)
	if feature.GcpNfsRestoreV2.Value(context.Background()) {
		reconciler := gcpnfsvolumerestorev2.NewReconciler(
			args.ScopeProvider,
			args.KcpCluster,
			args.SkrCluster,
			f.fileRestoreClientProviderV2,
			f.fileBackupClientProviderV2,
		)
		return &GcpNfsVolumeRestoreReconciler{reconciler: &reconciler}
	}

	reconciler := gcpnfsvolumerestorev1.NewReconciler(
		args.ScopeProvider,
		args.KcpCluster,
		args.SkrCluster,
		f.fileRestoreClientProviderV1,
		f.fileBackupClientProviderV1,
		f.env,
	)
	return &GcpNfsVolumeRestoreReconciler{reconciler: &reconciler}
}

func SetupGcpNfsVolumeRestoreReconciler(
	reg skrruntime.SkrRegistry,
	fileRestoreClientProviderV1 gcpclient.ClientProvider[gcpnfsrestoreclientv1.FileRestoreClient],
	fileBackupClientProviderV1 gcpclient.ClientProvider[gcpnfsbackupclientv1.FileBackupClient],
	fileRestoreClientProviderV2 gcpclient.GcpClientProvider[gcpnfsrestoreclientv2.FileRestoreClient],
	fileBackupClientProviderV2 gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient],
	env abstractions.Environment,
) error {
	return reg.Register().
		WithFactory(&GcpNfsVolumeRestoreReconcilerFactory{
			fileRestoreClientProviderV1: fileRestoreClientProviderV1,
			fileBackupClientProviderV1:  fileBackupClientProviderV1,
			fileRestoreClientProviderV2: fileRestoreClientProviderV2,
			fileBackupClientProviderV2:  fileBackupClientProviderV2,
			env:                         env,
		}).
		For(&cloudresourcesv1beta1.GcpNfsVolumeRestore{}).
		Complete()
}
