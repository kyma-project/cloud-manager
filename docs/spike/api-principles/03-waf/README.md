# Cloud Manager WAF — Spike Findings

## Purpose

WAF is the test case for spike bullet #3: *apply the pattern from step 2 and check whether it holds*.

The question is not "what should the WAF API look like?" It is: **does the [progressive API design model](../02-industry-patterns/progressive-api-design/progressive-api-design.md) hold for WAF, and what does that tell us about Cloud Manager API design principles in general?**

---

## How the Progressive API Layers Map to WAF

| Layer | Progressive model | WAF implementation |
|-------|------------------|--------------------|
| **Layer 0** | Cloud detection — controller concern, no user-facing field | No `spec.cloud` or `spec.provider` field. Cloud is read from Kyma Scope at runtime. |
| **Layer 1** | Zero-config intent — neutral vocabulary, same on all clouds | **Deferred.** `WafConfiguration.spec.intents` was designed for this role but is not yet implementable — the fragment-merge model it depends on has no clean cross-provider solution, and the intent vocabulary lacks a defined override path. See [waf-configuration-design.md](waf-configuration-design.md). |
| **Layer 2** | Neutral tuning — optional shared fields, same semantics on all clouds | **Deliberately absent.** No WAF tuning concept has identical semantics on all three providers. See [design-rationale.md](design-rationale.md). |
| **Layer 3 / unstructured payload (P6)** | Provider escape hatch — unstructured payload, controller-validated only | `WafPolicy.spec.data` carries complete provider-native JSON inline. No typed sub-struct. Currently the only specified entry point. |

The resulting resource flow (current state — Layer 1 deferred):

```
AppLoadBalancer
    ↓ references
WafPolicy        (Layer 3 unstructured payload — complete provider-native JSON inline in spec.data)
    ↓ Cloud Manager SKR reconciler
Cloud WAF (AWS WAFv2 / Azure Front Door WAF / GCP Cloud Armor)
```

---

## Findings

### What the model validates

| Feature | AWS | Azure | GCP | Result |
|---------|-----|-------|-----|--------|
| Managed rule groups | ✅ | ✅ | ✅ | Layer 1 via `spec.intents` — deferred; no implementation path yet |
| Managed rules override | ✅ | ✅ | ⚠️ Degrades entire ruleset | Not portable — unstructured payload (P6) only |
| Custom rules (path/header/IP) | ✅ | ✅ | ✅ | Portable concept — unstructured payload (P6) via `WafPolicy.spec.data` |

Path, header, and IP-based custom rules translate cleanly across all three providers. They are not exposed as typed Layer 2 fields because doing so covers only one portable concept while leaving geographic blocking, size filtering, rate limiting, and managed rule tuning all at Layer 3 anyway. A partial Layer 2 adds API surface without eliminating the escape hatch — the cleaner boundary is `spec.intents` for built-in goals (once Layer 1 is implementable) and `WafPolicy.spec.data` for everything else.

### Why Layer 2 is absent for WAF

Layer 2 requires that a field concept genuinely exists on all clouds with the same semantics. No WAF field meets this bar:

- Managed rule group names and bundling are completely provider-specific — no common vocabulary exists across AWS, Azure, and GCP.
- Rule overrides use provider-specific rule IDs and behave differently per provider (GCP silently degrades the entire ruleset instead of the targeted rule).

A field that silently behaves differently per provider is worse than no abstraction. The `spec.intents` vocabulary sidesteps this by operating at the level of user goals, not provider fields.

### Why unstructured payload (P6) and not typed sub-struct (P5) for Layer 3

Provider WAF JSON schemas are large, versioned, and structurally different. A typed sub-struct (`spec.aws`, `spec.gcp`, `spec.azure`) would require Cloud Manager to mirror three provider WAF schemas. The unstructured payload approach — inline JSON in `spec.data` — lets the user write provider-native content directly, with the controller validating only the envelope.

---

## Conclusion for the Spike

The progressive API design model holds for WAF, with two forced adaptations: Layer 2 is empty by design, and Layer 1 is deferred. The `spec.intents` Layer 1 vocabulary was designed to substitute for what would normally be Layer 2 neutral tuning — by operating above the provider API surface entirely rather than mapping across it. However, the intent vocabulary is not yet implementable: the fragment-merge model it depends on has no clean cross-provider solution, and the intents have no defined override path for false positives. Use case collection is a prerequisite before Layer 1 can be designed — see [waf-configuration-design.md](waf-configuration-design.md).

**The finding for the broader API principles question:**

> A portable abstraction is justified when provider capabilities genuinely overlap for the use case. When they do not, the abstraction must operate at a higher level — user intent rather than provider fields — with Cloud Manager owning the translation. A leaky field-level abstraction is worse than no abstraction.

---

## Documentation

**[api-specification.md](api-specification.md)** — The `WafPolicy` API design and design decisions.

**[waf-configuration-design.md](waf-configuration-design.md)** — The `WafConfiguration` intent-based design as explored, and why the fragment-merge model stalls across providers. Deferred from the current spec.

**[design-rationale.md](design-rationale.md)** — How each layer of the progressive API model maps to WAF, and where the model holds.

**[research/](research/)** — Cross-provider analysis of boolean logic, label chaining, condition types, and managed rule structures.
