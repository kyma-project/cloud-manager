# Industry Patterns for Separating User Intent from Provider Implementation

## Pattern Taxonomy

| Pattern | How it separates intent | User sees provider? | Portability |
|---------|------------------------|---------------------|-------------|
| **PVC / StorageClass / PV** | User writes intent (`capacity`, `accessMode`); StorageClass carries opaque `parameters` for provisioner | No — picks class name | High within cluster |
| **Gardener DNS** | `DNSEntry` has only `dnsName`, `ttl`, `targets`; controller matches domain to `DNSProvider` at runtime | No — fully oblivious | High — automatic routing |
| **Gateway API** | 3-tier role split: GatewayClass → Gateway → HTTPRoute; `parametersRef` is the provider escape hatch | Yes — picks `gatewayClassName` | Medium (conformance-dependent) |
| **Crossplane** | XRD defines user schema; Composition maps fields to provider CRD via patches/transforms | Optional | Very high |
| **Cluster API** | Abstract `Cluster` holds `infrastructureRef` pointing to provider-specific `AWSCluster` | Yes — must create both objects | Low at manifest level |
| **ACK / ASO** | No separation; CRD mirrors the cloud API directly | Entirely | Zero |

---

## Key Lessons per Pattern

### PVC / StorageClass / PV

The canonical Kubernetes split: `PVC` expresses *what* (`10Gi`, `ReadWriteOnce`), `StorageClass` names the provisioner and carries opaque `parameters`, `PV` is the provisioned result owned by the controller.

**What Cloud Manager can take from it:**
- User-facing spec carries intent (capacity, access mode, version). Provider-specific sizing vocabulary (`db.t3.micro`, `BASIC_HDD`) belongs in the controller mapping, not in the user spec.
- The `parameters: map[string]string` escape hatch is valid for provider-specific config that cannot be abstracted — but it sacrifices CRD-level validation.

**Limit:** Opaque `parameters` break schema validation and make the API undiscoverable. Cloud Manager's typed provider sub-structs in KCP (`instance.gcp`, `instance.aws`) are strictly better here.

---

### Gardener DNS

`DNSEntry` carries only `dnsName`, `ttl`, `targets` — no provider field at all. The controller matches the domain against `DNSProvider.spec.domains.include` at runtime and routes to Route53, Cloud DNS, or Azure DNS transparently.

**What Cloud Manager can take from it:**
- `IpRange` already follows this pattern: one neutral `cidr` field, provider routing handled internally. This is the target shape for any resource where the user's decision is genuinely identical across providers.
- Implicit routing only works when provider domain ownership is non-overlapping. Cloud Manager's runtime binding (Kyma cluster scope determines provider) is the equivalent mechanism.

**Limit:** Provider-specific routing policies (latency-based, geolocation) cannot be expressed in the neutral resource — they require provider-specific extension or a separate resource.

---

### Gateway API

Three-tier role split where each tier is owned by a different persona: GatewayClass (infrastructure operator) → Gateway (cluster operator) → HTTPRoute (application developer). `GatewayClass.spec.parametersRef` is the provider-specific escape hatch for the infrastructure tier.

**What Cloud Manager can take from it:**
- Role separation maps cleanly onto Cloud Manager's architecture: the Kyma module operator owns the infrastructure binding (equivalent to GatewayClass), the SKR resource is owned by the application developer, and the KCP resource is the platform-side intermediary.
- The `parametersRef` pattern — a typed reference to a provider-specific config object — is worth considering for resources where the user legitimately needs to tune provider behavior without Cloud Manager having to enumerate every option.

**Limit:** `parametersRef` immediately breaks portability for any consumer that sets it. It is an escape valve, not a first-class pattern.

---

### Crossplane

Five-layer model: XRD (schema) → Composition (field mapping) → XR (platform resource) → Claim (user namespaced resource) → Managed Resource (provider CRD mirroring cloud API).

The Composition maps `spec.size: small/medium/large` to `spec.instanceType: db.t3.micro` at the Composition layer — in YAML, not compiled code.

