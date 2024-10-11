package iprange

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func kcpNetworkInit(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	// defaults to Kyma Shoot network if not specified in IpRange
	// since before KCP IpRange got network property in spec
	// the ipranges were created in the kyma shoot network
	state.networkKey = client.ObjectKey{
		Name:      common.KcpNetworkKymaCommonName(state.Scope().Name),
		Namespace: state.Scope().Namespace,
	}
	if state.ObjAsIpRange().Spec.Network != nil {
		state.networkKey.Name = state.ObjAsIpRange().Spec.Network.Name
	}

	state.isCloudManagerNetwork = common.IsKcpNetworkCM(state.networkKey.Name, state.Scope().Name)
	state.isKymaNetwork = common.IsKcpNetworkKyma(state.networkKey.Name, state.Scope().Name)

	logger := composed.LoggerFromCtx(ctx).
		WithValues(
			"kcpNetwork", state.networkKey.String(),
			"isCloudManagerNetwork", state.isCloudManagerNetwork,
			"isKymaNetwork", state.isKymaNetwork,
		)
	ctx = composed.LoggerIntoCtx(ctx, logger)

	return nil, ctx
}
