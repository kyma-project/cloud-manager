// Package v2 provides the modern, streamlined GCP NfsInstance implementation.
//
// This package follows the OLD reconciler pattern (multi-provider CRD) but with
// improved organization and maintainability. It is designed to replace the v1
// implementation when enabled via the gcpNfsInstanceV2 feature flag.
//
// Key improvements over v1:
// - Better code organization (operations, validation, state packages)
// - Cleaner state management
// - Improved validation logic
// - Enhanced client abstraction
// - Focused testing on business logic
// - Consistent error handling
//
// Architecture:
// - State hierarchy: focal.State → types.State → v2.State
// - Action composition: Sequential execution with explicit flow control
// - Client: Abstracted FilestoreClient interface
// - Operations: Separate packages for CRUD, validation, state management
package v2

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance/types"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	v2client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/v2/client"
	"google.golang.org/api/file/v1"
)

// OperationType represents the type of operation to perform on a GCP Filestore instance.
type OperationType int

const (
	// OpNone indicates no operation is needed.
	OpNone OperationType = iota
	// OpAdd indicates a new instance should be created.
	OpAdd
	// OpModify indicates an existing instance should be updated.
	OpModify
	// OpDelete indicates an instance should be deleted.
	OpDelete
)

// String returns a string representation of the operation type.
func (o OperationType) String() string {
	switch o {
	case OpNone:
		return "NONE"
	case OpAdd:
		return "ADD"
	case OpModify:
		return "MODIFY"
	case OpDelete:
		return "DELETE"
	default:
		return "UNKNOWN"
	}
}

// State represents the GCP-specific state for NfsInstance reconciliation.
// It extends the shared types.State (OLD pattern) with GCP-specific fields.
type State struct {
	types.State

	// client is the GCP Filestore client for API operations.
	client v2client.FilestoreClient

	// gcpInstance is the cached GCP Filestore instance loaded from the API.
	gcpInstance *file.Instance

	// pendingOp tracks an in-progress GCP operation.
	pendingOp *file.Operation

	// opType indicates what operation needs to be performed.
	opType OperationType

	// currentState tracks the current lifecycle state of the resource.
	currentState v1beta1.StatusState

	// updateMask contains the list of fields that need to be updated.
	updateMask []string
}

// StateFactory creates new State instances for GCP NfsInstance reconciliation.
type StateFactory interface {
	// NewState creates a new State from the shared types.State.
	// Returns an error if the GCP client cannot be initialized.
	NewState(ctx context.Context, nfsInstanceState types.State) (*State, error)
}

// stateFactory implements StateFactory.
type stateFactory struct {
	filestoreClientProvider gcpclient.ClientProvider[v2client.FilestoreClient]
	env                     abstractions.Environment
}

// NewStateFactory creates a new StateFactory.
func NewStateFactory(
	filestoreClientProvider gcpclient.ClientProvider[v2client.FilestoreClient],
	env abstractions.Environment,
) StateFactory {
	return &stateFactory{
		filestoreClientProvider: filestoreClientProvider,
		env:                     env,
	}
}

// NewState creates a new State instance.
func (f *stateFactory) NewState(ctx context.Context, nfsInstanceState types.State) (*State, error) {
	fc, err := f.filestoreClientProvider(
		ctx,
		config.GcpConfig.CredentialsFile,
	)
	if err != nil {
		return nil, err
	}

	return newState(nfsInstanceState, fc), nil
}

// newState creates a new State with the provided client.
func newState(nfsInstanceState types.State, fc v2client.FilestoreClient) *State {
	return &State{
		State:  nfsInstanceState,
		client: fc,
		opType: OpNone,
	}
}

// Helper methods for State

// GetGcpLocation returns the GCP location (region/zone) for the instance.
// Falls back to the Scope region if not explicitly specified.
func (s *State) GetGcpLocation() string {
	nfsInstance := s.ObjAsNfsInstance()
	location := nfsInstance.Spec.Instance.Gcp.Location
	if location == "" {
		location = s.Scope().Spec.Region
	}
	return location
}

