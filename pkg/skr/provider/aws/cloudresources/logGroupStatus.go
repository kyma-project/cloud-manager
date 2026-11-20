package cloudresources

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// logGroupStatus sets wafLogGroupImplementation to the cloudresources.status.subscription.wafLogGroup
// as a record that log groups are created. If implementation changes, the reconciliation will be redone
func logGroupStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	return composed.NewStatusPatcherComposed(state.ObjAsCloudResources()).
		MutateStatus(func(cr *cloudresourcesv1beta1.CloudResources) {
			cr.Status.Subscription.WafLogGroup = wafLogGroupImplementation
		}).
		Run(ctx, state.Cluster().K8sClient())
}
