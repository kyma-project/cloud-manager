package wafpolicy

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	webAcl := state.ObjAsWafPolicy()

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, ctx
	}

	return composed.NewStatusPatcherComposed(webAcl).
		MutateStatus(func(acl *cloudresourcesv1beta1.WafPolicy) {
			acl.Status.ProviderId = ptr.Deref(state.awsWebAcl.ARN, "")
			acl.SetStatusReady()
		}).
		OnStatusChanged(
			composed.Log("WafPolicy is Ready"),
		).
		Run(ctx, state.Cluster().K8sClient())
}
