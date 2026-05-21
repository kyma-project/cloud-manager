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

func securityEnabledDetermine(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	state.securityDataSourceEnabledOnRuntime = common.IsSecurityScanEnabledOnRuntime(state.ObjAsRuntime())

	// load all runtimes in subscription
	list := &infrastructuremanagerv1.RuntimeList{}
	opts := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(cloudcontrolv1beta1.RuntimeFiledBindingName, state.ObjAsRuntime().Spec.Shoot.SecretBindingName),
	}
	if err := state.Cluster().K8sClient().List(ctx, list, opts); err != nil {
		return composed.LogErrorAndReturn(err, "Error listing runtimes by binding name", composed.StopWithRequeue, ctx)
	}

	// determine if security is on for any runtime in the subscription
	// assume not, and set true on first with enabled security
	state.securityServiceEnabledOnSubscription = false
	for _, runtime := range list.Items {
		if common.IsSecurityScanEnabledOnRuntime(&runtime) {
			state.securityServiceEnabledOnSubscription = true
			break
		}
	}

	return nil, ctx
}
