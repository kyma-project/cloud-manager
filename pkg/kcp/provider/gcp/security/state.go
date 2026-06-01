package security

import (
	"context"

	"cloud.google.com/go/securitycentermanagement/apiv1/securitycentermanagementpb"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	securityclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/security/client"
	runtimetypes "github.com/kyma-project/cloud-manager/pkg/kcp/runtime/types"
)

type StateFactory interface {
	NewState(ctx context.Context, runtimeState runtimetypes.State) (context.Context, composed.State, error)
}

func NewStateFactory(gcpClientProvider gcpclient.GcpClientProvider[securityclient.Client]) StateFactory {
	return &stateFactory{
		gcpClientProvider: gcpClientProvider,
	}
}

type stateFactory struct {
	gcpClientProvider gcpclient.GcpClientProvider[securityclient.Client]
}

func (f *stateFactory) NewState(ctx context.Context, runtimeState runtimetypes.State) (context.Context, composed.State, error) {
	project := runtimeState.Subscription().Status.SubscriptionInfo.Gcp.Project
	clnt := f.gcpClientProvider(project)

	logger := composed.LoggerFromCtx(ctx).WithValues("gcpProjectId", project)
	ctx = composed.LoggerIntoCtx(ctx, logger)

	return ctx, newState(runtimeState, clnt), nil
}

func newState(runtimeState runtimetypes.State, gcpClient securityclient.Client) *State {
	return &State{
		State:     runtimeState,
		gcpClient: gcpClient,
	}
}

type State struct {
	runtimetypes.State

	gcpClient securityclient.Client

	// sccServices keyed by service ID; nil until sccServicesLoad runs
	sccServices map[string]*securitycentermanagementpb.SecurityCenterService
}
