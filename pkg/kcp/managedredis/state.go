package managedredis

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/kcp/managedredis/types"
)

type state struct {
	focal.State
}

func (s *state) ObjAsAzureManagedRedis() *cloudcontrolv1beta1.AzureManagedRedis {
	return s.Obj().(*cloudcontrolv1beta1.AzureManagedRedis)
}

func newState(focalState focal.State) types.State {
	return &state{State: focalState}
}
