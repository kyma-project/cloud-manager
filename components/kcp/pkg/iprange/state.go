package iprange

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/iprange/types"
)

type state struct {
	focal.State
}

func (s *state) ObjAsIpRange() *cloudresourcesv1beta1.IpRange {
	return s.Obj().(*cloudresourcesv1beta1.IpRange)
}

func newState(focalState focal.State) types.State {
	return &state{State: focalState}
}
