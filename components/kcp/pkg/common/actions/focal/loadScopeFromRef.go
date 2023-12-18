package focal

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/kcp/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadScopeFromRef(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	if state.CommonObj().ScopeRef() == nil {
		logger.Info("Object has no scope reference")
		return nil, nil
	}

	logger.Info("Loading scope from reference")

	scope := &cloudresourcesv1beta1.Scope{}
	err := state.Client().Get(ctx, types.NamespacedName{
		Name:      state.CommonObj().ScopeRef().Name,
		Namespace: state.CommonObj().GetNamespace(),
	}, scope)
	if client.IgnoreNotFound(err) != nil {
		err = fmt.Errorf("error getting Scope from reference: %w", err)
		logger.Error(err, "Error loading scope from ref")
		return composed.StopWithRequeue, nil
	}
	if apierrors.IsNotFound(err) {
		logger.Error(err, "Scope the object refers to does not exist")
		return nil, nil // fixInvalidScopeRef will fix the invalid reference
	}

	logger.Info("Loaded Scope from reference")

	state.Scope = scope

	return nil, nil
}
