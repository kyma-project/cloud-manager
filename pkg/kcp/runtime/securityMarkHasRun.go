package runtime

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func securityMarkHasRun(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	state.securityCooldown.MarkHasRun(state.ObjAsRuntime().Name, state.SecurityServiceEnabledOnSubscription())

	return nil, ctx
}
