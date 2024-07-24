package v2

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func validateCidr(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()
	logger.WithValues("ipRange", ipRange.Name).Info("Validating CIDR")
	if len(ipRange.Status.Cidr) == 0 {
		return composed.UpdateStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonInvalidCidr,
				Message: "Cidr is required",
			}).
			SuccessError(composed.StopAndForget).
			SuccessLogMsg("Error updating IpRange status due to missing cidr.").
			Run(ctx, state)
	}

	//Parse CIDR.
	addr, prefix, err := util.CidrParseIPnPrefix(ipRange.Status.Cidr)
	if err != nil {
		return composed.UpdateStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonInvalidCidr,
				Message: err.Error(),
			}).
			SuccessError(composed.StopAndForget).
			SuccessLogMsg("Error updating IpRange status due to cidr overlap.").
			Run(ctx, state)
	}

	//Store the parsed values in the state object.
	state.ipAddress = addr
	state.prefix = prefix

	return nil, nil
}
