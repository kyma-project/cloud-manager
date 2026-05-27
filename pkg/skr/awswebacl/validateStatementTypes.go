package awswebacl

import (
	"context"
	"fmt"
	"strings"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
)

func validateStatementTypes(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)
	obj := state.ObjAsAwsWebAcl()

	// Check if feature flag is enabled
	if !feature.WafManagedRuleGroupOnly.Value(ctx) {
		return nil, ctx
	}

	// Collect all violations
	var violations []string

	// Validate each rule's statements
	for i, rule := range obj.Spec.Rules {
		for j, stmt := range rule.Statements {
			stmtPath := fmt.Sprintf("spec.rules[%d].statements[%d]", i, j)
			violations = append(violations, validateStatement(&stmt, stmtPath)...)
		}
	}

	// If violations found, update status and stop
	if len(violations) > 0 {
		errMsg := fmt.Sprintf("WebACL contains restricted statement types (only ManagedRuleGroup allowed): %s",
			strings.Join(violations, "; "))
		logger.Info("Statement type validation failed", "violations", violations)

		return composed.NewStatusPatcherComposed(obj).
			MutateStatus(func(acl *cloudresourcesv1beta1.AwsWebAcl) {
				acl.SetStatusStatementTypeRestricted(errMsg)
			}).
			OnSuccess(composed.Forget).
			Run(ctx, state.Cluster().K8sClient())
	}

	logger.Info("Statement type validation passed")
	return nil, ctx
}

// validateStatement recursively validates a statement
// Returns list of violation descriptions
func validateStatement(stmt *cloudresourcesv1beta1.AwsWebAclStatement, path string) []string {
	var violations []string

	// Check if only ManagedRuleGroup is set
	if !isOnlyManagedRuleGroup(stmt) {
		violations = append(violations, fmt.Sprintf("%s has non-ManagedRuleGroup statement type", path))
	}

	return violations
}

// isOnlyManagedRuleGroup checks if the statement has only ManagedRuleGroup set
func isOnlyManagedRuleGroup(stmt *cloudresourcesv1beta1.AwsWebAclStatement) bool {
	return stmt.ManagedRuleGroup != nil &&
		stmt.RateBased == nil &&
		stmt.GeoMatch == nil &&
		stmt.ByteMatch == nil &&
		stmt.LabelMatch == nil &&
		stmt.SizeConstraint == nil &&
		stmt.SqliMatch == nil &&
		stmt.XssMatch == nil &&
		stmt.RegexMatch == nil &&
		stmt.AsnMatch == nil
}
