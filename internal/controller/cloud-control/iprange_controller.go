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
	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
	sapiprange "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/iprange"
	sapiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/iprange/client"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func SetupIpRangeReconciler(
	ctx context.Context,
	kcpManager manager.Manager,
	awsProvider awsclient.SkrClientProvider[awsiprangeclient.Client],
	azureProvider azureclient.ClientProvider[azureiprangeclient.Client],
	gcpSvcNetProvider gcpclient.GcpClientProvider[gcpiprangeclient.ServiceNetworkingClient],
	gcpComputeProvider gcpclient.GcpClientProvider[gcpiprangeclient.ComputeClient],
	sapProvider sapclient.SapClientProvider[sapiprangeclient.Client],
	env abstractions.Environment,
	gcpOldProviders ...interface{}, // Optional for testing: [0] = OldComputeClient, [1] = ServiceNetworkingClient (v2)
) error {
	if env == nil {
		env = abstractions.NewOSEnvironment()
	}

	// Create v3 GCP state factory (NEW pattern with clean actions)
	gcpV3StateFactory := gcpiprange.NewV3StateFactory(gcpSvcNetProvider, gcpComputeProvider)

	// For v2 ServiceNetworking, use provided mock or production provider
	var gcpV2SvcNetProvider gcpclient.ClientProvider[gcpiprangeclient.ServiceNetworkingClient]
	if len(gcpOldProviders) > 1 {
		if provider, ok := gcpOldProviders[1].(gcpclient.ClientProvider[gcpiprangeclient.ServiceNetworkingClient]); ok {
			gcpV2SvcNetProvider = provider // Use provided mock
		}
	}
	if gcpV2SvcNetProvider == nil {
		//nolint:staticcheck // SA1019: Using deprecated function intentionally for v2 legacy implementation
		gcpV2SvcNetProvider = gcpiprangeclient.NewServiceNetworkingClient() // Use production (ClientProvider) for v2
	}

	// For OldComputeClient, use provided mock or production provider
	var gcpV2ComputeProvider gcpclient.ClientProvider[gcpiprangeclient.OldComputeClient]
	if len(gcpOldProviders) > 0 {
		if provider, ok := gcpOldProviders[0].(gcpclient.ClientProvider[gcpiprangeclient.OldComputeClient]); ok {
			gcpV2ComputeProvider = provider // Use provided mock
		}
	}
	if gcpV2ComputeProvider == nil {
		gcpV2ComputeProvider = gcpiprangeclient.NewOldComputeClientProvider() // Use production
	}

	gcpV2StateFactory := gcpiprange.NewV2StateFactory(gcpV2SvcNetProvider, gcpV2ComputeProvider, env)

	return NewIpRangeReconciler(
		iprange.NewIPRangeReconciler(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(kcpManager)),
			focal.NewStateFactory(),
			awsiprange.NewStateFactory(awsProvider),
			azureiprange.NewStateFactory(azureProvider),
			gcpV3StateFactory,
			gcpV2StateFactory,
			sapiprange.NewStateFactory(sapProvider),
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

	if err := mgr.GetFieldIndexer().IndexField(
		ctx,
		&cloudcontrolv1beta1.RedisCluster{},
		cloudcontrolv1beta1.IpRangeField,
		func(obj client.Object) []string {
			redisCluster := obj.(*cloudcontrolv1beta1.RedisCluster)
			return []string{fmt.Sprintf("%s/%s", redisCluster.Namespace, redisCluster.Spec.IpRange.Name)}
		}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudcontrolv1beta1.IpRange{}).
		Complete(r)
}
