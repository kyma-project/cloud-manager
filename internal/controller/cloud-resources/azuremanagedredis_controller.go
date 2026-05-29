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
	"github.com/kyma-project/cloud-manager/pkg/skr/azuremanagedredis"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	skrreconciler "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type AzureManagedRedisReconcilerFactory struct{}

func (f *AzureManagedRedisReconcilerFactory) New(args skrreconciler.ReconcilerArguments) reconcile.Reconciler {
	return &AzureManagedRedisReconciler{
		reconciler: azuremanagedredis.NewReconcilerFactory().New(args),
	}
}

// AzureManagedRedisReconciler reconciles an AzureManagedRedis object on the SKR side.
type AzureManagedRedisReconciler struct {
	reconciler reconcile.Reconciler
}

//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=azuremanagedredis,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=azuremanagedredis/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=azuremanagedredis/finalizers,verbs=update

func (r *AzureManagedRedisReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.reconciler.Reconcile(ctx, req)
}

func SetupAzureManagedRedisReconciler(reg skrruntime.SkrRegistry) error {
	return reg.Register().
		WithFactory(&AzureManagedRedisReconcilerFactory{}).
		For(&cloudresourcesv1beta1.AzureManagedRedis{}).
		Complete()
}
