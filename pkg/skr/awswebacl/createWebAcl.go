package awswebacl

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func createWebAcl(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	webAcl := state.ObjAsAwsWebAcl()

	// Skip if already exists
	if webAcl.Status.Arn != "" {
		return nil, ctx
	}

	logger.Info("Creating AWS WebACL")

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

	// Determine scope
	scope, err := convertScope(state.Scope())
	if err != nil {
		webAcl.Status.State = cloudresourcesv1beta1.StateError
		return composed.PatchStatus(webAcl).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonInvalidSpec,
				Message: fmt.Sprintf("Error determining scope: %v", err),
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	// Create WebACL
	createdWebACL, lockToken, err := state.awsClient.CreateWebACL(
		ctx,
		webAcl.Name,
		webAcl.Spec.Description,
		scope,
		defaultAction,
		rules,
		visibilityConfig,
		convertTags(webAcl),
	)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating WebACL", composed.StopWithRequeue, ctx)
	}

	// Update status with created resource info
	webAcl.Status.Arn = ptr.Deref(createdWebACL.ARN, "")
	webAcl.Status.Capacity = createdWebACL.Capacity
	webAcl.Status.State = cloudresourcesv1beta1.StateReady

	// Store lock token in state (transient, not persisted)
	state.lockToken = lockToken

	return composed.PatchStatus(webAcl).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonReady,
			Message: "WebACL created successfully",
		}).
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
