package sapnfsvolumesnapshot

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
	sapconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/config"
)

func clientCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	pp := sapclient.NewProviderParamsFromConfig(sapconfig.SapConfig).
		WithDomain(state.Scope.Spec.Scope.OpenStack.DomainName).
		WithProject(state.Scope.Spec.Scope.OpenStack.TenantName).
		WithRegion(state.Scope.Spec.Region)

	client, err := state.provider(ctx, pp)
	if err != nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("error creating SAP snapshot client: %w", err),
			"Error creating SAP snapshot client",
			composed.StopWithRequeue,
			ctx,
		)
	}

	state.snapshotClient = client

	return nil, ctx
}
