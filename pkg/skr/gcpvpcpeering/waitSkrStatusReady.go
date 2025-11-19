package gcpvpcpeering

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitSkrStatusReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.ObjAsGcpVpcPeering().Status.Conditions == nil || state.ObjAsGcpVpcPeering().Status.Conditions[0].Status != cloudresourcesv1beta1.StateReady {
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return nil, nil
}
