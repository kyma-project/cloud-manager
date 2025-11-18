package awsvpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
)

func waitNetworkReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, ctx
	}

	if state.RemoteNetwork != nil && meta.IsStatusConditionTrue(state.RemoteNetwork.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady) {
		return nil, ctx
	}

	return composed.StopWithRequeue, ctx
}
