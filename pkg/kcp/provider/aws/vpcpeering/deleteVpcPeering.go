package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

func deleteVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	lll := logger.WithValues("vpcPeeringName", obj.Name)

	if len(obj.Status.Id) == 0 {
		lll.Info("VpcPeering deleted before AWS peering is created")
		return nil, nil
	}

	lll = lll.WithValues("vpcPeeringId", obj.Status.Id)
	lll.Info("Deleting VpcPeering")

	err := state.client.DeleteVpcPeeringConnection(ctx, ptr.To(obj.Status.Id))

	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error deleting vpc peering", composed.LoggerIntoCtx(ctx, lll))
	}

	return nil, nil
}
