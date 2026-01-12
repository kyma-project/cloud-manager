// Package v2 provides the modern, streamlined GCP NfsInstance implementation.
//
// This package follows the OLD reconciler pattern (multi-provider CRD) but with
// improved organization and maintainability. It is designed to replace the v1
// implementation when enabled via the gcpNfsInstanceV2 feature flag.
//
// Key improvements over v1:
// - Modern GCP client (cloud.google.com/go/filestore) with protobuf types
// - Better code organization (operations, validation, state packages)
// - Cleaner state management
// - Improved validation logic
// - Enhanced client abstraction
// - Focused testing on business logic
// - Consistent error handling
//
// Architecture:
// - State hierarchy: focal.State → types.State → v2.State (OLD pattern)
// - Action composition: Sequential execution with explicit flow control
// - Client: Modern FilestoreClient using cloud.google.com/go/filestore
// - Operations: Separate packages for CRUD, validation, state management
package v2

import (
	"context"
	"fmt"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance/types"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	v2client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/v2/client"
)

// State represents the GCP-specific state for NfsInstance reconciliation.
// It extends the shared types.State (OLD pattern) with GCP-specific fields.
// Uses modern GCP protobuf types from cloud.google.com/go/filestore.
type State struct {
	types.State // Embedded shared state (OLD pattern compatibility)

	// Client for GCP Filestore operations
	filestoreClient v2client.FilestoreClient

	// Cached GCP resources (using modern protobuf types)
	instance *filestorepb.Instance // Current Filestore instance from GCP
}

// StateFactory creates State instances for reconciliation.
// This interface supports dependency injection and testing.
type StateFactory interface {
	NewState(ctx context.Context, nfsInstanceState types.State) (*State, error)
}

// stateFactory is the default implementation of StateFactory.
type stateFactory struct {
	filestoreClientProvider gcpclient.GcpClientProvider[v2client.FilestoreClient]
	env                     abstractions.Environment
}

// NewStateFactory creates a new StateFactory.
// Follows the NEW pattern for GCP client initialization (GcpClientProvider).
func NewStateFactory(
	filestoreClientProvider gcpclient.GcpClientProvider[v2client.FilestoreClient],
	env abstractions.Environment,
) StateFactory {
	return &stateFactory{
		filestoreClientProvider: filestoreClientProvider,
		env:                     env,
	}
}

// NewState creates a new State instance with initialized GCP client.
// This method is called at the start of each reconciliation loop.
// Uses NEW pattern client initialization (no credentials file parameter needed).
func (f *stateFactory) NewState(ctx context.Context, nfsInstanceState types.State) (*State, error) {
	filestoreClient := f.filestoreClientProvider()

	return &State{
		State:           nfsInstanceState,
		filestoreClient: filestoreClient,
	}, nil
}

// ============================================================================
// State Query Methods
// ============================================================================

// GetFilestoreClient returns the GCP Filestore client.
func (s *State) GetFilestoreClient() v2client.FilestoreClient {
	return s.filestoreClient
}

// GetInstance returns the cached GCP Filestore instance.
func (s *State) GetInstance() *filestorepb.Instance {
	return s.instance
}

// SetInstance caches the GCP Filestore instance.
func (s *State) SetInstance(instance *filestorepb.Instance) {
	s.instance = instance
}

// ============================================================================
// Helper Methods - GCP Resource Mapping
// ============================================================================

// GetGcpProjectId returns the GCP project ID from the Scope.
func (s *State) GetGcpProjectId() string {
	return s.Scope().Spec.Scope.Gcp.Project
}

// GetGcpLocation returns the GCP location (zone) for the Filestore instance.
// Falls back to the region from Scope if not specified in NfsInstance.
func (s *State) GetGcpLocation() string {
	nfsInstance := s.ObjAsNfsInstance()
	location := nfsInstance.Spec.Instance.Gcp.Location
	if location == "" {
		location = s.Scope().Spec.Region
	}
	return location
}

// GetGcpVpcNetwork returns the VPC network name from the Scope.
func (s *State) GetGcpVpcNetwork() string {
	return s.Scope().Spec.Scope.Gcp.VpcNetwork
}

// GetGcpInstanceId returns the Filestore instance ID.
// This is the name used in GCP APIs.
func (s *State) GetGcpInstanceId() string {
	return s.ObjAsNfsInstance().Name
}

// ============================================================================
// Helper Methods - State Comparison
// ============================================================================

// DoesFilestoreMatch returns true if the cached GCP instance matches the desired spec.
// Currently only checks capacity, can be extended to check other fields.
func (s *State) DoesFilestoreMatch() bool {
	if s.instance == nil || len(s.instance.FileShares) == 0 {
		return false
	}

	nfsInstance := s.ObjAsNfsInstance()
	desiredCapacityGb := int64(nfsInstance.Spec.Instance.Gcp.CapacityGb)
	actualCapacityGb := s.instance.FileShares[0].CapacityGb

	return actualCapacityGb == desiredCapacityGb
}

// NeedsUpdate returns true if the Filestore instance needs to be updated.
func (s *State) NeedsUpdate() bool {
	return !s.DoesFilestoreMatch()
}

// ============================================================================
// Helper Methods - GCP API Conversion
// ============================================================================

// ToGcpInstance converts the NfsInstance CRD spec to a GCP Filestore Instance.
// Uses modern protobuf types from cloud.google.com/go/filestore.
func (s *State) ToGcpInstance() *filestorepb.Instance {
	nfsInstance := s.ObjAsNfsInstance()
	gcpOptions := nfsInstance.Spec.Instance.Gcp

	// Collect GCP details from Scope
	project := s.GetGcpProjectId()
	vpc := s.GetGcpVpcNetwork()

	return &filestorepb.Instance{
		Description: nfsInstance.Name,
		Tier:        convertTier(gcpOptions.Tier),
		FileShares: []*filestorepb.FileShareConfig{
			{
				Name:       gcpOptions.FileShareName,
				CapacityGb: int64(gcpOptions.CapacityGb),
				// SourceBackup is not yet supported in v2 - will be added when needed
			},
		},
		Networks: []*filestorepb.NetworkConfig{
			{
				Network:         fmt.Sprintf("projects/%s/global/networks/%s", project, vpc),
				Modes:           []filestorepb.NetworkConfig_AddressMode{filestorepb.NetworkConfig_MODE_IPV4},
				ReservedIpRange: s.IpRange().Status.Id,
				ConnectMode:     filestorepb.NetworkConfig_PRIVATE_SERVICE_ACCESS,
			},
		},
	}
}
