package network

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func preventDeleteWhenUsedByPeeringLocalNetwork(ctx context.Context, st composed.State) (error, context.Context) {
	return composed.PreventDeleteWhenUsed(
		&cloudcontrolv1beta1.VpcPeeringList{},
		st.Name().String(),
		cloudcontrolv1beta1.VpcPeeringLocalNetworkField,
		func(ctx context.Context, st composed.State, _ client.ObjectList, usedByNames []string) (error, context.Context) {
			state := st.(*state)
			state.ObjAsNetwork().Status.State = string(cloudcontrolv1beta1.StateWarning)
			return composed.PatchStatus(state.ObjAsNetwork()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeWarning,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ReasonDeleteWhileUsed,
					Message: fmt.Sprintf("Can not be deleted while used by VpcPeering local network: %v", usedByNames),
				}).
				ErrorLogMessage("Error patching KCP Network status with DeleteWhileUsed by VpcPeering local network Warning").
				SuccessLogMsg("Delaying KCP Network deleting while used by VpcPeering local network").
				SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
				Run(ctx, state)
		},
	)(ctx, st)
}
