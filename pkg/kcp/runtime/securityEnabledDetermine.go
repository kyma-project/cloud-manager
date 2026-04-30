package runtime

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func securityEnabledDetermine(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	state.securityServiceEnabledOnSubscription = false
	for _, v := range state.allRuntimesInSubscription {
		if v {
			state.securityServiceEnabledOnSubscription = true
			break
		}
	}

	state.securityDataSourceEnabledOnRuntime = common.IsSecurityScanEnabledOnRuntime(state.ObjAsRuntime())

	return nil, ctx
}
