package nfsinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/nfsinstance/types"
	gcpclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/nfsinstance/client"
	"google.golang.org/api/file/v1"
)

type State struct {
	types.State
	curState        v1beta1.StatusState
	operation       gcpclient.OperationType
	updateMask      []string
	fsInstance      *file.Instance
	filestoreClient client.FilestoreClient
}

type StateFactory interface {
	NewState(ctx context.Context, nfsInstanceState types.State) (*State, error)
}

type stateFactory struct {
	filestoreClientProvider gcpclient.ClientProvider[client.FilestoreClient]
	env                     abstractions.Environment
}

func NewStateFactory(filestoreClientProvider gcpclient.ClientProvider[client.FilestoreClient], env abstractions.Environment) StateFactory {
	return &stateFactory{
		filestoreClientProvider: filestoreClientProvider,
		env:                     env,
	}
}

func (f *stateFactory) NewState(ctx context.Context, nfsInstanceState types.State) (*State, error) {
	httpClient, err := gcpclient.GetCachedGcpClient(ctx, f.env.Get("GCP_SA_JSON_KEY_PATH"))
	if err != nil {
		return nil, err
	}
	fc, err := f.filestoreClientProvider(
		ctx,
		httpClient,
	)
	if err != nil {
		return nil, err
	}
	return newState(nfsInstanceState, fc), nil
}

func newState(nfsInstanceState types.State, fc client.FilestoreClient) *State {
	return &State{
		State:           nfsInstanceState,
		filestoreClient: fc,
	}
}

func (s State) doesFilestoreMatch() bool {
	nfsInstance := s.ObjAsNfsInstance()
	return s.fsInstance != nil &&
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
				Name:       gcpOptions.FileShareName,
				CapacityGb: int64(gcpOptions.CapacityGb),
			},
		},
		Networks: []*file.NetworkConfig{
			{
				Network:     gcpclient.GetVPCPath(project, vpc),
				ConnectMode: string(gcpOptions.ConnectMode),
			},
		},
	}
}
