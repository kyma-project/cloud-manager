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

	"github.com/kyma-project/cloud-manager/pkg/skr/gcprediscluster"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	skrreconciler "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
)

type GcpRedisClusterReconcilerFactory struct{}

func (f *GcpRedisClusterReconcilerFactory) New(args skrreconciler.ReconcilerArguments) reconcile.Reconciler {
	return &GcpRedisClusterReconciler{
		reconciler: gcprediscluster.NewReconcilerFactory().New(args),
	}
}

// GcpRedisClusterReconciler reconciles a GcpRedisCluster object
type GcpRedisClusterReconciler struct {
	reconciler reconcile.Reconciler
}

// +kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=gcpredisclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=gcpredisclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=gcpredisclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GcpRedisCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *GcpRedisClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.reconciler.Reconcile(ctx, req)
}

func SetupGcpRedisClusterReconciler(reg skrruntime.SkrRegistry) error {
	return reg.Register().
		WithFactory(&GcpRedisClusterReconcilerFactory{}).
		For(&cloudresourcesv1beta1.GcpRedisCluster{}).
		Complete()
}
