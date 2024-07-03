package vpcpeering

import (
	computepb "cloud.google.com/go/compute/apiv1/computepb"
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

	peeringName          *string
	vpcPeeringConnection *computepb.NetworkPeering
	remoteVpc            *string
	remoteProject        *string
	importCustomRoutes   *bool
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
		f.env.Get("GCP_SA_JSON_KEY_PATH"),
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
		peeringName:        &vpcPeeringState.ObjAsVpcPeering().Spec.VpcPeering.Gcp.PeeringName,
		remoteVpc:          &vpcPeeringState.ObjAsVpcPeering().Spec.VpcPeering.Gcp.RemoteVpc,
		remoteProject:      &vpcPeeringState.ObjAsVpcPeering().Spec.VpcPeering.Gcp.RemoteProject,
		importCustomRoutes: &vpcPeeringState.ObjAsVpcPeering().Spec.VpcPeering.Gcp.ImportCustomRoutes,
	}
}
