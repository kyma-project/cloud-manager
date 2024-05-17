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
	"sort"
	"time"
)

func checkQuota(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.KcpIpRange != nil {
		// Can not enforce quota on SKR IpRanges with already created KCP IpRanges
		// we must let them pass, even when breaching quota, since they were created
		// probably in time when quota was higher or not implemented.
		// WARNING!!! These objects during deprovisioning will lose their KCP copy
		// eventually and must be allowed to be deleted in SKR as well.
		return nil, nil
	}

	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	list := &cloudresourcesv1beta1.IpRangeList{}
	err := state.Cluster().K8sClient().List(ctx, list)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing SKR IpRanges for quota check", composed.StopWithRequeue, ctx)
	}

	// Must be sorted by create date since overlap error is set on NEWER and pass-on with OLDER
	// so quota must allow OLDER resource, and put quota error on NEWER as well
	sort.Slice(list.Items, func(i, j int) bool {
		if list.Items[i].CreationTimestamp.Equal(&list.Items[j].CreationTimestamp) {
			return list.Items[i].Name < list.Items[j].Name
		}
		return list.Items[i].CreationTimestamp.Before(&list.Items[j].CreationTimestamp)
	})

	// Find out what's the quota for this Kind
	totalCountQuota := quota.SkrQuota.TotalCountForObj(state.Obj(), state.Cluster().Scheme(), state.KymaRef.Name)

	// If count(items) < quota, then it's all fine
	if len(list.Items) <= totalCountQuota {
		return nil, nil
	}

	// There are more objects than quota allows.
	// Evaluate if this object should be set with QuotaExceeded condition.

	// Find out
	// * count of "valid" objects (atm all objects are treated as valid)
	// * index of this object in the list sorted by the creationTimestamp ASC (older first)
	validObjectCount := 0
	myIndexInTheList := -1
	for idx, obj := range list.Items {
		validObjectCount++
		if obj.Name == state.Obj().GetName() && obj.Namespace == state.Obj().GetNamespace() {
			myIndexInTheList = idx
		}
	}

	if validObjectCount <= totalCountQuota {
		return nil, nil
	}

	// list index: 0 1 2 3
	// quota=2  =>  0 1 are allowed, 2 3 are exceeding quota

	if myIndexInTheList < totalCountQuota {
		// this object is allowed under the quota due to its age

		// clear QuotaExceeded if it has one
		quotaCond := meta.FindStatusCondition(state.ObjAsIpRange().Status.Conditions, cloudresourcesv1beta1.ConditionTypeQuotaExceeded)
		if quotaCond == nil {
			return nil, nil
		}

		logger.
			WithValues(
				"totalCountQuota", totalCountQuota,
				"resources", fmt.Sprintf("%v", pie.Map(list.Items, func(obj cloudresourcesv1beta1.IpRange) string {
					return fmt.Sprintf("%s/%s/%s", obj.Namespace, obj.Name, obj.CreationTimestamp.Format(time.RFC3339Nano))
				})),
			).
			Info("Clearing SKR IpRange QuotaExceeded condition back to processing")

		state.ObjAsIpRange().Status.State = cloudresourcesv1beta1.StateProcessing

		return composed.UpdateStatus(state.ObjAsIpRange()).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeQuotaExceeded).
			ErrorLogMessage("Error clearing QuotaExceeded condition for SKR IpRange").
			SuccessLogMsg("Cleared SKR IpRange QuotaExceeded condition back to processing").
			SuccessErrorNil(). // continue afterward
			Run(ctx, st)
	}

	// Put QuotaExceeded on this object
	logger.
		WithValues(
			"totalCountQuota", totalCountQuota,
			"resources", fmt.Sprintf("%v", pie.Map(list.Items, func(obj cloudresourcesv1beta1.IpRange) string {
				return fmt.Sprintf("%s/%s/%s", obj.Namespace, obj.Name, obj.CreationTimestamp.Format(time.RFC3339Nano))
			})),
		).
		Info("Quota exceeded")

	state.ObjAsIpRange().Status.State = cloudresourcesv1beta1.ConditionTypeQuotaExceeded
	return composed.UpdateStatus(state.ObjAsIpRange()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeQuotaExceeded,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionTypeQuotaExceeded,
			Message: fmt.Sprintf("Maximum number of %d resources exceeded", totalCountQuota),
		}).
		ErrorLogMessage("Error updating SKR IpRange status with quota exceeded status").
		SuccessLogMsg("Forgetting SKR IpRange with quota exceeded status").
		SuccessError(composed.StopWithRequeueDelay(300*time.Millisecond)).
		FailedError(composed.StopWithRequeue).
		Run(ctx, st)
}
