package criprange

import (
	"context"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	labelKymaName        = "cloud-resources.kyma-project.io/kymaName"
	labelRemoteName      = "cloud-resources.kyma-project.io/remoteName"
	labelRemoteNamespace = "cloud-resources.kyma-project.io/remoteNamespace"
)

func loadKcpIpRange(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	list := &cloudcontrolv1beta1.IpRangeList{}
	err := state.KcpCluster.K8sClient().List(
		ctx,
		list,
		client.MatchingLabels{
			labelKymaName:        state.KymaRef.Name,
			labelRemoteName:      state.Name().Name,
			labelRemoteNamespace: state.Name().Namespace,
		},
		client.InNamespace(state.KymaRef.Namespace),
	)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading KCP IpRange", composed.StopWithRequeue, nil)
	}

	if len(list.Items) == 0 {
		return nil, nil
	}

	if len(list.Items) == 1 {
		state.KcpIpRange = &list.Items[0]
		return nil, nil
	}

	// more than one IpRange found in KCP, log warning and pick one
	names := pie.Map(list.Items, func(x cloudcontrolv1beta1.IpRange) string {
		return x.Name
	})
	names = pie.Sort(names)
	logger := composed.LoggerFromCtx(ctx)
	// TODO: log as warning
	logger.
		WithValues("objectKind", "IpRange").
		WithValues("names", names).
		Info("Found more then one KCP object")
	selectedName := names[0]
	for _, i := range list.Items {
		if i.Name == selectedName {
			state.KcpIpRange = &i
			break
		}
	}

	return nil, nil
}
