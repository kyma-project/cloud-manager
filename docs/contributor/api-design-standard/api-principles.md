# Cloud Manager API Modeling Strategy

## Governing Rule


Every modeling decision - a new resource, new field, new type split - must pass the following governing rule first:

> **If your addition to the codebase changes what a user decides, expose it. If it only changes how the user request gets fulfilled, Cloud Manager handles it internally.**


---

## Resource Shapes

### Portable Intent Resources

Use portable intent resources when the user's requirement is the same regardless of which cloud provider fulfills it. The resource describes what the user needs using concepts they already know: protocols, capacity, topology, lifecycle policy. It contains no cloud-provider vocabulary.

Create portable intent resources when the following criteria apply:
- You can have the same spec for any cloud provider.
- You can derive provider-specific details from the user's input or hold these details in a separate referenced resource.
- Forcing the user to name a provider concept would not change their requirement, only their vocabulary.

### Provider-Specific Resources

Use provider-specific resources when the user's requirement is inherently tied to a provider—because concepts differ in meaningful ways, or because the user must choose among provider-native options to get the outcome they need.

Create provider-specific resources when the following criteria apply:
- You have a spec where the same field means different things, has different valid values, or requires different trade-offs for different cloud providers.
- A shared schema would either drop fidelity or produce large sections that are empty or meaningless for any given provider.
- The user is expected to bring provider knowledge to their requirement.

Naming follows the pattern `<Provider><Resource>` (for example, `AwsRedisInstance`, `GcpSubnet`). This convention makes the provider boundary visible at the type level, not buried in a field.

### Portable Containers with Provider-Specific Payload

Use portable containers with a provider-specific payload when the resource envelope is the same across providers, but the content inside it cannot be generalized. A single **spec.data** field holds the provider-native payload as an unstructured value. Local schema validation is limited; full validation happens at the cloud provider.

Create portable containers with a provider-specific payload when the following criteria apply:
- The act of creating, referencing, and managing the resource is provider-neutral.
- The payload itself (rule sets, policies, scripts) has no meaningful cross-provider representation.
- Abstracting the payload would require users to learn a new domain specific language instead of the provider's own, well-documented format.
- The provider schema is large, content-rich (for example, complex policy documents), or cyclical. Such structures cannot be faithfully represented as a Custom Resource Definition's (CRD's) hierarchical JSON schema without loss or artificial flattening.

---

## Decision Criteria

Before modeling any resource or field:

| Question | Yes → | No → |
|---|---|---|
| Does the user make the same decision regardless of provider? | Portable intent resource | Provider-specific resource |
| Is the schema meaningful without knowing the active provider? | Portable intent resource | Provider-specific resource |
| Can the provider detail be derived from what the user already expressed? | Derive it internally | Expose it in spec |
| Would learning the provider vocabulary change the user's decision? | Provider-specific resource | Derive or omit it |
| Is the envelope portable but the content not? | Portable container + `data` escape hatch | Model the content directly |

---

## Structural Rules

**Split a feature across multiple resources only when those resources have independent ownership, lifecycle, permissions, or reuse requirements.** A split that only exists to separate "intent" from "config" adds indirection without benefit. A split is justified when users will independently create, version, share, or permission the two parts.

**Use object references for all relationships.** References make lifecycle boundaries explicit and inspectable. Embedding one resource's concerns inside another's spec couples their lifecycles invisibly.

**Provide the escape hatch before blocking expert users.** When a feature has a well-defined common path but provider-specific depth that some users will need, an unstructured `data` field with provider-side validation is the right safety valve. Do not force a false abstraction to avoid it.

**Pre-deploy visible defaults.** Where sensible defaults exist, deploy them as named resources in the cluster so they are discoverable, reusable, and overridable. The common outcome should not require expert knowledge to achieve.

