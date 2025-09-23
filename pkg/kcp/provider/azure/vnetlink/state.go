package vnetlink

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vnetlink/types"
)

type State struct {
	focal.State
}

func (s *State) ObjAsAzureVNetLink() *cloudcontrolv1beta1.AzureVNetLink {
	return s.Obj().(*cloudcontrolv1beta1.AzureVNetLink)
}

func newState(focalState focal.State) types.State {
	return &State{State: focalState}
}
