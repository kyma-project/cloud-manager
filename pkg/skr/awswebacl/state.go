package awswebacl

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	webaclclient "github.com/kyma-project/cloud-manager/pkg/skr/awswebacl/client"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

type State struct {
	composed.State
	KymaRef           klog.ObjectRef
	KcpCluster        composed.StateCluster
	Scope             *cloudcontrolv1beta1.Scope
	awsClientProvider awsclient.SkrClientProvider[webaclclient.Client]
	env               abstractions.Environment
}

type StateFactory interface {
	NewState(ctx context.Context, baseState composed.State) (*State, error)
}

func newStateFactory(
	baseStateFactory composed.StateFactory,
	kymaRef klog.ObjectRef,
	kcpCluster composed.StateCluster,
	awsClientProvider awsclient.SkrClientProvider[webaclclient.Client],
) *stateFactory {
	return &stateFactory{
		baseStateFactory:  baseStateFactory,
		kymaRef:           kymaRef,
		kcpCluster:        kcpCluster,
		awsClientProvider: awsClientProvider,
	}
}

type stateFactory struct {
	baseStateFactory  composed.StateFactory
	kymaRef           klog.ObjectRef
	kcpCluster        composed.StateCluster
	awsClientProvider awsclient.SkrClientProvider[webaclclient.Client]
}

func (f *stateFactory) NewState(req ctrl.Request) *State {
	return &State{
		State:             f.baseStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.AwsWebAcl{}),
		KymaRef:           f.kymaRef,
		KcpCluster:        f.kcpCluster,
		awsClientProvider: f.awsClientProvider,
	}
}

func (s *State) ObjAsAwsWebAcl() *cloudresourcesv1beta1.AwsWebAcl {
	return s.Obj().(*cloudresourcesv1beta1.AwsWebAcl)
}
