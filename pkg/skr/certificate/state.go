package certificate

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	certificatetypes "github.com/kyma-project/cloud-manager/pkg/skr/certificate/types"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
)

// State is the shared implementation of certificatetypes.State interface.
// It wraps commonscope.State and provides Certificate-specific functionality.
type State struct {
	commonscope.State
}

var _ certificatetypes.State = &State{}

func (s *State) ObjAsCertificate() *cloudresourcesv1beta1.Certificate {
	return s.Obj().(*cloudresourcesv1beta1.Certificate)
}

func newState(scopeState commonscope.State) certificatetypes.State {
	return &State{State: scopeState}
}
