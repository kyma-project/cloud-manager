package types

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// State is the shared interface for IpRange state that all provider-specific states extend.
// It provides common functionality needed by all cloud providers.
type State interface {
	focal.State
	ObjAsIpRange() *cloudcontrolv1beta1.IpRange

	// CIDR range management
	ExistingCidrRanges() []string
	SetExistingCidrRanges(v []string)

	// Network references
	Network() *cloudcontrolv1beta1.Network
	SetNetwork(network *cloudcontrolv1beta1.Network)
	NetworkKey() client.ObjectKey
	SetNetworkKey(key client.ObjectKey)

	// Network type flags
	IsCloudManagerNetwork() bool
	SetIsCloudManagerNetwork(v bool)
	IsKymaNetwork() bool
	SetIsKymaNetwork(v bool)

	// Kyma network and peering (used in multi-cluster scenarios)
	KymaNetwork() *cloudcontrolv1beta1.Network
	SetKymaNetwork(network *cloudcontrolv1beta1.Network)
	KymaPeering() *cloudcontrolv1beta1.VpcPeering
	SetKymaPeering(peering *cloudcontrolv1beta1.VpcPeering)
}
