package focal

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CommonObject interface {
	client.Object
	composed.ObjWithConditions

	ScopeRef() cloudresourcesv1beta1.ScopeRef
	SetScopeRef(scopeRef cloudresourcesv1beta1.ScopeRef)
}
