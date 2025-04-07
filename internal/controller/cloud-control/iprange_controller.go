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
	"fmt"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsiprange "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/iprange"
	awsiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/iprange/client"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azureiprange "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/iprange"
	azureiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/iprange/client"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange"
	gcpiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	v3gcpiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v3/client"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func SetupIpRangeReconciler(
	ctx context.Context,
	kcpManager manager.Manager,
	awsProvider awsclient.SkrClientProvider[awsiprangeclient.Client],
	azureProvider azureclient.ClientProvider[azureiprangeclient.Client],
	gcpSvcNetProvider gcpclient.ClientProvider[gcpiprangeclient.ServiceNetworkingClient],
	gcpComputeProvider gcpclient.ClientProvider[gcpiprangeclient.ComputeClient],
	v3ComputeProvider gcpclient.ClientProvider[v3gcpiprangeclient.ComputeClient],
	v3NetworkConnectivityClient gcpclient.ClientProvider[v3gcpiprangeclient.NetworkConnectivityClient],
	env abstractions.Environment,
) error {
	if env == nil {
		env = abstractions.NewOSEnvironment()
	}
	return NewIpRangeReconciler(
		iprange.NewIPRangeReconciler(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(kcpManager)),
			focal.NewStateFactory(),
			awsiprange.NewStateFactory(awsProvider),
			azureiprange.NewStateFactory(azureProvider),
			gcpiprange.NewStateFactory(
				gcpSvcNetProvider,
				gcpComputeProvider,
				v3ComputeProvider,
				v3NetworkConnectivityClient,
				env,
			),
		),
	).SetupWithManager(ctx, kcpManager)
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
func (r *IpRangeReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	if err := mgr.GetFieldIndexer().IndexField(
		ctx,
		&cloudcontrolv1beta1.NfsInstance{},
		cloudcontrolv1beta1.IpRangeField,
		func(obj client.Object) []string {
			nfsInstance := obj.(*cloudcontrolv1beta1.NfsInstance)
			return []string{fmt.Sprintf("%s/%s", nfsInstance.Namespace, nfsInstance.Spec.IpRange.Name)}
		}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(
		ctx,
		&cloudcontrolv1beta1.RedisInstance{},
		cloudcontrolv1beta1.IpRangeField,
		func(obj client.Object) []string {
			redisInstance := obj.(*cloudcontrolv1beta1.RedisInstance)
			return []string{fmt.Sprintf("%s/%s", redisInstance.Namespace, redisInstance.Spec.IpRange.Name)}
		}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudcontrolv1beta1.IpRange{}).
		Complete(r)
}
