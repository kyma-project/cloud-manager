package exposedData

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/types"
)

func kcpNetworkLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	net := &cloudcontrolv1beta1.Network{}
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.ObjAsScope().Namespace,
		Name:      common.KcpNetworkKymaCommonName(state.ObjAsScope().Name),
	}, net)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading KCP Kyma Network for Azure expose cloud data", composed.StopWithRequeue, ctx)
	}

	state.kcpNetwork = net

	return nil, ctx
}
