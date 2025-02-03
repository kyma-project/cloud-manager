package rediscluster

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/kcp/rediscluster/types"
)

type state struct {
	focal.State

	ipRange *cloudcontrolv1beta1.IpRange
}

func (s *state) ObjAsRedisCluster() *cloudcontrolv1beta1.RedisCluster {
	return s.Obj().(*cloudcontrolv1beta1.RedisCluster)
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
