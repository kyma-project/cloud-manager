package nfsinstance

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance/types"
)

type state struct {
	focal.State

	ipRange *cloudresourcesv1beta1.IpRange
}

func (s *state) ObjAsNfsInstance() *cloudresourcesv1beta1.NfsInstance {
	return s.Obj().(*cloudresourcesv1beta1.NfsInstance)
}

func (s *state) IpRange() *cloudresourcesv1beta1.IpRange {
	return s.ipRange
}

func (s *state) SetIpRange(r *cloudresourcesv1beta1.IpRange) {
	s.ipRange = r
}

func newState(focalState focal.State) types.State {
	return &state{State: focalState}
}
