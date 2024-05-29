package focal

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CommonObject interface {
	client.Object
	composed.ObjWithConditions

	ScopeRef() cloudcontrolv1beta1.ScopeRef
	SetScopeRef(scopeRef cloudcontrolv1beta1.ScopeRef)
}
