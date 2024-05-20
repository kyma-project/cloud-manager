package vpcpeering

import (
	computepb "cloud.google.com/go/compute/apiv1/computepb"
	"context"
	"github.com/go-logr/logr"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/cloudclient"
	vpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/vpcpeering/client"
	vpcpeeringtypes "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/types"
)

type State struct {
	vpcpeeringtypes.State

	client   vpcpeeringclient.Client
	provider gcpclient.SkrClientProvider[vpcpeeringclient.Client]

	saJsonKeyPath string

	vpc                  *computepb.Network
	vpcPeeringConnection *computepb.NetworkPeering
	remoteVpc            *computepb.Network
}

type StateFactory interface {
	NewState(ctx context.Context, state vpcpeeringtypes.State, logger logr.Logger) (*State, error)
}

func NewStateFactory(skrProvider gcpclient.SkrClientProvider[vpcpeeringclient.Client]) StateFactory {
	return &stateFactory{
		skrProvider: skrProvider,
	}
}

type stateFactory struct {
	skrProvider gcpclient.SkrClientProvider[vpcpeeringclient.Client]
}

func (f *stateFactory) NewState(ctx context.Context, vpcPeeringState vpcpeeringtypes.State, logger logr.Logger) (*State, error) {
	//roleName := awsconfig.AwsConfig.AssumeRoleName
	//awsAccessKeyId := awsconfig.AwsConfig.AccessKeyId
	//awsSecretAccessKey := awsconfig.AwsConfig.SecretAccessKey

	//roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", vpcPeeringState.Scope().Spec.Scope.Aws.AccountId, roleName)

	//logger.WithValues(
	//	"awsRegion", vpcPeeringState.Scope().Spec.Region,
	//	"awsRole", roleArn,
	//).Info("Assuming AWS role")

	//c, err := f.skrProvider(
	//	ctx,
	//	vpcPeeringState.Scope().Spec.Region,
	//	awsAccessKeyId,
	//	awsSecretAccessKey,
	//	roleArn,
	//)

	//if err != nil {
	//	return nil, err
	//}
	//
	//return newState(vpcPeeringState, c, f.skrProvider, awsAccessKeyId, awsSecretAccessKey, roleName), nil
	return nil, nil
}

func newState(vpcPeeringState vpcpeeringtypes.State,
	client vpcpeeringclient.Client,
	provider gcpclient.SkrClientProvider[vpcpeeringclient.Client],
	saJsonKeyPath string) *State {
	return &State{
		State:         vpcPeeringState,
		client:        client,
		provider:      provider,
		saJsonKeyPath: saJsonKeyPath,
	}
}
