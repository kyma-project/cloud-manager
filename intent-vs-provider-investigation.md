# Investigation: User Intent vs Provider-Specific Implementation

## 1. Honest Inventory

Cloud Manager uses a two-layer model:

- **SKR** (`cloud-resources.kyma-project.io/v1beta1`) — user-facing CRDs on the Kyma cluster
- **KCP** (`cloud-control.kyma-project.io/v1beta1`) — control-plane CRDs that trigger cloud API calls

### Resource-by-Resource Breakdown

| Resource | User Intent Fields | Provider-Specific Fields | Ambiguous / Inconsistent |
|----------|--------------------|--------------------------|--------------------------|
| **IpRange** (SKR) | `cidr` | — | — |
| **IpRange** (KCP) | `cidr` | `options.gcp.purpose`, `options.gcp.psaService` | `status.vpcId`, `status.addressSpaceId` — derived, but named in cloud terms |
| **GcpSubnet** (SKR) | `cidr` | "Gcp" prefix in type name — GCP-only but styled as if generic | — |
| **GcpRedisInstance** (SKR) | `redisTier`, `redisVersion`, `authEnabled`, `redisConfigs`, `ipRange` | `maintenancePolicy.dayOfWeek` (GCP-specific shape) | `redisTier` uses GCP tier names (S1–S8, P1–P6) |
| **AwsRedisInstance** (SKR) | `cacheNodeType`, `engineVersion`, `authEnabled` | `cacheNodeType` is an AWS-specific term | `readReplicas` immutability differs from GCP |
| **AzureRedisInstance** (SKR) | `sku`, `replicasPerPrimary` | `sku.family` (S/P), `sku.capacity` — Azure-specific naming | — |
| **RedisInstance** (KCP) | — | `instance.gcp`, `instance.azure`, `instance.aws`, `instance.alicloud` — discriminated union | `status.nodeType`, `status.memorySizeGb`, `status.replicaCount` — normalized cross-provider ✓ |
| **GcpNfsVolume** (SKR) | `tier`, `capacityGb`, `ipRange` | `fileShareName` (GCP concept), `location` (deprecated), `connectMode` | K8s PV/PVC spec embedded in SKR type — mixes infrastructure and K8s lifecycle |
| **NfsInstance** (KCP) | — | `instance.gcp`, `instance.aws`, `instance.azure`, `instance.openStack`, `instance.alicloud` | `status.hosts[]` (legacy) vs `status.host`/`status.path` (new) — backward-compat noise |
| **GcpVpcPeering** (SKR) | `remoteVpc`, `remoteProject`, `importCustomRoutes` | GCP-specific field names and validation (project ID length, etc.) | `deleteRemotePeering` — implementation detail surfaced to the user |
| **VpcPeering** (KCP) | `details.localNetwork`, `details.remoteNetwork`, `details.importCustomRoutes` | `vpcPeering.gcp.*`, `vpcPeering.azure.*`, `vpcPeering.aws.*` | Two modes (`details` vs `vpcPeering`) with no documented decision rule |
| **GcpNfsVolumeBackup** (SKR) | `source.volume`, `accessibleFrom` | `location` (GCP region, immutable) | `fileStoreBackupLabels` — GCP metadata detail in status |
| **GcpNfsBackupSchedule** (SKR) | `nfsVolumeRef`, `schedule`, `maxRetentionDays`, `maxReadyBackups` | — (schedule is provider-neutral) | `accessibleFrom` is a Cloud Manager concept but only present on GCP types |

### Where We Are Consistent

- **KCP discriminated unions** (`instance.gcp`, `instance.aws`, etc.) correctly isolate provider-specific fields
- **KCP Redis status normalization** (`memorySizeGb`, `replicaCount`, `primaryEndpoint`) is the right pattern — maps provider-specific sizing to neutral output fields
- **IpRange** `options.gcp.*` / `options.aws.*` is the right shape for provider-specific spec extension

### Where We Are Inconsistent

- SKR uses provider-specific type names (`GcpRedisInstance`) for some resources and generic names (`IpRange`) for others — no clear principle governing the choice
- VpcPeering KCP has two modes (`details` neutral + `vpcPeering` provider-specific) without a documented decision rule
- NFS embeds K8s PV/PVC spec in the SKR type — unique to NFS, mixes infrastructure provisioning with Kubernetes object lifecycle
- Status normalization is thorough for Redis, incomplete for NFS (legacy `hosts[]` + new `host`/`path` coexist), absent for VpcPeering

