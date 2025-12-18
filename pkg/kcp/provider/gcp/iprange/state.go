package iprange

import (
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	gcpiprangev2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v2"
	gcpiprangev3 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v3"
)

// V3StateFactory is an alias for gcpiprangev3.StateFactory to be used by the reconciler.
type V3StateFactory = gcpiprangev3.StateFactory

// NewV3StateFactory is a wrapper for gcpiprangev3.NewStateFactory to be called from controller setup.
func NewV3StateFactory(
	serviceNetworkingClientProvider gcpclient.GcpClientProvider[gcpiprangeclient.ServiceNetworkingClient],
	computeClientProvider gcpclient.GcpClientProvider[gcpiprangeclient.ComputeClient],
) V3StateFactory {
	return gcpiprangev3.NewStateFactory(serviceNetworkingClientProvider, computeClientProvider)
}

// V2StateFactory is an alias for gcpiprangev2.StateFactory to be used by the reconciler.
type V2StateFactory = gcpiprangev2.StateFactory

// NewV2StateFactory is a wrapper for gcpiprangev2.NewStateFactory to be called from controller setup.
func NewV2StateFactory(
	serviceNetworkingClientProvider gcpclient.ClientProvider[gcpiprangeclient.ServiceNetworkingClient],
	oldComputeClientProvider gcpclient.ClientProvider[gcpiprangeclient.OldComputeClient],
	env abstractions.Environment,
) V2StateFactory {
	return gcpiprangev2.NewStateFactory(serviceNetworkingClientProvider, oldComputeClientProvider, env)
}
