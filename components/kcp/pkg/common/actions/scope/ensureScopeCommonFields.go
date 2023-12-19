package scope

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/kcp/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ensureScopeCommonFields(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)

	state.Scope().Name = state.CommonObj().KymaName()
	state.Scope().Namespace = state.Obj().GetNamespace()

	// set kyma name in label
	metav1.SetMetaDataLabel(&state.Scope().ObjectMeta, cloudresourcesv1beta1.KymaLabel, state.CommonObj().KymaName())

	// set kyma name in spec
	state.Scope().Spec.KymaName = state.CommonObj().KymaName()

	// set shoot name
	state.Scope().Spec.ShootName = state.ShootName()

	return nil, nil
}
