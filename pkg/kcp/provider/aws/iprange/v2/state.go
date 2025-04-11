package v2

import (
	"context"
	"fmt"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/go-logr/logr"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	awsiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/iprange/client"
)

type State struct {
	iprangetypes.State

	awsClient awsiprangeclient.Client

	vpc                  *ec2types.Vpc
	associatedCidrBlock  *ec2types.VpcCidrBlockAssociation
	allSubnets           []ec2types.Subnet
	cloudResourceSubnets []ec2types.Subnet
}

type StateFactory interface {
	NewState(ctx context.Context, ipRangeState iprangetypes.State, logger logr.Logger) (*State, error)
}

func NewStateFactory(skrProvider awsclient.SkrClientProvider[awsiprangeclient.Client]) StateFactory {
	return &stateFactory{
		skrProvider: skrProvider,
	}
}

type stateFactory struct {
	skrProvider awsclient.SkrClientProvider[awsiprangeclient.Client]
}

func (f *stateFactory) NewState(ctx context.Context, ipRangeState iprangetypes.State, logger logr.Logger) (*State, error) {
	roleName := fmt.Sprintf("arn:aws:iam::%s:role/%s", ipRangeState.Scope().Spec.Scope.Aws.AccountId, awsconfig.AwsConfig.Default.AssumeRoleName)

	logger.
		WithValues(
			"awsRegion", ipRangeState.Scope().Spec.Region,
			"awsRole", roleName,
		).
		Info("Assuming AWS role")

	c, err := f.skrProvider(
		ctx,
		ipRangeState.Scope().Spec.Scope.Aws.AccountId,
		ipRangeState.Scope().Spec.Region,
		awsconfig.AwsConfig.Default.AccessKeyId,
		awsconfig.AwsConfig.Default.SecretAccessKey,
		roleName,
	)
	if err != nil {
		return nil, err
	}

	return newState(ipRangeState, c), nil
}

func newState(ipRangeState iprangetypes.State, c awsiprangeclient.Client) *State {
	return &State{
		State:     ipRangeState,
		awsClient: c,
	}
}
