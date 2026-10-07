# WAF Research - Background Analysis

This directory contains cross-provider research that informed the WAF API design.

## Research Files

### [rule-granularity-comparison.md](rule-granularity-comparison.md)
**What it proves**: Boolean logic capabilities differ significantly across providers

- **AWS WAFv2**: Full nested AND/OR/NOT statements, unlimited nesting
- **GCP Cloud Armor**: CEL expressions with AND/OR/NOT, ~2048 char limit
- **Azure WAF**: Flat AND-only conditions, OR requires multiple rules

**Impact on API design**: WafConfiguration uses hybrid model (flat AND + limited OR via `anyOf`) to avoid Azure rule explosion while supporting common use cases.

---

### [label-based-chaining-cross-provider.md](label-based-chaining-cross-provider.md)
**What it proves**: Label chaining is AWS-native, requires workarounds elsewhere

- **AWS WAFv2**: Native `LabelMatchStatement` + `RuleLabels` (optimal performance)
- **Azure WAF**: No native support, workaround via duplicated custom rules
- **GCP Cloud Armor**: No native support, workaround via priority-based bypass or CEL composition

**Impact on API design**: WafConfiguration includes `classifications` feature for reusable condition sets. Controller translates to:
- AWS: Native labels (best performance)
- Azure: Auto-generated duplicated rules
- GCP: Priority-based bypass rules

---

### [conditional-override-conditions-cross-provider.md](conditional-override-conditions-cross-provider.md)
**What it proves**: Core condition types (path, headers, IP, query params) are universally supported

- **17 condition types analyzed** across AWS/Azure/GCP
- Most core conditions have native support on all providers
- Some advanced conditions (geo, cookies, time-based) have limitations

**Impact on API design**: WafConfiguration Phase 2.1 focuses on universally-supported conditions. Advanced conditions (geo, time-based) deferred to Phase 2.2+.

---

### [managed-rules-structure-comparison.md](managed-rules-structure-comparison.md)
**What it proves**: Managed rule group structures differ but can be abstracted

- **AWS WAFv2**: Nested 2 levels deep (Rules → Statement → ManagedRuleGroupStatement)
- **Azure WAF**: Flat array (`managedRuleSets[]`)
- **GCP Cloud Armor**: Flat array (`rules[]`) with CEL expressions

**Impact on API design**: WafConfiguration uses flat `managedRuleGroups[]` array. Controller handles provider-specific nesting/structure translation.

---

## How Research Informed API Design

### Phase 1 Decision: Start with WafPolicy
- Research showed significant provider differences in advanced features
- Terraform/Crossplane use provider-specific approach successfully
- **Decision**: Ship WafPolicy (provider-specific) first, add portable layer after validation

### Phase 2 Design: WafConfiguration
- Research identified universally-supported features (rule overrides, basic conditions)
- Research identified AWS-specific optimizations (label chaining)
- Research identified Azure limitations (flat AND-only, 100 rule limit)
- **Decision**: Base + overrides pattern, hybrid boolean logic, classifications for label chaining

### Feature Prioritization
- **Phase 2.1**: Rule overrides, conditional overrides, classifications (validated across all 3 providers)
- **Phase 2.2**: Geographic conditions, time-based rules, advanced rate limiting (AWS/GCP only, Azure limitations)
- **Phase 2.3**: Custom rules, A/B testing, dynamic updates (requires more analysis)

---

## Cross-References

- See [../api-specification.md](../api-specification.md) for the final API design
- See [../design-rationale.md](../design-rationale.md) for why we chose this approach
- See [../use-cases/](../use-cases/) for provider capability validation per use case

---

## Research Methodology

Each research document follows this structure:

1. **Provider capability analysis** - What each provider supports natively
2. **Structural comparison** - How providers model the same concept
3. **Translation strategies** - How WafConfiguration controller bridges differences
4. **Limitations and workarounds** - What works, what doesn't, what requires compromise
5. **Recommendations** - What to include in portable API, what to defer

This research ensures the WafConfiguration API is grounded in reality, not wishful thinking about provider portability.
