package awswebacl

import (
	"fmt"

	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
)

// StatementConverter defines the interface for converting Kyma statements to AWS WAF statements
// This allows polymorphic handling of all statement types (Statement, Statement1, Statement2, Statement3, Statement4)
type StatementConverter interface {
	ToWafStatement() (*wafv2types.Statement, error)
}

// Ensure all statement types implement StatementConverter
var (
	_ StatementConverter = (*statementWrapper)(nil)
	_ StatementConverter = (*statement1Wrapper)(nil)
	_ StatementConverter = (*statement2Wrapper)(nil)
	_ StatementConverter = (*statement3Wrapper)(nil)
	_ StatementConverter = (*statement4Wrapper)(nil)
)

// ===== Level 0: Root Statement (has RateBased + ManagedRuleGroup) =====

type statementWrapper struct {
	stmt cloudresourcesv1beta1.AwsWebAclStatement
}

func (w *statementWrapper) ToWafStatement() (*wafv2types.Statement, error) {
	result := &wafv2types.Statement{}

	// Count how many statement types are set
	count := 0

	// Root-level specific: RateBased and ManagedRuleGroup
	if w.stmt.RateBased != nil {
		count++
	}
	if w.stmt.ManagedRuleGroup != nil {
		count++
	}

	// Logical operators
	if w.stmt.AndStatement != nil {
		count++
	}
	if w.stmt.OrStatement != nil {
		count++
	}
	if w.stmt.NotStatement != nil {
		count++
	}

	// Leaf statements
	if w.stmt.GeoMatch != nil {
		count++
	}
	if w.stmt.ByteMatch != nil {
		count++
	}
	if w.stmt.LabelMatch != nil {
		count++
	}
	if w.stmt.SizeConstraint != nil {
		count++
	}
	if w.stmt.SqliMatch != nil {
		count++
	}
	if w.stmt.XssMatch != nil {
		count++
	}
	if w.stmt.RegexMatch != nil {
		count++
	}
	if w.stmt.AsnMatch != nil {
		count++
	}

	// Validate exactly one is set
	if count == 0 {
		return nil, fmt.Errorf("statement must have exactly one condition set")
	}
	if count > 1 {
		return nil, fmt.Errorf("statement must have exactly one condition set, found %d", count)
	}

	// Now convert the single statement that was set
	if w.stmt.RateBased != nil {
		rateStmt, err := convertRateBasedStatement(w.stmt.RateBased)
		if err != nil {
			return nil, err
		}
		result.RateBasedStatement = rateStmt
		return result, nil
	}

	if w.stmt.ManagedRuleGroup != nil {
		managedStmt, err := convertManagedRuleGroupStatement(w.stmt.ManagedRuleGroup)
		if err != nil {
			return nil, err
		}
		result.ManagedRuleGroupStatement = managedStmt
		return result, nil
	}

	// Logical operators
	if w.stmt.AndStatement != nil {
		andStmt, err := convertLogicalAnd(w.stmt.AndStatement.Statements, func(s cloudresourcesv1beta1.AwsWebAclStatement1) StatementConverter {
			return WrapStatement1(s)
		})
		if err != nil {
			return nil, err
		}
		result.AndStatement = andStmt
		return result, nil
	}

	if w.stmt.OrStatement != nil {
		orStmt, err := convertLogicalOr(w.stmt.OrStatement.Statements, func(s cloudresourcesv1beta1.AwsWebAclStatement1) StatementConverter {
			return WrapStatement1(s)
		})
		if err != nil {
			return nil, err
		}
		result.OrStatement = orStmt
		return result, nil
	}

	if w.stmt.NotStatement != nil {
		notStmt, err := convertLogicalNot(w.stmt.NotStatement.Statement, WrapStatement1)
		if err != nil {
			return nil, err
		}
		result.NotStatement = notStmt
		return result, nil
	}

	// Leaf statements (shared logic via helper) - no need to check count again
	return convertLeafStatements(result, leafStatements{
		GeoMatch:       w.stmt.GeoMatch,
		ByteMatch:      w.stmt.ByteMatch,
		LabelMatch:     w.stmt.LabelMatch,
		SizeConstraint: w.stmt.SizeConstraint,
		SqliMatch:      w.stmt.SqliMatch,
		XssMatch:       w.stmt.XssMatch,
		RegexMatch:     w.stmt.RegexMatch,
		AsnMatch:       w.stmt.AsnMatch,
	})
}

// WrapStatement wraps a Level 0 statement
func WrapStatement(stmt cloudresourcesv1beta1.AwsWebAclStatement) StatementConverter {
	return &statementWrapper{stmt: stmt}
}

