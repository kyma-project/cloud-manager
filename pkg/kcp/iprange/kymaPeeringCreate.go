package iprange

import (
	"context"
	cloudcontrol1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func kymaPeeringCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, ctx
	}

	peering := &cloudcontrol1beta1.VpcPeering{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: state.ObjAsIpRange().Namespace,
			Name:      state.Scope().Name,
		},
		Spec: cloudcontrol1beta1.VpcPeeringSpec{
			RemoteRef: cloudcontrol1beta1.RemoteRef{
				Namespace: "none",
				Name:      "none",
			},
			Scope: cloudcontrol1beta1.ScopeRef{Name: state.Scope().Name},
			Details: &cloudcontrol1beta1.VpcPeeringDetails{
				LocalNetwork: klog.ObjectRef{
					Name:      state.kymaNetwork.Name,
				},
				RemoteNetwork: klog.ObjectRef{
					Name:      state.kymaNetwork.Name,
				},
				PeeringName: ""
			},
		},
	}

	return nil, nil
}
