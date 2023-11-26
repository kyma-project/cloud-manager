package genericActions

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/apis/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func HasFinalizer(ctx context.Context, state composed.State) bool {
	result := controllerutil.ContainsFinalizer(state.Obj(), cloudresourcesv1beta1.Finalizer)
	return result
}
