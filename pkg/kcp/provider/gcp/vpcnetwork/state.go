package vpcnetwork

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpvpcnetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/vpcnetwork/client"
	vpcnetworktypes "github.com/kyma-project/cloud-manager/pkg/kcp/vpcnetwork/types"
)

type StateFactory interface {
	NewState(ctx context.Context, baseState vpcnetworktypes.State) (context.Context, composed.State, error)
}

func NewStateFactory(
	gcpClientProvider gcpclient.GcpClientProvider[gcpvpcnetworkclient.Client],
) StateFactory {
	return &stateFactory{
		gcpClientProvider: gcpClientProvider,
	}
}

func (f *stateFactory) NewState(ctx context.Context, baseState vpcnetworktypes.State) (context.Context, composed.State, error) {
	clnt := f.gcpClientProvider(baseState.Subscription().Status.SubscriptionInfo.Gcp.Project)
	return ctx, newState(baseState, clnt), nil
}

type stateFactory struct {
	gcpClientProvider gcpclient.GcpClientProvider[gcpvpcnetworkclient.Client]
}

func newState(baseState vpcnetworktypes.State, gcpClient gcpvpcnetworkclient.Client) *State {
	return &State{
		State:     baseState,
		gcpClient: gcpClient,
	}
}

type State struct {
	vpcnetworktypes.State

	gcpClient gcpvpcnetworkclient.Client
}
