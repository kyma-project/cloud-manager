package awswebacl

import (
	"fmt"

	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
)

// convertRuleStatement converts a rule's statements based on its logical operator
func convertRuleStatement(rule *cloudresourcesv1beta1.AwsWebAclRule) (*wafv2types.Statement, error) {
	switch rule.LogicalOperator {
	case cloudresourcesv1beta1.LogicalOperatorNone:
		return convertNoneOperator(rule.Statements)

	case cloudresourcesv1beta1.LogicalOperatorAnd:
		return convertAndOperator(rule.Statements)

	case cloudresourcesv1beta1.LogicalOperatorOr:
		return convertOrOperator(rule.Statements)

	case cloudresourcesv1beta1.LogicalOperatorNot:
		return convertNotOperator(rule.Statements)

	default:
		return nil, fmt.Errorf("unknown logical operator: %s", rule.LogicalOperator)
	}
}

// convertNoneOperator handles single statement (no logical operation)
func convertNoneOperator(statements []cloudresourcesv1beta1.AwsWebAclStatement) (*wafv2types.Statement, error) {
	if len(statements) != 1 {
		return nil, fmt.Errorf("NONE operator requires exactly 1 statement, got %d", len(statements))
	}

	return convertSingleStatement(&statements[0])
}

// convertAndOperator handles AND logical operation (all statements must match)
func convertAndOperator(statements []cloudresourcesv1beta1.AwsWebAclStatement) (*wafv2types.Statement, error) {
	if len(statements) < 2 {
		return nil, fmt.Errorf("AND operator requires at least 2 statements, got %d", len(statements))
	}

	// Validate no special statements (ManagedRuleGroup, RateBased) in logical operators
	for i, stmt := range statements {
		if stmt.ManagedRuleGroup != nil || stmt.RateBased != nil {
			return nil, fmt.Errorf("statement[%d]: ManagedRuleGroup and RateBased cannot be used with AND operator", i)
		}
	}

	wafStatements := make([]wafv2types.Statement, 0, len(statements))
	for i, stmt := range statements {
		wafStmt, err := convertSingleStatement(&stmt)
		if err != nil {
			return nil, fmt.Errorf("error converting AND statement[%d]: %w", i, err)
		}
		wafStatements = append(wafStatements, *wafStmt)
	}

	return &wafv2types.Statement{
		AndStatement: &wafv2types.AndStatement{
			Statements: wafStatements,
		},
	}, nil
}

// convertOrOperator handles OR logical operation (at least one statement must match)
func convertOrOperator(statements []cloudresourcesv1beta1.AwsWebAclStatement) (*wafv2types.Statement, error) {
	if len(statements) < 2 {
		return nil, fmt.Errorf("OR operator requires at least 2 statements, got %d", len(statements))
	}

	// Validate no special statements (ManagedRuleGroup, RateBased) in logical operators
	for i, stmt := range statements {
		if stmt.ManagedRuleGroup != nil || stmt.RateBased != nil {
			return nil, fmt.Errorf("statement[%d]: ManagedRuleGroup and RateBased cannot be used with OR operator", i)
		}
	}

	wafStatements := make([]wafv2types.Statement, 0, len(statements))
	for i, stmt := range statements {
		wafStmt, err := convertSingleStatement(&stmt)
		if err != nil {
			return nil, fmt.Errorf("error converting OR statement[%d]: %w", i, err)
		}
		wafStatements = append(wafStatements, *wafStmt)
	}

	return &wafv2types.Statement{
		OrStatement: &wafv2types.OrStatement{
			Statements: wafStatements,
		},
	}, nil
}

// convertNotOperator handles NOT logical operation (negates the statement)
func convertNotOperator(statements []cloudresourcesv1beta1.AwsWebAclStatement) (*wafv2types.Statement, error) {
	if len(statements) != 1 {
		return nil, fmt.Errorf("NOT operator requires exactly 1 statement, got %d", len(statements))
	}

	// Validate no special statements (ManagedRuleGroup, RateBased) in NOT
	stmt := &statements[0]
	if stmt.ManagedRuleGroup != nil || stmt.RateBased != nil {
		return nil, fmt.Errorf("ManagedRuleGroup and RateBased cannot be used with NOT operator")
	}

	wafStmt, err := convertSingleStatement(stmt)
	if err != nil {
		return nil, fmt.Errorf("error converting NOT statement: %w", err)
	}

	return &wafv2types.Statement{
		NotStatement: &wafv2types.NotStatement{
			Statement: wafStmt,
		},
	}, nil
}

