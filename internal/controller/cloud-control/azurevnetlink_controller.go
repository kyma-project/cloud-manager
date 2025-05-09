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

	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/manager"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"

	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	vnetlinkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vnetlink/client"

	vnetlink "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vnetlink"
)

func SetupAzureVNetLinkReconciler(
	kcpManager manager.Manager,
	azureProvider azureclient.ClientProvider[vnetlinkclient.Client],
	env abstractions.Environment) error {

	if env == nil {
		abstractions.NewOSEnvironment()
	}
	return NewAzureVNetLinkReconciler(
		vnetlink.NewAzureVNetLinkReconciler(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(kcpManager)),
			focal.NewStateFactory(),
			vnetlink.NewStateFactory(azureProvider),
		),
	).SetupWithManager(kcpManager)

}

func NewAzureVNetLinkReconciler(reconciler vnetlink.AzureVNetLinkReconciler,
) *AzureVNetLinkReconciler {
	return &AzureVNetLinkReconciler{
		Reconciler: reconciler,
	}
}

// AzureVNetLinkReconciler reconciles a AzureVNetLink object
type AzureVNetLinkReconciler struct {
	Reconciler vnetlink.AzureVNetLinkReconciler
}

// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=azurevnetlinks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=azurevnetlinks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=azurevnetlinks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AzureVNetLink object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *AzureVNetLinkReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.Reconciler.Reconcile(ctx, req)
}

// SetupWithManager sets up the controller with the Manager.
func (r *AzureVNetLinkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudcontrolv1beta1.AzureVNetLink{}).
		Complete(r)
}
