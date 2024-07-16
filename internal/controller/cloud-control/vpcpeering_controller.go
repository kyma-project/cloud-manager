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
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/cloudclient"

	awsvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/vpcpeering/client"
	azurevpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vpcpeering/client"
	gcpvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/vpcpeering/client"

	"sigs.k8s.io/controller-runtime/pkg/manager"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"

	awsVpCPeering "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/vpcpeering"
	azurevpcpeering "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vpcpeering"
	gcpvpcpeering "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/vpcpeering"
	"github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering"
)

func SetupVpcPeeringReconciler(
	kcpManager manager.Manager,
	awsSkrProvider awsclient.SkrClientProvider[awsvpcpeeringclient.Client],
	azureSkrProvider azureclient.SkrClientProvider[azurevpcpeeringclient.Client],
	gcpSkrProvider gcpclient.ClientProvider[gcpvpcpeeringclient.Client],
	env abstractions.Environment,
) error {
	if env == nil {
		env = abstractions.NewOSEnvironment()
	}
	return NewVpcPeeringReconciler(
		vpcpeering.NewVpcPeeringReconciler(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(kcpManager)),
			focal.NewStateFactory(),
			awsVpCPeering.NewStateFactory(awsSkrProvider),
			azurevpcpeering.NewStateFactory(azureSkrProvider),
			gcpvpcpeering.NewStateFactory(gcpSkrProvider, env),
		),
	).SetupWithManager(kcpManager)
}

func NewVpcPeeringReconciler(
	reconciler vpcpeering.VPCPeeringReconciler,
) *VpcPeeringReconciler {
	return &VpcPeeringReconciler{
		Reconciler: reconciler,
	}
}

type VpcPeeringReconciler struct {
	Reconciler vpcpeering.VPCPeeringReconciler
}

//+kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=vpcpeerings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=vpcpeerings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=vpcpeerings/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VpcPeering object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *VpcPeeringReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.Reconciler.Reconcile(ctx, req)
}

// SetupWithManager sets up the controller with the Manager.
func (r *VpcPeeringReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudcontrolv1beta1.VpcPeering{}).
		Complete(r)
}
