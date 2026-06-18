package iprange

import (
	"context"
	"errors"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func vSwitchCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.vSwitch != nil {
		return nil, ctx
	}

	// Determine the zone to use
	zoneId, err := resolveZone(ctx, state)
	if err != nil {
		logger.Error(err, "Error resolving zone for AliCloud VSwitch")
		state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.StateError
		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: fmt.Sprintf("Error resolving zone: %s", err),
			}).
			ErrorLogMessage("Error patching AliCloud KCP IpRange status after failed zone resolution").
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	logger.Info("Creating AliCloud VSwitch for IpRange", "vpcId", state.vpcId, "zoneId", zoneId, "cidr", state.ObjAsIpRange().Status.Cidr)

	vSwitchId, err := state.client.CreateVSwitch(ctx, state.vpcId, zoneId, state.ObjAsIpRange().Status.Cidr, state.VSwitchName())
	if err != nil {
		logger.Error(err, "Error creating AliCloud VSwitch for IpRange")
		return composed.StopWithRequeue, ctx
	}

	state.vSwitchId = vSwitchId

	return nil, ctx
}

func resolveZone(ctx context.Context, state *State) (string, error) {
	// Use zone from scope if available
	zones := state.Scope().Spec.Scope.Alicloud.Network.Zones
	if len(zones) > 0 {
		return zones[0], nil
	}

	// Otherwise query available zones from AliCloud
	availableZones, err := state.client.DescribeZones(ctx)
	if err != nil {
		return "", fmt.Errorf("error querying available zones: %w", err)
	}
	if len(availableZones) == 0 {
		return "", errors.New("no available zones found in region")
	}

	return availableZones[0], nil
}
