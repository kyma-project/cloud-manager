package iprange

import (
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
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
