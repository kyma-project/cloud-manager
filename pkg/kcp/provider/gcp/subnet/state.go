package subnet

import (
	"context"

	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/networkconnectivity/apiv1/networkconnectivitypb"
	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"

	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/subnet/client"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

type State struct {
	focal.State

	computeClient              client.ComputeClient
	networkComnnectivityClient client.NetworkConnectivityClient

	subnet                  *computepb.Subnetwork
	serviceConnectionPolicy *networkconnectivitypb.ServiceConnectionPolicy

	network *cloudcontrolv1beta1.Network

	updateMask []string
}

type StateFactory interface {
	NewState(ctx context.Context, focalState focal.State) (*State, error)
}

type stateFactory struct {
	computeClientProvider             gcpclient.GcpClientProvider[client.ComputeClient]
	networkConnectivityClientProvider gcpclient.GcpClientProvider[client.NetworkConnectivityClient]
	env                               abstractions.Environment
}

func NewStateFactory(
	computeClientProvider gcpclient.GcpClientProvider[client.ComputeClient],
	networkConnectivityClientProvider gcpclient.GcpClientProvider[client.NetworkConnectivityClient],
	env abstractions.Environment) StateFactory {
	return &stateFactory{
		computeClientProvider:             computeClientProvider,
		networkConnectivityClientProvider: networkConnectivityClientProvider,
		env:                               env,
	}
}

func (statefactory *stateFactory) NewState(ctx context.Context, focalState focal.State) (*State, error) {

	computeClient := statefactory.computeClientProvider()
	networkConnectivityClient := statefactory.networkConnectivityClientProvider()

	return newState(focalState, computeClient, networkConnectivityClient), nil
}

func newState(focalState focal.State, computeClient client.ComputeClient, networkConnectivityClient client.NetworkConnectivityClient) *State {
	return &State{
		State:                      focalState,
		computeClient:              computeClient,
		networkComnnectivityClient: networkConnectivityClient,
	}
}

func (s *State) ObjAsGcpSubnet() *cloudcontrolv1beta1.GcpSubnet {
	return s.Obj().(*cloudcontrolv1beta1.GcpSubnet)
}

func (s *State) ShouldUpdateConnectionPolicy() bool {
	return len(s.updateMask) > 0
}

func (s *State) ConnectionPolicySubnetsContainCurrent() bool {
	project := s.Scope().Spec.Scope.Gcp.Project
	region := s.Scope().Spec.Region
	currentSubnetName := GetSubnetFullName(project, region, s.ObjAsGcpSubnet().Status.Id)
	return pie.Contains(s.serviceConnectionPolicy.PscConfig.Subnetworks, currentSubnetName)
}

func (s *State) ConnectionPolicySubnetsLen() int {
	return len(s.serviceConnectionPolicy.PscConfig.Subnetworks)
}

func (s *State) AddCurrentSubnetToConnectionPolicy() {
	project := s.Scope().Spec.Scope.Gcp.Project
	region := s.Scope().Spec.Region
	currentSubnetName := GetSubnetFullName(project, region, s.ObjAsGcpSubnet().Status.Id)
	s.serviceConnectionPolicy.PscConfig.Subnetworks = append(s.serviceConnectionPolicy.PscConfig.Subnetworks, currentSubnetName)
	s.updateMask = append(s.updateMask, "psc_config")
}

func (s *State) RemoveCurrentSubnetFromConnectionPolicy() {
	project := s.Scope().Spec.Scope.Gcp.Project
	region := s.Scope().Spec.Region
	currentSubnetName := GetSubnetFullName(project, region, s.ObjAsGcpSubnet().Status.Id)
	s.serviceConnectionPolicy.PscConfig.Subnetworks = pie.FilterNot(s.serviceConnectionPolicy.PscConfig.Subnetworks, func(name string) bool {
		return name == currentSubnetName
	})
	s.updateMask = append(s.updateMask, "psc_config")
}

func (s *State) ShouldDeleteConnectionPolicy() bool {
	return s.ConnectionPolicySubnetsLen() == 1 && s.ConnectionPolicySubnetsContainCurrent()
}
