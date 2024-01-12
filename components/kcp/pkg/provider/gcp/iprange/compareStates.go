package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/components/lib/composed"
)

func compareStates(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	//TBD: Check and see whether the desiredState == actualState
	//and set the inSync flag in state object
	state.inSync = true

	return nil, nil
}
