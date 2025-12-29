package commonAction

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func StateProviderPredicate(provider cloudcontrolv1beta1.ProviderType) composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		state, ok := st.(State)
		if !ok {
			composed.LoggerFromCtx(ctx).
				WithValues("stateType", fmt.Sprintf("%T", st)).
				Error(common.ErrLogical, "Non kcpcommonaction state given to kcpcommonaction.StateProviderPredicate")
			return false
		}
		return state.Subscription().Status.Provider == provider
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
