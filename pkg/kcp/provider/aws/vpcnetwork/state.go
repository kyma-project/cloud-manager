package vpcnetwork

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	awsvpcnetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/vpcnetwork/client"
	vpcnetworktypes "github.com/kyma-project/cloud-manager/pkg/kcp/vpcnetwork/types"
)

func NewStateFactory(awsClientProvider awsclient.SkrClientProvider[awsvpcnetworkclient.Client]) StateFactory {
	return &stateFactory{
		awsClientProvider: awsClientProvider,
	}
}

type StateFactory interface {
	NewState(ctx context.Context, baseState vpcnetworktypes.State) (context.Context, composed.State, error)
}

var _ StateFactory = (*stateFactory)(nil)

type stateFactory struct {
	awsClientProvider awsclient.SkrClientProvider[awsvpcnetworkclient.Client]
}

func (f *stateFactory) NewState(ctx context.Context, baseState vpcnetworktypes.State) (context.Context, composed.State, error) {
	roleName := awsutil.RoleArnDefault(baseState.Subscription().Status.SubscriptionInfo.Aws.Account)

	logger := composed.LoggerFromCtx(ctx)
	logger = logger.WithValues("awsAssumeRoleName", roleName)
	ctx = composed.LoggerIntoCtx(ctx, logger)

	c, err := f.awsClientProvider(
		ctx,
		baseState.Subscription().Status.SubscriptionInfo.Aws.Account,
		baseState.ObjAsVpcNetwork().Spec.Region,
		awsconfig.AwsConfig.Default.AccessKeyId,
		awsconfig.AwsConfig.Default.SecretAccessKey,
		roleName,
	)
	if err != nil {
		return ctx, nil, err
	}

	return ctx, newState(baseState, c), nil
}

func newState(baseState vpcnetworktypes.State, awsClient awsvpcnetworkclient.Client) *State {
	return &State{
		State: baseState,
		awsClient: awsClient,
	}
}

type State struct {
	vpcnetworktypes.State

	awsClient awsvpcnetworkclient.Client
}
