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
	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	restoreclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client"
	"github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolumerestore"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	reconcile2 "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// GcpNfsVolumeRestoreReconciler reconciles a GcpNfsVolumeRestore object
type GcpNfsVolumeRestoreReconciler struct {
	Reconciler gcpnfsvolumerestore.Reconciler
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
	_ = log.FromContext(ctx)
	return r.Reconciler.Run(ctx, req)
}

type GcpNfsVolumeRestoreReconcilerFactory struct {
	fileRestoreClientProvider gcpclient.ClientProvider[restoreclient.FileRestoreClient]
	env                       abstractions.Environment
}

func (f *GcpNfsVolumeRestoreReconcilerFactory) New(args reconcile2.ReconcilerArguments) reconcile.Reconciler {
	return &GcpNfsVolumeRestoreReconciler{
		Reconciler: gcpnfsvolumerestore.NewReconciler(
			args.KymaRef,
			args.KcpCluster,
			args.SkrCluster,
			f.fileRestoreClientProvider, f.env),
	}
}

func SetupGcpNfsVolumeRestoreReconciler(reg skrruntime.SkrRegistry, fileRestoreClientProvider gcpclient.ClientProvider[restoreclient.FileRestoreClient],
	env abstractions.Environment, logger logr.Logger) error {
	// "_" + crd + ".yaml" should be the suffix for the yaml present in config/crd/bases
	if util.IsCrdDisabled(env, "GcpNfsVolumeRestores") {
		logger.Info("GcpNfsVolumeRestore CRD is disabled. Skipping controller setup.")
		return nil
	} else {
		return reg.Register().
			WithFactory(&GcpNfsVolumeRestoreReconcilerFactory{fileRestoreClientProvider: fileRestoreClientProvider, env: env}).
			For(&cloudresourcesv1beta1.GcpNfsVolumeRestore{}).
			Complete()
	}
}
