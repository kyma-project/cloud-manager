package vpcnetwork

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	kcpcommonaction "github.com/kyma-project/cloud-manager/pkg/kcp/commonAction"
	vpcnetworktypes "github.com/kyma-project/cloud-manager/pkg/kcp/vpcnetwork/types"
)

func newState(kcpCommonState kcpcommonaction.State) *State {
	return &State{
		State: kcpCommonState,
	}
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

func (s *State) IsKymaTypePredicate(_ context.Context, _ composed.State) bool {
	return s.ObjAsVpcNetwork().Spec.Type == cloudcontrolv1beta1.VpcNetworkTypeKyma
}

func (s *State) IsGardenerTypePredicate(_ context.Context, _ composed.State) bool {
	return s.ObjAsVpcNetwork().Spec.Type == cloudcontrolv1beta1.VpcNetworkTypeGardener
}