// ===== Level 1: Statement1 (no RateBased, no ManagedRuleGroup) =====

type statement1Wrapper struct {
	stmt cloudresourcesv1beta1.AwsWebAclStatement1
}

func (w *statement1Wrapper) ToWafStatement() (*wafv2types.Statement, error) {
	result := &wafv2types.Statement{}

	// Count how many statement types are set
	count := 0
	if w.stmt.AndStatement != nil {
		count++
	}
	if w.stmt.OrStatement != nil {
		count++
	}
	if w.stmt.NotStatement != nil {
		count++
	}
	if w.stmt.GeoMatch != nil {
		count++
	}
	if w.stmt.ByteMatch != nil {
		count++
	}
	if w.stmt.LabelMatch != nil {
		count++
	}
	if w.stmt.SizeConstraint != nil {
		count++
	}
	if w.stmt.SqliMatch != nil {
		count++
	}
	if w.stmt.XssMatch != nil {
		count++
	}
	if w.stmt.RegexMatch != nil {
		count++
	}
	if w.stmt.AsnMatch != nil {
		count++
	}

	// Validate exactly one is set
	if count == 0 {
		return nil, fmt.Errorf("level 1 statement must have exactly one condition set")
	}
	if count > 1 {
		return nil, fmt.Errorf("level 1 statement must have exactly one condition set, found %d", count)
	}

	// Logical operators
	if w.stmt.AndStatement != nil {
		andStmt, err := convertLogicalAnd(w.stmt.AndStatement.Statements, func(s cloudresourcesv1beta1.AwsWebAclStatement2) StatementConverter {
			return WrapStatement2(s)
		})
		if err != nil {
			return nil, err
		}
		result.AndStatement = andStmt
		return result, nil
	}

	if w.stmt.OrStatement != nil {
		orStmt, err := convertLogicalOr(w.stmt.OrStatement.Statements, func(s cloudresourcesv1beta1.AwsWebAclStatement2) StatementConverter {
			return WrapStatement2(s)
		})
		if err != nil {
			return nil, err
		}
		result.OrStatement = orStmt
		return result, nil
	}

	if w.stmt.NotStatement != nil {
		notStmt, err := convertLogicalNot(w.stmt.NotStatement.Statement, WrapStatement2)
		if err != nil {
			return nil, err
		}
		result.NotStatement = notStmt
		return result, nil
	}

	// Leaf statements
	return convertLeafStatements(result, leafStatements{
		GeoMatch:       w.stmt.GeoMatch,
		ByteMatch:      w.stmt.ByteMatch,
		LabelMatch:     w.stmt.LabelMatch,
		SizeConstraint: w.stmt.SizeConstraint,
		SqliMatch:      w.stmt.SqliMatch,
		XssMatch:       w.stmt.XssMatch,
		RegexMatch:     w.stmt.RegexMatch,
		AsnMatch:       w.stmt.AsnMatch,
	})
}

// WrapStatement1 wraps a Level 1 statement
func WrapStatement1(stmt cloudresourcesv1beta1.AwsWebAclStatement1) StatementConverter {
	return &statement1Wrapper{stmt: stmt}
}

// ===== Level 2: Statement2 =====

type statement2Wrapper struct {
	stmt cloudresourcesv1beta1.AwsWebAclStatement2
}

func (w *statement2Wrapper) ToWafStatement() (*wafv2types.Statement, error) {
	result := &wafv2types.Statement{}

	// Logical operators
	if w.stmt.AndStatement != nil {
		andStmt, err := convertLogicalAnd(w.stmt.AndStatement.Statements, func(s cloudresourcesv1beta1.AwsWebAclStatement3) StatementConverter {
			return WrapStatement3(s)
		})
		if err != nil {
			return nil, err
		}
		result.AndStatement = andStmt
		return result, nil
	}

	if w.stmt.OrStatement != nil {
		orStmt, err := convertLogicalOr(w.stmt.OrStatement.Statements, func(s cloudresourcesv1beta1.AwsWebAclStatement3) StatementConverter {
			return WrapStatement3(s)
		})
		if err != nil {
			return nil, err
		}
		result.OrStatement = orStmt
		return result, nil
	}

	if w.stmt.NotStatement != nil {
		notStmt, err := convertLogicalNot(w.stmt.NotStatement.Statement, WrapStatement3)
		if err != nil {
			return nil, err
		}
		result.NotStatement = notStmt
		return result, nil
	}

	// Leaf statements
	return convertLeafStatements(result, leafStatements{
		GeoMatch:       w.stmt.GeoMatch,
		ByteMatch:      w.stmt.ByteMatch,
		LabelMatch:     w.stmt.LabelMatch,
		SizeConstraint: w.stmt.SizeConstraint,
		SqliMatch:      w.stmt.SqliMatch,
		XssMatch:       w.stmt.XssMatch,
		RegexMatch:     w.stmt.RegexMatch,
		AsnMatch:       w.stmt.AsnMatch,
	})
}

