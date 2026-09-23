# WAF Spike — Design Rationale and Findings

## What This Document Is

This is the reasoning behind the two-level WAF architecture that was explored as the spike test case. It records what was tried, what worked, what did not, and what the findings mean for Cloud Manager API design principles.

---

## The Question Being Tested

Can a Gateway API-style two-level architecture — portable intent resource + provider-specific passthrough — work for WAF?

```
WafConfiguration (portable intent)
    ↓ SKR controller translates
WafPolicy (provider-specific JSON passthrough)
    ↓ KCP remote reconciler provisions
Cloud WAF (AWS WAFv2 / Azure Front Door WAF / GCP Cloud Armor)
```

The analogy to Kubernetes Gateway API:
- `Gateway` / `HTTPRoute` (portable intent) → `WafConfiguration`
- `GatewayClass` (provider-specific) → `WafPolicy`

---

## What the Architecture Gets Right

### `WafPolicy` — provider-specific passthrough

`WafPolicy` with a single `spec.data` field (raw provider JSON) is the right answer for full control. It is honest, simple, and follows the same pattern as Terraform and Crossplane. No translation, no leaky abstraction.

Pre-deployed presets (`owasp-basic`, `owasp-moderate`, `owasp-strict`, `owasp-detection`) shipped as `WafPolicy` resources give users a curated starting point without hiding what they contain.

### `customRules` — the portable layer that works

Custom rules with path, header, and IP conditions are genuinely portable. All three providers support the same semantics:

| Condition | AWS | Azure | GCP |
|-----------|-----|-------|-----|
| Path exact | `ByteMatchStatement EXACTLY` | `Equals` | `request.path == '/x'` |
| Path prefix | `ByteMatchStatement STARTS_WITH` | `BeginsWith` | `request.path.matches('/x.*')` |
| Header exists | `SizeConstraintStatement >= 1` | `Exists` | `has(request.headers['x'])` |
| IP CIDR | `IPSetReferenceStatement` | `IPMatch` | `inIpRange(origin.ip, 'cidr')` |
| Negate | `NotStatement` | `negationConditon: true` | `!condition` |
| AND logic | `AndStatement` | array of `matchConditions` | `&&` in CEL |

The abstraction holds. A `customRules` entry means the same thing on every provider.

---

## Where the Architecture Breaks Down

### `ruleOverrides` — the portable layer that does not work

`ruleOverrides` was included as a way to tune managed rules (suppress false positives). It does not work as a portable abstraction:

1. **Provider-specific names**: `managedRuleGroup` and `ruleId` are provider-specific identifiers. `AWSManagedRulesCommonRuleSet` / `SizeRestrictions_BODY` does not exist on Azure or GCP. A configuration written for AWS is silently meaningless on other providers.

2. **GCP cannot implement it**: GCP Cloud Armor has no per-rule override. Any `ruleOverrides` entry degrades the **entire matched ruleset** to `preview: true` (count-all). This is a fundamentally different behaviour from what the user expressed — and it is silent.

3. **The field looks portable but is not**: Placing provider-specific identifiers inside a "portable" resource is worse than having no portable resource at all — it gives users false confidence.

**Conclusion**: `ruleOverrides` in `WafConfiguration` is a leaky abstraction. It should either be removed from `WafConfiguration` (users who need rule tuning drop to `WafPolicy`) or the field should not exist in a portable resource at all.

### Managed rule groups — cannot be abstracted

AWS, Azure, and GCP have completely different managed rule group names, bundling, and versioning. There is no portable vocabulary. The right answer is provider-specific presets in `WafPolicy`.

---

## The Key Finding for API Principles

> **A portable abstraction is only valid at the intersection of provider capabilities. When that intersection is narrow, a provider-specific passthrough with curated presets is more honest and more maintainable.**

For WAF, the intersection is: *custom rules with basic conditions*. Everything else — managed rule selection, rule-level tuning — is provider-specific.

This maps directly to spike output #4: *where provider-specific APIs stay the right answer*. WAF is a case where the portable layer is justified only for one narrow feature (`customRules`), and even the two-level architecture is questionable if `customRules` could be expressed more simply (e.g., directly on `AppLoadBalancer` or on `WafPolicy`).

---

## Open Questions Left by the Spike

These are not answered by the spike — they are what the spike surfaces for the broader API principles work:

1. **Is a dedicated portable resource worth it for one feature?** If `customRules` is the only genuinely portable capability in `WafConfiguration`, is a second resource type justified? Or should `customRules` live directly on `WafPolicy` or `AppLoadBalancer`?

2. **Where does rule tuning belong?** `ruleOverrides` is provider-specific. It belongs in `WafPolicy`. But then `WafPolicy` grows structured fields — is it still a passthrough, or does it become a structured provider-specific resource?

3. **Does this pattern generalise?** For Redis or VPC Peering, the provider capability overlap is higher (most Redis config fields mean the same thing everywhere). The portable-intent layer may be more justified there than for WAF.

---

## What Was Validated

| Feature | Portable? | Notes |
|---------|-----------|-------|
| Managed rule groups | ❌ | Provider names/structure incompatible |
| Managed rules override | ❌ | Provider-specific names; GCP degrades entire ruleset |
| Custom rules (path/header/IP) | ✅ | Clean translation on all providers |
| Rate limiting | ✅ | Universal support |
| IP allowlist/blocklist | ✅ | Universal support |
| Geographic blocking | ⚠️ | Azure lacks WAF-level geo support |
| Size-based filtering | ⚠️ | GCP limited |
| Bot protection | ✅ | Universal support |
