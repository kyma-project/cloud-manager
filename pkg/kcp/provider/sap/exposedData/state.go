package exposedData

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/routers"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
	sapconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/config"
	sapexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/exposedData/client"
	scopetypes "github.com/kyma-project/cloud-manager/pkg/kcp/scope/types"
)

func NewStateFactory(sapProvider sapclient.SapClientProvider[sapexposeddataclient.Client]) StateFactory {
	return &stateFactory{
		sapProvider: sapProvider,
	}
}

type StateFactory interface {
	NewState(ctx context.Context, baseState scopetypes.State) (composed.State, error)
}

var _ StateFactory = &stateFactory{}

func (f *stateFactory) NewState(ctx context.Context, baseState scopetypes.State) (composed.State, error) {
	pp := sapclient.NewProviderParamsFromConfig(sapconfig.SapConfig).
		WithDomain(baseState.ObjAsScope().Spec.Scope.OpenStack.DomainName).
		WithProject(baseState.ObjAsScope().Spec.Scope.OpenStack.TenantName).
		WithRegion(baseState.ObjAsScope().Spec.Region)

	sapClient, err := f.sapProvider(ctx, pp)
	if err != nil {
		return nil, fmt.Errorf("error creating sap client: %w", err)
	}

	return &State{
		State:     baseState,
		sapClient: sapClient,
	}, nil
}

type stateFactory struct {
	sapProvider sapclient.SapClientProvider[sapexposeddataclient.Client]
}

type State struct {
	scopetypes.State

	sapClient sapexposeddataclient.Client

	router *routers.Router
}
