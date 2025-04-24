package statewithscope

import (
	"context"
	"errors"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func StateProviderPredicate(provider cloudcontrolv1beta1.ProviderType) composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		scope, ok := ScopeFromState(st)
		if !ok {
			logger := log.FromContext(ctx)
			logger.
				WithValues("stateType", fmt.Sprintf("%T", st)).
				Error(errors.New("logical error"), "Could not find the Scope in the State to determine provider")
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
