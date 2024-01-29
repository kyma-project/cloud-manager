package crgcpnfsvolume

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validateSpec(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	existing := meta.FindStatusCondition(state.ObjAsGcpNfsVolume().Status.Conditions, cloudresourcesv1beta1.ConditionTypeCidrValid)
	if existing != nil && existing.Status == metav1.ConditionTrue {
		// already valid
		return nil, nil
	}
	if existing != nil && existing.Status == metav1.ConditionFalse {
		// already not valid
		return composed.StopAndForget, nil
	}

	//TODO validate all fields in the spec
	return nil, ctx
}
