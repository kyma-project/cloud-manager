# WAF Use Cases - Provider Capability Validation

This directory contains use case validations that test specific WAF scenarios across AWS, Azure, and GCP.

## Purpose

**These are NOT API design documents.**  
The API design is in [../api-specification.md](../api-specification.md).

These files validate:
- ✅ Which providers support specific use cases natively
- ✅ What workarounds are needed per provider
- ✅ Whether WafConfiguration Phase 2 features are feasible

---

## Use Case Index

| Use Case | AWS | Azure | GCP | Phase 2 Feature |
|----------|-----|-------|-----|-----------------|
| [01. Unconditional Rule Override](01-unconditional-rule-override.md) | ✅ Perfect | ✅ Perfect | ⚠️ Limited | `ruleOverrides` |
| [02. Custom Rule with Conditions](02-custom-rule-with-conditions.md) | ✅ Native | ✅ Native | ✅ Native | `customRules` ⭐ |
| [03. Rate Limiting with Path Thresholds](03-rate-limiting-with-path-thresholds.md) | ✅ Native | ✅ Native | ✅ Native | `customRules` with rate limit |
| [04. IP Allowlist/Blocklist](04-ip-allowlist-blocklist.md) | ✅ Native | ✅ Native | ✅ Native | `customRules` |
| [05. Geographic Blocking](05-geographic-blocking.md) | ✅ Native | ❌ Limited | ✅ Native | Phase 2.2+ `geoBlocking` |
| [06. Size-Based Filtering](06-size-based-filtering.md) | ✅ Native | ✅ Native | ⚠️ Limited | Phase 2.2 `sizeLimits` |
| [07. Bot Protection](07-bot-protection.md) | ✅ Native | ✅ Native | ✅ Native | `managedRuleGroups.BotProtection` |

---

## Validation Matrix

### Phase 2.1 Features (Validated ✅)

**Supported across all 3 providers:**
- ✅ Unconditional rule overrides (`ruleOverrides`)
- ✅ Custom rules with conditions (`customRules`) ⭐ Most portable feature
- ✅ Path, header, IP, method conditions
- ✅ Rate limiting with conditions
- ✅ Bot protection via managed rules

**Implementation notes:**
- AWS has best native support (nested conditions, full boolean logic)
- Azure supports AND-only conditions (flat structure)
- GCP uses CEL expressions (most flexible syntax)

### Phase 2.2 Features (Requires Further Analysis ⚠️)

**Partial support:**
- ⚠️ Size-based filtering (AWS/Azure native, GCP limited)
- ❌ Geographic blocking (Azure doesn't support at WAF level)

**Recommendation**: Phase 2.2 should focus on features with universal support. Geographic blocking should be Phase 2.3+ with documented Azure limitation.

---

## How to Read Use Case Files

Each use case file follows this structure:

```markdown
# Use Case N: [Name]

## User Story
"As a [role], I want [goal]..."

## Requirements
1. Requirement 1
2. Requirement 2
...

## Provider Capability Check

### AWS WAFv2
**Status:** ✅/⚠️/❌
**How it works:** [Native mechanism or workaround]
[Example AWS JSON]

### Azure WAF
**Status:** ✅/⚠️/❌
**How it works:** [Native mechanism or workaround]
[Example Azure JSON]

### GCP Cloud Armor
**Status:** ✅/⚠️/❌
**How it works:** [Native mechanism or workaround]
[Example GCP JSON]

## Validation Result
Summary of whether this use case can be supported in Phase 2.
Links to api-specification.md for how it maps to WafConfiguration.
```

---

## Use Cases vs API Design

**Use cases validate provider capabilities.**  
**API design (in api-specification.md) decides how to expose those capabilities.**

Example:
- **Use Case 03** validates that all providers support custom rules with conditions natively
- **API Design** includes `customRules` in Phase 2.1 with full confidence
- **Implementation** (controller) translates to provider-specific format (AWS Statement, Azure matchConditions, GCP CEL)

---

## Cross-References

- See [../api-specification.md](../api-specification.md) for WafConfiguration API design
- See [../research/](../research/) for detailed cross-provider analysis
- See [../implementation-examples.md](../implementation-examples.md) for complete working examples

---

## Adding New Use Cases

To add a new use case:

1. **Describe the user need** (not the API)
2. **List requirements** (what must work)
3. **Validate on each provider** (AWS, Azure, GCP)
4. **Show provider-specific examples** (actual JSON/YAML that works)
5. **Summarize feasibility** (can Phase 2 support this?)
6. **Link to API design** (don't design API here, reference main spec)

Remember: Use cases validate **provider capabilities**, not **API design**.
