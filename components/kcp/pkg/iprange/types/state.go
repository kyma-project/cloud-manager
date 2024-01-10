package types

import (
	"github.com/kyma-project/cloud-resources-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-resources-manager/components/kcp/pkg/common/actions/focal"
)

type State interface {
	focal.State
	ObjAsIpRange() *v1beta1.IpRange
}
