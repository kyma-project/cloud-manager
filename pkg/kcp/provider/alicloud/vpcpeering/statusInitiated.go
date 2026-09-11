package vpcpeering

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func statusInitiated(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsVpcPeering()

	if obj.Status.State != "" {
		return nil, ctx
	}

	obj.Status.State = string(cloudcontrolv1beta1.StateProcessing)

	return composed.PatchStatus(obj).
		ErrorLogMessage("Error patching AliCloud VpcPeering initiated status").
		SuccessErrorNil().
		Run(ctx, state)
}