// GetInstanceName returns the name to use for the GCP Filestore instance.
func (s *State) GetInstanceName() string {
	return s.ObjAsNfsInstance().Name
}

// GetProjectId returns the GCP project ID from the Scope.
func (s *State) GetProjectId() string {
	return s.Scope().Spec.Scope.Gcp.Project
}

// GetVpcNetwork returns the VPC network from the Scope.
func (s *State) GetVpcNetwork() string {
	return s.Scope().Spec.Scope.Gcp.VpcNetwork
}

// DoesInstanceMatch checks if the cached GCP instance matches the desired state.
func (s *State) DoesInstanceMatch() bool {
	if s.gcpInstance == nil || len(s.gcpInstance.FileShares) == 0 {
		return false
	}

	nfsInstance := s.ObjAsNfsInstance()
	desiredCapacity := int64(nfsInstance.Spec.Instance.Gcp.CapacityGb)
	actualCapacity := s.gcpInstance.FileShares[0].CapacityGb

	return actualCapacity == desiredCapacity
}

// ToGcpInstance converts the NfsInstance CRD spec to a GCP Filestore Instance.
func (s *State) ToGcpInstance() *file.Instance {
	nfsInstance := s.ObjAsNfsInstance()
	gcpOptions := nfsInstance.Spec.Instance.Gcp

	// Collect GCP details from Scope
	project := s.GetProjectId()
	vpc := s.GetVpcNetwork()

	return &file.Instance{
		Description: nfsInstance.Name,
		Tier:        string(gcpOptions.Tier),
		FileShares: []*file.FileShareConfig{
			{
				Name:         gcpOptions.FileShareName,
				CapacityGb:   int64(gcpOptions.CapacityGb),
				SourceBackup: gcpOptions.SourceBackup,
			},
		},
		Networks: []*file.NetworkConfig{
			{
				Network:         fmt.Sprintf("projects/%s/global/networks/%s", project, vpc),
				Modes:           []string{"MODE_IPV4"},
				ReservedIpRange: s.IpRange().Status.Cidr,
			},
		},
	}
}

// SetOperationType sets the operation type for this reconciliation cycle.
func (s *State) SetOperationType(opType OperationType) {
	s.opType = opType
}

// GetOperationType returns the current operation type.
func (s *State) GetOperationType() OperationType {
	return s.opType
}

// SetCurrentState sets the current lifecycle state.
func (s *State) SetCurrentState(state v1beta1.StatusState) {
	s.currentState = state
}

// GetCurrentState returns the current lifecycle state.
func (s *State) GetCurrentState() v1beta1.StatusState {
	return s.currentState
}

// SetGcpInstance caches the GCP instance loaded from the API.
func (s *State) SetGcpInstance(instance *file.Instance) {
	s.gcpInstance = instance
}

// GetGcpInstance returns the cached GCP instance.
func (s *State) GetGcpInstance() *file.Instance {
	return s.gcpInstance
}

// SetPendingOperation sets the pending GCP operation.
func (s *State) SetPendingOperation(op *file.Operation) {
	s.pendingOp = op
}

// GetPendingOperation returns the pending GCP operation.
func (s *State) GetPendingOperation() *file.Operation {
	return s.pendingOp
}

// HasPendingOperation returns true if there is a pending operation.
func (s *State) HasPendingOperation() bool {
	return s.pendingOp != nil && !s.pendingOp.Done
}

// SetUpdateMask sets the list of fields to update.
func (s *State) SetUpdateMask(mask []string) {
	s.updateMask = mask
}

// GetUpdateMask returns the list of fields to update.
func (s *State) GetUpdateMask() []string {
	return s.updateMask
}

// HasUpdateMask returns true if there are fields to update.
func (s *State) HasUpdateMask() bool {
	return len(s.updateMask) > 0
}

// GetClient returns the GCP Filestore client.
func (s *State) GetClient() v2client.FilestoreClient {
	return s.client
}