// WrapStatement2 wraps a Level 2 statement
func WrapStatement2(stmt cloudresourcesv1beta1.AwsWebAclStatement2) StatementConverter {
	return &statement2Wrapper{stmt: stmt}
}

// ===== Level 3: Statement3 =====

type statement3Wrapper struct {
	stmt cloudresourcesv1beta1.AwsWebAclStatement3
}

func (w *statement3Wrapper) ToWafStatement() (*wafv2types.Statement, error) {
	result := &wafv2types.Statement{}

	// Logical operators
	if w.stmt.AndStatement != nil {
		andStmt, err := convertLogicalAnd(w.stmt.AndStatement.Statements, func(s cloudresourcesv1beta1.AwsWebAclStatement4) StatementConverter {
			return WrapStatement4(s)
		})
		if err != nil {
			return nil, err
		}
		result.AndStatement = andStmt
		return result, nil
	}

	if w.stmt.OrStatement != nil {
		orStmt, err := convertLogicalOr(w.stmt.OrStatement.Statements, func(s cloudresourcesv1beta1.AwsWebAclStatement4) StatementConverter {
			return WrapStatement4(s)
		})
		if err != nil {
			return nil, err
		}
		result.OrStatement = orStmt
		return result, nil
	}

	if w.stmt.NotStatement != nil {
		notStmt, err := convertLogicalNot(w.stmt.NotStatement.Statement, WrapStatement4)
		if err != nil {
			return nil, err
		}
		result.NotStatement = notStmt
		return result, nil
	}

	// Leaf statements
	return convertLeafStatements(result, leafStatements{
		GeoMatch:       w.stmt.GeoMatch,
		ByteMatch:      w.stmt.ByteMatch,
		LabelMatch:     w.stmt.LabelMatch,
		SizeConstraint: w.stmt.SizeConstraint,
		SqliMatch:      w.stmt.SqliMatch,
		XssMatch:       w.stmt.XssMatch,
		RegexMatch:     w.stmt.RegexMatch,
		AsnMatch:       w.stmt.AsnMatch,
	})
}

// WrapStatement3 wraps a Level 3 statement
func WrapStatement3(stmt cloudresourcesv1beta1.AwsWebAclStatement3) StatementConverter {
	return &statement3Wrapper{stmt: stmt}
}

// ===== Level 4: Statement4 (leaf only, no logical operators) =====

type statement4Wrapper struct {
	stmt cloudresourcesv1beta1.AwsWebAclStatement4
}

func (w *statement4Wrapper) ToWafStatement() (*wafv2types.Statement, error) {
	result := &wafv2types.Statement{}

	// NO logical operators at Level 4 - leaf statements only
	return convertLeafStatements(result, leafStatements{
		GeoMatch:       w.stmt.GeoMatch,
		ByteMatch:      w.stmt.ByteMatch,
		LabelMatch:     w.stmt.LabelMatch,
		SizeConstraint: w.stmt.SizeConstraint,
		SqliMatch:      w.stmt.SqliMatch,
		XssMatch:       w.stmt.XssMatch,
		RegexMatch:     w.stmt.RegexMatch,
		AsnMatch:       w.stmt.AsnMatch,
	})
}

// WrapStatement4 wraps a Level 4 statement
func WrapStatement4(stmt cloudresourcesv1beta1.AwsWebAclStatement4) StatementConverter {
	return &statement4Wrapper{stmt: stmt}
}

// ===== Generic Logical Operator Converters (using generics) =====

// convertLogicalAnd is a generic converter for AND statements at any level
func convertLogicalAnd[T any](statements []T, wrapper func(T) StatementConverter) (*wafv2types.AndStatement, error) {
	if len(statements) < 2 {
		return nil, fmt.Errorf("and statement requires at least 2 nested statements")
	}

	result := make([]wafv2types.Statement, 0, len(statements))
	for i, stmt := range statements {
		converter := wrapper(stmt)
		wafStmt, err := converter.ToWafStatement()
		if err != nil {
			return nil, fmt.Errorf("error converting and statement[%d]: %w", i, err)
		}
		result = append(result, *wafStmt)
	}

	return &wafv2types.AndStatement{
		Statements: result,
	}, nil
}

