package statewithscope

import (
	"context"
	"errors"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type StateWithObjAsScope interface {
	ObjAsScope() *cloudcontrolv1beta1.Scope
}

func ScopeFromState(st composed.State) (*cloudcontrolv1beta1.Scope, bool) {
	if state, ok := st.(focal.State); ok {
		return state.Scope(), true
	}
	if state, ok := st.(StateWithObjAsScope); ok {
		return state.ObjAsScope(), true
	}
	if scope, ok := st.Obj().(*cloudcontrolv1beta1.Scope); ok {
		return scope, true
	}

	return nil, false
}

func StateProviderPredicate(provider cloudcontrolv1beta1.ProviderType) composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		scope, ok := ScopeFromState(st)
		if !ok {
			logger := log.FromContext(ctx)
			logger.
				WithValues("stateType", fmt.Sprintf("%T", st)).
				Error(errors.New("logical error"), "Could not find the Scope in the State")
			return false
		}
		if scope == nil {
			return false
		}
		return scope.Spec.Provider == provider
	}
}

func AwsProviderPredicate(ctx context.Context, st composed.State) bool {
	return StateProviderPredicate(cloudcontrolv1beta1.ProviderAws)(ctx, st)
}

func AzureProviderPredicate(ctx context.Context, st composed.State) bool {
	return StateProviderPredicate(cloudcontrolv1beta1.ProviderAzure)(ctx, st)
}

func GcpProviderPredicate(ctx context.Context, st composed.State) bool {
	return StateProviderPredicate(cloudcontrolv1beta1.ProviderGCP)(ctx, st)
}

func OpenStackProviderPredicate(ctx context.Context, st composed.State) bool {
	return StateProviderPredicate(cloudcontrolv1beta1.ProviderOpenStack)(ctx, st)
}
