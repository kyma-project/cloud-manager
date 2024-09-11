package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"k8s.io/utils/ptr"
)

func loadVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	virtualNetworkName := state.Scope().Spec.Scope.Azure.VpcNetwork
	resourceGroupName := virtualNetworkName // resourceGroup name have the same name as VPC
	virtualNetworkPeeringName := obj.Name

	peering, err := state.client.GetPeering(ctx, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName)

	if azuremeta.IsNotFound(err) {
		return nil, nil
	}

	if err != nil {
		return azuremeta.LogErrorAndReturn(err, "Error loading VPC Peering", ctx)
	}

	logger = logger.WithValues("id", ptr.Deref(peering.ID, ""))

	ctx = composed.LoggerIntoCtx(ctx, logger)

	state.peering = peering

	logger.Info("Azure VPC Peering loaded")

	id := ptr.Deref(peering.ID, "")

	var peeringState string

	if peering.Properties.PeeringState != nil {
		peeringState = string(*peering.Properties.PeeringState)
	}

	if obj.Status.Id == id && obj.Status.State == peeringState {
		return nil, ctx
	}

	obj.Status.Id = id
	obj.Status.State = peeringState

	return composed.PatchStatus(obj).
		ErrorLogMessage("Error updating VpcPeering status after loading vpc peering connection").
		SuccessErrorNil().
		Run(ctx, state)
}
