package types

import (
	"github.com/kyma-project/cloud-resources/components/kcp/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/actions/focal"
)

type State interface {
	focal.State
	ObjAsIpRange() *v1beta1.IpRange
}
