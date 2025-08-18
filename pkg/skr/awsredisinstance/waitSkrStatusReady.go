package awsredisinstance

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitSkrStatusReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.ObjAsAwsRedisInstance().Status.State != cloudresourcesv1beta1.StateReady {
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return nil, ctx
}
