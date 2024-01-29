package criprange

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func preventCidrChange(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.ObjAsIpRange().Spec.Cidr == state.ObjAsIpRange().Status.Cidr {
		return nil, nil
	}

	return composed.UpdateStatus(state.ObjAsIpRange()).
		SetCondition(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeCidrCanNotChange,
			Status:  "False",
			Reason:  cloudresourcesv1beta1.ConditionReasonCidrChanged,
			Message: "IpRange CIDR can not change",
		}).
		ErrorLogMessage("Error updating IpRange status with CIDR changed condition").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
