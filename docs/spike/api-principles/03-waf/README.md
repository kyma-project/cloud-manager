# Cloud Manager WAF — Spike Findings

## Purpose

WAF is the test case for spike bullet #3: *apply the pattern from step 2 and check whether it holds*.

The question is not "what should the WAF API look like?" It is: **does a two-level portable-intent + provider-specific-passthrough pattern hold for WAF, and what does that tell us about Cloud Manager API design principles in general?**

---

## What Was Tested

A two-level architecture modelled on Kubernetes Gateway API:

```
[SKR cluster]
AppLoadBalancer
    ↓ references
WafConfiguration (portable intent)
    ↓ SKR controller translates
WafPolicy (provider-specific JSON passthrough)
    ↓ remote reconciliation (KCP watches SKR resources)
```

Three specific features were validated across AWS, Azure, and GCP:

| Feature | AWS | Azure | GCP | Portable? |
|---------|-----|-------|-----|-----------|
| Managed rule groups | ✅ | ✅ | ✅ | ❌ Different names, structure, bundling |
| Managed rules override (`ruleOverrides`) | ✅ | ✅ | ⚠️ Degrades entire ruleset | ❌ Not portable |
| Custom rules with conditions (`customRules`) | ✅ | ✅ | ✅ | ✅ Genuinely portable |

---

## Findings

### What works: `customRules`

Path, header, and IP-based custom rules translate cleanly across all three providers. The portable abstraction holds here. A user writing:

```yaml
customRules:
  - name: health-check-bypass
    priority: 10
    action: allow
    conditions:
      path:
        exact: "/health"
```

gets the correct behaviour on AWS, Azure, and GCP with no semantic drift.

### What fails: `ruleOverrides`

Managed rule overrides are not portable:

- The `managedRuleGroup` and `ruleId` fields require **provider-specific names** — an AWS rule ID does not exist on Azure or GCP.
- GCP has no per-rule override mechanism. Any override entry causes the **entire matched ruleset** to degrade to preview mode (count-all), not the specific rule the user intended.
- The field looks portable but silently behaves differently per provider.

### What cannot be abstracted: managed rule groups

Managed rule group names, structure, and bundling are completely different across providers. AWS, Azure, and GCP have no common vocabulary. A portable `managedRuleGroups` field is not viable. Provider-specific presets (`WafPolicy` with `spec.data`) are the correct answer here.

---

## Conclusion for the Spike

The two-level pattern **partially holds** for WAF:

- ✅ It works for the additive, condition-based layer (`customRules`) where provider capabilities genuinely overlap.
- ❌ It breaks down for managed rule tuning (`ruleOverrides`) — provider-specific names and GCP's coarse granularity mean the abstraction leaks.
- ❌ It cannot abstract managed rule groups at all.

**The portable-intent layer is only as strong as the intersection of provider capabilities.** For WAF, that intersection is narrow: custom rules with basic conditions. Everything else requires provider-specific configuration.

This is a useful finding for the broader API principles question:

> A portable abstraction is justified when provider capabilities genuinely overlap for the use case. When they do not, a provider-specific passthrough (`WafPolicy`-style) with curated presets is more honest and more maintainable than a leaky abstraction.

---

## Documentation

**[api-specification.md](api-specification.md)** — The two-level API design, including the `ruleOverrides` portability failure and open design questions.

**[design-rationale.md](design-rationale.md)** — Why this architecture was explored, what it gets right, and where it breaks down.

**[implementation-examples.md](implementation-examples.md)** — Full provider translations for the three main use cases (managed rules override, custom rules with conditions, IP allowlist/blocklist).

**[use-cases/](use-cases/)** — Per-use-case provider capability validation (8 use cases, AWS/Azure/GCP).

**[research/](research/)** — Cross-provider analysis of boolean logic, label chaining, condition types, and managed rule structures.
