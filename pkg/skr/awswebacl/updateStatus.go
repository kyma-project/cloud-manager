package awswebacl

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"k8s.io/utils/ptr"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	webAcl := state.ObjAsAwsWebAcl()

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, ctx
	}

	return composed.NewStatusPatcherComposed(webAcl).
		MutateStatus(func(acl *cloudresourcesv1beta1.AwsWebAcl) {
			acl.Status.Arn = ptr.Deref(state.awsWebAcl.ARN, "")
			acl.Status.Capacity = state.awsWebAcl.Capacity
			acl.SetStatusReady()
		}).
		OnStatusChanged(
			composed.Log("AwsWebAcl is Ready"),
		).
		Run(ctx, state.Cluster().K8sClient())
}
