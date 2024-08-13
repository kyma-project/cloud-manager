package vpcpeering

import (
	pb "cloud.google.com/go/compute/apiv1/computepb"
	"context"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/cloudclient"
	vpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/vpcpeering/client"
	vpcpeeringtypes "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/types"
)

type State struct {
	vpcpeeringtypes.State

	client   vpcpeeringclient.VpcPeeringClient
	provider gcpclient.ClientProvider[vpcpeeringclient.VpcPeeringClient]

	//gcp config
	gcpConfig *gcpclient.GcpConfig

	remotePeeringName  string
	remoteVpc          string
	remoteProject      string
	importCustomRoutes bool

	//Peerings on both sides
	remoteVpcPeering *pb.NetworkPeering
	kymaVpcPeering   *pb.NetworkPeering
}

type StateFactory interface {
	NewState(ctx context.Context, state vpcpeeringtypes.State, logger logr.Logger) (*State, error)
}

type stateFactory struct {
	skrProvider gcpclient.ClientProvider[vpcpeeringclient.VpcPeeringClient]
	env         abstractions.Environment
	gcpConfig   *gcpclient.GcpConfig
}

func NewStateFactory(skrProvider gcpclient.ClientProvider[vpcpeeringclient.VpcPeeringClient], env abstractions.Environment) StateFactory {
	return &stateFactory{
		skrProvider: skrProvider,
		env:         env,
		gcpConfig:   gcpclient.GetGcpConfig(env),
	}
}

func (f *stateFactory) NewState(ctx context.Context, vpcPeeringState vpcpeeringtypes.State, logger logr.Logger) (*State, error) {
	c, err := f.skrProvider(
		ctx,
		f.env.Get(vpcpeeringclient.GcpVpcPeeringPath),
	)

	if err != nil {
		return nil, err
	}

	return newState(vpcPeeringState, c, f.skrProvider, f.gcpConfig), nil
}

func newState(vpcPeeringState vpcpeeringtypes.State,
	client vpcpeeringclient.VpcPeeringClient,
	provider gcpclient.ClientProvider[vpcpeeringclient.VpcPeeringClient],
	gcpConfig *gcpclient.GcpConfig) *State {
	return &State{
		State:              vpcPeeringState,
		client:             client,
		provider:           provider,
		gcpConfig:          gcpConfig,
		remotePeeringName:  vpcPeeringState.ObjAsVpcPeering().Spec.VpcPeering.Gcp.RemotePeeringName,
		remoteVpc:          vpcPeeringState.ObjAsVpcPeering().Spec.VpcPeering.Gcp.RemoteVpc,
		remoteProject:      vpcPeeringState.ObjAsVpcPeering().Spec.VpcPeering.Gcp.RemoteProject,
		importCustomRoutes: vpcPeeringState.ObjAsVpcPeering().Spec.VpcPeering.Gcp.ImportCustomRoutes,
	}
}

// Using cm- prefix to make it clear it's a cloud manager resource
// Using obj name suffix as the peering name since it is unique within the kcp namespace
func (s *State) getKymaVpcPeeringName() string {
	return "cm-" + s.Obj().GetName()
}