### Where Provider-Specific Modeling Is the Right Call

- GCP `fileShareName`, `purpose`, `psaService`, `connectMode` — no cross-provider equivalents exist
- AWS `cacheNodeType` — this is the user-facing sizing unit on AWS; the mapping to `status.memorySizeGb` is what Cloud Manager correctly does in status
- Azure `sku.family` (S/P) — the fundamental Azure Redis product tier distinction

---

## 2. Industry Patterns

### Pattern Taxonomy

| Pattern | Mechanism | Class Resource | User Sees Provider? | Portability |
|---------|-----------|---------------|---------------------|-------------|
| **PVC / StorageClass / PV** | Named class profile; provisioner interprets opaque `parameters` | Yes (StorageClass) | No — picks class name | High within cluster |
| **Gateway API** | Role-oriented 3-tier split; GatewayClass names controller | Yes (GatewayClass) | Yes — picks gatewayClassName | Medium (conformance-dependent) |
| **Gardener DNS** | Implicit domain-matching; no explicit provider reference in user intent | No | No — fully oblivious | High — automatic |
| **Crossplane** | XRD defines schema; Composition maps fields via patches/transforms | No (Composition plays this role) | Optional | Very high |
| **Cluster API** | Abstract `Cluster` holds `infrastructureRef` pointer to provider-specific `AWSCluster` | No | Yes — must create both objects | Low at manifest level |
| **ACK / ASO** | No separation; CRD = cloud API | No | Entirely | Zero |
| **OSBAPI / Service Catalog** | Catalog of named plans; broker handles all implementation out of cluster | Yes (ClusterServiceClass / Plan) | No | High, but broker is external |

### Key Lessons per Pattern

**PVC / StorageClass / PV**
The canonical separation: `PVC` expresses what (10 GiB, ReadWriteOnce), `StorageClass` names the provisioner and carries opaque `parameters`, `PV` is the provisioned result. Provider-specific capabilities (provisioned IOPS, Ultra Disk) escape only through the `parameters` map, which breaks portability. The class resource is owned by the cluster operator, not the user.

**Gateway API**
Three-tier role split: GatewayClass (infrastructure provider) → Gateway (cluster operator) → HTTPRoute (app developer). `GatewayClass.spec.parametersRef` is the escape hatch for provider-specific config. The route has no idea whether the LB is NGINX, Envoy, or GCP. Limit: `GatewayClass.parametersRef` immediately breaks portability for any provider-specific tuning.

**Gardener DNS**
DNSEntry expresses intent (`dnsName`, `ttl`, `targets`) with no provider reference. The controller matches `dnsName` against `spec.domains.include` across DNSProviders at runtime. User is fully oblivious to which backend (Route53, Cloud DNS, Azure DNS) handles the record. Limit: implicit matching requires non-overlapping domain ownership; provider-specific routing policies (latency-based, geolocation) cannot be expressed portably.

**Crossplane**
Five-layer model: XRD (schema) → Composition (mapping) → XR (platform-side abstract resource) → Claim (user-facing namespaced resource) → Managed Resource (provider CRD, mirrors cloud API). The Composition maps Claim fields to Managed Resource fields via declarative YAML patches and transforms. Platform teams can expose `spec.size: small/medium/large` and map it to `db.t3.micro` / `db.m5.large` at the Composition level. Limit: patch-based mapping is verbose; complex logic requires Composition Functions (arbitrary code). The XR/Claim duality is a recurring source of confusion.

**Cluster API**
`Cluster` holds `spec.infrastructureRef` pointing to `AWSCluster`. CAPI core orchestrates lifecycle; the provider controller watches `AWSCluster` and sets `status.ready: true`. They communicate via a well-defined contract of status fields. Limit: user must create both objects; portability requires replacing the entire provider object, not just changing a field.

**ACK / ASO**
No abstraction. CRD fields mirror the cloud API. Maximum completeness, zero portability. The right choice when: users already know the cloud API, full feature surface is more important than portability, and there is no need for a multi-cloud abstraction.

**OSBAPI / Service Catalog**
User picks a named plan from a catalog; the broker handles everything. The `ServiceInstance.spec.parameters` map is unstructured (no CRD-level validation). Kubernetes Service Catalog was archived in 2022 — pattern lives on in OpenShift and SAP BTP broker implementations.

