package nfsinstance

import (
	"context"
	"fmt"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	nfsinstancetypes "github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	awsnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nfsinstance/client"
)

type State struct {
	nfsinstancetypes.State

	awsClient awsnfsinstanceclient.Client

	efs                       *efstypes.FileSystemDescription
	mountTargets              []efstypes.MountTargetDescription
	mountTargetSecurityGroups map[string][]string
	securityGroupId           string
	securityGroup             *ec2types.SecurityGroup
}

type StateFactory interface {
	NewState(ctx context.Context, nfsInstanceState nfsinstancetypes.State) (*State, error)
}

func NewStateFactory(skrProvider awsclient.SkrClientProvider[awsnfsinstanceclient.Client]) StateFactory {
	return &stateFactory{
		skrProvider: skrProvider,
	}
}

type stateFactory struct {
	skrProvider awsclient.SkrClientProvider[awsnfsinstanceclient.Client]
}

func (f *stateFactory) NewState(ctx context.Context, nfsInstanceState nfsinstancetypes.State) (*State, error) {
	roleName := fmt.Sprintf("arn:aws:iam::%s:role/%s", nfsInstanceState.Scope().Spec.Scope.Aws.AccountId, awsconfig.AwsConfig.Default.AssumeRoleName)

	c, err := f.skrProvider(
		ctx,
		nfsInstanceState.Scope().Spec.Scope.Aws.AccountId,
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

func newState(nfsInstanceState nfsinstancetypes.State, c awsnfsinstanceclient.Client) *State {
	return &State{
		State:     nfsInstanceState,
		awsClient: c,
	}
}

func stopAndRequeueForCapacity() error {
	return composed.StopWithRequeueDelay(awsconfig.AwsConfig.EfsCapacityCheckInterval)
}

func StopAndRequeueForCapacityAction() composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		return stopAndRequeueForCapacity(), nil
	}
}
