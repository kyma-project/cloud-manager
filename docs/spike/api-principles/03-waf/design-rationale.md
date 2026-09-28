# WAF Spike — Design Rationale and Findings

## What This Document Is

This is the reasoning behind the WAF architecture that this spike validates as the test case for the [progressive API design model](../02-industry-patterns/progressive-api-design/progressive-api-design.md). It explains how each layer of that model maps to WAF, where the model holds, and where the WAF domain forces deliberate choices.

---

## The Question Being Tested

Can a Gateway API-style two-level architecture — portable intent resource + provider-specific passthrough — implement the progressive API design model for WAF?

```
WafConfiguration (Layer 1 intent — deferred, see waf-configuration-design.md)
    ↓ SKR controller translates
WafPolicy (Layer 3 unstructured payload (P6) — stable, currently specified)
    ↓ Cloud Manager SKR reconciler calls cloud provider API
Cloud WAF (AWS WAFv2 / Azure Front Door WAF / GCP Cloud Armor)
```

`WafPolicy` is the currently specified resource. `WafConfiguration` as a Layer 1 intent resource was explored but is deferred — the fragment-merge model it depends on does not have a clean cross-provider solution. See [waf-configuration-design.md](waf-configuration-design.md).

---

## Layer 0 — Cloud Detection

`WafPolicy` has no `spec.cloud`, `spec.provider`, or `compositionSelector` field. The controller reads the cloud provider from Kyma Scope at runtime and routes to the correct provider reconciler. The user writes the same resource on any cluster.

---

## Layer 1 — Zero-Config Intent (Explored, Not Yet Specified)

The intent of `WafConfiguration.spec.intents` was to be the Layer 1 surface — cloud-neutral intent names (`OwaspTop10`, `BotProtection`, `RateLimiting`, `IpReputation`, `OwaspTop10Detection`) that the controller resolves to provider-specific preset policy fragments in `kyma-system`.

A user who writes:

```yaml
kind: WafConfiguration
spec:
  intents:
    - OwaspTop10
    - BotProtection
```

would get correctly provisioned WAF protection on AWS, Azure, and GCP without knowing any provider-specific rule names. `OwaspTop10` on AWS becomes `AWSManagedRulesCommonRuleSet` + `AWSManagedRulesSQLiRuleSet`. On Azure it becomes `Microsoft_DefaultRuleSet`. On GCP it becomes `owasp-crs-v030301-id`. The abstraction holds because it operates above the provider API surface, not across it.

The Layer 1 goal is sound. The problem is in how the controller assembles a complete policy from multiple intent fragments — see [waf-configuration-design.md](waf-configuration-design.md) for why the merge model stalls.

---

## Layer 2 — Deliberately Absent

Layer 2 requires that a field concept genuinely exists on all clouds with the same semantics. The spike validated that no WAF tuning field meets this bar:

- **Managed rule group names** — provider names and bundling are incompatible (AWS `AWSManagedRulesCommonRuleSet`, Azure `Microsoft_DefaultRuleSet`, GCP `owasp-crs-v030301-id`)
- **Rule overrides** — not portable; uses provider-specific rule IDs and degrades silently on GCP
- **Custom rule conditions** — path/header/IP conditions are portable in concept but each provider uses structurally different JSON

A partial Layer 2 adds API surface without eliminating the need for the escape hatch. No Layer 2 fields are specified for WAF.

---

## Layer 3 — Unstructured Payload (P6) via spec.data

WAF rule content is schemaless — provider JSON structures vary significantly across versions and features. A typed sub-struct (P5) would require Cloud Manager to schema every provider's WAF JSON, which is not feasible for a feature of this complexity.

Instead, the design uses the **unstructured payload (P6)** pattern: the user supplies a complete provider-native WAF policy inline in `spec.data`, and Cloud Manager passes it through as an opaque payload.

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: WafPolicy
metadata:
  name: my-aws-policy
spec:
  data:                        # complete provider-native WAF policy JSON
    Name: "my-web-acl"
    DefaultAction:
      Allow: {}
    Rules: [...]
