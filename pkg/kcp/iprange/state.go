package iprange

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// State is the shared implementation of types.State interface.
// It contains fields and logic common to all cloud providers.
type State struct {
	focal.State

	// CIDR range tracking
	existingCidrRanges []string

	// Network references
	networkKey            client.ObjectKey
	isCloudManagerNetwork bool
	isKymaNetwork         bool
	network               *cloudcontrolv1beta1.Network

	// Kyma network and peering (multi-cluster)
	kymaNetwork *cloudcontrolv1beta1.Network
	kymaPeering *cloudcontrolv1beta1.VpcPeering
}

// Ensure State implements types.State interface
var _ types.State = &State{}

func (s *State) ObjAsIpRange() *cloudcontrolv1beta1.IpRange {
	return s.Obj().(*cloudcontrolv1beta1.IpRange)
}

// CIDR range methods

func (s *State) ExistingCidrRanges() []string {
	return s.existingCidrRanges
}

func (s *State) SetExistingCidrRanges(v []string) {
	s.existingCidrRanges = v
}

// Network methods

func (s *State) Network() *cloudcontrolv1beta1.Network {
	return s.network
}

func (s *State) SetNetwork(network *cloudcontrolv1beta1.Network) {
	s.network = network
}

func (s *State) NetworkKey() client.ObjectKey {
	return s.networkKey
}

func (s *State) SetNetworkKey(key client.ObjectKey) {
	s.networkKey = key
}

// Network type flag methods

func (s *State) IsCloudManagerNetwork() bool {
	return s.isCloudManagerNetwork
}

func (s *State) SetIsCloudManagerNetwork(v bool) {
	s.isCloudManagerNetwork = v
}

func (s *State) IsKymaNetwork() bool {
	return s.isKymaNetwork
}

func (s *State) SetIsKymaNetwork(v bool) {
	s.isKymaNetwork = v
}

// Kyma network and peering methods

func (s *State) KymaNetwork() *cloudcontrolv1beta1.Network {
	return s.kymaNetwork
}

func (s *State) SetKymaNetwork(network *cloudcontrolv1beta1.Network) {
	s.kymaNetwork = network
}

func (s *State) KymaPeering() *cloudcontrolv1beta1.VpcPeering {
	return s.kymaPeering
}

func (s *State) SetKymaPeering(peering *cloudcontrolv1beta1.VpcPeering) {
	s.kymaPeering = peering
}

// newState creates a new shared IpRange state from a focal state.
// This is used by the shared reconciler before provider-specific state is created.
func newState(focalState focal.State) types.State {
	return &State{State: focalState}
}
