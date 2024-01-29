package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func ignoreScopeObj(ctx context.Context, st composed.State) (error, context.Context) {
	if st.Obj().GetObjectKind().GroupVersionKind().Kind == "Scope" {
		return composed.StopAndForget, nil
	}
	return nil, nil
}
