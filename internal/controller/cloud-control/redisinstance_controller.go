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

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azureredisinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/redisinstance"
	azureredisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/redisinstance/client"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpredisinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/redisinstance"
	gcpredisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/redisinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/redisinstance"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func SetupRedisInstanceReconciler(
	kcpManager manager.Manager,
	gcpFilestoreClientProvider gcpclient.ClientProvider[gcpredisinstanceclient.MemorystoreClient],
	azureFilestoreClientProvider azureclient.SkrClientProvider[azureredisinstanceclient.Client],
	env abstractions.Environment,
) error {
	if env == nil {
		env = abstractions.NewOSEnvironment()
	}
	return NewRedisInstanceReconciler(
		redisinstance.NewRedisInstanceReconciler(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(kcpManager)),
			focal.NewStateFactory(),
			gcpredisinstance.NewStateFactory(gcpFilestoreClientProvider, env),
			azureredisinstance.NewStateFactory(azureFilestoreClientProvider),
		),
	).SetupWithManager(kcpManager)
}

func NewRedisInstanceReconciler(
	reconciler redisinstance.RedisInstanceReconciler,
) *RedisInstanceReconciler {
	return &RedisInstanceReconciler{
		Reconciler: reconciler,
	}
}

type RedisInstanceReconciler struct {
	Reconciler redisinstance.RedisInstanceReconciler
}

//+kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=redisinstances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=redisinstances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=redisinstances/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RedisInstance object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *RedisInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.Reconciler.Reconcile(ctx, req)
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudcontrolv1beta1.RedisInstance{}, builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Complete(r)
}
