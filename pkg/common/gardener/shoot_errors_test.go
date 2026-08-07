package gardener

import (
	"testing"

	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestIsTransientShootErrors(t *testing.T) {
	tests := []struct {
		name       string
		lastErrors []gardenertypes.LastError
		want       bool
	}{
		{
			name:       "empty slice is transient",
			lastErrors: nil,
			want:       true,
		},
		{
			name: "no codes is transient",
			lastErrors: []gardenertypes.LastError{
				{Description: "some error", Codes: nil},
			},
			want: true,
		},
		{
			name: "rate limits exceeded is transient",
			lastErrors: []gardenertypes.LastError{
				{Codes: []gardenertypes.ErrorCode{gardenertypes.ErrorInfraRateLimitsExceeded}},
			},
			want: true,
		},
		{
			name: "retryable infra dependencies is transient",
			lastErrors: []gardenertypes.LastError{
				{Codes: []gardenertypes.ErrorCode{gardenertypes.ErrorRetryableInfraDependencies}},
			},
			want: true,
		},
		{
			name: "both transient codes is transient",
			lastErrors: []gardenertypes.LastError{
				{Codes: []gardenertypes.ErrorCode{
					gardenertypes.ErrorInfraRateLimitsExceeded,
					gardenertypes.ErrorRetryableInfraDependencies,
				}},
			},
			want: true,
		},
		{
			name: "quota exceeded is not transient",
			lastErrors: []gardenertypes.LastError{
				{Codes: []gardenertypes.ErrorCode{gardenertypes.ErrorInfraQuotaExceeded}},
			},
			want: false,
		},
		{
			name: "unauthenticated is not transient",
			lastErrors: []gardenertypes.LastError{
				{Codes: []gardenertypes.ErrorCode{gardenertypes.ErrorInfraUnauthenticated}},
			},
			want: false,
		},
		{
			name: "unauthorized is not transient",
			lastErrors: []gardenertypes.LastError{
				{Codes: []gardenertypes.ErrorCode{gardenertypes.ErrorInfraUnauthorized}},
			},
			want: false,
		},
		{
			name: "configuration problem is not transient",
			lastErrors: []gardenertypes.LastError{
				{Codes: []gardenertypes.ErrorCode{gardenertypes.ErrorConfigurationProblem}},
			},
			want: false,
		},
		{
			name: "retryable configuration problem is not transient",
			lastErrors: []gardenertypes.LastError{
				{Codes: []gardenertypes.ErrorCode{gardenertypes.ErrorRetryableConfigurationProblem}},
			},
			want: false,
		},
		{
			name: "problematic webhook is not transient",
			lastErrors: []gardenertypes.LastError{
				{Codes: []gardenertypes.ErrorCode{gardenertypes.ErrorProblematicWebhook}},
			},
			want: false,
		},
		{
			name: "mixed transient and non-transient codes is not transient",
			lastErrors: []gardenertypes.LastError{
				{Codes: []gardenertypes.ErrorCode{
					gardenertypes.ErrorInfraRateLimitsExceeded,
					gardenertypes.ErrorInfraQuotaExceeded,
				}},
			},
			want: false,
		},
		{
			name: "non-transient code in second LastError is not transient",
			lastErrors: []gardenertypes.LastError{
				{Codes: []gardenertypes.ErrorCode{gardenertypes.ErrorInfraRateLimitsExceeded}},
				{Codes: []gardenertypes.ErrorCode{gardenertypes.ErrorConfigurationProblem}},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsTransientShootErrors(tt.lastErrors))
		})
	}
}

func TestIsTerminalShootLastOperation(t *testing.T) {
	tests := []struct {
		name string
		op   *gardenertypes.LastOperation
		want bool
	}{
		{
			name: "nil operation is not terminal",
			op:   nil,
			want: false,
		},
		{
			name: "Processing is not terminal",
			op:   &gardenertypes.LastOperation{State: gardenertypes.LastOperationStateProcessing},
			want: false,
		},
		{
			name: "Error is not terminal (Gardener retrying)",
			op:   &gardenertypes.LastOperation{State: gardenertypes.LastOperationStateError},
			want: false,
		},
		{
			name: "Succeeded is not terminal",
			op:   &gardenertypes.LastOperation{State: gardenertypes.LastOperationStateSucceeded},
			want: false,
		},
		{
			name: "Pending is not terminal",
			op:   &gardenertypes.LastOperation{State: gardenertypes.LastOperationStatePending},
			want: false,
		},
		{
			name: "Failed is terminal",
			op:   &gardenertypes.LastOperation{State: gardenertypes.LastOperationStateFailed},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsTerminalShootLastOperation(tt.op))
		})
	}
}
