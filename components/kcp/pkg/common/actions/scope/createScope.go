package scope

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-resources-manager/components/lib/composed"
	"strings"
)

func createScope(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)
	switch state.Provider() {
	case cloudresourcesv1beta1.ProviderGCP:
		return createScopeGcp(ctx, state)
	case cloudresourcesv1beta1.ProviderAzure:
		return createScopeAzure(ctx, state)
	case cloudresourcesv1beta1.ProviderAws:
		return createScopeAws(ctx, state)
	}

	err := fmt.Errorf("unable to handle unknown provider '%s'", state.Provider())
	logger := composed.LoggerFromCtx(ctx)
	logger.Error(err, "Error defining scope")
	return composed.StopAndForget, nil // no requeue

}

func commonVpcName(shootNamespace, shootName string) string {
	project := strings.TrimPrefix(shootNamespace, "garden-")
	return fmt.Sprintf("shoot--%s--%s", project, shootName)
}
