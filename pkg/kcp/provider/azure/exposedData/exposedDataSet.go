package exposedData

import (
	"context"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func exposedDataSet(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	var list []string
	for _, ip := range state.publicIPAddresses {
		addr := ptr.Deref(ip.Properties.IPAddress, "")
		if addr != "" {
			list = append(list, addr)
		}
	}

	state.ExposedData().NatGatewayIps = pie.Sort(pie.Unique(list))

	logger := composed.LoggerFromCtx(ctx)
	logger.
		WithValues("natGatewayIps", fmt.Sprintf("%v", state.ExposedData().NatGatewayIps)).
		Info("Exposed Data Azure Nat Gateway IP addresses")

	return nil, ctx
}
