package exposedData

import (
	"context"
	"errors"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
)

func kcpNetworkVerify(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.KcpNetworkKyma() == nil {
		return composed.LogErrorAndReturn(common.LogicalError, "Azure ExposedData must have KCP Network Kym loaded", composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	logger := composed.LoggerFromCtx(ctx)

	readyCond := meta.FindStatusCondition(state.KcpNetworkKyma().Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)
	if readyCond == nil {
		logger.Info("Waiting for KCP Network to be ready - azure exposed data")
		return composed.StopWithRequeue, ctx
	}

	if state.KcpNetworkKyma().Status.Network == nil || state.KcpNetworkKyma().Status.Network.Aws == nil {
		err := errors.New("logical error")
		return composed.LogErrorAndReturn(err, "KCP Kyma network is ready but w/out aws status reference", composed.StopAndForget, ctx)
	}

	if state.KcpNetworkKyma().Status.Network.Aws.NetworkName == "" {
		err := errors.New("logical error")
		return composed.LogErrorAndReturn(err, "KCP Kyma network is ready but has invalid aws status reference vpc name", composed.StopAndForget, ctx)
	}

	state.vpcName = state.KcpNetworkKyma().Status.Network.Aws.NetworkName

	return nil, ctx
}
