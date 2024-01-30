package crgcpnfsvolume

import (
	"context"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	labelKymaName        = "cloud-resources.kyma-project.io/kymaName"
	labelRemoteName      = "cloud-resources.kyma-project.io/remoteName"
	labelRemoteNamespace = "cloud-resources.kyma-project.io/remoteNamespace"
	labelNfsVolName      = "cloud-resources.kyma-project.io/nfsVolumeName"
	labelNfsVolNS        = "cloud-resources.kyma-project.io/nfsVolumeNamespace"
)

func loadKcpNfsInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	list := &cloudcontrolv1beta1.NfsInstanceList{}
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
		return composed.LogErrorAndReturn(err, "Error loading KCP NfsInstance", composed.StopWithRequeue, nil)
	}

	if len(list.Items) == 0 {
		return nil, nil
	}

	if len(list.Items) == 1 {
		state.KcpNfsInstance = &list.Items[0]
		return nil, nil
	}

	// more than one NfsInstance found in KCP, log warning and pick one
	names := pie.Map(list.Items, func(x cloudcontrolv1beta1.NfsInstance) string {
		return x.Name
	})
	names = pie.Sort(names)
	logger := composed.LoggerFromCtx(ctx)
	// TODO: log as warning
	logger.
		WithValues("objectKind", "NfsInstance").
		WithValues("names", names).V(-1).
		Info("Found more then one KCP object")
	selectedName := names[0]
	for _, i := range list.Items {
		if i.Name == selectedName {
			state.KcpNfsInstance = &i
			break
		}
	}

	return nil, nil
}
