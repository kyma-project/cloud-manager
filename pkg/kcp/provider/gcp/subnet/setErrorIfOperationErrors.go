package subnet

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func setErrorIfOperationErrors(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.subnetCreationOperation == nil {
		return nil, nil
	}

	if state.subnetCreationOperation.Error == nil {
		return nil, nil
	}

	subnet := state.ObjAsGcpSubnet()

	for _, err := range state.subnetCreationOperation.Error.Errors {
		logger.Error(fmt.Errorf("%s", err.GetMessage()), fmt.Sprintf("[%s] %s", err.GetCode(), err.GetMessage()))
	}

	hasCondErr := meta.FindStatusCondition(subnet.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError) != nil
	if !hasCondErr {
		meta.SetStatusCondition(subnet.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
			Message: "KCP GcpSubnet creation operation failed",
		})
		subnet.Status.State = cloudcontrolv1beta1.StateError

		return composed.UpdateStatus(subnet).
			SuccessError(composed.StopAndForget).
			SuccessLogMsg("successfully updated GcpSubnet status to error state due to creation operation failure").
			ErrorLogMessage("failed to update GcpSubnet status to error").
			Run(ctx, st)
	}

	return composed.StopAndForget, nil
}
