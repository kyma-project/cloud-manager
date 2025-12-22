package vpcnetwork

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	kcpcommonaction "github.com/kyma-project/cloud-manager/pkg/kcp/commonAction"
	vpcnetworktypes "github.com/kyma-project/cloud-manager/pkg/kcp/vpcnetwork/types"
)

func newState(kcpCommonState kcpcommonaction.State) *State {
	return &State{}
}

var _ vpcnetworktypes.State = (*State)(nil)

type State struct {
	kcpcommonaction.State

	normalizedSpecCidrs []string
}

func (s *State) ObjAsVpcNetwork() *cloudcontrolv1beta1.VpcNetwork {
	return s.Obj().(*cloudcontrolv1beta1.VpcNetwork)
}

func (s *State) NormalizedSpecCidrs() []string {
	return s.normalizedSpecCidrs
}
