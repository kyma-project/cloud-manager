package v3

import (
	"context"
	"strings"

	"cloud.google.com/go/compute/apiv1/computepb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	"google.golang.org/api/servicenetworking/v1"
)

// State is the GCP-specific state for IpRange that extends the shared iprange state.
// It follows the multi-provider pattern (RedisInstance style) where:
//   - composed.State (base K8s operations)
//   - focal.State (adds Scope management)
//   - iprangetypes.State (shared IpRange state for all providers)
//   - iprange.State (GCP-specific state with GCP clients and resources)
type State struct {
	iprangetypes.State // Extends shared iprange state (which extends focal.State)

	// GCP API clients
	serviceNetworkingClient gcpiprangeclient.ServiceNetworkingClient
	computeClient           gcpiprangeclient.ComputeClient

	// GCP-specific remote resources
	address           *computepb.Address            // Global address resource
	serviceConnection *servicenetworking.Connection // PSA connection
	operation         interface{}                   // Can be compute or servicenetworking operation

	// GCP-specific state tracking
	inSync          bool                            // Whether remote state matches desired state
	ipAddress       string                          // IP address from CIDR allocation
	prefix          int                             // Prefix length from CIDR allocation
	peeringIpRanges []string                        // IP ranges used in PSA peering
	addressOp       gcpclient.OperationType         // Address operation state
	connectionOp    gcpclient.OperationType         // PSA connection operation state
	curState        cloudcontrolv1beta1.StatusState // Current reconciliation state
}

// Ensure State implements iprangetypes.State interface
var _ iprangetypes.State = &State{}

// StateFactory creates GCP-specific IpRange state from shared iprange state.
type StateFactory interface {
	NewState(ctx context.Context, ipRangeState iprangetypes.State) (*State, error)
}

type stateFactory struct {
	serviceNetworkingClientProvider gcpclient.GcpClientProvider[gcpiprangeclient.ServiceNetworkingClient]
	computeClientProvider           gcpclient.GcpClientProvider[gcpiprangeclient.ComputeClient]
}

// NewStateFactory creates a new GCP IpRange state factory.
// Follows NEW pattern with GcpClientProvider for compute, OLD pattern for service networking.
func NewStateFactory(
	serviceNetworkingClientProvider gcpclient.GcpClientProvider[gcpiprangeclient.ServiceNetworkingClient],
	computeClientProvider gcpclient.GcpClientProvider[gcpiprangeclient.ComputeClient],
) StateFactory {
	return &stateFactory{
		serviceNetworkingClientProvider: serviceNetworkingClientProvider,
		computeClientProvider:           computeClientProvider,
	}
}

// NewState creates a new GCP-specific state by extending the shared iprange state.
// This follows the multi-provider pattern where provider-specific state wraps shared state.
func (f *stateFactory) NewState(ctx context.Context, ipRangeState iprangetypes.State) (*State, error) {
	// Get GCP clients using provider functions
	serviceNetworkingClient := f.serviceNetworkingClientProvider()
	computeClient := f.computeClientProvider()

	return &State{
		State:                   ipRangeState,
		serviceNetworkingClient: serviceNetworkingClient,
		computeClient:           computeClient,
	}, nil
}

// ObjAsGcpIpRange is a convenience method for type assertion (following GCP pattern).
func (s *State) ObjAsGcpIpRange() *cloudcontrolv1beta1.IpRange {
	return s.ObjAsIpRange()
}

// GCP-specific state methods

// DoesAddressMatch checks if the existing GCP address matches the desired configuration.
func (s *State) DoesAddressMatch() bool {
	vpc := s.Scope().Spec.Scope.Gcp.VpcNetwork
	return s.address != nil &&
		s.address.Address != nil && *s.address.Address == s.ipAddress &&
		s.address.PrefixLength != nil && *s.address.PrefixLength == int32(s.prefix) &&
		s.address.Network != nil && strings.HasSuffix(*s.address.Network, vpc)
}

// DoesConnectionIncludeRange checks if the PSA connection includes the address range.
// Returns the index of the range in the connection, or -1 if not found.
func (s *State) DoesConnectionIncludeRange() int {
	if s.serviceConnection == nil {
		return -1
	}
	if s.address == nil {
		return -1
	}

	for i, name := range s.serviceConnection.ReservedPeeringRanges {
		if s.address.Name != nil && *s.address.Name == name {
			return i
		}
	}
	return -1
}

// DoesConnectionMatchPeeringRanges checks if the PSA connection's reserved ranges
// match the desired peering IP ranges. Returns true if they match (no update needed).
func (s *State) DoesConnectionMatchPeeringRanges() bool {
	if s.serviceConnection == nil {
		return false
	}

	// Check if lengths match
	if len(s.serviceConnection.ReservedPeeringRanges) != len(s.peeringIpRanges) {
		return false
	}

	// Create a map of existing ranges for quick lookup
	existingRanges := make(map[string]bool)
	for _, name := range s.serviceConnection.ReservedPeeringRanges {
		existingRanges[name] = true
	}

	// Check if all desired ranges are in the connection
	for _, name := range s.peeringIpRanges {
		if !existingRanges[name] {
			return false
		}
	}

	return true
}

// Getters and setters for GCP-specific fields

func (s *State) InSync() bool {
	return s.inSync
}

func (s *State) SetInSync(v bool) {
	s.inSync = v
}

func (s *State) IpAddress() string {
	return s.ipAddress
}

func (s *State) SetIpAddress(v string) {
	s.ipAddress = v
}

func (s *State) Prefix() int {
	return s.prefix
}

func (s *State) SetPrefix(v int) {
	s.prefix = v
}

func (s *State) PeeringIpRanges() []string {
	return s.peeringIpRanges
}

func (s *State) SetPeeringIpRanges(v []string) {
	s.peeringIpRanges = v
}

func (s *State) Address() *computepb.Address {
	return s.address
}

func (s *State) SetAddress(v *computepb.Address) {
	s.address = v
}

func (s *State) ServiceConnection() *servicenetworking.Connection {
	return s.serviceConnection
}

func (s *State) SetServiceConnection(v *servicenetworking.Connection) {
	s.serviceConnection = v
}

func (s *State) Operation() interface{} {
	return s.operation
}

func (s *State) SetOperation(v interface{}) {
	s.operation = v
}

func (s *State) AddressOp() gcpclient.OperationType {
	return s.addressOp
}

func (s *State) SetAddressOp(v gcpclient.OperationType) {
	s.addressOp = v
}

func (s *State) ConnectionOp() gcpclient.OperationType {
	return s.connectionOp
}

func (s *State) SetConnectionOp(v gcpclient.OperationType) {
	s.connectionOp = v
}

func (s *State) CurState() cloudcontrolv1beta1.StatusState {
	return s.curState
}

func (s *State) SetCurState(v cloudcontrolv1beta1.StatusState) {
	s.curState = v
}

func (s *State) ComputeClient() gcpiprangeclient.ComputeClient {
	return s.computeClient
}

func (s *State) ServiceNetworkingClient() gcpiprangeclient.ServiceNetworkingClient {
	return s.serviceNetworkingClient
}
