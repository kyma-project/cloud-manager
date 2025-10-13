package types

import (
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
)

type State interface {
	focal.State
	ObjAsVpcPeering() *v1beta1.VpcPeering
	LocalNetwork() *v1beta1.Network
	RemoteNetwork() *v1beta1.Network
}
