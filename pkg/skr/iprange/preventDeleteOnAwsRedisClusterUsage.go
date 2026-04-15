package iprange

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/util"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func preventDeleteOnAwsRedisClusterUsage(ctx context.Context, st composed.State) (error, context.Context) {
	return composed.PreventDeleteWhenUsed(
		&cloudresourcesv1beta1.AwsRedisClusterList{},
		// IpRange is global scope so it's indexed by its name only in internal/controller/cloud-resources/iprange_controller.go
		st.Name().Name,
		cloudresourcesv1beta1.IpRangeField,
		func(ctx context.Context, st composed.State, _ client.ObjectList, usedByNames []string) (error, context.Context) {
			state := st.(*State)
			msg := fmt.Sprintf("Can not be deleted while used by: %s", usedByNames)
			existing := meta.FindStatusCondition(*state.ObjAsIpRange().Conditions(), cloudresourcesv1beta1.ConditionTypeWarning)
			if existing != nil && existing.Status == metav1.ConditionTrue && existing.Reason == cloudresourcesv1beta1.ConditionTypeDeleteWhileUsed && existing.Message == msg {
				return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
			}
			return composed.UpdateStatus(state.ObjAsIpRange()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeWarning,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionTypeDeleteWhileUsed,
					Message: msg,
				}).
				DeriveStateFromConditions(state.MapConditionToState()).
				ErrorLogMessage("Error updating IpRange status with Warning condition for delete while in use").
				SuccessLogMsg("Forgetting SKR IpRange marked for deleting that is in use").
				SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
				Run(ctx, state)
		},
	)(ctx, st)

}
