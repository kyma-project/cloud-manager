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
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsnetwork "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/network"
	provider "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azurenetwork "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/network"
	networkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/network/client"
	gcpnetwork "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/network"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/kcp/network"
	ctrl "sigs.k8s.io/controller-runtime"
)

func SetupNetworkReconciler(
	kcpManager manager.Manager,
	azureProvider provider.ClientProvider[networkclient.Client],
) error {
	return NewNetworkReconciler(
		network.NewNetworkReconciler(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(kcpManager)),
			focal.NewStateFactory(),
			awsnetwork.NewStateFactory(),
			azurenetwork.NewStateFactory(azureProvider),
			gcpnetwork.NewStateFactory(),
		),
	).SetupWithManager(kcpManager)
}

func NewNetworkReconciler(reconciler reconcile.Reconciler) *NetworkReconciler {
	return &NetworkReconciler{
		reconciler: reconciler,
	}
}

// NetworkReconciler reconciles a Network object
type NetworkReconciler struct {
	reconciler reconcile.Reconciler
}

// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=networks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=networks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=networks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *NetworkReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.reconciler.Reconcile(ctx, req)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NetworkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// index networks by scope name
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if err := mgr.GetFieldIndexer().IndexField(ctx, &cloudcontrolv1beta1.Network{}, cloudcontrolv1beta1.NetworkFieldScope, func(obj client.Object) []string {
		net := obj.(*cloudcontrolv1beta1.Network)
		return []string{net.Spec.Scope.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudcontrolv1beta1.Network{}).
		Complete(r)
}
