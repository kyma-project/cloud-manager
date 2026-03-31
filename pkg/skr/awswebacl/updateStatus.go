package awswebacl

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	webAcl := state.ObjAsAwsWebAcl()

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	// If WebACL exists and is ready
	if webAcl.Status.Arn != "" && webAcl.Status.State != cloudresourcesv1beta1.StateError {
		webAcl.Status.State = cloudresourcesv1beta1.StateReady

		return composed.PatchStatus(webAcl).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonReady,
				Message: "WebACL is ready",
			}).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	return nil, ctx
}
