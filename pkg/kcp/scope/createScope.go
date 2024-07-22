package scope

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"strings"
)

func createScope(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	switch state.provider {
	case cloudcontrolv1beta1.ProviderGCP:
		return createScopeGcp(ctx, state)
	case cloudcontrolv1beta1.ProviderAzure:
		return createScopeAzure(ctx, state)
	case cloudcontrolv1beta1.ProviderAws:
		return createScopeAws(ctx, state)
	case cloudcontrolv1beta1.ProviderOpenStack:
		return createScopeOpenStack(ctx, state)
	}

	err := fmt.Errorf("unable to handle unknown provider '%s'", state.provider)
	logger := composed.LoggerFromCtx(ctx)
	logger.Error(err, "Error defining scope")
	return composed.StopAndForget, nil // no requeue

}

func commonVpcName(shootNamespace, shootName string) string {
	project := strings.TrimPrefix(shootNamespace, "garden-")
	return fmt.Sprintf("shoot--%s--%s", project, shootName)
}
