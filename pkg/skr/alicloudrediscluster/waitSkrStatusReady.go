package alicloudrediscluster

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
)

// waitSkrStatusReady waits until the SKR object has either a Ready or Error
// condition before continuing. Checking for conditions is more reliable than
// checking the state string - it allows both happy-path (Ready) and error-path
// (Error) to advance the pipeline immediately.
func waitSkrStatusReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	conditions := state.ObjAsAlicloudRedisCluster().Status.Conditions

	hasReady := meta.FindStatusCondition(conditions, cloudresourcesv1beta1.ConditionTypeReady) != nil
	hasError := meta.FindStatusCondition(conditions, cloudresourcesv1beta1.ConditionTypeError) != nil

	if hasReady || hasError {
		return nil, ctx
	}

	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
}
