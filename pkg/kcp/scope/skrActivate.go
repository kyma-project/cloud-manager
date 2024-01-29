package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func skrActivate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	state.activeSkrCollection.AddKymaName(state.kyma.GetName())

	return nil, nil
}
