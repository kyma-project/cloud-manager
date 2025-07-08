package exposedData

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"k8s.io/utils/ptr"
)

func publicIpAddressesLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	for _, natGateway := range state.natGateways {
		for _, ipRef := range natGateway.Properties.PublicIPAddresses {
			rd, err := azureutil.ParseResourceID(ptr.Deref(ipRef.ID, ""))
			if err != nil {
				logger.Error(err, "Error parsing public ip address id")
				continue
			}

			ip, err := state.azureClient.GetPublicIpAddress(ctx, rd.ResourceGroup, rd.ResourceName)
			if err != nil {
				return azuremeta.LogErrorAndReturn(err, "Error loading public ip address", ctx)
			}

			state.publicIPAddresses = append(state.publicIPAddresses, ip)
		}
	}

	return nil, ctx
}
