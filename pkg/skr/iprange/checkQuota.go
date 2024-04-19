package iprange

import (
	"context"
	"fmt"
	"github.com/elliotchance/pie/v2"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/quota"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func checkQuota(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.KcpIpRange != nil {
		// can not enforce quota on SKR IpRanges with already created KCP IpRanges
		return nil, nil
	}

	list := &cloudresourcesv1beta1.IpRangeList{}
	err := state.Cluster().K8sClient().List(ctx, list)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing SKR IpRanges for quota check", composed.StopWithRequeue, ctx)
	}

	totalCountQuota := quota.SkrQuota.TotalCountForObj(state.Obj(), state.Cluster().Scheme(), state.KymaRef.Name)

	if len(list.Items) <= totalCountQuota {
		return nil, nil
	}

	errCond := meta.FindStatusCondition(*state.ObjAsIpRange().Conditions(), cloudresourcesv1beta1.ConditionTypeError)
	if errCond != nil && errCond.Reason == cloudresourcesv1beta1.ConditionReasonQuotaExceeded {
		return composed.StopAndForget, nil
	}

	// Quota exceeded
	logger := composed.LoggerFromCtx(ctx)
	logger.
		WithValues(
			"totalCountQuota", totalCountQuota,
			"resources", fmt.Sprintf("%v", pie.Map(list.Items, func(obj cloudresourcesv1beta1.IpRange) string {
				return fmt.Sprintf("%s/%s", obj.Namespace, obj.Name)
			})),
		).
		Info("Quota exceeded")

	// go through the list
	// skip objects that already have QuotaExceeded reason
	// determine if the reconciled obj should be set with QuotaExceeded reason

	setQuotaExceededReason := false
	allowedObjectCount := 0
	for _, obj := range list.Items {
		objErrCond := meta.FindStatusCondition(obj.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
		if objErrCond != nil && objErrCond.Reason == cloudresourcesv1beta1.ConditionReasonQuotaExceeded {
			continue
		}
		allowedObjectCount++

		if allowedObjectCount > totalCountQuota {
			setQuotaExceededReason = true
			break
		}
	}

	if !setQuotaExceededReason {
		return nil, nil
	}

	state.ObjAsIpRange().Status.State = cloudresourcesv1beta1.ConditionReasonQuotaExceeded
	return composed.UpdateStatus(state.ObjAsIpRange()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonQuotaExceeded,
			Message: fmt.Sprintf("Maximum number of %d resources exceeded", totalCountQuota),
		}).
		ErrorLogMessage("Error updating SKR IpRange status with quota exceeded status").
		SuccessLogMsg("Forgetting SKR IpRange with quota exceeded status").
		SuccessError(composed.StopAndForget).
		FailedError(composed.StopWithRequeue).
		Run(ctx, st)
}
