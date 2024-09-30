package azurevpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
)

func waitStatusReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !meta.IsStatusConditionTrue(*state.ObjAsAzureVpcPeering().Conditions(), cloudcontrolv1beta1.ConditionTypeReady) ||
		state.ObjAsAzureVpcPeering().Status.State != cloudcontrolv1beta1.VirtualNetworkPeeringStateConnected {
		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
	}

	return nil, nil
}
