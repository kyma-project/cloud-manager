package nfsinstance

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance/types"
	client2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/client"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"google.golang.org/api/file/v1"
)

type State struct {
	types.State
	curState        v1beta1.StatusState
	operation       client2.OperationType
	updateMask      []string
	validations     []string
	fsInstance      *file.Instance
	filestoreClient client.FilestoreClient
}

type StateFactory interface {
	NewState(ctx context.Context, nfsInstanceState types.State) (*State, error)
}

type stateFactory struct {
	filestoreClientProvider client2.ClientProvider[client.FilestoreClient]
	env                     abstractions.Environment
}

func NewStateFactory(filestoreClientProvider client2.ClientProvider[client.FilestoreClient], env abstractions.Environment) StateFactory {
	return &stateFactory{
		filestoreClientProvider: filestoreClientProvider,
		env:                     env,
	}
}

func (f *stateFactory) NewState(ctx context.Context, nfsInstanceState types.State) (*State, error) {
	httpClient, err := client2.GetCachedGcpClient(ctx, f.env.Get("GCP_SA_JSON_KEY_PATH"))
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
				Network:         client2.GetVPCPath(project, vpc),
				ConnectMode:     string(gcpOptions.ConnectMode),
				ReservedIpRange: s.IpRange().Spec.RemoteRef.Name,
			},
		},
	}
}
