package vpcpeering

import (
	"context"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/go-logr/logr"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	awsvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/vpcpeering/client"
	vpcpeeringtypes "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/types"
)

type State struct {
	vpcpeeringtypes.State

	client       awsvpcpeeringclient.Client
	remoteClient awsvpcpeeringclient.Client
	provider     awsclient.SkrClientProvider[awsvpcpeeringclient.Client]

	awsAccessKeyid     string
	awsSecretAccessKey string

	vpc              *ec2types.Vpc
	vpcPeering       *ec2types.VpcPeeringConnection
	remoteVpc        *ec2types.Vpc
	remoteVpcPeering *ec2types.VpcPeeringConnection

	routeTables       []ec2types.RouteTable
	remoteRouteTables []ec2types.RouteTable
}

type StateFactory interface {
	NewState(ctx context.Context, state vpcpeeringtypes.State, logger logr.Logger) (*State, error)
}

func NewStateFactory(skrProvider awsclient.SkrClientProvider[awsvpcpeeringclient.Client]) StateFactory {
	return &stateFactory{
		skrProvider: skrProvider,
	}
}

type stateFactory struct {
	skrProvider awsclient.SkrClientProvider[awsvpcpeeringclient.Client]
}

func (f *stateFactory) NewState(ctx context.Context, vpcPeeringState vpcpeeringtypes.State, logger logr.Logger) (*State, error) {
	awsAccessKeyId := awsconfig.AwsConfig.Peering.AccessKeyId
	awsSecretAccessKey := awsconfig.AwsConfig.Peering.SecretAccessKey

	roleArn := awsutil.RoleArnPeering(vpcPeeringState.Scope().Spec.Scope.Aws.AccountId)

	logger.WithValues(
		"awsRegion", vpcPeeringState.Scope().Spec.Region,
		"awsRole", roleArn,
	).Info("Assuming AWS role")

	c, err := f.skrProvider(
		ctx,
		vpcPeeringState.Scope().Spec.Scope.Aws.AccountId,
		vpcPeeringState.Scope().Spec.Region,
		awsAccessKeyId,
		awsSecretAccessKey,
		roleArn,
	)

	if err != nil {
		return nil, err
	}

	return newState(vpcPeeringState, c, f.skrProvider, awsAccessKeyId, awsSecretAccessKey), nil
}

func newState(vpcPeeringState vpcpeeringtypes.State,
	client awsvpcpeeringclient.Client,
	provider awsclient.SkrClientProvider[awsvpcpeeringclient.Client],
	key string,
	secret string,
) *State {
	return &State{
		State:              vpcPeeringState,
		client:             client,
		provider:           provider,
		awsAccessKeyid:     key,
		awsSecretAccessKey: secret,
	}
}