**Where Cloud Manager sits:**
Cloud Manager is Crossplane without the XRD/Composition layer. The SKR resource is the Claim, the KCP resource is the XR, and the provider action pipelines in Go are the Composition logic — compiled rather than declarative. The provider sub-structs (`instance.gcp`, `instance.aws`) in KCP are the Managed Resource fields.

**What Cloud Manager can take from it:**
- The Composition principle: user-facing spec carries neutral concepts; the mapping to provider vocabulary is the controller's responsibility, not the user's.
- Applied to Cloud Manager: `engineVersion: "7.0"` is the user concept. `REDIS_7_0` is the GCP API vocabulary. The mapping belongs in the action pipeline, not in the CRD field value.

**Advantage over Crossplane:** Go pipelines are type-safe, testable, and easier to reason about than YAML patch chains. They cannot be reconfigured at runtime without a redeploy — which is a correct and deliberate constraint for a product operator.

**Limit of pure Crossplane:** Patch-based Compositions are verbose for complex logic; Composition Functions (arbitrary code) are required for anything non-trivial, at which point the YAML abstraction is no longer cheaper than Go.

---

### Cluster API

`Cluster` holds `spec.infrastructureRef` pointing to a provider-specific `AWSCluster`. The two objects communicate through a defined status contract: the provider controller watches `AWSCluster` and sets `status.ready: true` when provisioning is complete. CAPI core orchestrates lifecycle without knowing the cloud provider.

**What Cloud Manager can take from it:**
- The SKR → KCP relationship is structurally identical: the SKR resource is the abstract `Cluster`, the KCP resource is the `AWSCluster`, and they communicate through a status contract. Cloud Manager improves on CAPI by managing the second object internally rather than requiring the user to create it.
- The explicit status contract between layers (a defined set of fields the provider must populate before the orchestrator proceeds) is the right model for Cloud Manager's `waitKcpStatusUpdate` gates.

**Limit:** In CAPI, the user creates both objects — portability requires replacing the entire provider object. Cloud Manager avoids this by generating the KCP object from the SKR spec.

---

### ACK / ASO

No abstraction. CRD fields mirror the cloud API. Maximum completeness, zero portability. The right choice when users already know the cloud API, full feature surface matters more than portability, and there is no multi-cloud abstraction goal.

**When Cloud Manager uses this correctly:**
- `AwsVpcPeering`, `GcpVpcPeering`, `AzureVpcPeering` — the underlying peering mechanism, identity model, and routing options differ structurally across providers. A portable wrapper would either drop fidelity or produce a union type where most fields are irrelevant on any given provider. Provider-specific resources are the right call here.

---

## Core Dimensions

**Abstraction degree**
None (ACK) → Class-based (StorageClass) → Schema-mapped (Crossplane) → Implicit routing (Gardener DNS)

**Who owns the mapping**
Compiled code (ACK, Cloud Manager) · Declarative YAML patches (Crossplane) · Opaque provisioner parameters (StorageClass) · Domain-matching algorithm (Gardener DNS)

**Single-object vs. multi-object**
One object (ACK) → Two paired objects (CAPI) → Three layers (PVC + StorageClass + PV) → Five layers (Crossplane)

**Design-time vs. runtime binding**
CAPI and Crossplane bind provider at manifest-authoring time. Gardener DNS and StorageClass dynamic provisioning bind at runtime from cluster configuration. Cloud Manager uses runtime binding: the Kyma cluster's provider scope determines which action pipeline runs.

---

## Decision Guide

| User decision is the same across providers? | Model |
|---------------------------------------------|-------|
| Yes, and the controller can route silently | Provider-neutral resource (Gardener DNS pattern) — `IpRange` |
| Yes, but provider-specific config is sometimes needed | Provider-neutral spec + typed provider sub-struct in KCP (Crossplane Claim pattern) |
| No — concepts differ structurally | Provider-specific resource per provider (ACK pattern) — VpcPeering |
| Envelope is neutral, but payload content is not | Portable container + unstructured `data` escape hatch |
| Provider detail can be derived from user input | Derive internally; omit from spec |
