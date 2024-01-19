package nfsinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/nfsinstance/types"
	gcpclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/nfsinstance/client"
	"google.golang.org/api/file/v1"
)

type State struct {
	types.State
	curState        v1beta1.StatusState
	operation       focal.OperationType
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
	name := nfsInstance.Spec.RemoteRef.Name
	return s.fsInstance != nil && s.fsInstance.Name == name
}

func (s State) toInstance() *file.Instance {
	nfsInstance := s.ObjAsNfsInstance()

	//GCP Scope Details
	gcpScope := s.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	vpc := gcpScope.VpcNetwork

	return &file.Instance{
		Name:        nfsInstance.Spec.RemoteRef.Name,
		Description: nfsInstance.Name,
		Tier:        "BASIC_HDD",

		FileShares: []*file.FileShareConfig{
			{
				Name:       "vol",
				CapacityGb: 1024,
			},
		},
		Networks: []*file.NetworkConfig{
			{
				Network:         gcpclient.GetVPCPath(project, vpc),
				ReservedIpRange: s.IpRange().Spec.Cidr,
				ConnectMode:     "PRIVATE_SERVICE_ACCESS",
			},
		},
	}
}
