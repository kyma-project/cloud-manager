package nfsinstance

import (
	"context"
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	efsTypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/abstractions"
	nfsinstancetypes "github.com/kyma-project/cloud-manager/components/kcp/pkg/nfsinstance/types"
	awsclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/aws/client"
	nfsinstanceclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/aws/nfsinstance/client"
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

func NewStateFactory(skrProvider awsclient.SkrClientProvider[nfsinstanceclient.Client], env abstractions.Environment) StateFactory {
	return &stateFactory{
		skrProvider: skrProvider,
		env:         env,
	}
}

type stateFactory struct {
	skrProvider awsclient.SkrClientProvider[nfsinstanceclient.Client]
	env         abstractions.Environment
}

func (f *stateFactory) NewState(ctx context.Context, nfsInstanceState nfsinstancetypes.State) (*State, error) {
	roleName := fmt.Sprintf("arn:aws:iam::%s:role/%s", nfsInstanceState.Scope().Spec.Scope.Aws.AccountId, f.env.Get("AWS_ROLE_NAME"))

	c, err := f.skrProvider(
		ctx,
		nfsInstanceState.Scope().Spec.Region,
		f.env.Get("AWS_ACCESS_KEY_ID"),
		f.env.Get("AWS_SECRET_ACCESS_KEY"),
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
