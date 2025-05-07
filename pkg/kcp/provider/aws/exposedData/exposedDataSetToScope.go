package exposedData

import (
	"context"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func exposedDataSetToScope(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	var list []string
	for _, gw := range state.natGayteways {
		for _, addr := range gw.NatGatewayAddresses {
			ip := ptr.Deref(addr.PublicIp, "")
			if ip != "" {
				list = append(list, ip)
			}
		}
	}

	state.ObjAsScope().Status.ExposedData.NatGatewayIps = pie.Sort(pie.Unique(list))

	logger := composed.LoggerFromCtx(ctx)
	logger.
		WithValues("natGatewayIps", fmt.Sprintf("%v", state.ObjAsScope().Status.ExposedData.NatGatewayIps)).
		Info("Exposed Data AWS Nat Gateway IP addresses")

	return nil, ctx
}
