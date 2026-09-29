# WafConfiguration — Design Exploration and Why It Stalls

## What This Document Is

This document captures the `WafConfiguration` design as explored during the spike, and explains why the fragment-merge model it depends on does not have a clean solution across AWS, Azure, and GCP. `WafConfiguration` is deferred from the current API specification — see [api-specification.md](api-specification.md).

---

## The Intended Design

`WafConfiguration` was designed as a higher-level SKR resource that lets users express protection goals as named intents, without writing provider-specific JSON. The controller would resolve intents to preset policy fragments, merge them with any user-supplied fragments, and generate a complete `WafPolicy` with the merged content inline in `spec.data`.

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafConfiguration
metadata:
  name: my-config
  namespace: my-namespace
spec:
  intents:
    - OwaspTop10
    - BotProtection
  customIntents:
    - dataRef:
        name: my-custom-waf-rules   # user-owned WafPolicy spec.data fragment, same namespace
status:
  generatedPolicy:
    name: my-config
  conditions:
    - type: Ready
      status: True
      reason: PolicyGenerated
      observedGeneration: 1
```

### Intent vocabulary

| Intent | Description |
|--------|-------------|
| `OwaspTop10` | OWASP Top 10 protection (block mode) |
| `OwaspTop10Detection` | OWASP Top 10 detection only (count mode) |
| `BotProtection` | Bot and crawler protection |
| `RateLimiting` | Basic rate limiting |
| `IpReputation` | Known malicious IP blocking |

Each intent maps to a preset policy fragment per provider stored in `kyma-system`:

```
kyma-system/waf-preset-owasp-top10-aws
kyma-system/waf-preset-owasp-top10-azure
kyma-system/waf-preset-owasp-top10-gcp
kyma-system/waf-preset-bot-protection-aws
...
```

These presets are stored as ConfigMaps in `kyma-system` (as infrastructure data, not as the user-facing contract) — only the controller reads them. Users never reference preset ConfigMaps directly.

### How the controller was intended to work

```
1. For each intent in spec.intents:
   - Read preset policy fragment from kyma-system for current provider

2. For each entry in spec.customIntents:
   - Read the referenced policy fragment from the WafConfiguration namespace

3. Merge all fragments into a single complete policy
   - customIntents wins on conflict
   - Controller overwrites unconditionally on every reconcile

4. Generate WafPolicy with the merged content in spec.data
   - ownerReference: WafConfiguration
```

The analogy was Kubernetes Gateway API: `WafConfiguration` as `Gateway`/`HTTPRoute` (portable intent), `WafPolicy` as `GatewayClass` (provider-specific).

---

## Why It Stalls: The Fragment Merge Problem

The model depends on merging multiple JSON fragments into a single valid provider WAF policy. This is where it breaks down.

### AWS WAFv2 — array merge has no natural key

An AWS WAF policy is a flat `Rules` array. Two fragments both contribute rules:

```json
// preset: OwaspTop10
{ "Rules": [
    { "Name": "AWSManagedRulesCommonRuleSet", "Priority": 1000, ... },
    { "Name": "AWSManagedRulesSQLiRuleSet",   "Priority": 1010, ... }
]}

// customIntents fragment
{ "Rules": [
    { "Name": "HealthCheckBypass", "Priority": 10, ... },
    { "Name": "AdminInternalBypass", "Priority": 20, ... }
]}
```

To merge these, the controller needs a merge strategy:

- **Merge by `Name`?** — rules with the same `Name` in `customIntents` replace preset rules; others append. Requires the user to know preset rule names to override them. Works if names are stable.
- **Append-only?** — `customIntents` rules always append after presets. Users can add rules but cannot modify or suppress preset rules. Cannot be used to put an allow-rule before a block-rule from a preset.
- **JSON Merge Patch (RFC 7396)?** — arrays replace entirely. A `customIntents` fragment that touches `Rules` must supply the full replacement array, defeating the purpose of fragments.
- **Strategic merge patch?** — requires a patch key annotation per array element. WAF JSON is not Kubernetes YAML; there is no `$patch: merge` or `x-kubernetes-list-map-keys` to leverage.

None of these strategies is both ergonomic and correct for the general case.

### Azure Application Gateway WAF — different array keys per section

Azure WAF policy has two separate arrays that rules land in depending on type:

```json
{
  "properties": {
    "customRules": [ ... ],           // user-defined match rules
    "managedRules": {
      "managedRuleSets": [ ... ],     // managed rule set references
      "exclusions": [ ... ]
    }
  }
}
```

A preset fragment for `OwaspTop10` writes into `managedRules.managedRuleSets`. A `customIntents` fragment for custom rules writes into `customRules`. These are structurally separate — a merge strategy that works for one does not apply to the other. The controller would need to understand Azure's schema structure to route each fragment to the right array, which re-introduces the schema knowledge the ConfigMap approach was meant to avoid.

### GCP Cloud Armor — priority namespace collision

GCP policies are also a `rules` array, keyed by integer `priority`. Two fragments can independently assign the same priority to different rules:

```json
// preset: OwaspTop10
{ "rules": [{ "priority": 1000, ... }] }

