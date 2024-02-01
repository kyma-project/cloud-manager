package gcpnfsvolume

import (
	"context"
	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadPersistenceVolume(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	list := &v1.PersistentVolumeList{}
	err := state.SkrCluster.K8sClient().List(
		ctx,
		list,
		client.MatchingLabels{
			labelNfsVolName: state.Name().Name,
			labelNfsVolNS:   state.Name().Namespace,
		},
	)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading Persistent Volume", composed.StopWithRequeue, nil)
	}

	if len(list.Items) == 0 {
		return nil, nil
	}

	if len(list.Items) == 1 {
		state.PV = &list.Items[0]
		return nil, nil
	}

	// more than one PersistentVolume found in SKR, log warning and pick one
	names := pie.Map(list.Items, func(x v1.PersistentVolume) string {
		return x.Name
	})
	names = pie.Sort(names)
	logger := composed.LoggerFromCtx(ctx)
	// TODO: log as warning
	logger.
		WithValues("objectKind", "PersistentVolume").
		WithValues("names", names).
		Info("Found more then one PersistentVolume")
	selectedName := names[0]
	for _, i := range list.Items {
		if i.Name == selectedName {
			state.PV = &i
			break
		}
	}

	return nil, nil

}
