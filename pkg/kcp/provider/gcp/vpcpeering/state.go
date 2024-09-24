package vpcpeering

import (
	pb "cloud.google.com/go/compute/apiv1/computepb"
	"context"
	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/cloudclient"
	vpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/vpcpeering/client"
	vpcpeeringtypes "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/types"
)

type State struct {
	vpcpeeringtypes.State

	client   vpcpeeringclient.VpcPeeringClient
	provider gcpclient.ClientProvider[vpcpeeringclient.VpcPeeringClient]

	localNetwork  *cloudcontrolv1beta1.Network
	remoteNetwork *cloudcontrolv1beta1.Network

	remotePeeringName  string
	importCustomRoutes bool

	//Peerings on both sides
	remoteVpcPeering *pb.NetworkPeering
	kymaVpcPeering   *pb.NetworkPeering
}

type StateFactory interface {
	NewState(ctx context.Context, state vpcpeeringtypes.State, logger logr.Logger) (*State, error)
}

type stateFactory struct {
	clientProvider gcpclient.ClientProvider[vpcpeeringclient.VpcPeeringClient]
	env            abstractions.Environment
}

func NewStateFactory(clientProvider gcpclient.ClientProvider[vpcpeeringclient.VpcPeeringClient], env abstractions.Environment) StateFactory {
	return &stateFactory{
		clientProvider: clientProvider,
		env:            env,
	}
}

func (f *stateFactory) NewState(ctx context.Context, vpcPeeringState vpcpeeringtypes.State, logger logr.Logger) (*State, error) {
	c, err := f.clientProvider(
		ctx,
		f.env.Get(vpcpeeringclient.GcpVpcPeeringPath),
	)

	if err != nil {
		return nil, err
	}

	return newState(vpcPeeringState, c, f.clientProvider), nil
}

func newState(vpcPeeringState vpcpeeringtypes.State,
	client vpcpeeringclient.VpcPeeringClient,
	provider gcpclient.ClientProvider[vpcpeeringclient.VpcPeeringClient],
) *State {
	return &State{
		State:              vpcPeeringState,
		client:             client,
		provider:           provider,
		remotePeeringName:  vpcPeeringState.ObjAsVpcPeering().Spec.Details.PeeringName,
		importCustomRoutes: vpcPeeringState.ObjAsVpcPeering().Spec.Details.ImportCustomRoutes,
	}
}

// Using cm- prefix to make it clear it's a cloud manager resource
// Using obj name suffix as the peering name since it is unique within the kcp namespace
func (s *State) getKymaVpcPeeringName() string {
	return "cm--" + s.Obj().GetName()
}
