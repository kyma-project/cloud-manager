package vpcpeering

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/types"
)

type State struct {
	focal.State
	localNetwork  *cloudcontrolv1beta1.Network
	remoteNetwork *cloudcontrolv1beta1.Network
}

func (s *State) ObjAsVpcPeering() *cloudcontrolv1beta1.VpcPeering {
	return s.Obj().(*cloudcontrolv1beta1.VpcPeering)
}

func (s *State) LocalNetwork() *cloudcontrolv1beta1.Network {
	return s.localNetwork
}

func (s *State) RemoteNetwork() *cloudcontrolv1beta1.Network {
	return s.remoteNetwork
}

func newState(focalState focal.State) types.State {
	return &State{State: focalState}
}
