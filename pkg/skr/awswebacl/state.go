package awswebacl

import (
	"context"
	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	webaclclient "github.com/kyma-project/cloud-manager/pkg/skr/awswebacl/client"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	ctrl "sigs.k8s.io/controller-runtime"
)

type State struct {
	commonscope.State
	awsClientProvider awsclient.SkrClientProvider[webaclclient.Client]
	env               abstractions.Environment

	awsClient    webaclclient.Client
	roleName     string
	awsWebAcl    *wafv2types.WebACL // Loaded AWS WebACL
	lockToken    string             // Transient lock token from loadWebAcl, not persisted
	updateNeeded bool               // Whether update is needed based on spec vs AWS state
}

func newStateFactory(
	composedStateFactory composed.StateFactory,
	commonScopeStateFactory commonscope.StateFactory,
	awsClientProvider awsclient.SkrClientProvider[webaclclient.Client],
	env abstractions.Environment,
) *stateFactory {
	return &stateFactory{
		composedStateFactory:    composedStateFactory,
		commonScopeStateFactory: commonScopeStateFactory,
		awsClientProvider:       awsClientProvider,
		env:                     env,
	}
}

type stateFactory struct {
	composedStateFactory    composed.StateFactory
	commonScopeStateFactory commonscope.StateFactory
	awsClientProvider       awsclient.SkrClientProvider[webaclclient.Client]
	env                     abstractions.Environment
}

func (f *stateFactory) NewState(ctx context.Context, req ctrl.Request) (*State, error) {
	scopeState, err := f.commonScopeStateFactory.NewState(ctx, req.NamespacedName,
		f.composedStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.AwsWebAcl{}),
	)
	if err != nil {
		return nil, err
	}

	return &State{
		State:             scopeState,
		awsClientProvider: f.awsClientProvider,
		env:               f.env,
	}, nil
}

func (s *State) ObjAsAwsWebAcl() *cloudresourcesv1beta1.AwsWebAcl {
	return s.Obj().(*cloudresourcesv1beta1.AwsWebAcl)
}
