package awswebacl

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateWebAcl(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	webAcl := state.ObjAsAwsWebAcl()

	// Skip if not created yet
	if webAcl.Status.Arn == "" {
		return nil, ctx
	}

	// Check if update is needed
	if !needsUpdate(webAcl) {
		return nil, ctx
	}

	logger.Info("Updating AWS WebACL")

	// Convert spec to AWS types
	defaultAction, err := convertDefaultAction(webAcl.Spec.DefaultAction)
	if err != nil {
		webAcl.Status.State = cloudresourcesv1beta1.StateError
		return composed.PatchStatus(webAcl).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonInvalidSpec,
				Message: fmt.Sprintf("Invalid default action: %v", err),
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	rules, err := convertRules(webAcl.Spec.Rules)
	if err != nil {
		webAcl.Status.State = cloudresourcesv1beta1.StateError
		return composed.PatchStatus(webAcl).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonInvalidSpec,
				Message: fmt.Sprintf("Invalid rules: %v", err),
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	visibilityConfig := convertVisibilityConfig(webAcl.Spec.VisibilityConfig, webAcl.Name)

	scope := ScopeRegional()

	// Get ID from loaded WebACL in state
	if state.awsWebAcl == nil || state.awsWebAcl.Id == nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("WebACL not loaded in state"),
			"Cannot update WebACL without loaded state",
			composed.StopWithRequeue,
			ctx,
		)
	}

	// Update WebACL
	err = state.awsClient.UpdateWebACL(
		ctx,
		webAcl.Name,
		*state.awsWebAcl.Id,
		scope,
		defaultAction,
		rules,
		visibilityConfig,
		state.lockToken,
	)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating WebACL", composed.StopWithRequeue, ctx)
	}

	logger.Info("WebACL updated successfully")

	return nil, ctx
}

func needsUpdate(webAcl *cloudresourcesv1beta1.AwsWebAcl) bool {
	// For now, skip updates to avoid optimistic lock issues
	// TODO: Add comparison logic to detect actual changes between spec and AWS state
	return false
}
