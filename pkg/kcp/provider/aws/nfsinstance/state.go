package nfsinstance

import (
	"context"
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	efsTypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	nfsinstancetypes "github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	nfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nfsinstance/client"
)

type State struct {
	nfsinstancetypes.State

	awsClient nfsinstanceclient.Client

	efs                       *efsTypes.FileSystemDescription
	mountTargets              []efsTypes.MountTargetDescription
	mountTargetSecurityGroups map[string][]string
	securityGroupId           string
	securityGroup             *ec2Types.SecurityGroup
}

type StateFactory interface {
	NewState(ctx context.Context, nfsInstanceState nfsinstancetypes.State) (*State, error)
}

func NewStateFactory(skrProvider awsclient.SkrClientProvider[nfsinstanceclient.Client]) StateFactory {
	return &stateFactory{
		skrProvider: skrProvider,
	}
}

type stateFactory struct {
	skrProvider awsclient.SkrClientProvider[nfsinstanceclient.Client]
}

func (f *stateFactory) NewState(ctx context.Context, nfsInstanceState nfsinstancetypes.State) (*State, error) {
	roleName := fmt.Sprintf("arn:aws:iam::%s:role/%s", nfsInstanceState.Scope().Spec.Scope.Aws.AccountId, awsconfig.AwsConfig.Default.AssumeRoleName)

	c, err := f.skrProvider(
		ctx,
		nfsInstanceState.Scope().Spec.Region,
		awsconfig.AwsConfig.Default.AccessKeyId,
		awsconfig.AwsConfig.Default.SecretAccessKey,
		roleName,
	)
	if err != nil {
		return nil, err
	}

	return newState(nfsInstanceState, c), nil
}

func newState(nfsInstanceState nfsinstancetypes.State, c nfsinstanceclient.Client) *State {
	return &State{
		State:     nfsInstanceState,
		awsClient: c,
	}
}
