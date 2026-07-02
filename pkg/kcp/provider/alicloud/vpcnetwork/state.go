package vpcnetwork

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/config"
	alicloudvpcnetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/vpcnetwork/client"
	vpcnetworktypes "github.com/kyma-project/cloud-manager/pkg/kcp/vpcnetwork/types"
)

type StateFactory interface {
	NewState(ctx context.Context, baseState vpcnetworktypes.State) (context.Context, *State, error)
}

func NewStateFactory(clientProvider alicloudvpcnetworkclient.ClientProvider) StateFactory {
	return &stateFactory{
		clientProvider: clientProvider,
	}
}

type stateFactory struct {
	clientProvider alicloudvpcnetworkclient.ClientProvider
}

func (f *stateFactory) NewState(ctx context.Context, baseState vpcnetworktypes.State) (context.Context, *State, error) {
	if baseState.Subscription().Status.Provider != cloudcontrolv1beta1.ProviderAlicloud {
		return ctx, nil, fmt.Errorf("subscription for VpcNetwork must be of provider AliCloud, but subscription %q is of provider %q",
			baseState.Subscription().Name, baseState.Subscription().Status.Provider)
	}

	accessKeyId := alicloudconfig.AlicloudConfig.AccessKeyId
	accessKeySecret := alicloudconfig.AlicloudConfig.AccessKeySecret

	region := baseState.ObjAsVpcNetwork().Spec.Region

	logger := composed.LoggerFromCtx(ctx)
	logger = logger.WithValues("alicloudRegion", region)
	ctx = composed.LoggerIntoCtx(ctx, logger)

	c, err := f.clientProvider(ctx, region, accessKeyId, accessKeySecret)
	if err != nil {
		return ctx, nil, fmt.Errorf("error creating alicloud vpc client: %w", err)
	}

	return ctx, newState(baseState, c), nil
}

func newState(baseState vpcnetworktypes.State, alicloudClient alicloudvpcnetworkclient.Client) *State {
	return &State{
		State:          baseState,
		alicloudClient: alicloudClient,
	}
}

type State struct {
	vpcnetworktypes.State

	alicloudClient alicloudvpcnetworkclient.Client
}

func (s *State) IsKymaTypePredicate(ctx context.Context, st composed.State) bool {
	return s.ObjAsVpcNetwork().Spec.Type == cloudcontrolv1beta1.VpcNetworkTypeKyma
}

func (s *State) IsGardenerTypePredicate(ctx context.Context, st composed.State) bool {
	return s.ObjAsVpcNetwork().Spec.Type == cloudcontrolv1beta1.VpcNetworkTypeGardener
}
