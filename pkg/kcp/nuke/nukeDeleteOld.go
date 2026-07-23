package nuke

import (
	"context"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/rate"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func nukeDeleteOld(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	readyCond := meta.FindStatusCondition(state.ObjAsNuke().Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)
	if readyCond == nil || readyCond.Status != metav1.ConditionTrue {
		return nil, ctx
	}

	age := time.Since(state.ObjAsNuke().CreationTimestamp.Time)
	if age < 30*24*time.Hour {
		return nil, ctx
	}

	err := state.Cluster().K8sClient().Delete(ctx, state.ObjAsNuke())
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting old Nuke", composed.StopWithRequeueDelay(rate.Slow1s.When(state.ObjAsNuke())), ctx)
	}

	return composed.StopAndForget, ctx
}
