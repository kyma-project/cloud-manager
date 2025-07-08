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

	"github.com/kyma-project/cloud-manager/pkg/skr/gcpsubnet"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	skrreconciler "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
)

type GcpSubnetReconcilerFactory struct{}

func (f *GcpSubnetReconcilerFactory) New(args skrreconciler.ReconcilerArguments) reconcile.Reconciler {
	return &GcpSubnetReconciler{
		reconciler: gcpsubnet.NewReconcilerFactory().New(args),
	}
}

// GcpSubnetReconciler reconciles a GcpSubnet object
type GcpSubnetReconciler struct {
	reconciler reconcile.Reconciler
}

// +kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=gcpsubnets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=gcpsubnets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=gcpsubnets/finalizers,verbs=update

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
	return r.reconciler.Reconcile(ctx, req)
}

func SetupGcpSubnetReconciler(reg skrruntime.SkrRegistry) error {

	reg.IndexField(&cloudresourcesv1beta1.GcpRedisCluster{}, cloudresourcesv1beta1.GcpSubnetField, func(object client.Object) []string {
		gcpRedisCluster, ok := object.(*cloudresourcesv1beta1.GcpRedisCluster)
		if !ok {
			return []string{}
		}
		if gcpRedisCluster.Spec.Subnet.Name == "" {
			return []string{"default"}
		}
		return []string{gcpRedisCluster.Spec.Subnet.Name}
	})

	return reg.Register().
		WithFactory(&GcpSubnetReconcilerFactory{}).
		For(&cloudresourcesv1beta1.GcpSubnet{}).
		Complete()
}
