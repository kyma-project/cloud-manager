package runtime

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

const tmpSecurityEnabledLabel = "cloud-manager.kyma-project.io/tmp-security-enabled"

func predicateSecurityEnabled(_ context.Context, st composed.State) bool {
	state := st.(*State)
	// a temporary mean to annotate runtime with security enabled until a filed is added
	_, ok := state.ObjAsRuntime().Labels[tmpSecurityEnabledLabel]
	return ok
}
