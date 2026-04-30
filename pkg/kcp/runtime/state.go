package runtime

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	runtimetypes "github.com/kyma-project/cloud-manager/pkg/kcp/runtime/types"
)

var _ runtimetypes.State = &State{}

type State struct {
	composed.State

	Subscription *cloudcontrolv1beta1.Subscription
	VpcNetwork   *cloudcontrolv1beta1.VpcNetwork
}

func (s *State) ObjAsRuntime() *infrastructuremanagerv1.Runtime {
	return s.Obj().(*infrastructuremanagerv1.Runtime)
}
