package genericActions

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/apis/cloud-resources/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CommonStatus interface {
	UpdateCondition(
		statusState cloudresourcesv1beta1.StatusState,
		conditionType cloudresourcesv1beta1.ConditionType,
		reason cloudresourcesv1beta1.ConditionReason,
		conditionStatus metav1.ConditionStatus,
		message string,
	)

	RemoveCondition(conditionType cloudresourcesv1beta1.ConditionType)

	FindStatusCondition(conditionType cloudresourcesv1beta1.ConditionType) *metav1.Condition
}
