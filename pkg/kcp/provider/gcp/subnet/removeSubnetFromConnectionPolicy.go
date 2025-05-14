package subnet

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func removeSubnetFromConnectionPolicy(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.serviceConnectionPolicy == nil {
		return nil, nil
	}
	if state.subnet == nil {
		return nil, nil
	}

	gcpScope := state.Scope().Spec.Scope.Gcp
	region := state.Scope().Spec.Region

	subnetName := GetSubnetFullName(gcpScope.Project, region, ptr.Deref(state.subnet.Name, ""))

	if !state.ConnectionPolicySubnetsContain(subnetName) {
		return nil, nil
	}

	if state.ConnectionPolicySubnetsLen() == 1 { // last one in, cant remove
		return nil, nil
	}

	state.RemoveFromConnectionPolicySubnets(subnetName)

	return nil, nil
}
