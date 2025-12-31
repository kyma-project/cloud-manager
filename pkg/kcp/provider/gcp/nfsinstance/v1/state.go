// Package v1 provides the legacy GCP NfsInstance implementation.
//
// Deprecated: This package is maintained for backward compatibility only.
// New code should use the v2 package when available via the gcpNfsInstanceV2 feature flag.
// This implementation uses the OLD reconciler pattern and will be removed in a future release.
package v1

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance/types"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	v1client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/v1/client"
	"google.golang.org/api/file/v1"
)

type State struct {
	types.State
	curState        v1beta1.StatusState
	operation       gcpclient.OperationType
	updateMask      []string
	validations     []string
	fsInstance      *file.Instance
	filestoreClient v1client.FilestoreClient
}

type StateFactory interface {
	NewState(ctx context.Context, nfsInstanceState types.State) (*State, error)
}

type stateFactory struct {
	filestoreClientProvider gcpclient.ClientProvider[v1client.FilestoreClient]
	env                     abstractions.Environment
}

func NewStateFactory(filestoreClientProvider gcpclient.ClientProvider[v1client.FilestoreClient], env abstractions.Environment) StateFactory {
	return &stateFactory{
		filestoreClientProvider: filestoreClientProvider,
		env:                     env,
	}
}

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

func newState(nfsInstanceState types.State, fc v1client.FilestoreClient) *State {
	return &State{
		State:           nfsInstanceState,
		filestoreClient: fc,
	}
}

func (s State) doesFilestoreMatch() bool {
	nfsInstance := s.ObjAsNfsInstance()
	return s.fsInstance != nil && len(s.fsInstance.FileShares) > 0 &&
		s.fsInstance.FileShares[0].CapacityGb == int64(nfsInstance.Spec.Instance.Gcp.CapacityGb)
}

func (s State) getGcpLocation() string {
	nfsInstance := s.ObjAsNfsInstance()
	gcpOptions := nfsInstance.Spec.Instance.Gcp
	location := gcpOptions.Location
	if location == "" {
		location = s.Scope().Spec.Region
	}
	return location
}

func (s State) toInstance() *file.Instance {
	nfsInstance := s.ObjAsNfsInstance()
	gcpOptions := nfsInstance.Spec.Instance.Gcp

	//Collect GCP Details
	gcpScope := s.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	vpc := gcpScope.VpcNetwork

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
				Network:         gcpclient.GetVPCPath(project, vpc),
				ConnectMode:     string(gcpOptions.ConnectMode),
				ReservedIpRange: s.IpRange().Status.Id,
			},
		},
	}
}
