package exposedData

import (
	"context"
	"fmt"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	awsexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/exposedData/client"
	scopetypes "github.com/kyma-project/cloud-manager/pkg/kcp/scope/types"
)

func NewStateFactory(awsClientProvider awsclient.SkrClientProvider[awsexposeddataclient.Client]) StateFactory {
	return &stateFactory{
		awsClientProvider: awsClientProvider,
	}
}

type StateFactory interface {
	NewState(ctx context.Context, baseState scopetypes.State) (context.Context, composed.State, error)
}

var _ StateFactory = &stateFactory{}

type stateFactory struct {
	awsClientProvider awsclient.SkrClientProvider[awsexposeddataclient.Client]
}

func (f *stateFactory) NewState(ctx context.Context, baseState scopetypes.State) (context.Context, composed.State, error) {
	roleName := fmt.Sprintf("arn:aws:iam::%s:role/%s", baseState.ObjAsScope().Spec.Scope.Aws.AccountId, awsconfig.AwsConfig.Default.AssumeRoleName)

	logger := composed.LoggerFromCtx(ctx)
	logger = logger.WithValues("awsAssumeRoleName", roleName)
	ctx = composed.LoggerIntoCtx(ctx, logger)

	c, err := f.awsClientProvider(
		ctx,
		baseState.ObjAsScope().Spec.Scope.Aws.AccountId,
		baseState.ObjAsScope().Spec.Region,
		awsconfig.AwsConfig.Default.AccessKeyId,
		awsconfig.AwsConfig.Default.SecretAccessKey,
		roleName,
	)
	if err != nil {
		return ctx, nil, err
	}

	return ctx, newState(baseState, c), nil
}

func newState(baseState scopetypes.State, awsClient awsexposeddataclient.Client) *State {
	return &State{
		State:     baseState,
		awsClient: awsClient,
	}
}

type State struct {
	scopetypes.State

	awsClient awsexposeddataclient.Client

	vpcName      string
	vpc          *ec2types.Vpc
	natGayteways []*ec2types.NatGateway
}