### Core Design Dimensions

**Dimension 1: Abstraction degree**
Zero (ACK/ASO) → Class-based (StorageClass) → Schema-based (Crossplane) → Implicit routing (Gardener DNS)

**Dimension 2: Feature completeness vs. portability tradeoff**
Every abstraction restricts vocabulary to the intersection of what all providers support. Provider-specific capabilities escape only through opaque extension points (StorageClass `parameters`, GatewayClass `parametersRef`).

**Dimension 3: Who owns the mapping**
- Compiled code (ACK, Cloud Manager action pipelines)
- Declarative YAML patches (Crossplane Composition)
- Opaque provisioner parameters (StorageClass)
- Domain-matching algorithm (Gardener DNS)

**Dimension 4: Single-object vs. multi-object**
One object (ACK) → Two paired objects (CAPI Cluster + InfraCluster) → Three layers (PVC + StorageClass + PV) → Five layers (Crossplane)

**Dimension 5: Design-time vs. runtime binding**
CAPI and Crossplane `compositionRef` bind at manifest-authoring time. Gardener DNS and StorageClass dynamic provisioning bind at runtime based on current cluster configuration.

### Where Cloud Manager Sits

Cloud Manager's architecture is **Crossplane without the XRD/Composition layer**. The SKR resource is the Claim, the KCP resource is the XR, and the provider action pipelines in Go are the Composition logic — compiled rather than declarative. The provider sub-structs (`instance.gcp`, `instance.aws`) in KCP are the Managed Resource fields.

The tradeoff vs. Crossplane: Go pipelines are type-safe, testable, and easier to reason about; they cannot be reconfigured at runtime without a redeploy. This is a deliberate and correct choice for a product operator.

---

## 3. WAF as a Test Case

### Per-Provider WAF Model

| Concept | AWS | Azure | GCP | Alibaba Cloud | SCI (OpenStack) |
|---------|-----|-------|-----|---------------|-----------------|
| Product name | WAF v2 (WebACL) | WAFPolicy (two types) | Cloud Armor (SecurityPolicy) | WAF 3.0 | None native |
| Attaches to | ALB, CloudFront, API GW, Cognito, AppSync, App Runner | App Gateway **or** Front Door (separate ARM types) | Backend Service | Domain name (DNS intercept) | Third-party or N/A |
| Rule source | AWS ManagedRuleGroups, custom | OWASP ruleset, custom | Pre-configured rules, custom | Rule groups | N/A |
| Attachment model | ARN reference | Policy linked to gateway resource | Policy linked to backend service | DNS intercept | N/A |

### Where the Pattern Holds

A generic `WafPolicy` SKR resource could expose:
- `spec.targetRef` — what to protect (a service, gateway, or ingress)
- `spec.rules` — provider-neutral rule categories: OWASP Top 10 on/off, rate limiting, IP allowlist/blocklist
- `spec.instance.aws`, `spec.instance.azure`, `spec.instance.gcp`, `spec.instance.alicloud` — provider-specific sub-structs for anything without a cross-provider equivalent

The KCP `WafPolicy` would follow the same discriminated union pattern as `RedisInstance` and `NfsInstance`.

### Where It Breaks Down

The `targetRef` problem is structural, not terminological:
- AWS attaches to an infrastructure ARN
- GCP attaches to a backend service
- Azure has two different WAF resource types (`ApplicationGatewayWAFPolicy` vs `FrontDoorWAFPolicy`) — not the same thing
- AliCloud intercepts at DNS, not at infrastructure

There is no single neutral "what to protect" concept that maps cleanly across all providers. Kubernetes Gateway API's `HTTPRoute → Gateway` model is the closest industry solution: if Cloud Manager routes user traffic through a `Gateway` abstraction, a WAF policy could attach to the `Gateway` regardless of which cloud hosts it.

### Comparison with Redis as a Test Case

Redis works cleanly because the user intent fields (capacity, version, auth, configs) have functional equivalents across all providers. The tier naming is provider-specific (S1–S8 GCP, node type AWS, SKU capacity Azure) but the underlying concepts — memory size, replica count, version — are universal. Cloud Manager maps them to a normalized status correctly.

WAF is harder because the **attachment model** differs structurally, not just terminologically. WAF rules are also less uniformly translatable: an AWS `AWSManagedRulesCommonRuleSet` has no direct Azure or GCP equivalent by name, only by function.

