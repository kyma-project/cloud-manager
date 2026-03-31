package awswebacl

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteWebAcl(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	webAcl := state.ObjAsAwsWebAcl()

	// Skip if never created
	if webAcl.Status.Arn == "" {
		return nil, ctx
	}

	logger.Info("Deleting AWS WebACL")

	scope, err := convertScope(state.Scope())
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error determining scope", composed.StopWithRequeue, ctx)
	}

	// Delete WebACL
	id := extractIdFromArn(webAcl.Status.Arn)
	err = state.awsClient.DeleteWebACL(ctx, webAcl.Name, id, scope, state.lockToken)
	if err != nil {
		// If not found, consider it deleted
		if isNotFoundError(err) {
			logger.Info("WebACL not found in AWS, considering as deleted")
			webAcl.Status.Arn = ""
			return nil, ctx
		}
		return composed.LogErrorAndReturn(err, "Error deleting WebACL", composed.StopWithRequeue, ctx)
	}

	logger.Info("WebACL deleted successfully")

	// Clear status
	webAcl.Status.Arn = ""
	webAcl.Status.Capacity = 0
	webAcl.Status.State = cloudresourcesv1beta1.StateDeleting

	return composed.PatchStatus(webAcl).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeDeleting,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonProcessing,
			Message: "WebACL deleted",
		}).
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
