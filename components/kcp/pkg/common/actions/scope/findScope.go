package scope

import (
	"context"
	"fmt"
	"github.com/elliotchance/pie/v2"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func findScope(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(State)

	list := &cloudresourcesv1beta1.ScopeList{}
	err := state.Cluster().K8sClient().List(
		ctx,
		list,
		client.InNamespace(state.Obj().GetNamespace()),
		client.MatchingLabels{
			cloudresourcesv1beta1.KymaLabel: state.ObjAsCommonObj().KymaName(),
		},
	)
	if err != nil {
		err = fmt.Errorf("error listing scopes: %w", err)
		logger.Error(err, "Error when no scope ref")
		return composed.StopWithRequeue, nil
	}

	if len(list.Items) == 0 {
		logger.Info("No matching Scope found, proceeding to create one")
		return nil, nil
	}

	var scope *cloudresourcesv1beta1.Scope
	if len(list.Items) > 1 {
		for _, s := range list.Items {
			if s.Name == state.ObjAsCommonObj().KymaName() {
				scope = &s
				break
			}
		}
		if scope == nil {
			scope = &list.Items[0]
		}
		logger.
			WithValues(
				"scopes", pie.Map(list.Items, func(s cloudresourcesv1beta1.Scope) string {
					return s.Name
				}),
				"pickedScope", scope.Name,
			).
			Info("Found more then one scope!")
	} else {
		// only one Scope found, all fine
		scope = &list.Items[0]
	}

	state.SetScope(scope)

	return nil, nil
}
