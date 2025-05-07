package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func networkReferenceKymaLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	var kymaName string
	if state.gardenerCluster != nil {
		kymaName = state.gardenerCluster.GetName()
	} else {
		kymaName = state.Name().Name
	}

	netList := &cloudcontrolv1beta1.NetworkList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(cloudcontrolv1beta1.NetworkFieldScope, kymaName),
	}
	if err := state.Cluster().K8sClient().List(ctx, netList, listOps); err != nil {
		return composed.LogErrorAndReturn(err, "Error listing scope networks", composed.StopWithRequeue, ctx)
	}

	kymaNetworkName := common.KcpNetworkKymaCommonName(kymaName)

	for _, net := range netList.Items {
		netCopy := net
		if net.Name == kymaNetworkName {
			state.kcpNetworkKyma = &netCopy
		}
	}

	return nil, ctx
}
