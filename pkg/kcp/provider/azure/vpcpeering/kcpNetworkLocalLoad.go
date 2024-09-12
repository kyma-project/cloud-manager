package vpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func kcpNetworkLocalLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	net := &cloudcontrolv1beta1.Network{}
	namespace := state.ObjAsVpcPeering().Spec.Details.LocalNetwork.Namespace
	if namespace == "" {
		namespace = state.ObjAsVpcPeering().Namespace
	}

	err := state.Cluster().K8sClient().Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      state.ObjAsVpcPeering().Spec.Details.LocalNetwork.Name,
	}, net)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading local KCP Network for KCP VpcPeering", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	state.localNetworkId = azureutil.NewVirtualNetworkResourceIdFromNetworkReference(net.Status.Network)
	logger := composed.LoggerFromCtx(ctx)
	logger.WithValues(
		"localNetwork", net.Name,
		"localNetworkAzureId", state.localNetworkId.String(),
	)
	ctx = composed.LoggerIntoCtx(ctx, logger)
	state.localNetwork = net

	logger.Info("KCP VpcPeeing local network loaded")

	return nil, ctx
}
