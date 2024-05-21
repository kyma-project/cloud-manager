package v1

import (
	"context"
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/go-logr/logr"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	iprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/iprange/client"
)

// State is state of the IpRange reconciliation for AWS
// Though there's no need for it to be public, still made like that
// to avoid conflict with variable names - DO NOT LOWERCASE IT TO MAKE IT PRIVATE
type State struct {
	iprangetypes.State

	client iprangeclient.Client

	vpc                  *ec2Types.Vpc
	associatedCidrBlock  *ec2Types.VpcCidrBlockAssociation
	allSubnets           []ec2Types.Subnet
	cloudResourceSubnets []ec2Types.Subnet
}

type StateFactory interface {
	NewState(ctx context.Context, ipRangeState iprangetypes.State, logger logr.Logger) (*State, error)
}

func NewStateFactory(skrProvider awsclient.SkrClientProvider[iprangeclient.Client]) StateFactory {
	return &stateFactory{
		skrProvider: skrProvider,
	}
}

type stateFactory struct {
	skrProvider awsclient.SkrClientProvider[iprangeclient.Client]
}

func (f *stateFactory) NewState(ctx context.Context, ipRangeState iprangetypes.State, logger logr.Logger) (*State, error) {
	roleName := fmt.Sprintf("arn:aws:iam::%s:role/%s", ipRangeState.Scope().Spec.Scope.Aws.AccountId, awsconfig.AwsConfig.AssumeRoleName)

	logger.
		WithValues(
			"awsRegion", ipRangeState.Scope().Spec.Region,
			"awsRole", roleName,
		).
		Info("Assuming AWS role")

	c, err := f.skrProvider(
		ctx,
		ipRangeState.Scope().Spec.Region,
		awsconfig.AwsConfig.AccessKeyId,
		awsconfig.AwsConfig.SecretAccessKey,
		roleName,
	)
	if err != nil {
		return nil, err
	}

	return newState(ipRangeState, c), nil
}

func newState(ipRangeState iprangetypes.State, c iprangeclient.Client) *State {
	return &State{
		State:  ipRangeState,
		client: c,
	}
}
