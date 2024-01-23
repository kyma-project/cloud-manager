package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func ensureScopeCommonFields(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	// set name
	state.ObjAsScope().Name = state.kyma.GetName() // same name as Kyma CR
	state.ObjAsScope().Namespace = state.Obj().GetNamespace()

	// set finalizer
	controllerutil.AddFinalizer(state.ObjAsScope(), cloudcontrolv1beta1.FinalizerName)

	// set kyma name in label
	metav1.SetMetaDataLabel(&state.ObjAsScope().ObjectMeta, cloudcontrolv1beta1.KymaLabel, state.Obj().GetName())

	// set kyma name in spec
	state.ObjAsScope().Spec.KymaName = state.Obj().GetName()

	// set shoot name
	state.ObjAsScope().Spec.ShootName = state.shootName

	// set region
	state.ObjAsScope().Spec.Region = state.shoot.Spec.Region

	// set provider
	state.ObjAsScope().Spec.Provider = state.provider

	return nil, nil
}
