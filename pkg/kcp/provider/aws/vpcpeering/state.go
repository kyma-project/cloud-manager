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

	client   vpcpeeringclient.Client
	provider awsclient.SkrClientProvider[vpcpeeringclient.Client]

	awsAccessKeyid     string
	awsSecretAccessKey string
	roleName           string

	vpc                  *ec2Types.Vpc
	vpcPeeringConnection *ec2Types.VpcPeeringConnection
	remoteVpc            *ec2Types.Vpc
	remoteRegion         string
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

	roleName := f.env.Get("AWS_ROLE_NAME")
	awsAccessKeyId := f.env.Get("AWS_ACCESS_KEY_ID")
	awsSecretAccessKey := f.env.Get("AWS_SECRET_ACCESS_KEY")

	roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", vpcPeeringState.Scope().Spec.Scope.Aws.AccountId, roleName)

	logger.WithValues(
		"awsRegion", vpcPeeringState.Scope().Spec.Region,
		"awsRole", roleArn,
	).Info("Assuming AWS role")

	c, err := f.skrProvider(
		ctx,
		vpcPeeringState.Scope().Spec.Region,
		awsAccessKeyId,
		awsSecretAccessKey,
		roleArn,
	)

	if err != nil {
		return nil, err
	}

	return newState(vpcPeeringState, c, f.skrProvider, awsAccessKeyId, awsSecretAccessKey, roleName), nil
}

func newState(vpcPeeringState vpcpeeringtypes.State,
	client vpcpeeringclient.Client,
	provider awsclient.SkrClientProvider[vpcpeeringclient.Client],
	key string,
	secret string,
	roleName string) *State {
	return &State{
		State:              vpcPeeringState,
		client:             client,
		provider:           provider,
		awsAccessKeyid:     key,
		awsSecretAccessKey: secret,
		roleName:           roleName,
	}
}