// customIntents fragment
{ "rules": [{ "priority": 1000, ... }] }   // conflict
```

Priority is the only identity for a GCP rule — there is no name field. A merge strategy must either reject collisions (bad UX), silently overwrite one (non-deterministic), or require the user to coordinate priority namespaces across all fragments they combine (defeats the abstraction).

### The deeper problem: intents have the same issue

The merge problem is not specific to `customIntents`. It applies equally to `spec.intents`: when a user declares `[OwaspTop10, BotProtection]`, the controller must merge the two preset fragments for those intents. On AWS, both fragments contribute to the same `Rules` array. On GCP, both contribute to the same `rules` array with integer priorities. The merge strategy question applies to every pair of intents, not just to `customIntents` vs presets.

This means the problem is not solvable by removing `customIntents`. The entire `WafConfiguration` fragment-assembly model depends on a merge strategy that does not have a clean cross-provider answer.

---

## What Would Be Needed to Unblock This

Any of the following would resolve the merge stall:

1. **Restrict `WafConfiguration` to a single intent at a time** — no merging required; the controller maps one intent directly to a complete provider policy (not a fragment). Loses the ability to combine protections (e.g. OWASP + BotProtection together) unless each intent produces a self-contained policy that the provider can accept alongside others (not how AWS WAFv2 works — it uses a single WebACL).

2. **Define a provider-aware merge DSL** — the fragments include merge annotations (e.g. `$mergeKey: Name` for AWS, `$mergeKey: priority` for GCP). The controller interprets these per provider. Adds controller complexity and a new fragment format to document and validate.

3. **Drop the fragment model entirely for `customIntents`** — `spec.customIntents` accepts a complete policy replacement, not additions. Users who need customisation write a full policy inline in `WafPolicy.spec.data` directly. This collapses `WafConfiguration` to `spec.intents`-only (no customisation path), with the expert path being `WafPolicy`.

4. **Make intents produce independent provider resources** — if the cloud provider supports attaching multiple independent WAF rule groups to a load balancer (rather than a single policy), each intent could be a separate resource with no merging. AWS WAFv2 does not support this model; all rules must be in a single WebACL.

However, resolving the merge problem is not sufficient to unblock `WafConfiguration`. There is a deeper prerequisite.

---

## The Deeper Prerequisite: Use Cases and Override Path

Even if merging were solved, the intent vocabulary (`OwaspTop10`, `BotProtection`, etc.) gives users no way to tune or override the rules those intents resolve to. When `OwaspTop10` blocks a legitimate request — a false positive — the only recourse in the current design is to abandon `WafConfiguration` entirely and write a full `WafPolicy.spec.data`. That is a cliff, not a ladder.

A usable Layer 1 must answer: **what does a user do when an intent produces the wrong behaviour?** Without a defined override or exclusion path, the intent vocabulary is a one-size-fits-all wrapper that breaks the progressive API model's core promise — that each layer adds capability without forcing the user to abandon what lower layers gave them.

**This means use case collection is a prerequisite before `WafConfiguration` can be meaningfully designed**, not a follow-on activity. The use cases must answer:

- What is the most common reason a user would need to override or tune an intent's default rules?
- What is the minimum override surface that covers the majority of real cases (e.g. rule exclusions, action overrides, priority adjustments)?
- Can that override surface be expressed in a cloud-neutral way, or does it inevitably require provider-specific syntax?

Until those questions are answered with real use cases, any `WafConfiguration` design risks shipping an abstraction that works only for the happy path and forces power users to the escape hatch immediately.

---

## Possible Solution: Provider-Specific WafConfiguration Kinds

The merge problem exists because a single `WafConfiguration` must work across providers with structurally incompatible policy schemas. One way to remove that constraint entirely is to split into provider-specific Kinds: `AwsWafConfiguration`, `AzureWafConfiguration`, `GcpWafConfiguration`.

### What this solves

Each Kind owns one provider's schema. The controller for `AwsWafConfiguration` knows the AWS WAFv2 shape — it can use `Name` as the merge key for the `Rules` array, manage `Priority` assignment, and handle `OverrideAction` correctly. The controller for `GcpWafConfiguration` knows GCP Cloud Armor's integer `priority` namespace and can enforce uniqueness. `AzureWafConfiguration` can route fragments to `customRules` vs `managedRules.managedRuleSets` because the schema is known at compile time.

The fragment-merge problem is reduced to a per-provider problem with a known structure — solvable with typed Go structs and deterministic merge logic — rather than a cross-provider problem requiring a neutral DSL.

### API sketch

```yaml
# AWS
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: AwsWafConfiguration
metadata:
  name: my-aws-config
