package iprange

import (
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	v2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v2"
	v3 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v3"
)

// V3StateFactory is an alias for v3.StateFactory to be used by the reconciler.
type V3StateFactory = v3.StateFactory

// NewV3StateFactory is a wrapper for v3.NewStateFactory to be called from controller setup.
func NewV3StateFactory(
	serviceNetworkingClientProvider gcpclient.GcpClientProvider[gcpiprangeclient.ServiceNetworkingClient],
	computeClientProvider gcpclient.GcpClientProvider[gcpiprangeclient.ComputeClient],
	env abstractions.Environment,
) V3StateFactory {
	return v3.NewStateFactory(serviceNetworkingClientProvider, computeClientProvider, env)
}

// V2StateFactory is an alias for v2.StateFactory to be used by the reconciler.
type V2StateFactory = v2.StateFactory

// NewV2StateFactory is a wrapper for v2.NewStateFactory to be called from controller setup.
func NewV2StateFactory(
	serviceNetworkingClientProvider gcpclient.ClientProvider[gcpiprangeclient.ServiceNetworkingClient],
	oldComputeClientProvider gcpclient.ClientProvider[gcpiprangeclient.OldComputeClient],
	env abstractions.Environment,
) V2StateFactory {
	return v2.NewStateFactory(serviceNetworkingClientProvider, oldComputeClientProvider, env)
}
