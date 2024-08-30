package v2

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func identifyPeeringIpRanges(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	ipRange := state.ObjAsIpRange()

	//Check and see whether the object is being deleted
	deleting := composed.IsMarkedForDeletion(ipRange)

	//If the address object doesn't exist in GCP, continue.
	if state.address == nil {
		return nil, nil
	}

	//If Service Connection doesn't exist in GCP
	if state.serviceConnection == nil {
		//Add this IpAddress to be included while creating Service Connection.
		if !deleting {
			state.ipRanges = []string{state.address.Name}
		}
		return nil, nil
	}

	logger.WithValues("ipRange :", ipRange.Name).Info("Loading IpRanges List")

	//Get GCP scope specific values
	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	vpc := gcpScope.VpcNetwork

	list, err := state.computeClient.ListGlobalAddresses(ctx, project, vpc)
	if err != nil {
		return composed.UpdateStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: "Error listing Global Addresses from GCP",
			}).
			SuccessError(composed.StopWithRequeue).
			SuccessLogMsg("Error listing Global Addresses from GCP").
			Run(ctx, state)
	}

	//load the address names into a map.
	tmpMap := map[string]struct{}{}
	for _, addr := range list.Items {
		if addr.Purpose != string(v1beta1.GcpPurposePSA) {
			tmpMap[addr.Name] = struct{}{}
		}
	}

	//Iterate over the list peering ranges, and include required ones.
	for _, name := range state.serviceConnection.ReservedPeeringRanges {
		//If this object is getting deleted, do not include its name.
		if deleting && name == state.address.Name {
			continue
		}

		//If the IpRange exists, include it.
		if _, ok := tmpMap[name]; ok {
			state.ipRanges = append(state.ipRanges, name)
		}
	}

	//If not deleting, add the name of this IpRange
	if !deleting {
		state.ipRanges = append(state.ipRanges, state.address.Name)
	}

	logger.WithValues("ipRange :", ipRange.Name).Info(fmt.Sprintf("IpRanges List :: %v", state.ipRanges))

	return nil, nil
}
