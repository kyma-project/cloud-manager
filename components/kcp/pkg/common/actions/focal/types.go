package focal

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/kcp/api/cloud-resources/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CommonObject interface {
	client.Object

	KymaName() string

	ScopeRef() *cloudresourcesv1beta1.ScopeRef
	SetScopeRef(scopeRef *cloudresourcesv1beta1.ScopeRef)

	Conditions() *[]metav1.Condition
}
