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
	"github.com/kyma-project/cloud-manager/pkg/skr/awsrediscluster"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	skrreconciler "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
)

type AwsRedisClusterReconcilerFactory struct{}

func (f *AwsRedisClusterReconcilerFactory) New(args skrreconciler.ReconcilerArguments) reconcile.Reconciler {
	return &AwsRedisClusterReconciler{
		reconciler: awsrediscluster.NewReconcilerFactory().New(args),
	}
}

// AwsRedisClusterReconciler reconciles a AwsRedisCluster object
type AwsRedisClusterReconciler struct {
	reconciler reconcile.Reconciler
}

//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=awsredisclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=awsredisclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=awsredisclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AwsRedisCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *AwsRedisClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.reconciler.Reconcile(ctx, req)
}

func SetupAwsRedisClusterReconciler(reg skrruntime.SkrRegistry) error {
	return reg.Register().
		WithFactory(&AwsRedisClusterReconcilerFactory{}).
		For(&cloudresourcesv1beta1.AwsRedisCluster{}).
		Complete()
}
