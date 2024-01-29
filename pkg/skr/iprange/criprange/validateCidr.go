package criprange

import (
	"context"
	"fmt"
	"github.com/3th1nk/cidr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validateCidr(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	existing := meta.FindStatusCondition(state.ObjAsIpRange().Status.Conditions, cloudresourcesv1beta1.ConditionTypeCidrValid)
	if existing != nil && existing.Status == metav1.ConditionTrue {
		// already valid
		return nil, nil
	}
	if existing != nil && existing.Status == metav1.ConditionFalse {
		// already not valid
		return composed.StopAndForget, nil
	}

	rng, err := cidr.Parse(state.ObjAsIpRange().Spec.Cidr)
	if err != nil {
		return composed.UpdateStatus(state.ObjAsIpRange()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeCidrValid,
				Status:  metav1.ConditionFalse,
				Reason:  cloudresourcesv1beta1.ConditionReasonCidrInvalidSyntax,
				Message: fmt.Sprintf("CIDR %s has invalid syntax", state.ObjAsIpRange().Spec.Cidr),
			}).
			ErrorLogMessage("Error updating IpRange status with invalid CIDR syntax").
			Run(ctx, state)
	}

	ones, bits := rng.CIDR().Mask.Size()

	if bits != 32 {
		return composed.UpdateStatus(state.ObjAsIpRange()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeCidrValid,
				Status:  metav1.ConditionFalse,
				Reason:  cloudresourcesv1beta1.ConditionReasonCidrInvalidSize,
				Message: fmt.Sprintf("CIDR %s is not IPv4", state.ObjAsIpRange().Spec.Cidr),
			}).
			ErrorLogMessage("Error updating IpRange status with CIDR not an IPv4 condition").
			Run(ctx, state)
	}

	maxOnes := 31
	if ones > maxOnes {
		return composed.UpdateStatus(state.ObjAsIpRange()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeCidrValid,
				Status:  metav1.ConditionFalse,
				Reason:  cloudresourcesv1beta1.ConditionReasonCidrInvalidSize,
				Message: fmt.Sprintf("CIDR %s block size must not be greater than %d", state.ObjAsIpRange().Spec.Cidr, maxOnes),
			}).
			ErrorLogMessage("Error updating IpRange status with invalid CIDR mask").
			Run(ctx, state)
	}

	return composed.UpdateStatus(state.ObjAsIpRange()).
		SetCondition(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeCidrValid,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonCidrValid,
			Message: fmt.Sprintf("CIDR %s is valid", state.ObjAsIpRange().Spec.Cidr),
		}).
		SuccessError(nil).
		ErrorLogMessage("Error updating IpRange status with valid CIDR condition").
		Run(ctx, state)
}
