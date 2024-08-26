package nfsinstance

import (
	"context"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	nfsinstancetypes "github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance/types"
)

func copyStatusHostsToHost(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(nfsinstancetypes.State)

	if state.ObjAsNfsInstance().Status.Host != "" && state.ObjAsNfsInstance().Status.Path != "" {
		return nil, nil
	}

	changed := false

	if state.Scope().Spec.Provider == cloudcontrolv1beta1.ProviderAws && len(state.ObjAsNfsInstance().Status.Hosts) > 0 {
		changed = true
		state.ObjAsNfsInstance().Status.Host = pie.First(state.ObjAsNfsInstance().Status.Hosts)
		state.ObjAsNfsInstance().Status.Path = "/"
	} else if state.Scope().Spec.Provider == cloudcontrolv1beta1.ProviderGCP && len(state.ObjAsNfsInstance().Status.Hosts) > 0 {
		changed = true
		state.ObjAsNfsInstance().Status.Host = pie.First(state.ObjAsNfsInstance().Status.Hosts)
		state.ObjAsNfsInstance().Status.Path = state.ObjAsNfsInstance().Spec.Instance.Gcp.FileShareName
	}

	if !changed {
		return nil, nil
	}

	return composed.PatchStatus(state.ObjAsNfsInstance()).
		ErrorLogMessage("Error patching NfsInstance status after copying hosts to host").
		SuccessErrorNil().
		Run(ctx, state)
}
