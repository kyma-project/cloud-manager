package gcpnfsvolume

import (
	"context"
	"errors"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadKcpIpRange(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	logger.Info("Loading KCP IpRange")

	list := &cloudcontrolv1beta1.IpRangeList{}

	// This condition should never happen. Adding this check to have proper error handling instead of panic.
	if state.SkrIpRange == nil {
		return composed.LogErrorAndReturn(errors.New("SkrIpRange is not set in gcpNfsVolume Status"), "Error loading KCP IpRange", composed.StopWithRequeue, ctx)
	}
	err := state.KcpCluster.K8sClient().List(
		ctx,
		list,
		client.MatchingLabels{
			cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
			cloudcontrolv1beta1.LabelRemoteName:      state.SkrIpRange.Name,
			cloudcontrolv1beta1.LabelRemoteNamespace: state.SkrIpRange.Namespace,
		},
		client.InNamespace(state.KymaRef.Namespace),
	)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading KCP IpRange", composed.StopWithRequeue, ctx)
	}

	if len(list.Items) == 0 {
		logger.Info("KCP IpRange not found")
		return nil, nil
	}

	if len(list.Items) == 1 {
		logger.Info("KCP IpRange is loaded")
		state.KcpIpRange = &list.Items[0]
		return nil, nil
	}

	// more than one IpRange found in KCP, log warning and pick one
	names := pie.Map(list.Items, func(x cloudcontrolv1beta1.IpRange) string {
		return x.Name
	})
	names = pie.Sort(names)

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
