package iprange

import (
	"context"
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/kcp/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/abstractions"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/provider/aws/client"
)

// State is state of the IpRange reconciliation for AWS
// Though there's no need for it to be public, still made like that
// to avoid conflict with variable names - DO NOT LOWERCASE IT TO MAKE IT PRIVATE
type State struct {
	focal.State

	networkClient client.NetworkClient
	efsClient     client.EfsClient

	vpc                  *ec2Types.Vpc
	associatedCidrBlock  *ec2Types.VpcCidrBlockAssociation
	allSubnets           []ec2Types.Subnet
	cloudResourceSubnets []ec2Types.Subnet
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

// ==================================================================

func (s *State) IpRange() *cloudresourcesv1beta1.IpRange {
	return s.Obj().(*cloudresourcesv1beta1.IpRange)
}
