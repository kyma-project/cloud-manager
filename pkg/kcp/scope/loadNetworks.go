package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadNetworks(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	netList := &cloudcontrolv1beta1.NetworkList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(cloudcontrolv1beta1.NetworkFieldScope, state.kyma.GetName()),
	}
	if err := state.Cluster().K8sClient().List(ctx, netList, listOps); err != nil {
		return composed.LogErrorAndReturn(err, "Error listing scope networks", composed.StopWithRequeue, ctx)
	}

	logger := composed.LoggerFromCtx(ctx)

	kymaNetworkName := common.KcpNetworkKymaCommonName(state.kyma.GetName())
	cmNetworkName := common.KcpNetworkCMCommonName(state.kyma.GetName())

	for _, net := range netList.Items {
		netCopy := net
		if net.Name == kymaNetworkName {
			state.kcpNetworkKyma = &netCopy
			logger.
				WithValues("kcpNetworkKyma", state.kcpNetworkKyma.Name).
				Info("Kyma network found")
		}
		if net.Name == cmNetworkName {
			state.kcpNetworkCm = &netCopy
			logger.
				WithValues("kcpNetworkCm", state.kcpNetworkCm.Name).
				Info("CM network found")
		}
	}

	return nil, nil
}
