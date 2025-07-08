package statewithscope

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

type StateWithObjAsScope interface {
	ObjAsScope() *cloudcontrolv1beta1.Scope
}

func ScopeFromState(st composed.State) (*cloudcontrolv1beta1.Scope, bool) {
	if state, ok := st.(focal.State); ok {
		return state.Scope(), true
	}
	if state, ok := st.(StateWithObjAsScope); ok {
		return state.ObjAsScope(), true
	}
	if scope, ok := st.Obj().(*cloudcontrolv1beta1.Scope); ok {
		return scope, true
	}

	return nil, false
}
