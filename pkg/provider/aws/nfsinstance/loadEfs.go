package nfsinstance

import (
	"context"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/composed"
	"k8s.io/utils/pointer"
)

func loadEfs(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	list, err := state.awsClient.DescribeFileSystems(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing AWS file systems", composed.StopWithRequeue, nil)
	}

	name := state.ObjAsNfsInstance().Spec.RemoteRef.String()
	for _, fs := range list {
		if pointer.StringDeref(fs.Name, "") == name {
			state.efs = &fs
			return nil, nil
		}
	}

	return nil, nil
}
