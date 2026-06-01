package awswebacl

import (
	"fmt"

	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
)

// convertRuleStatement converts a rule's single ManagedRuleGroup statement
func convertRuleStatement(rule *cloudresourcesv1beta1.AwsWebAclRule) (*wafv2types.Statement, error) {
	if len(rule.Statements) != 1 {
		return nil, fmt.Errorf("rule must have exactly 1 statement, got %d", len(rule.Statements))
	}

	return convertSingleStatement(&rule.Statements[0])
}

// convertSingleStatement converts a single ManagedRuleGroup statement to AWS WAF format
func convertSingleStatement(stmt *cloudresourcesv1beta1.AwsWebAclStatement) (*wafv2types.Statement, error) {
	result := &wafv2types.Statement{}

	if stmt.ManagedRuleGroup == nil {
		return nil, fmt.Errorf("statement must have ManagedRuleGroup set")
	}

	managedStmt, err := convertManagedRuleGroupStatement(stmt.ManagedRuleGroup)
	if err != nil {
		return nil, err
	}
	result.ManagedRuleGroupStatement = managedStmt
	return result, nil
}
