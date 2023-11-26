package genericActions

import (
	"context"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func FinalizerAdd(finalizer string) composed.Action {
	return func(ctx context.Context, state composed.State) error {
		controllerutil.AddFinalizer(state.Obj(), finalizer)
		return nil
	}
}
