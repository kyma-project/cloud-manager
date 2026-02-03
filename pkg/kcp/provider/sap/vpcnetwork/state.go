package vpcnetwork

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
	sapconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/config"
	sapmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/meta"
	sapvpcnetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/vpcnetwork/client"
	vpcnetworktypes "github.com/kyma-project/cloud-manager/pkg/kcp/vpcnetwork/types"
)

func NewStateFactory(sapClientProvider sapclient.SapClientProvider[sapvpcnetworkclient.Client]) StateFactory {
	return &stateFactory{
		sapClientProvider: sapClientProvider,
	}
}

type StateFactory interface {
	NewState(ctx context.Context, baseState vpcnetworktypes.State) (context.Context, composed.State, error)
}

var _ StateFactory = (*stateFactory)(nil)

type stateFactory struct {
	sapClientProvider sapclient.SapClientProvider[sapvpcnetworkclient.Client]
}

func (f *stateFactory) NewState(ctx context.Context, baseState vpcnetworktypes.State) (context.Context, composed.State, error) {
	if baseState.Subscription().Status.Provider != cloudcontrolv1beta1.ProviderOpenStack {
		return ctx, nil, fmt.Errorf("subscription for VpcNetwork must be of provider OpenStack, but subscription %q is of provider %q", baseState.Subscription().Name, baseState.Subscription().Status.Provider)
	}
	pp := sapclient.NewProviderParamsFromConfig(sapconfig.SapConfig).
		WithDomain(baseState.Subscription().Status.SubscriptionInfo.OpenStack.DomainName).
		WithProject(baseState.Subscription().Status.SubscriptionInfo.OpenStack.TenantName).
		WithRegion(baseState.ObjAsVpcNetwork().Spec.Region)
	sapClient, err := f.sapClientProvider(ctx, pp)
	if err != nil {
		return ctx, nil, fmt.Errorf("error creating SAP client for vpcnetwork: %w", err)
	}

	ctx = sapmeta.SetSapDomainProjectRegion(ctx, pp.DomainName, pp.ProjectName, pp.RegionName)

	return ctx, newState(baseState, sapClient), nil
}

func newState(baseState vpcnetworktypes.State, sapClient sapvpcnetworkclient.Client) *State {
	return &State{
		State:     baseState,
		sapClient: sapClient,
	}
}

type State struct {
	vpcnetworktypes.State

	sapClient sapvpcnetworkclient.Client
}
