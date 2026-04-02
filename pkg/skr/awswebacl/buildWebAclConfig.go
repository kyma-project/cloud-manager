package awswebacl

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func buildWebAclConfig(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	webAcl := state.ObjAsAwsWebAcl()

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

	// Store in state
	state.defaultAction = defaultAction
	state.rules = rules
	state.visibilityConfig = visibilityConfig

	return nil, ctx
}
