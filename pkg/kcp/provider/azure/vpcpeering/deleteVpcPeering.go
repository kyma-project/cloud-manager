package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"k8s.io/utils/ptr"
)

func deleteVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	if state.localPeering == nil {
		logger.Info("Skipping local peering deletion")
		return nil, nil
	}

	resourceId, err := util.ParseResourceID(ptr.Deref(state.localPeering.ID, ""))

	if err != nil {
		logger.Error(err, "Failed parsing localPeering.ID while deleting local peering")
		return nil, nil
	}

	logger.Info("Deleting VpcPeering")

	err = state.localClient.DeletePeering(
		ctx,
		resourceId.ResourceGroup,
		resourceId.ResourceName,
		obj.GetLocalPeeringName(),
	)

	if err != nil {
		return azuremeta.LogErrorAndReturn(err, "Error deleting local vpc peering", ctx)
	}

	logger.Info("Local VpcPeering deleted")

	return nil, nil
}
