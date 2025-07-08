package statewithscope

import (
	"context"
	"errors"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func IsTrialPredicate(ctx context.Context, st composed.State) bool {
	scope, ok := ScopeFromState(st)
	if !ok || scope == nil {
		logger := log.FromContext(ctx)
		logger.
			WithValues("stateType", fmt.Sprintf("%T", st)).
			Error(errors.New("logical error"), "Could not find the non-nil Scope in the State to determine trial")
		return true
	}
	plan := scope.Labels[cloudcontrolv1beta1.LabelScopeBrokerPlanName]
	return plan == "trial"
}
