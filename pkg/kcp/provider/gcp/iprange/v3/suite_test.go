package v3

import (
	"context"
	"testing"

	"github.com/kyma-project/cloud-manager/pkg/feature"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGcpIpRange(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GCP IpRange Suite")
}

// Helper to create context with feature flag for testing
// useRefactored: true = use NEW refactored implementation, false = use v2 legacy implementation
func contextWithIpRangeRefactoredFlag(ctx context.Context, useRefactored bool) context.Context {
	ffCtx := feature.ContextBuilderFromCtx(ctx).Build(ctx)

	// Override the feature flag value
	if useRefactored {
		ffCtx = feature.ContextBuilderFromCtx(ctx).Custom("ipRangeRefactored", true).Build(ctx)
	} else {
		ffCtx = feature.ContextBuilderFromCtx(ctx).Custom("ipRangeRefactored", false).Build(ctx)
	}

	return ffCtx
}

// Helper to get context with legacy implementation (v2)
func contextWithLegacy(ctx context.Context) context.Context {
	return contextWithIpRangeRefactoredFlag(ctx, false)
}

// Helper to get context with refactored implementation
func contextWithRefactored(ctx context.Context) context.Context {
	return contextWithIpRangeRefactoredFlag(ctx, true)
}

// Implementation type for parameterized tests
type IpRangeImplementation string

const (
	ImplementationLegacy     IpRangeImplementation = "legacy"
	ImplementationRefactored IpRangeImplementation = "refactored"
)

// Helper to get context for implementation type
func contextForImplementation(ctx context.Context, impl IpRangeImplementation) context.Context {
	switch impl {
	case ImplementationLegacy:
		return contextWithLegacy(ctx)
	case ImplementationRefactored:
		return contextWithRefactored(ctx)
	default:
		return ctx
	}
}
