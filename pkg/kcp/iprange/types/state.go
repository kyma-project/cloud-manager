package types

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
)

type State interface {
	focal.State
	ObjAsIpRange() *cloudcontrolv1beta1.IpRange
	ExistingCidrRanges() []string
	SetExistingCidrRanges(v []string)
	Network() *cloudcontrolv1beta1.Network
}
