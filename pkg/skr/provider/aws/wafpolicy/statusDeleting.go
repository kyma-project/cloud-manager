package wafpolicy

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
)

func statusDeleting(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	policy := state.ObjAsWafPolicy()

	readyCondition := meta.FindStatusCondition(policy.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)
	if readyCondition != nil && (readyCondition.Reason == cloudresourcesv1beta1.ReasonDeleting || readyCondition.Reason == cloudresourcesv1beta1.ReasonDeletionBlockedByDependents) {
		return nil, ctx
	}

	return composed.NewStatusPatcherComposed(policy).
		MutateStatus(func(acl *cloudresourcesv1beta1.WafPolicy) {
			acl.SetStatusDeleting()
		}).
		OnSuccess(composed.Continue).
		OnStatusChanged(composed.Log("WafPolicy Deleting")).
		Run(ctx, state.Cluster().K8sClient())
}
