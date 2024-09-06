package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadNetworks(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	netList := &cloudcontrolv1beta1.NetworkList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(cloudcontrolv1beta1.NetworkFieldScope, st.Name().Name),
	}
	if err := state.Cluster().K8sClient().List(ctx, netList, listOps); err != nil {
		return composed.LogErrorAndReturn(err, "Error listing scope networks", composed.StopWithRequeue, ctx)
	}

	state.allNetworks = netList

	return nil, nil
}
