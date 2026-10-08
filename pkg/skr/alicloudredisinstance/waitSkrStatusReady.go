package alicloudredisinstance

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
)

// waitSkrStatusReady waits until the SKR object has either a Ready or Error
// condition before continuing. It does not poll the state string - it checks
// for the condition so that both happy-path (Ready) and error-path (Error)
// allow the pipeline to advance and react immediately.
func waitSkrStatusReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	conditions := state.ObjAsAlicloudRedisInstance().Status.Conditions

	hasReady := meta.FindStatusCondition(conditions, cloudresourcesv1beta1.ConditionTypeReady) != nil
	hasError := meta.FindStatusCondition(conditions, cloudresourcesv1beta1.ConditionTypeError) != nil

	if hasReady || hasError {
		return nil, ctx
	}

	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
}
