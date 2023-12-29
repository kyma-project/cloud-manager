package types

import (
	"github.com/kyma-project/cloud-resources/components/kcp/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/actions/focal"
)

type State interface {
	focal.State
	ObjAsNfsInstance() *v1beta1.NfsInstance

	IpRange() *v1beta1.IpRange
	SetIpRange(r *v1beta1.IpRange)
}
