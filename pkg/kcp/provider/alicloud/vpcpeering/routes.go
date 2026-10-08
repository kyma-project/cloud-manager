package vpcpeering

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createRoutes(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.vpcPeering == nil || len(state.vpcPeering.RemoteIpv4Cidrs) == 0 {
		return nil, ctx
	}
	if state.ObjAsVpcPeering().Spec.Details == nil {
		return nil, ctx
	}

	// Use the shoot name (not the CIDR) as the tag key for MATCHED/UNMATCHED filtering,
	// consistent with the AWS pattern (ShouldUpdateRouteTable uses shoot name as tag key).
	shootName := state.Scope().Spec.ShootName
	strategy := string(state.ObjAsVpcPeering().Spec.Details.RemoteRouteTableUpdateStrategy)
	tables := routeTablesForStrategy(state.routeTables, strategy, shootName)

	instanceId := state.ObjAsVpcPeering().Status.Id
	for _, table := range tables {
		for _, cidr := range state.vpcPeering.RemoteIpv4Cidrs {
			if err := state.client.CreateRouteEntry(ctx, table.RouteTableId, cidr, instanceId); err != nil {
				return composed.LogErrorAndReturn(err, "Error creating local route entry", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
			}
		}
	}

	return nil, ctx
}

func createRemoteRoutes(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	// The destination CIDR for remote routes pointing at the local (Kyma shoot) VPC.
	// Scope.Scope.Alicloud.VpcNetwork is the VPC name, not the CIDR.
	// The actual CIDR is in Scope.Scope.Alicloud.Network.VPC.CIDR.
	localVpcCidr := state.Scope().Spec.Scope.Alicloud.Network.VPC.CIDR
	if localVpcCidr == "" {
		return nil, ctx
	}
	if state.ObjAsVpcPeering().Spec.Details == nil {
		return nil, ctx
	}

	shootName := state.Scope().Spec.ShootName
	strategy := string(state.ObjAsVpcPeering().Spec.Details.RemoteRouteTableUpdateStrategy)
	tables := routeTablesForStrategy(state.remoteRouteTables, strategy, shootName)

	instanceId := state.ObjAsVpcPeering().Status.Id
	for _, table := range tables {
		if err := state.remoteClient.CreateRouteEntry(ctx, table.RouteTableId, localVpcCidr, instanceId); err != nil {
			return composed.LogErrorAndReturn(err, "Error creating remote route entry", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
		}
	}

	return nil, ctx
}

func updateSuccessStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsVpcPeering()

	if state.vpcPeering == nil {
		return nil, ctx
	}

	if len(obj.Status.Id) > 0 &&
		len(obj.Status.RemoteId) > 0 &&
		meta.IsStatusConditionTrue(*obj.Conditions(), cloudcontrolv1beta1.ConditionTypeReady) {
		return nil, ctx
	}

	meta.RemoveStatusCondition(obj.Conditions(), cloudcontrolv1beta1.ConditionTypeError)

	obj.Status.State = string(cloudcontrolv1beta1.StateReady)
	if state.vpcPeering.RemoteVpcId != "" {
		obj.Status.RemoteId = state.vpcPeering.RemoteVpcId
	}

	return composed.PatchStatus(obj).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonReady,
			Message: "VpcPeering is provisioned",
		}).
		ErrorLogMessage("Error patching AliCloud VpcPeering with success status").
		SuccessLogMsg("AliCloud VpcPeering is ready").
		FailedError(composed.StopWithRequeue).
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
