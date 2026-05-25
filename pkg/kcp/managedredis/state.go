package managedredis

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/kcp/managedredis/types"
)

type state struct {
	focal.State

	ipRange *cloudcontrolv1beta1.IpRange
}

func (s *state) ObjAsAzureManagedRedis() *cloudcontrolv1beta1.AzureManagedRedis {
	return s.Obj().(*cloudcontrolv1beta1.AzureManagedRedis)
}

func (s *state) IpRange() *cloudcontrolv1beta1.IpRange {
	return s.ipRange
}

func (s *state) SetIpRange(r *cloudcontrolv1beta1.IpRange) {
	s.ipRange = r
}

func newState(focalState focal.State) types.State {
	return &state{State: focalState}
}
