package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"k8s.io/utils/pointer"
)

func loadVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	if len(obj.Status.Id) == 0 {
		return nil, nil
	}

	resource, err := util.ParseResourceID(obj.Status.Id)

	if err != nil {
		return azuremeta.LogErrorAndReturn(err, "Error parsing virtual network peering ID", ctx)
	}

	peering, err := state.client.Get(ctx, resource.ResourceGroup, resource.ResourceName, resource.SubResourceName)

	if err != nil {
		return azuremeta.LogErrorAndReturn(err, "Error loading VPC Peering", ctx)
	}

	logger = logger.WithValues("id", pointer.StringDeref(peering.ID, ""))

	ctx = composed.LoggerIntoCtx(ctx, logger)

	state.peering = peering

	logger.Info("Azure VPC Peering loaded")

	return nil, ctx
}
