package iprange

import (
	"context"
	"fmt"
	"github.com/3th1nk/cidr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validateCidr(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if len(state.ObjAsIpRange().Status.Cidr) > 0 {
		logger.Info("Cidr already checked and valid")
		return nil, nil
	}

	rng, err := cidr.Parse(state.ObjAsIpRange().Spec.Cidr)
	if err != nil {
		return composed.UpdateStatus(state.ObjAsIpRange()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionFalse,
				Reason:  cloudresourcesv1beta1.ConditionReasonInvalidCidr,
				Message: fmt.Sprintf("CIDR %s has invalid syntax", state.ObjAsIpRange().Spec.Cidr),
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error updating IpRange status with invalid CIDR syntax").
			SuccessLogMsg("Forgetting IpRange with invalid Cidr syntax").
			Run(ctx, state)
	}

	ones, bits := rng.CIDR().Mask.Size()

	if bits != 32 {
		return composed.UpdateStatus(state.ObjAsIpRange()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionFalse,
				Reason:  cloudresourcesv1beta1.ConditionReasonInvalidCidr,
				Message: fmt.Sprintf("CIDR %s is not IPv4", state.ObjAsIpRange().Spec.Cidr),
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error updating IpRange status with CIDR not an IPv4 condition").
			SuccessLogMsg("Forgetting IpRange with invalid non IPv4 Cidr").
			Run(ctx, state)
	}

	maxOnes := 30
	if ones > maxOnes {
		return composed.UpdateStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonInvalidCidr,
				Message: fmt.Sprintf("CIDR %s block size must not be greater than %d", state.ObjAsIpRange().Spec.Cidr, maxOnes),
			}).
			ErrorLogMessage("Error updating IpRange status with too big CIDR mask").
			SuccessLogMsg("Forgetting IpRange with too big Cidr mask").
			Run(ctx, state)
	}

	minOnes := 16
	if ones < minOnes {
		return composed.UpdateStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonInvalidCidr,
				Message: fmt.Sprintf("CIDR %s block size must not be less than %d", state.ObjAsIpRange().Spec.Cidr, minOnes),
			}).
			ErrorLogMessage("Error updating IpRange status with too small CIDR mask").
			SuccessLogMsg("Forgetting IpRange with too small Cidr mask").
			Run(ctx, state)
	}

	state.ObjAsIpRange().Status.Cidr = state.ObjAsIpRange().Spec.Cidr
	err = state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "error updating IpRange status after cidr successful validation", composed.StopWithRequeue, ctx)
	}
	logger.Info("SKR IpRange Cidr validated")

	return nil, nil
}
