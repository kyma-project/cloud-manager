package runtime

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func runtimesLoadAllInSubscription(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	list := &infrastructuremanagerv1.RuntimeList{}
	opts := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(cloudcontrolv1beta1.RuntimeFiledBindingName, state.ObjAsRuntime().Spec.Shoot.SecretBindingName),
	}
	if err := state.Cluster().K8sClient().List(ctx, list, opts); err != nil {
		return composed.LogErrorAndReturn(err, "Error listing runtimes by binding name", composed.StopWithRequeue, ctx)
	}

	allRuntimesInSubscription := make(map[string]bool, len(list.Items))
	for _, runtime := range list.Items {
		allRuntimesInSubscription[runtime.Name] = common.IsSecurityScanEnabledOnRuntime(&runtime)
	}
	state.allRuntimesInSubscription = allRuntimesInSubscription

	return nil, ctx
}
