package gardener

import gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"

// transientShootErrorCodes lists the only two Gardener error codes that are truly auto-recovering.
// All other codes indicate user misconfiguration or permanent infrastructure failures.
// Source: Gardener dashboard frontend/src/utils/errorCodes.js (temporaryError flag).
var transientShootErrorCodes = map[gardenertypes.ErrorCode]bool{
	gardenertypes.ErrorInfraRateLimitsExceeded:    true, // ERR_INFRA_RATE_LIMITS_EXCEEDED
	gardenertypes.ErrorRetryableInfraDependencies: true, // ERR_RETRYABLE_INFRA_DEPENDENCIES
}

// IsTransientShootErrors returns true when all present error codes are known-transient and
// Gardener is expected to recover without any manual action.
// An empty codes list within a LastError is treated as transient (Gardener retrying, code not yet written).
// An empty lastErrors slice is also transient (no errors).
func IsTransientShootErrors(lastErrors []gardenertypes.LastError) bool {
	for _, le := range lastErrors {
		for _, code := range le.Codes {
			if !transientShootErrorCodes[code] {
				return false
			}
		}
	}
	return true
}

// IsTerminalShootLastOperation returns true when Gardener has exhausted its automatic retry window
// and will not attempt the operation again without manual intervention.
func IsTerminalShootLastOperation(op *gardenertypes.LastOperation) bool {
	return op != nil && op.State == gardenertypes.LastOperationStateFailed
}
