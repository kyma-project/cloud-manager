package gcpsubnet

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/util"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func preventDeleteOnGcpRedisClusterUsage(ctx context.Context, st composed.State) (error, context.Context) {
	return composed.PreventDeleteWhenUsed(
		&cloudresourcesv1beta1.GcpRedisClusterList{},
		// GcpSubnet is global scope so it's indexed by its name only in internal/controller/cloud-resources/gcpsubnet_controller.go
		st.Name().Name,
		cloudresourcesv1beta1.GcpSubnetField,
		func(ctx context.Context, st composed.State, _ client.ObjectList, usedByNames []string) (error, context.Context) {
			state := st.(*State)
			gcpSubnet := state.ObjAsGcpSubnet()
			gcpSubnet.Status.State = cloudresourcesv1beta1.StateError
			return composed.UpdateStatus(gcpSubnet).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeWarning,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionTypeDeleteWhileUsed,
					Message: fmt.Sprintf("Can not be deleted while used by: %s", usedByNames),
				}).
				ErrorLogMessage("Error updating GcpSubnet status with Warning condition for delete while in use").
				SuccessLogMsg("Forgetting SKR GcpSubnet marked for deleting that is in use").
				SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
				Run(ctx, state)
		},
	)(ctx, st)

}
