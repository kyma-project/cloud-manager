package subnet

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func addSubnetToConnectionPolicy(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.serviceConnectionPolicy == nil {
		return composed.StopWithRequeue, nil
	}
	if state.subnet == nil {
		return composed.StopWithRequeue, nil
	}

	gcpScope := state.Scope().Spec.Scope.Gcp
	region := state.Scope().Spec.Region

	subnetName := GetSubnetFullName(gcpScope.Project, region, ptr.Deref(state.subnet.Name, ""))

	if state.ConnectionPolicySubnetsContain(subnetName) {
		return nil, nil
	}

	state.AddToConnectionPolicySubnets(subnetName)

	return nil, nil
}
