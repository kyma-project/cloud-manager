package iprange

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func preventDeleteOnNfsInstanceUsage(ctx context.Context, st composed.State) (error, context.Context) {
	return composed.PreventDeleteWhenUsed(
		&cloudcontrolv1beta1.NfsInstanceList{},
		st.Name().String(),
		cloudcontrolv1beta1.IpRangeField,
		func(ctx context.Context, st composed.State, _ client.ObjectList, usedByNames []string) (error, context.Context) {
			state := st.(*State)
			state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.WarningState
			return composed.PatchStatus(state.ObjAsIpRange()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeWarning,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ReasonDeleteWhileUsed,
					Message: fmt.Sprintf("Can not be deleted while used by NfsInstance: %v", usedByNames),
				}).
				ErrorLogMessage("Error patching KCP IpRange status with DeleteWhileUsed by NfsInstance Warning").
				SuccessLogMsg("Delaying KCP IpRange deleting while used by NfsInstance").
				SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
				Run(ctx, state)
		},
	)(ctx, st)
}
