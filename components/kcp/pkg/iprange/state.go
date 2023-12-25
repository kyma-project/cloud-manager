package iprange

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/kcp/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/actions/focal"
)

type State interface {
	focal.State
	IpRange() *cloudresourcesv1beta1.IpRange
}

type state struct {
	focal.State
}

func (s *state) IpRange() *cloudresourcesv1beta1.IpRange {
	return s.Obj().(*cloudresourcesv1beta1.IpRange)
}

func newState(scopeState focal.State) State {
	return &state{State: scopeState}
}
