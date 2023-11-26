package genericActions

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/apis/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func UpdateCondition(
	statusState cloudresourcesv1beta1.StatusState,
	conditionType cloudresourcesv1beta1.ConditionType,
	reason cloudresourcesv1beta1.ConditionReason,
	conditionStatus metav1.ConditionStatus,
	message string,
) composed.Action {
	return func(ctx context.Context, state composed.State) error {
		commonStatus := state.Obj().(CommonStatus)
		commonStatus.UpdateCondition(
			statusState,
			conditionType,
			reason,
			conditionStatus,
			message,
		)
		return nil
	}
}
