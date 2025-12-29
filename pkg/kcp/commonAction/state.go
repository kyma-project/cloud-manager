package commonAction

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

type State interface {
	composed.State

	ObjAsObjWithStatus() composed.ObjWithStatus

	Subscription() *cloudcontrolv1beta1.Subscription
	VpcNetwork() *cloudcontrolv1beta1.VpcNetwork
	IpRange() *cloudcontrolv1beta1.IpRange
	GcpSubnet() *cloudcontrolv1beta1.GcpSubnet
}

// factory ========================================================================

type StateFactory interface {
	NewState(base composed.State) State
}

func NewStateFactory() StateFactory {
	return &stateFactory{}
}

type stateFactory struct{}

func (f *stateFactory) NewState(base composed.State) State {
	return &stateImpl{
		State: base,
	}
}

// state ========================================================================

type stateImpl struct {
	composed.State

	subscription *cloudcontrolv1beta1.Subscription
	vpcNetwork   *cloudcontrolv1beta1.VpcNetwork
	// azureGardenerVpcNetwork exists only on Azure provider when Runtime is created in Gardener created network
	//azureGardenerVpcNetwork *cloudcontrolv1beta1.VpcNetwork
	ipRange   *cloudcontrolv1beta1.IpRange
	gcpSubnet *cloudcontrolv1beta1.GcpSubnet
}

func (s *stateImpl) ObjAsObjWithStatus() composed.ObjWithStatus {
	return s.Obj().(composed.ObjWithStatus)
}

func (s *stateImpl) Subscription() *cloudcontrolv1beta1.Subscription {
	return s.subscription
}

func (s *stateImpl) VpcNetwork() *cloudcontrolv1beta1.VpcNetwork {
	return s.vpcNetwork
}

func (s *stateImpl) IpRange() *cloudcontrolv1beta1.IpRange {
	return s.ipRange
}

func (s *stateImpl) GcpSubnet() *cloudcontrolv1beta1.GcpSubnet {
	return s.gcpSubnet
}

