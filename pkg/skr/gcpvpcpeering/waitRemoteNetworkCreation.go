package gcpvpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
)

func waitNetworkReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.KcpRemoteNetwork != nil && meta.IsStatusConditionTrue(*state.KcpRemoteNetwork.Conditions(), cloudcontrolv1beta1.ConditionTypeReady) {
		return nil, ctx
	}

	return composed.StopWithRequeue, ctx
}
