package vpcpeering

import (
	"context"
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	vpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/vpcpeering/client"
	vpcpeeringtypes "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/types"
)

type State struct {
	vpcpeeringtypes.State

	client       vpcpeeringclient.Client
	remoteClient vpcpeeringclient.Client

	// TODO state
	vpc       *ec2Types.Vpc
	remoteVpc *ec2Types.Vpc
}

type StateFactory interface {
	NewState(ctx context.Context, state vpcpeeringtypes.State, logger logr.Logger) (*State, error)
}

func NewStateFactory(skrProvider awsclient.SkrClientProvider[vpcpeeringclient.Client], env abstractions.Environment) *stateFactory {
	return &stateFactory{
		skrProvider: skrProvider,
		env:         env,
	}
}

type stateFactory struct {
	skrProvider awsclient.SkrClientProvider[vpcpeeringclient.Client]
	env         abstractions.Environment
}

func (f *stateFactory) NewState(ctx context.Context, vpcPeeringState vpcpeeringtypes.State, logger logr.Logger) (*State, error) {

	roleName := fmt.Sprintf("arn:aws:iam::%s:role/%s", vpcPeeringState.Scope().Spec.Scope.Aws.AccountId, f.env.Get("AWS_ROLE_NAME"))

	logger.WithValues(
		"awsRegion", vpcPeeringState.Scope().Spec.Region,
		"awsRole", roleName,
	).Info("Assuming AWS role")

	c, err := f.skrProvider(
		ctx,
		vpcPeeringState.Scope().Spec.Region,
		f.env.Get("AWS_ACCESS_KEY_ID"),
		f.env.Get("AWS_SECRET_ACCESS_KEY"),
		roleName,
	)

	if err != nil {
		return nil, err
	}

	return newState(vpcPeeringState, c), nil
}

func newState(vpcPeeringState vpcpeeringtypes.State, c vpcpeeringclient.Client) *State {
	return &State{
		State:  vpcPeeringState,
		client: c,
	}
}
