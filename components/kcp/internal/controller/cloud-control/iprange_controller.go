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
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/composed"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/iprange"
	awsclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/aws/client"
	awsiprange "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/aws/iprange"
	iprangeclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/aws/iprange/client"
	azureiprange "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/azure/iprange"
	gcpclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/client"
	gcpiprange "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/iprange"
	gcpiprangeclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/iprange/client"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func SetupIpRangeReconciler(
	kcpManager manager.Manager,
	awsProvider awsclient.SkrClientProvider[iprangeclient.Client],
	gcpSvcNetProvider gcpclient.ClientProvider[gcpiprangeclient.ServiceNetworkingClient],
	gcpComputeProvider gcpclient.ClientProvider[gcpiprangeclient.ComputeClient],
) error {
	return NewIpRangeReconciler(
		iprange.NewIPRangeReconciler(
			composed.NewStateFactory(composed.NewStateClusterFromManager(kcpManager)),
			focal.NewStateFactory(),
			awsiprange.NewStateFactory(awsProvider, abstractions.NewOSEnvironment()),
			azureiprange.NewStateFactory(nil),
			gcpiprange.NewStateFactory(gcpSvcNetProvider, gcpComputeProvider, abstractions.NewOSEnvironment()),
		),
	).SetupWithManager(kcpManager)
}

func NewIpRangeReconciler(
	reconciler iprange.IPRangeReconciler,
) *IpRangeReconciler {
	return &IpRangeReconciler{
		Reconciler: reconciler,
	}
}

type IpRangeReconciler struct {
	Reconciler iprange.IPRangeReconciler
}

//+kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=ipranges,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=ipranges/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=ipranges/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *IpRangeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.Reconciler.Reconcile(ctx, req)
}

// SetupWithManager sets up the controller with the Manager.
func (r *IpRangeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudresourcesv1beta1.IpRange{}).
		Complete(r)
}
