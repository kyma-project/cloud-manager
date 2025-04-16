package subnet

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func preventDeleteOnGcpRedisClusterUsage(ctx context.Context, st composed.State) (error, context.Context) {
	return composed.PreventDeleteWhenUsed(
		&cloudcontrolv1beta1.GcpRedisClusterList{},
		st.Name().String(),
		cloudcontrolv1beta1.GcpSubnetField,
		func(ctx context.Context, st composed.State, _ client.ObjectList, usedByNames []string) (error, context.Context) {
			state := st.(*State)
			state.ObjAsGcpSubnet().Status.State = cloudcontrolv1beta1.StateWarning
			return composed.PatchStatus(state.ObjAsGcpSubnet()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeWarning,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ReasonDeleteWhileUsed,
					Message: fmt.Sprintf("Can not be deleted while used by GcpRedisCluster: %v", usedByNames),
				}).
				ErrorLogMessage("Error patching KCP GcpSubnet status with DeleteWhileUsed by GcpRedisCluster Warning").
				SuccessLogMsg("Delaying KCP GcpSubnet deleting while used by GcpRedisCluster").
				SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
				Run(ctx, state)
		},
	)(ctx, st)
}
