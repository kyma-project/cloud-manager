package scope

import (
	"context"
	"errors"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func shootNameMustHave(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.shootName == "" {
		return composed.LogErrorAndReturn(
			errors.New("shoot name error"),
			"Unable to find the shoot name, can not proceed",
			composed.StopAndForget,
			ctx,
		)
	}

	return nil, ctx
}
