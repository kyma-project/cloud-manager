package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func networkExistsPredicate(_ context.Context, st composed.State) bool {
	return st.(*State).net != nil
}
