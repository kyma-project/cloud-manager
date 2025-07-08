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

package cloudcontrol

import (
	"context"
	"fmt"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"

	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/subnet"
	gcpsubnetclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/subnet/client"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func SetupGcpSubnetReconciler(
	ctx context.Context,
	kcpManager manager.Manager,
	computeClientProvider gcpclient.GcpClientProvider[gcpsubnetclient.ComputeClient],
	networkConnectivityClientProvider gcpclient.GcpClientProvider[gcpsubnetclient.NetworkConnectivityClient],
	env abstractions.Environment,
) error {
	if env == nil {
		env = abstractions.NewOSEnvironment()
	}
	return NewGcpSubnetReconciler(
		subnet.NewGcpSubnetReconciler(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(kcpManager)),
			focal.NewStateFactory(),
			subnet.NewStateFactory(computeClientProvider, networkConnectivityClientProvider, env),
		),
	).SetupWithManager(ctx, kcpManager)
}

func NewGcpSubnetReconciler(
	reconciler subnet.GcpSubnetReconciler,
) *GcpSubnetReconciler {
	return &GcpSubnetReconciler{
		Reconciler: reconciler,
	}
}

type GcpSubnetReconciler struct {
	Reconciler subnet.GcpSubnetReconciler
}

// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=gcpsubnets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=gcpsubnets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=gcpsubnets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GcpSubnet object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *GcpSubnetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.Reconciler.Reconcile(ctx, req)
}

// SetupWithManager sets up the controller with the Manager.
func (r *GcpSubnetReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	if err := mgr.GetFieldIndexer().IndexField(
		ctx,
		&cloudcontrolv1beta1.GcpRedisCluster{},
		cloudcontrolv1beta1.GcpSubnetField,
		func(obj client.Object) []string {
			gcpRedisCluster := obj.(*cloudcontrolv1beta1.GcpRedisCluster)
			return []string{fmt.Sprintf("%s/%s", gcpRedisCluster.Namespace, gcpRedisCluster.Spec.Subnet.Name)}
		}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudcontrolv1beta1.GcpSubnet{}, builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Complete(r)
}
