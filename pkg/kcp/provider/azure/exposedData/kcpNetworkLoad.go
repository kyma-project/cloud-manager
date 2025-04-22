package exposedData

import (
	"context"
	"errors"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
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

	logger := composed.LoggerFromCtx(ctx)

	readyCond := meta.FindStatusCondition(net.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)
	if readyCond == nil {
		logger.Info("Waiting for KCP Network to be ready - azure expose")
		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx
	}
	if net.Status.Network == nil || net.Status.Network.Azure == nil {
		err := errors.New("logical error")
		return composed.LogErrorAndReturn(err, "KCP Kyma network is ready but w/out azure status reference", composed.StopAndForget, ctx)
	}
	state.networkId = azureutil.NewVirtualNetworkResourceIdFromNetworkReference(net.Status.Network)

	if !state.networkId.IsValid() {
		err := errors.New("logical error")
		return composed.LogErrorAndReturn(err, "KCP Kyma network is ready but has invalid azure status reference id", composed.StopAndForget, ctx)
	}

	state.kcpNetwork = net

	return nil, ctx
}
