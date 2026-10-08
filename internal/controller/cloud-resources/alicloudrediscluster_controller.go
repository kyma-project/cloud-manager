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

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/skr/alicloudrediscluster"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	skrreconciler "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
)

type AlicloudRedisClusterReconcilerFactory struct{}

func (f *AlicloudRedisClusterReconcilerFactory) New(args skrreconciler.ReconcilerArguments) reconcile.Reconciler {
	return &AlicloudRedisClusterReconciler{
		reconciler: alicloudrediscluster.NewReconcilerFactory().New(args),
	}
}

// AlicloudRedisClusterReconciler reconciles a AlicloudRedisCluster object
type AlicloudRedisClusterReconciler struct {
	reconciler reconcile.Reconciler
}

//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=alicloudredisclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=alicloudredisclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=alicloudredisclusters/finalizers,verbs=update

func (r *AlicloudRedisClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.reconciler.Reconcile(ctx, req)
}

func SetupAlicloudRedisClusterReconciler(reg skrruntime.SkrRegistry) error {
	return reg.Register().
		WithFactory(&AlicloudRedisClusterReconcilerFactory{}).
		For(&cloudresourcesv1beta1.AlicloudRedisCluster{}).
		Complete()
}
