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

	if len(obj.Status.Id) == 0 {
		logger.Info("VpcPeering deleted before AWS peering is created")
		return nil, nil
	}

	logger.Info("Deleting VpcPeering")

	err := state.client.DeleteVpcPeeringConnection(ctx, ptr.To(obj.Status.Id))

	if err != nil {

		if awsmeta.IsNotFound(err) {
			logger.Info("VpcPeeringConnection not found")
			return nil, nil
		}

		return awsmeta.LogErrorAndReturn(err, "Error deleting vpc peering", ctx)
	}

	logger.Info("VpcPeering deleted")

	return nil, nil
}
