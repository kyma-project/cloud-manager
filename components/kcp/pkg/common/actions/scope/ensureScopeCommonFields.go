package scope

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ensureScopeCommonFields(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)

	state.Scope().Name = state.ObjAsCommonObj().KymaName()
	state.Scope().Namespace = state.Obj().GetNamespace()

	// set kyma name in label
	metav1.SetMetaDataLabel(&state.Scope().ObjectMeta, cloudresourcesv1beta1.KymaLabel, state.ObjAsCommonObj().KymaName())

	// set kyma name in spec
	state.Scope().Spec.KymaName = state.ObjAsCommonObj().KymaName()

	// set shoot name
	state.Scope().Spec.ShootName = state.ShootName()

	// set region
	state.Scope().Spec.Region = state.Shoot().Spec.Region

	// set provider
	state.Scope().Spec.Provider = state.Provider()

	return nil, nil
}
