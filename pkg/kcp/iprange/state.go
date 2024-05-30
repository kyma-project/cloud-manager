package iprange

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
)

type State struct {
	focal.State
}

func (s *State) ObjAsIpRange() *cloudcontrolv1beta1.IpRange {
	return s.Obj().(*cloudcontrolv1beta1.IpRange)
}

func newState(focalState focal.State) types.State {
	return &State{State: focalState}
}