// convertSingleStatement converts a single statement to AWS WAF format
// Validates exactly one statement type is set
func convertSingleStatement(stmt *cloudresourcesv1beta1.AwsWebAclStatement) (*wafv2types.Statement, error) {
	result := &wafv2types.Statement{}

	// Count how many statement types are set
	count := 0
	if stmt.RateBased != nil {
		count++
	}
	if stmt.ManagedRuleGroup != nil {
		count++
	}
	if stmt.GeoMatch != nil {
		count++
	}
	if stmt.ByteMatch != nil {
		count++
	}
	if stmt.LabelMatch != nil {
		count++
	}
	if stmt.SizeConstraint != nil {
		count++
	}
	if stmt.SqliMatch != nil {
		count++
	}
	if stmt.XssMatch != nil {
		count++
	}
	if stmt.RegexMatch != nil {
		count++
	}
	if stmt.AsnMatch != nil {
		count++
	}

	// Validate exactly one is set
	if count == 0 {
		return nil, fmt.Errorf("statement must have exactly one condition set")
	}
	if count > 1 {
		return nil, fmt.Errorf("statement must have exactly one condition set, found %d", count)
	}

	// Convert the single statement that was set
	if stmt.RateBased != nil {
		rateStmt, err := convertRateBasedStatement(stmt.RateBased)
		if err != nil {
			return nil, err
		}
		result.RateBasedStatement = rateStmt
		return result, nil
	}

	if stmt.ManagedRuleGroup != nil {
		managedStmt, err := convertManagedRuleGroupStatement(stmt.ManagedRuleGroup)
		if err != nil {
			return nil, err
		}
		result.ManagedRuleGroupStatement = managedStmt
		return result, nil
	}

	if stmt.GeoMatch != nil {
		geoStmt, err := convertGeoMatchStatement(stmt.GeoMatch)
		if err != nil {
			return nil, err
		}
		result.GeoMatchStatement = geoStmt
		return result, nil
	}

	if stmt.ByteMatch != nil {
		byteStmt, err := convertByteMatchStatement(stmt.ByteMatch)
		if err != nil {
			return nil, err
		}
		result.ByteMatchStatement = byteStmt
		return result, nil
	}

	if stmt.LabelMatch != nil {
		labelStmt := convertLabelMatchStatement(stmt.LabelMatch)
		result.LabelMatchStatement = labelStmt
		return result, nil
	}

	if stmt.SizeConstraint != nil {
		sizeStmt, err := convertSizeConstraintStatement(stmt.SizeConstraint)
		if err != nil {
			return nil, err
		}
		result.SizeConstraintStatement = sizeStmt
		return result, nil
	}

	if stmt.SqliMatch != nil {
		sqliStmt, err := convertSqliMatchStatement(stmt.SqliMatch)
		if err != nil {
			return nil, err
		}
		result.SqliMatchStatement = sqliStmt
		return result, nil
	}

	if stmt.XssMatch != nil {
		xssStmt, err := convertXssMatchStatement(stmt.XssMatch)
		if err != nil {
			return nil, err
		}
		result.XssMatchStatement = xssStmt
		return result, nil
	}

	if stmt.RegexMatch != nil {
		regexStmt, err := convertRegexMatchStatement(stmt.RegexMatch)
		if err != nil {
			return nil, err
		}
		result.RegexMatchStatement = regexStmt
		return result, nil
	}

	if stmt.AsnMatch != nil {
		asnStmt, err := convertAsnMatchStatement(stmt.AsnMatch)
		if err != nil {
			return nil, err
		}
		result.AsnMatchStatement = asnStmt
		return result, nil
	}

	// Should never reach here if validation was done correctly
	return nil, fmt.Errorf("no statement set (internal error)")
}
