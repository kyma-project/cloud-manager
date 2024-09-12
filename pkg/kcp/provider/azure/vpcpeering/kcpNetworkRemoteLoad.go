package vpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func kcpNetworkRemoteLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	net := &cloudcontrolv1beta1.Network{}
	namespace := state.ObjAsVpcPeering().Spec.Details.RemoteNetwork.Namespace
	if namespace == "" {
		namespace = state.ObjAsVpcPeering().Namespace
	}

	err := state.Cluster().K8sClient().Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      state.ObjAsVpcPeering().Spec.Details.RemoteNetwork.Name,
	}, net)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading remote KCP Network for KCP VpcPeering", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	state.remoteNetworkId = azureutil.NewVirtualNetworkResourceIdFromNetworkReference(net.Status.Network)
	logger := composed.LoggerFromCtx(ctx)
	logger.WithValues(
		"remoteNetwork", net.Name,
		"remoteNetworkAzureId", state.remoteNetworkId.String(),
	)
	ctx = composed.LoggerIntoCtx(ctx, logger)
	state.remoteNetwork = net

	logger.Info("KCP VpcPeeing remote network loaded")

	return nil, ctx
}