**Conclusion for WAF:** The discriminated union pattern (provider sub-structs in KCP) applies correctly to rule configuration. The `targetRef` problem requires a prior decision about how Cloud Manager models the traffic entry point — that decision gates the WAF design, and should likely align with Gateway API's `parentRef` pattern.

### Generalization Beyond WAF

The WAF test reveals a generalizable principle: **the pattern holds for resource properties (what the resource is) but strains at attachment models (what the resource connects to)**. This applies beyond WAF:

- NFS volume attachment to a network (IpRange vs. Subnet) has similar structural differences per provider
- VpcPeering's two modes (`details` vs `vpcPeering`) exist precisely because the peering connectivity model differs structurally between AWS (VPC peering = bilateral) and GCP (peering = two unilateral operations)
- Any future resource that attaches to provider-specific networking constructs will face the same challenge

The right response is not to force a single `targetRef` field, but to accept that attachment is a provider-specific concern and model it in the provider sub-struct — the same way `connectMode` is in `instance.gcp` for NFS, not at the top level.

---

## 4. Making Cloud Manager CRDs AI-First

An AI-first CRD is one where a language model — or any tool that consumes the OpenAPI schema — can answer three questions without leaving the schema:

1. **What does this field do?** (semantic description)
2. **What values are valid, and what do they mean?** (enum semantics, constraints)
3. **When is this field set, and what does it contain?** (lifecycle and format)

Cloud Manager's CRDs are currently good at structure and validation, and weak at all three. The problems are systematic, not scattered — they can be fixed with clear conventions applied consistently.

### Current State

**Strengths (keep these):**
- CEL validation rules are present and have human-readable `message:` fields — this is the right pattern
- Enum values are validated against explicit lists
- Spec fields for user-facing resources (GcpRedisInstance, GcpNfsVolume) have partial descriptions
- Type-level struct comments exist in Go source

**Gaps (fix these):**

| Gap | Example | Impact |
|-----|---------|--------|
| Status fields have no descriptions | `primaryEndpoint`, `authString`, `hosts`, `stateData`, `opIdentifier` — all undocumented | LLM cannot tell a user what to read after provisioning |
| Enum values are listed but not explained | `redisTier` enumerates S1–S8 but the schema says nothing about what S vs P means | LLM must guess from names |
| No lifecycle documentation | `memorySizeGb` — is it populated before or after Ready? Is `hosts` always a list? | LLM cannot tell a user when fields are safe to read |
| No format documentation | `authString` — is this a token, a password, a connection string? `cidr` — validated by CEL but the description says nothing | LLM must infer from field name |
| Go comments don't reach the CRD | The redisTier comment explains S vs P tiers in Go source; none of it appears in the generated CRD schema | Only humans reading the source benefit |
| No published schema artifact | No OpenAPI 3.0 export, no JSON schema, no CRD reference docs | AI tooling must parse raw YAML CRD files |

### What "AI-First" Means in Practice

AI-first is not a special mode. It is the outcome of writing documentation at the layer where tooling reads it: the generated CRD schema, not the Go source. Kubebuilder propagates Go struct field comments into CRD `description:` fields automatically. The work is adding those comments — the infrastructure already exists.

**Rule 1: Every spec field gets a description that answers "what does the user put here?"**

Bad (current state):
```go
// RedisTier defines the service and capacity tier.
// +kubebuilder:validation:Enum=S1;S2;...
RedisTier string `json:"redisTier"`
```

Good:
```go
// RedisTier sets the service tier (S = Standard, P = Premium/HA) and capacity.
// S-tiers are single-zone; P-tiers are multi-zone with automatic failover.
// S1–S4 and P1–P4 share the same memory sizes; S5–S8 and P5–P6 are larger capacities.
// Immutable: service tier (S vs P) cannot be changed after creation; capacity tier can.
// +kubebuilder:validation:Enum=S1;S2;S3;S4;S5;S6;S7;S8;P1;P2;P3;P4;P5;P6
RedisTier string `json:"redisTier"`
```

**Rule 2: Every status field gets a description that answers "what is this, when is it set, what format?"**

Bad (current state):
```go
PrimaryEndpoint string `json:"primaryEndpoint,omitempty"`
AuthString      string `json:"authString,omitempty"`
```