// convertLogicalOr is a generic converter for OR statements at any level
func convertLogicalOr[T any](statements []T, wrapper func(T) StatementConverter) (*wafv2types.OrStatement, error) {
	if len(statements) < 2 {
		return nil, fmt.Errorf("or statement requires at least 2 nested statements")
	}

	result := make([]wafv2types.Statement, 0, len(statements))
	for i, stmt := range statements {
		converter := wrapper(stmt)
		wafStmt, err := converter.ToWafStatement()
		if err != nil {
			return nil, fmt.Errorf("error converting or statement[%d]: %w", i, err)
		}
		result = append(result, *wafStmt)
	}

	return &wafv2types.OrStatement{
		Statements: result,
	}, nil
}

// convertLogicalNot is a generic converter for NOT statements at any level
func convertLogicalNot[T any](statement T, wrapper func(T) StatementConverter) (*wafv2types.NotStatement, error) {
	converter := wrapper(statement)
	wafStmt, err := converter.ToWafStatement()
	if err != nil {
		return nil, fmt.Errorf("error converting not statement: %w", err)
	}

	return &wafv2types.NotStatement{
		Statement: wafStmt,
	}, nil
}

// ===== Leaf Statement Helpers =====

// leafStatements is a helper struct to pass leaf statement pointers
type leafStatements struct {
	GeoMatch       *cloudresourcesv1beta1.AwsWebAclGeoMatchStatement
	ByteMatch      *cloudresourcesv1beta1.AwsWebAclByteMatchStatement
	LabelMatch     *cloudresourcesv1beta1.AwsWebAclLabelMatchStatement
	SizeConstraint *cloudresourcesv1beta1.AwsWebAclSizeConstraintStatement
	SqliMatch      *cloudresourcesv1beta1.AwsWebAclSqliMatchStatement
	XssMatch       *cloudresourcesv1beta1.AwsWebAclXssMatchStatement
	RegexMatch     *cloudresourcesv1beta1.AwsWebAclRegexMatchStatement
	AsnMatch       *cloudresourcesv1beta1.AwsWebAclAsnMatchStatement
}

// convertLeafStatements handles all leaf statement conversions
// Assumes validation has already been done by the wrapper (exactly one is set)
// Returns the statement with the single leaf converted
func convertLeafStatements(result *wafv2types.Statement, leaves leafStatements) (*wafv2types.Statement, error) {
	if leaves.GeoMatch != nil {
		geoStmt, err := convertGeoMatchStatement(leaves.GeoMatch)
		if err != nil {
			return nil, err
		}
		result.GeoMatchStatement = geoStmt
		return result, nil
	}

	if leaves.ByteMatch != nil {
		byteStmt, err := convertByteMatchStatement(leaves.ByteMatch)
		if err != nil {
			return nil, err
		}
		result.ByteMatchStatement = byteStmt
		return result, nil
	}

	if leaves.LabelMatch != nil {
		labelStmt := convertLabelMatchStatement(leaves.LabelMatch)
		result.LabelMatchStatement = labelStmt
		return result, nil
	}

	if leaves.SizeConstraint != nil {
		sizeStmt, err := convertSizeConstraintStatement(leaves.SizeConstraint)
		if err != nil {
			return nil, err
		}
		result.SizeConstraintStatement = sizeStmt
		return result, nil
	}

	if leaves.SqliMatch != nil {
		sqliStmt, err := convertSqliMatchStatement(leaves.SqliMatch)
		if err != nil {
			return nil, err
		}
		result.SqliMatchStatement = sqliStmt
		return result, nil
	}

	if leaves.XssMatch != nil {
		xssStmt, err := convertXssMatchStatement(leaves.XssMatch)
		if err != nil {
			return nil, err
		}
		result.XssMatchStatement = xssStmt
		return result, nil
	}

	if leaves.RegexMatch != nil {
		regexStmt, err := convertRegexMatchStatement(leaves.RegexMatch)
		if err != nil {
			return nil, err
		}
		result.RegexMatchStatement = regexStmt
		return result, nil
	}

	if leaves.AsnMatch != nil {
		asnStmt, err := convertAsnMatchStatement(leaves.AsnMatch)
		if err != nil {
			return nil, err
		}
		result.AsnMatchStatement = asnStmt
		return result, nil
	}

	// Should never reach here if validation was done correctly
	return nil, fmt.Errorf("no leaf statement set (internal error)")
}

// ===== Public API =====
// Note: The main convertStatement() function remains in util.go for backward compatibility
// It can optionally be updated to call WrapStatement(stmt).ToWafStatement()
