package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func ensureScopeCommonFields(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	// set name
	state.ObjAsScope().Name = state.kyma.GetName() // same name as Kyma CR
	state.ObjAsScope().Namespace = state.kyma.GetNamespace()

	// set finalizer
	controllerutil.AddFinalizer(state.ObjAsScope(), api.CommonFinalizerDeletionHook)

	// set kyma name in label
	metav1.SetMetaDataLabel(&state.ObjAsScope().ObjectMeta, cloudcontrolv1beta1.LabelKymaName, state.Obj().GetName())

	// set kyma name in spec
	state.ObjAsScope().Spec.KymaName = state.Obj().GetName()

	// set shoot name
	state.ObjAsScope().Spec.ShootName = state.shootName

	// set region
	state.ObjAsScope().Spec.Region = state.shoot.Spec.Region

	// set provider
	state.ObjAsScope().Spec.Provider = state.provider

	// copy kyma labels to scope that are used in metrics and features
	if state.ObjAsScope().Labels == nil {
		state.ObjAsScope().Labels = make(map[string]string, len(cloudcontrolv1beta1.ScopeLabels))
	}
	for _, label := range cloudcontrolv1beta1.ScopeLabels {
		state.ObjAsScope().Labels[label] = state.kyma.GetLabels()[label]
	}

	return nil, nil
}