spec:
  intents:
    - OwaspTop10
    - BotProtection
  customRules:
    - name: HealthCheckBypass        # typed; Name is the merge key
      priority: 10
      action: Allow
      statement:
        byteMatch:
          fieldToMatch: URI
          positionalConstraint: STARTS_WITH
          searchString: /health
  defaultAction: Allow
```

```yaml
# GCP
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: GcpWafConfiguration
spec:
  intents:
    - OwaspTop10
  customRules:
    - priority: 1000               # typed; priority is the merge key; controller enforces uniqueness
      action: allow
      match:
        expr: "request.path == '/health'"
  defaultAction: allow
```

```yaml
# Azure
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: AzureWafConfiguration
spec:
  intents:
    - OwaspTop10
  customRules:
    - name: HealthCheckBypass      # routes to customRules array; typed
      priority: 10
      action: Allow
      matchConditions:
        - matchVariable: RequestUri
          operator: BeginsWith
          matchValues: ["/health"]
  managedRuleOverrides: []
  policyMode: Prevention
```

### How it interacts with WafPolicy

Each provider-specific `WafConfiguration` generates a `WafPolicy` with `spec.data` set to the assembled provider-native JSON — the same way the unified `WafConfiguration` was intended to work. `AppLoadBalancer.spec.policyRef` still resolves to a `WafPolicy` name.

```
AwsWafConfiguration / AzureWafConfiguration / GcpWafConfiguration
    ↓ controller assembles provider-native JSON, creates/updates
WafPolicy (spec.data: complete provider-native JSON)
    ↓
AppLoadBalancer (spec.policyRef: WafPolicy)
```

`WafPolicy` remains the stable foundation and the only cloud-provisioning Kind. The provider-specific `WafConfiguration` Kinds are a generation layer — they never call cloud provider APIs directly.

### Trade-offs

| | Unified WafConfiguration | Provider-specific WafConfiguration |
|---|---|---|
| Portability | Single Kind, any cloud | User picks the right Kind per cluster |
| Merge problem | No clean cross-provider solution | Solved per provider with typed structs |
| Override/tuning surface | Requires cloud-neutral DSL | Can use provider-native field names |
| API size | One Kind | Three Kinds; intent vocabulary shared |
| Migration cost | One Kind to learn | Users on multi-cloud must manage three |

### Remaining open question

Provider-specific Kinds solve the merge problem but do not resolve the deeper prerequisite described above: the override path. A user who deploys `AwsWafConfiguration` with `OwaspTop10` and gets a false positive still needs a way to suppress or tune specific rules within the intent. That override surface (`customRules` in the sketch above) is now expressible in a typed way — but the vocabulary, defaults, and priority assignment conventions still need to be defined from real use cases before the Kind can be specified.

The split also moves the "which Kind do I use?" decision to the user at authoring time, which is a trade-off relative to the single `WafConfiguration` goal of transparent provider routing.

---

## What Remains Valid

The `WafPolicy` design is unaffected by this problem. A `WafPolicy` carries a single complete policy inline in `spec.data` — no merging, no fragments. It is the stable foundation regardless of how `WafConfiguration` is eventually resolved.

The preset policy content (per intent per provider) remains valid as reference material — the question is only how to combine them, not what they contain.
