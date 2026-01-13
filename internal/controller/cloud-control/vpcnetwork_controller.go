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

	"github.com/kyma-project/cloud-manager/pkg/composed"
	kcpcommonaction "github.com/kyma-project/cloud-manager/pkg/kcp/commonAction"
	awsvpcnetwork "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/vpcnetwork"
	kcpvpcnetwork "github.com/kyma-project/cloud-manager/pkg/kcp/vpcnetwork"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
)

func SetupVpcNetworkReconciler(
	kcpManager manager.Manager,
	awsStateFactory awsvpcnetwork.StateFactory,
) error {
	return NewVpcNetworkReconciler(
		kcpvpcnetwork.New(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(kcpManager)),
			kcpcommonaction.NewStateFactory(),
			awsStateFactory,
		),
	).SetupWithManager(kcpManager)
}

func NewVpcNetworkReconciler(r kcpvpcnetwork.VpcNetworkReconciler) *VpcNetworkReconciler {
	return &VpcNetworkReconciler{
		Reconciler: r,
	}
}

// VpcNetworkReconciler reconciles a VpcNetwork object
type VpcNetworkReconciler struct {
	Reconciler kcpvpcnetwork.VpcNetworkReconciler
}

// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=vpcnetworks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=vpcnetworks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=vpcnetworks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VpcNetwork object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.0/pkg/reconcile
func (r *VpcNetworkReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.Reconciler.Reconcile(ctx, req)
}

// SetupWithManager sets up the controller with the Manager.
func (r *VpcNetworkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudcontrolv1beta1.VpcNetwork{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 10,
		}).
		Named("cloud-control-vpcnetwork").
		Complete(r)
}
