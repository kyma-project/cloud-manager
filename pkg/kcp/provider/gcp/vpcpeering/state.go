package vpcpeering

import (
	pb "cloud.google.com/go/compute/apiv1/computepb"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/vpcpeering/client"
	vpcpeeringtypes "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/types"
)

type State struct {
	vpcpeeringtypes.State

	client gcpvpcpeeringclient.VpcPeeringClient

	remotePeeringName  string
	importCustomRoutes bool

	//Peerings on both sides
	remoteVpcPeering *pb.NetworkPeering
	kymaVpcPeering   *pb.NetworkPeering
}

type StateFactory interface {
	NewState(state vpcpeeringtypes.State) (*State, error)
}

type stateFactory struct {
	clientProvider gcpclient.GcpClientProvider[gcpvpcpeeringclient.VpcPeeringClient]
	env            abstractions.Environment
}

func NewStateFactory(clientProvider gcpclient.GcpClientProvider[gcpvpcpeeringclient.VpcPeeringClient], env abstractions.Environment) StateFactory {
	return &stateFactory{
		clientProvider: clientProvider,
		env:            env,
	}
}

func (f *stateFactory) NewState(vpcPeeringState vpcpeeringtypes.State) (*State, error) {
	vpcPeeringClient := f.clientProvider(vpcPeeringState.Scope().Spec.Scope.Gcp.Project)
	return newState(vpcPeeringState, vpcPeeringClient), nil
}

func newState(vpcPeeringState vpcpeeringtypes.State,
	client gcpvpcpeeringclient.VpcPeeringClient,
) *State {
	return &State{
		State:              vpcPeeringState,
		client:             client,
		remotePeeringName:  vpcPeeringState.ObjAsVpcPeering().Spec.Details.PeeringName,
		importCustomRoutes: vpcPeeringState.ObjAsVpcPeering().Spec.Details.ImportCustomRoutes,
	}
}

// Using cm- prefix to make it clear it's a cloud manager resource
// Using obj name suffix as the peering name since it is unique within the kcp namespace
func (s *State) getKymaVpcPeeringName() string {
	return "cm-" + s.Obj().GetName()
}
