package iprange

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type State struct {
	focal.State

	existingCidrRanges []string

	networkKey            client.ObjectKey
	isCloudManagerNetwork bool
	isKymaNetwork         bool

	network *cloudcontrolv1beta1.Network

	kymaNetwork *cloudcontrolv1beta1.Network
	kymaPeering *cloudcontrolv1beta1.VpcPeering
}

func (s *State) ObjAsIpRange() *cloudcontrolv1beta1.IpRange {
	return s.Obj().(*cloudcontrolv1beta1.IpRange)
}

func (s *State) Network() *cloudcontrolv1beta1.Network {
	return s.network
}

func (s *State) ExistingCidrRanges() []string {
	return s.existingCidrRanges
}

func (s *State) SetExistingCidrRanges(v []string) {
	s.existingCidrRanges = v
}

func newState(focalState focal.State) types.State {
	return &State{State: focalState}
}
