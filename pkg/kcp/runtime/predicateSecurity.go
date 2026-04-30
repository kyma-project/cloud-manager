package runtime

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func predicateSecurityIsCool(_ context.Context, st composed.State) bool {
	state := st.(*State)

	result := state.securityCooldown.CanRunNow(state.ObjAsRuntime().Name, state.SecurityServiceEnabledOnSubscription())

	return result
}