**Status is always portable.** All resources report state through the `Ready` condition. Provider-internal identifiers and state belong in `status` sub-fields, never in the top-level condition. See [**Status Conditions**](#status-conditions) for the standard states, observed-generation protocol, and `message` vs. sub-field guidance.

**User intent belongs in `spec`; platform outcomes belong in `status`.** Fields the platform assigns during fulfillment (provider IDs, derived endpoints, assigned IPs) are never written to `spec`.

---

## Status Conditions

### Ownership

Cloud Manager is the sole writer of its resources' **.status**. Always patch status using **client.MergeFrom** (merge-patch with the current object as base) so concurrent spec updates are not overwritten.

### Ready Condition

Every resource must declare a Ready condition. This condition is the machine-readable expression of the UX principle **Ready Means Ready to Use**.

Expose a printable `State` column in the CRD that prints the Ready reason:

```
.status.conditions[?(@.type=="Ready")].reason
```

Do **not** maintain a separate **status.state** field alongside the reconciler. The Ready condition reason is the authoritative state.

### Standard Ready States

| Status | Reason | Kind | Notes                                                                                                                                                  |
|---|---|---|--------------------------------------------------------------------------------------------------------------------------------------------------------|
| `Unknown` | `Processing` | transient | Observed generation is being reconciled                                                                                                                |
| `Unknown` | `WaitingForDependency` | transient | Waiting for a dependency; `message` must name what                                                                                                     |
| `True` | `Available` | final | Observed generation is reconciled and resource is usable                                                                                               |
| `False` | `ConfigurationError` | terminal (user-actionable) | Invalid spec or misconfigured reference; user must correct                                                                                             |
| `Unknown` | `Error` | transient | Retryable error; `message` must contain error detail                                                                                                   |
| `False` | `Failure` | terminal | Non-retryable failure; `message` must contain error detail                                                                                             |
| `False` | `Deleting` | transient | Resource is being deleted                                                                                                                              |
| `False` | `DeletionBlockedByDependents` | transient | Retryable error; resource can not be deleted since it's used; `message` must name resources blocking it; user may be needed to correct and delete them |

### Observed Generation Protocol

The authoritative source for observed generation is `status.conditions[type='Ready'].observedGeneration`. Do not introduce a separate `status.observedGeneration` field. When the Ready condition is absent (new resource), treat observedGeneration as zero — this guarantees the stale check triggers on first reconcile.

On reconcile **start**:
- If `metadata.generation != observedGeneration` (condition stale): patch Ready to `Unknown/Processing`, then continue.
- If `metadata.generation == observedGeneration` (condition current): continue without patching.

On reconcile **end**, patch Ready to the outcome state:
- Success → `True/Available`
- User-actionable problem → `False/ConfigurationError`
- Retryable error → `Unknown/Error`
- Terminal failure → `False/Failure`

Use `meta.SetStatusCondition` from `k8s.io/apimachinery/pkg/api/meta` to write conditions. This helper advances `lastTransitionTime` only when the condition `status` changes, leaving it unchanged on loops where status stays the same.

### Terminal Failure Recovery

`False/Failure` is not retried automatically. A new reconcile cycle requires a generation change — either a `spec` update or a force-reconciliation annotation. Document the supported mechanism in the resource's API reference.

### `message` vs. `status` Sub-Fields

`message` carries diagnostic narrative for humans: raw provider error responses, descriptions of what failed. It is displayed in `kubectl get` output and UI panels. It must not be parsed programmatically.

Structured values that workloads consume or tools query belong in typed `status` sub-fields:

| Value | Where it belongs |
|---|---|
| Raw provider API error response | `message` |
| Cloud provider resource ID (ARN, resource name, …) | `status.id` or equivalent sub-field |
| Name of a blocking dependency | `status.waitingFor` or equivalent |
| Provisioned endpoint | `status.endpoint` or equivalent |
| Provider error classification for retry routing | typed `status` sub-field |

---

## When Provider-Specific Modeling Is Correct

Provider-specific resources are not a failure of abstraction. They are the right answer when provider differences are real, user-visible, and decision-relevant. Forcing a portable shape onto genuinely different concepts produces worse UX than honest provider-specific types.

Signs that provider-specific modeling is correct:
- Two providers express the same capability through incompatible topological concepts (for example, peering, subnet routing, IAM attachment).
- A shared schema would require a union of all provider options, leaving most fields irrelevant for any given provider.
- The configuration the user writes would be different not just in values but in structure and vocabulary.
- Historical attempts to unify have produced abstractions that broke under new requirements or provider features.

---

## Applying the Strategy

When introducing a new resource or field, work through these steps:

1. **Name the user decision.** What is the user actually choosing? Write it in one sentence without using any provider name.
2. **Test portability.** Would a user write the same spec on AWS, GCP, and Azure? If yes, model it as portable. If not, identify what differs.
3. **Classify what differs.** Is it a different decision (provider-specific resource) or just a different fulfillment path (derive it internally)?
4. **Check the split justification.** If you're considering two resources, name the independent lifecycle or ownership requirement. If you can't, keep it one resource.
5. **Confirm status conformance.** Every resource reports `Ready` / `Processing` / `Error`. No exceptions.

---

## Summary

Cloud Manager modeling is **decision-first**, not portable-first or provider-first. The resource shape follows from what the user decides and their requirements, not from an architectural preference for unification or specificity.

| Scenario | Modeling choice |
|---|---|
| Same user decision across providers | Portable intent resource |
| Decision is inherently provider-specific | Provider-specific resource per provider |
| Envelope is portable, payload is not | Portable container with `data` escape hatch |
| Provider detail can be derived from user input | Derive it internally; omit from spec |
| Provider detail changes the user's outcome | Expose it explicitly in spec |
| Concepts differ in structure, not just values | Provider-specific resources |
