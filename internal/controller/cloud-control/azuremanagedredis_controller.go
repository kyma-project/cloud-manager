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

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/managedredis"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azuremanagedredis "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/managedredis"
	azuremanagedredisclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/managedredis/client"
)

func SetupAzureManagedRedisReconciler(
	kcpManager manager.Manager,
	azureManagedRedisClientProvider azureclient.ClientProvider[azuremanagedredisclient.Client],
) error {
	return NewAzureManagedRedisReconciler(
		managedredis.NewManagedRedisReconciler(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(kcpManager)),
			focal.NewStateFactory(),
			azuremanagedredis.NewStateFactory(azureManagedRedisClientProvider),
		),
	).SetupWithManager(kcpManager)
}

func NewAzureManagedRedisReconciler(
	reconciler managedredis.ManagedRedisReconciler,
) *AzureManagedRedisReconciler {
	return &AzureManagedRedisReconciler{
		Reconciler: reconciler,
	}
}

type AzureManagedRedisReconciler struct {
	Reconciler managedredis.ManagedRedisReconciler
}

// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=azuremanagedredis,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=azuremanagedredis/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=azuremanagedredis/finalizers,verbs=update

func (r *AzureManagedRedisReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.Reconciler.Reconcile(ctx, req)
}

func (r *AzureManagedRedisReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudcontrolv1beta1.AzureManagedRedis{}, builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Complete(r)
}
