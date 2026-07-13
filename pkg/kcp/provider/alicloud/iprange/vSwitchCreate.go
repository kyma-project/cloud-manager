package iprange

import (
	"context"
	"errors"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/iprange/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func vSwitchCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	zones := state.Scope().Spec.Scope.Alicloud.Network.Zones

	// Determine zone names; fall back to querying if scope has none
	var zoneNames []string
	if len(zones) > 0 {
		for _, z := range zones {
			zoneNames = append(zoneNames, z.Name)
		}
	} else {
		available, err := state.client.DescribeZones(ctx)
		if err != nil || len(available) == 0 {
			if err == nil {
				err = errors.New("no available zones found in region")
			}
			logger.Error(err, "Error resolving zones for AliCloud vSwitch")
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
		zoneNames = []string{available[0]}
	}

	// Build set of zones that already have a vSwitch, keyed by zone id.
	// This is deterministic regardless of the order vSwitches were loaded/created.
	existingZones := map[string]struct{}{}
	for _, vsw := range state.vSwitches {
		existingZones[vsw.ZoneId] = struct{}{}
	}

	anyCreated := false
	for i, zoneName := range zoneNames {
		if i >= len(state.zoneCidrs) {
			break
		}
		// Deterministic zone->CIDR pairing by index; skip zones already provisioned.
		if _, exists := existingZones[zoneName]; exists {
			continue
		}

		zoneCidr := state.zoneCidrs[i]
		name := state.VSwitchName(i)

		logger.Info("Creating AliCloud VSwitch for IpRange", "vpcId", state.vpcId, "zoneId", zoneName, "cidr", zoneCidr, "name", name)

		vSwitchId, err := state.client.CreateVSwitch(ctx, state.vpcId, zoneName, zoneCidr, name)
		if err != nil {
			logger.Error(err, "Error creating AliCloud VSwitch for IpRange")
			return composed.StopWithRequeue, ctx
		}

		state.vSwitches = append(state.vSwitches, &alicloudiprangeclient.VSwitchInfo{
			VSwitchId:   vSwitchId,
			VSwitchName: name,
			CidrBlock:   zoneCidr,
			VpcId:       state.vpcId,
			ZoneId:      zoneName,
			Status:      "Pending",
		})
		anyCreated = true
	}

	if anyCreated {
		return composed.StopWithRequeue, ctx
	}

	return nil, ctx
}