Good:
```go
// PrimaryEndpoint is the host:port of the Redis primary node. Set when state=Ready.
PrimaryEndpoint string `json:"primaryEndpoint,omitempty"`

// AuthString is the Redis AUTH password. Set when authEnabled=true and state=Ready.
// Store this value in a Secret; it does not change unless the instance is recreated.
AuthString string `json:"authString,omitempty"`
```

**Rule 3: CEL constraint messages explain the why, not just the what**

Bad:
```go
// +kubebuilder:validation:XValidation:rule="...",message="Service tier cannot be changed."
```

Good:
```go
// +kubebuilder:validation:XValidation:rule="...",message="Service tier (S vs P) is immutable because changing it requires destroying and recreating the instance."
```

**Rule 4: Internal-only status fields are marked as such**

Fields like `stateData`, `opIdentifier`, `subnetCreationOperationName` are reconciler bookkeeping, not user-readable output. They should be documented as such — or moved to annotations — so a tool doesn't suggest them to users.

```go
// OpIdentifier is an internal reconciler field tracking the in-flight cloud API operation.
// Not intended for user consumption.
OpIdentifier string `json:"opIdentifier,omitempty"`
```

**Rule 5: The discriminated union constraint is documented at the union level**

KCP resources use discriminated unions (`instance.gcp`, `instance.aws`, etc.) with MaxProperties=1. A tool reading the schema today cannot tell that exactly one sub-struct must be set. The XValidation message should explain this:

```go
// +kubebuilder:validation:XValidation:rule="...",message="Exactly one provider sub-struct (gcp, aws, azure, alicloud) must be set. The selected provider must match the Kyma cluster's cloud provider."
```

### Structural Improvements

Beyond comments, two structural changes would make the CRDs significantly more consumable:

**A. Publish a schema artifact**

Generate and commit an OpenAPI 3.0 document from the CRD YAML files as part of `make manifests`. This gives AI tooling, documentation generators, and client generators a single canonical schema endpoint without parsing Kubernetes CRD YAML.

A minimal target: one JSON schema file per CRD group (`cloud-resources.json`, `cloud-control.json`) committed to `config/schema/`. These can be generated with `kubectl get crd -o json | jq ...` or tools like `crd-ref-docs`.

**B. Separate user-observable status from reconciler bookkeeping**

The clearest AI-first signal is a schema where every field in `status` is something a user or automation should read. Currently `stateData`, `opIdentifier`, `subnetCreationOperationName`, and similar fields are in `status` alongside user-meaningful fields like `primaryEndpoint` and `hosts`. This forces a tool to treat all status fields as equally meaningful.

Options:
- Move reconciler bookkeeping to annotations (conventional in Kubernetes — see Cluster API's `cluster.x-k8s.io/paused` annotation pattern)
- Or add a nested `status.internal` struct and document it as non-user-facing

Either approach gives a tool a clean signal: fields in `status` (outside `internal`) are safe to surface to users; fields in `status.internal` or annotations are reconciler state.

### Prioritization

The gaps above are not equally valuable to fix. Ordered by impact on AI-assisted user workflows:

| Priority | Fix | Why |
|----------|-----|-----|
| 1 | Document all status fields on SKR types (GcpRedisInstance, GcpNfsVolume, etc.) | These are the fields users read after provisioning — highest LLM interaction surface |
| 2 | Explain enum semantics for redisTier, tier (NFS), sku.family | Users ask "which tier should I use?" — currently unanswerable from schema alone |
| 3 | Add lifecycle annotations ("set when state=Ready") to connection/credential fields | Prevents LLMs from suggesting fields that aren't populated yet |
| 4 | Document internal-only status fields as non-user-facing | Reduces noise in LLM-generated answers |
| 5 | Publish schema artifact | Enables tooling and documentation generation |
| 6 | Document KCP types | Lower priority — users don't interact with KCP directly |

### Connection to the Intent/Provider Separation

The documentation problem and the intent/provider separation problem are the same problem viewed from different angles. A field that is user intent should be documented as "you set this to express what you want." A field that is provider-specific should be documented as "this is AWS/GCP/Azure-specific; only relevant when running on X." A field that is reconciler bookkeeping should be documented as "internal; do not rely on this."

A well-described CRD is also a forcing function for the design. If you cannot write a clear one-sentence description of what a field does and who sets it, that is a signal the field's purpose is ambiguous — which usually means the intent/provider boundary in the design is also ambiguous at that point.
