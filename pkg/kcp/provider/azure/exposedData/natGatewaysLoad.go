package exposedData

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"k8s.io/utils/ptr"
)

func natGatewaysLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	for _, subnet := range state.subnets {
		if subnet.Properties.NatGateway == nil {
			continue
		}
		rd, err := azureutil.ParseResourceID(ptr.Deref(subnet.Properties.NatGateway.ID, ""))
		if err != nil {
			logger.Error(err, "Error parsing NAT Gateway ID")
			continue
		}

		natGateway, err := state.azureClient.GetNatGateway(ctx, rd.ResourceGroup, rd.ResourceName)
		if err != nil {
			return azuremeta.LogErrorAndReturn(err, "Error loading NAT Gateway", ctx)
		}

		state.natGateways = append(state.natGateways, natGateway)
	}

	return nil, ctx
}
