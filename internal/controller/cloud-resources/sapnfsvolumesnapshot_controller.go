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
	"github.com/kyma-project/cloud-manager/pkg/composed"
	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	reconcile2 "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	"github.com/kyma-project/cloud-manager/pkg/skr/sapnfsvolumesnapshot"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type SapNfsVolumeSnapshotReconcilerFactory struct {
	snapshotClientProvider sapclient.SapClientProvider[sapclient.SnapshotClient]
}

func (f *SapNfsVolumeSnapshotReconcilerFactory) New(args reconcile2.ReconcilerArguments) reconcile.Reconciler {
	kcpCluster := composed.NewStateClusterFromCluster(args.KcpCluster)
	skrCluster := composed.NewStateClusterFromCluster(args.SkrCluster)
	r := sapnfsvolumesnapshot.NewReconcilerFactory().New(
		args.ScopeProvider,
		kcpCluster,
		skrCluster,
		f.snapshotClientProvider,
	)
	return &SapNfsVolumeSnapshotReconciler{reconciler: r}
}

// SapNfsVolumeSnapshotReconciler reconciles a SapNfsVolumeSnapshot object
type SapNfsVolumeSnapshotReconciler struct {
	reconciler *sapnfsvolumesnapshot.Reconciler
}

// +kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=sapnfsvolumesnapshots,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=sapnfsvolumesnapshots/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=sapnfsvolumesnapshots/finalizers,verbs=update

func (r *SapNfsVolumeSnapshotReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.reconciler.Run(ctx, req)
}

func SetupSapNfsVolumeSnapshotReconciler(
	reg skrruntime.SkrRegistry,
	snapshotClientProvider sapclient.SapClientProvider[sapclient.SnapshotClient],
) error {
	return reg.Register().
		WithFactory(&SapNfsVolumeSnapshotReconcilerFactory{
			snapshotClientProvider: snapshotClientProvider,
		}).
		For(&cloudresourcesv1beta1.SapNfsVolumeSnapshot{}).
		Complete()
}