```

`WafPolicy` is the Layer 3 resource. It is honest, simple, and follows the same pattern as Terraform and Crossplane: no translation, no leaky abstraction. `spec.data` must contain a complete provider WAF policy (not a fragment).

### Why unstructured payload (P6) and not typed sub-struct (P5)

A typed sub-struct (P5) would require defining `spec.aws`, `spec.gcp`, `spec.azure` structs that mirror each provider's WAF JSON schema. Provider WAF schemas are large, versioned, and in AWS's case structurally recursive (logical operator nesting). The unstructured payload (P6) — the inline `runtime.RawExtension` — lets the user write provider-native JSON directly, with controller-side validation only.

---

## Design Decisions

### Why spec.data (inline) instead of spec.configMapRef

During the spike, `spec.configMapRef` was considered — a reference to a ConfigMap in the same namespace carrying the policy JSON. It was not adopted for the following reasons:

- **`WafPolicy` is the policy document.** A ConfigMap reference is pure indirection with no semantic value of its own. The policy content belongs with the resource that owns it.
- **Single object, single apply.** `spec.data` requires one `kubectl apply`; `spec.configMapRef` requires two coordinated objects. The coordination surface is a liability, not a feature.
- **Consistent observability.** With `spec.configMapRef`, validation errors appear on `WafPolicy` but the content is in the ConfigMap — the user must cross-reference two objects. With `spec.data`, the policy and its status are co-located.
- **No orphan risk.** An inline field has no orphaned ConfigMap edge case if the `WafPolicy` is deleted.
- **etcd size is equivalent.** The "large payload in a ConfigMap avoids bloating the CR" argument does not hold — etcd stores both; the total bytes are the same.

The one argument for `spec.configMapRef` — that the same ConfigMap could be shared by multiple `WafPolicy` objects — does not apply in practice. WAF policies are namespace-scoped and tightly coupled to the workload they protect; sharing a policy across resources via a reference adds no value and introduces implicit coupling.

### Why spec.data (inline) instead of spec.aws/spec.gcp/spec.azure sub-structs

Other Cloud Manager resources (`RedisInstance`, `VpcPeering`) use typed provider sub-structs — `spec.instance.aws`, `spec.instance.gcp`, etc. — with known, validated fields. This is the typed sub-struct (P5) pattern from the progressive design model: a typed Layer 3 sub-struct with full CRD schema validation. It works because the option space for those resources is bounded and fully expressible as a Go struct.

WAF policy schemas are a different category of problem:

- **Schema size** — a complete AWS WAFv2 `WebACL` definition covers hundreds of fields across deeply nested rule structures. Expressing this as a CRD would produce an unmanageably large schema.
- **Circular dependencies** — AWS WAFv2 allows logical operator nesting (`AndStatement` containing `OrStatement` containing `NotStatement` containing `AndStatement`...) that is structurally recursive. Kubernetes CRD validation does not support recursive or circular type references.
- **No shared structure** — AWS, Azure, and GCP WAF schemas are so different that a typed `spec.aws` / `spec.gcp` / `spec.azure` split would give users three completely unrelated APIs with no benefit over raw JSON.

A single `WafPolicy` with a `spec.data` inline field is the correct abstraction. `runtime.RawExtension` holds arbitrary JSON — the right primitive for a schemaless provider payload.

---

## Spike Findings: What the Portability Validation Showed

### Custom rules are genuinely portable

Custom rules with path, header, and IP conditions translate cleanly across all three providers:

| Condition | AWS | Azure | GCP |
|-----------|-----|-------|-----|
| Path exact | `ByteMatchStatement EXACTLY` | `Equals` | `request.path == '/x'` |
| Path prefix | `ByteMatchStatement STARTS_WITH` | `BeginsWith` | `request.path.matches('/x.*')` |
| Header exists | `SizeConstraintStatement >= 1` | `Exists` | `has(request.headers['x'])` |
| IP CIDR | `IPSetReferenceStatement` | `IPMatch` | `inIpRange(origin.ip, 'cidr')` |
| Negate | `NotStatement` | `negationConditon: true` | `!condition` |
| AND logic | `AndStatement` | array of `matchConditions` | `&&` in CEL |

The portable abstraction holds for this class of rule. These are expressed in provider-native JSON in `WafPolicy.spec.data`.

### Feature portability summary

| Feature | Portable? | Layer | Notes |
|---------|-----------|-------|-------|
| Managed rule groups | ❌ | Layer 1 via intents (deferred) | Provider names/structure incompatible — resolved by intent abstraction |
| Managed rules override | ❌ | unstructured payload (P6) only | Provider-specific names; GCP degrades entire ruleset |
| Custom rules (path/header/IP) | ✅ | unstructured payload (P6) via `WafPolicy.spec.data` | Portable concept — expressed as provider-native JSON |
| Rate limiting | ✅ | unstructured payload (P6) via `WafPolicy.spec.data` | Universal support |
| IP allowlist/blocklist | ✅ | unstructured payload (P6) via `WafPolicy.spec.data` | Universal support |
| Geographic blocking | ⚠️ | unstructured payload (P6) via `WafPolicy.spec.data` | Azure Application Gateway WAF lacks geo support |
| Size-based filtering | ⚠️ | unstructured payload (P6) via `WafPolicy.spec.data` | GCP limited; Azure global-only |
| Bot protection | ✅ | Layer 1 via intents (deferred) + unstructured payload (P6) | Basic via intent; advanced via `WafPolicy.spec.data` |

---

## Status

`WafPolicy` reports only `status.conditions` (standard Kubernetes condition list) and `status.providerId`. No proprietary top-level `state` field. Status is always normalised — the same fields on all clouds.
