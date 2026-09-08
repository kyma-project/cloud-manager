package wafpolicy

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	wafpolicytypes "github.com/kyma-project/cloud-manager/pkg/skr/wafpolicy/types"
)

// State is the shared implementation of wafpolicytypes.State interface.
// It wraps commonscope.State and provides WafPolicy-specific functionality.
type State struct {
	commonscope.State
}

// Ensure State implements wafpolicytypes.State
var _ wafpolicytypes.State = &State{}

func (s *State) ObjAsWafPolicy() *cloudresourcesv1beta1.WafPolicy {
	return s.Obj().(*cloudresourcesv1beta1.WafPolicy)
}

// newState creates a new shared WafPolicy state from a commonscope state.
func newState(scopeState commonscope.State) wafpolicytypes.State {
	return &State{State: scopeState}
}
