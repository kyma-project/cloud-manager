package iprange

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/abstractions"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/provider/aws/client"
)

type State struct {
	focal.State

	networkClient client.NetworkClient
	efsClient     client.EfsClient
}

type StateFactory interface {
	NewState(ctx context.Context, focalState focal.State) (*State, error)
}

func NewStateFactory(skrProvider client.SkrProvider, env abstractions.Environment) StateFactory {
	return &stateFactory{
		skrProvider: skrProvider,
		env:         env,
	}
}

type stateFactory struct {
	skrProvider client.SkrProvider
	env         abstractions.Environment
}

func (f *stateFactory) NewState(ctx context.Context, focalState focal.State) (*State, error) {
	roleName := fmt.Sprintf("arn:aws:iam::%s:role/%s", focalState.Scope().Spec.Scope.Aws.AccountId, f.env.Get("AWS_ROLE_NAME"))

	networkClient, err := f.skrProvider.Network()(
		ctx,
		focalState.Scope().Spec.Region,
		f.env.Get("AWS_ACCESS_KEY_ID"),
		f.env.Get("AWS_SECRET_ACCESS_KEY"),
		roleName,
	)
	if err != nil {
		return nil, err
	}

	efsClient, err := f.skrProvider.Efs()(
		ctx,
		focalState.Scope().Spec.Region,
		f.env.Get("AWS_ACCESS_KEY_ID"),
		f.env.Get("AWS_SECRET_ACCESS_KEY"),
		roleName,
	)
	if err != nil {
		return nil, err
	}

	return newState(focalState, networkClient, efsClient), nil
}

func newState(focalState focal.State, networkClient client.NetworkClient, efsClient client.EfsClient) *State {
	return &State{
		State:         focalState,
		networkClient: networkClient,
		efsClient:     efsClient,
	}
}
