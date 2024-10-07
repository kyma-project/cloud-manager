package nuke

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func deleteResources(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	for _, rks := range state.Resources {
		for _, obj := range rks.Objects {
			if obj.GetDeletionTimestamp().IsZero() {
				logger = logger.
					WithValues(
						"kind", rks.Kind,
						"orphanName", obj.GetName(),
					)
				logger.Info("Nuke deleting orphaned resource")

				err := state.Cluster().K8sClient().Delete(ctx, obj)
				if err != nil {
					logger.Error(err, "Error nuke deleting orphaned resource")
				}
			}
		}
	}

	return nil, nil
}
