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

	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpnfsbackupclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client"
	"github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolume"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	reconcile2 "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
)

type GcpNfsVolumeReconcilerFactory struct {
	fileBackupClientProvider gcpclient.ClientProvider[gcpnfsbackupclient.FileBackupClient]
	env                      abstractions.Environment
}

func (f *GcpNfsVolumeReconcilerFactory) New(args reconcile2.ReconcilerArguments) reconcile.Reconciler {
	return &GcpNfsVolumeReconciler{
		kymaRef:    args.KymaRef,
		kcpCluster: args.KcpCluster,
		skrCluster: args.SkrCluster,
		Reconciler: gcpnfsvolume.NewReconciler(args.KymaRef, args.KcpCluster, args.SkrCluster, f.fileBackupClientProvider, f.env),
	}
}

// GcpNfsVolumeReconciler reconciles a GcpNfsVolume object
type GcpNfsVolumeReconciler struct {
	kymaRef    klog.ObjectRef
	kcpCluster cluster.Cluster
	skrCluster cluster.Cluster
	Reconciler gcpnfsvolume.Reconciler
}

//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=gcpnfsvolumes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=gcpnfsvolumes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=gcpnfsvolumes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GcpNfsVolume object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *GcpNfsVolumeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	return r.Reconciler.Run(ctx, req)
}

func SetupGcpNfsVolumeReconciler(
	reg skrruntime.SkrRegistry,
	fileBackupClientProvider gcpclient.ClientProvider[gcpnfsbackupclient.FileBackupClient],
	env abstractions.Environment,
) error {
	return reg.Register().
		WithFactory(&GcpNfsVolumeReconcilerFactory{
			fileBackupClientProvider: fileBackupClientProvider,
			env:                      env,
		}).
		For(&cloudresourcesv1beta1.GcpNfsVolume{}).
		Watches(&corev1.PersistentVolume{}, gcpnfsvolume.PVEventHandler).
		Complete()
}
