package iprange

import (
	"context"
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/iprange/types"
	awsclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/aws/client"
	iprangeclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/aws/iprange/client"
)

// State is state of the IpRange reconciliation for AWS
// Though there's no need for it to be public, still made like that
// to avoid conflict with variable names - DO NOT LOWERCASE IT TO MAKE IT PRIVATE
type State struct {
	types.State

	client iprangeclient.Client

	vpc                  *ec2Types.Vpc
	associatedCidrBlock  *ec2Types.VpcCidrBlockAssociation
	allSubnets           []ec2Types.Subnet
	cloudResourceSubnets []ec2Types.Subnet
}

type StateFactory interface {
	NewState(ctx context.Context, ipRangeState types.State) (*State, error)
}

func NewStateFactory(skrProvider awsclient.SkrClientProvider[iprangeclient.Client], env abstractions.Environment) StateFactory {
	return &stateFactory{
		skrProvider: skrProvider,
		env:         env,
	}
}

type stateFactory struct {
	skrProvider awsclient.SkrClientProvider[iprangeclient.Client]
	env         abstractions.Environment
}

func (f *stateFactory) NewState(ctx context.Context, ipRangeState types.State) (*State, error) {
	roleName := fmt.Sprintf("arn:aws:iam::%s:role/%s", ipRangeState.Scope().Spec.Scope.Aws.AccountId, f.env.Get("AWS_ROLE_NAME"))

	c, err := f.skrProvider(
		ctx,
		ipRangeState.Scope().Spec.Region,
		f.env.Get("AWS_ACCESS_KEY_ID"),
		f.env.Get("AWS_SECRET_ACCESS_KEY"),
		roleName,
	)
	if err != nil {
		return nil, err
	}

	return newState(ipRangeState, c), nil
}

func newState(ipRangeState types.State, c iprangeclient.Client) *State {
	return &State{
		State:  ipRangeState,
		client: c,
	}
}
