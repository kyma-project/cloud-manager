package scope

import (
	"context"
	"errors"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
)

func networkReferenceKymaWaitReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.kcpNetworkKyma == nil {
		err := errors.New("logical error")
		return composed.LogErrorAndReturn(err, "kcpNetworkKyma should not be nil", composed.StopAndForget, ctx)
	}

	readyCond := meta.FindStatusCondition(state.kcpNetworkKyma.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)
	if readyCond != nil {
		return nil, ctx
	}

	logger := composed.LoggerFromCtx(ctx)
	logger.Info("Waiting for KCP Network to be ready")

	return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx
}
