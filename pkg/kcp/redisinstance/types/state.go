package types

import (
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
)

type State interface {
	focal.State
	ObjAsRedisInstance() *v1beta1.RedisInstance

	IpRange() *v1beta1.IpRange
	SetIpRange(r *v1beta1.IpRange)
}
