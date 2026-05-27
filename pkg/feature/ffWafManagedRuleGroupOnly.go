package feature

import (
	"context"
)

const wafManagedRuleGroupOnlyFlagName = "wafManagedRuleGroupOnly"

var WafManagedRuleGroupOnly = &wafManagedRuleGroupOnly{}

type wafManagedRuleGroupOnly struct{}

func (f *wafManagedRuleGroupOnly) Value(ctx context.Context) bool {
	v := provider.BoolVariation(ctx, wafManagedRuleGroupOnlyFlagName, false)
	return v
}
