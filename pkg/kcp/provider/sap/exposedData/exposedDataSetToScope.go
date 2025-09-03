package exposedData

import (
	"context"
	"fmt"

	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func exposedDataSetToScope(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	var list []string
	if state.router != nil {
		for _, ip := range state.router.GatewayInfo.ExternalFixedIPs {
			list = append(list, ip.IPAddress)
		}
		list = pie.Sort(pie.Unique(list))
	}

	state.ObjAsScope().Status.ExposedData.NatGatewayIps = list

	logger.
		WithValues("natGatewayIps", fmt.Sprintf("%v", state.ObjAsScope().Status.ExposedData.NatGatewayIps)).
		Info("Exposed Data SAP Nat Gateway IP addresses")

	return nil, ctx
}

