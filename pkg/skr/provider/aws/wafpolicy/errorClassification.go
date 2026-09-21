package wafpolicy

import (
	"errors"

	"github.com/aws/smithy-go"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
)

// isConfigurationError determines if an AWS error is a user-actionable configuration error
// that should be reported as ConfigurationError (False/ConfigurationError) rather than
// retried or treated as a terminal failure.
func isConfigurationError(err error) bool {
	if err == nil {
		return false
	}

	// Check for IAM permission errors - these are configuration issues
	if awsmeta.IsAccessDenied(err) || awsmeta.IsUnauthorized(err) {
		return true
	}

	// Check for specific AWS API error codes that indicate configuration issues
	var apiError smithy.APIError
	if errors.As(err, &apiError) {
		switch apiError.ErrorCode() {
		// WAF validation errors - invalid rule structure, invalid parameters
		case "WAFInvalidParameterException",
			"WAFInvalidOperationException",
			"WAFInvalidResourceException",
			"WAFNonexistentItemException",
			"WAFUnavailableEntityException",
			"WAFInvalidPermissionPolicyException",
			"WAFTagOperationException",
			"WAFTagOperationInternalErrorException":
			return true

		// Quota/limit errors that user needs to address
		case "WAFLimitsExceededException",
			"WAFSubscriptionNotFoundException",
			"WAFDuplicateItemException":
			return true

		// General validation errors
		case "ValidationException",
			"InvalidParameterValue",
			"InvalidParameter",
			"MalformedInput":
			return true
		}
	}

	return false
}
