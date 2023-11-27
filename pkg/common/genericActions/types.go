package genericActions

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/apis/cloud-resources/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ObjWithOutcome interface {
	GetOutcome() *cloudresourcesv1beta1.Outcome
}

type ObjWithStatus interface {
	GetConditions() *[]metav1.Condition
	GetStatusState() cloudresourcesv1beta1.StatusState
	SetStatusState(statusState cloudresourcesv1beta1.StatusState)
}
