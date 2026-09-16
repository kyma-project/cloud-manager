# Progressive API Design for Cloud-Neutral Resource Abstraction

> **Context:** This document synthesises the industry research on Crossplane, kro, Kratix, CAPI, Gateway API, StorageClass, and Kyma Cloud Manager's own API principles spike into a concrete design model. The goal is not to adopt any of those tools — it is to implement an API where users can deploy the same object across all cloud providers without learning anything to get started, while retaining the ability to personalise their specific needs when required.

---

## The Core Design Goal

A user should be able to write this on an AWS cluster, a GCP cluster, and an Azure cluster and have it work on all three — without changing a single character:

```yaml
kind: VpcPeering
metadata:
  name: app-to-data
spec:
  remoteVpcId: "my-peer-network"
```

That is the zero-config baseline. Everything else is optional layering on top.

---

## What This Pattern Is Called

The industry has several names for this approach. The underlying ideas are:

- **Progressive Disclosure** (UX/product design) — show the minimal surface first; reveal complexity on demand
- **Convention over Configuration** (Ruby on Rails, Kubernetes itself) — provide sensible defaults for everything that can be defaulted; only require what cannot be inferred
- **Golden Path / Paved Road** (Platform Engineering) — the path of least resistance does the right thing; escape hatches exist but are not forced on beginners
- **Capability-oriented API** — the user expresses a capability they need ("I need this network to talk to that one"), not a cloud operation ("call CreateVpcPeeringConnection with these parameters")

The combination of these in a Kubernetes CRD context produces a **three-layer API**:

```
Layer 0 — Cloud detection     (controller concern, invisible to user)
Layer 1 — Zero-config intent  (what every user writes)
Layer 2 — Neutral tuning      (shared vocabulary, same across clouds)
Layer 3 — Provider escape     (typed sub-struct, power users only)
```

---

## The Three Layers Explained

### Layer 0 — Cloud Detection (Not a User Concern)

The cloud provider is **never a user-facing field**. Kyma already knows which cloud the cluster runs on — this information is available from the Kyma Scope or cluster metadata. The controller reads it at runtime.

**What this means for the API:** There is no `spec.cloud: aws` field. There is no `spec.provider: gcp` field. There is no `compositionSelector`. The user writes the same CRD on any cluster and the controller figures out the rest.

```yaml
# The user does NOT write this:
spec:
  cloud: aws           # ← never appears in the user spec
  provider: gcp        # ← never appears
  compositionSelector:
    matchLabels:
      cloud: azure     # ← never appears
```

**Controller responsibility:** Read cloud provider from Kyma Scope/cluster context. Route to the correct provider implementation. The user is entirely insulated from this routing.

---

### Layer 1 — Zero-Config Intent

The minimal valid resource. Contains only fields whose meaning is identical on every cloud. The user expresses *what* they want, not *how* to provision it. Smart defaults handle everything else.

**Design rules for Layer 1 fields:**
1. Every required field must have the same name and the same semantic meaning on all clouds
2. Every optional field must also have the same name and semantic meaning on all clouds
3. Values use cloud-neutral formats (k8s `resource.Quantity` not `capacityGb: 1024`; `"7.0"` not `"REDIS_7_0"`)
4. The controller fills in all provider-specific details not expressible in neutral terms

**VpcPeering — Layer 1:**
```yaml
kind: VpcPeering
metadata:
  name: app-to-data
spec:
  remoteVpcId: "my-peer-network"    # the only required concept
                                    # controller fills: region, routing, auth,
                                    # peering connection ID, route table entries
```

**RedisInstance — Layer 1:**
```yaml
kind: RedisInstance
metadata:
  name: session-cache
spec:
  memorySizeGb: 6       # neutral — same unit across all clouds
  engineVersion: "7.0"  # neutral — controller maps to provider format
```

**NfsVolume — Layer 1:**
```yaml
kind: NfsVolume
metadata:
  name: shared-storage
spec:
  capacity: "1Ti"       # k8s Quantity — neutral, same on all clouds
                        # controller converts to capacityGb integer for GCP/SAP APIs
```

At Layer 1, the user learns **nothing about the cloud**. They learn the concept (I need object storage, I need a cache, I need a peering connection). The CRD documentation can be written entirely in product terms — no cloud-specific knowledge required.

---

### Layer 2 — Neutral Tuning

Optional fields that express a user decision which exists on all clouds but may use different names or value formats in each provider's API. The field name and value format in the CRD are cloud-neutral. The controller translates to the provider's vocabulary.

**Design rules for Layer 2 fields:**
1. Field name is the same across all clouds (Pattern 3: Consistent Field Naming)
2. Value format is cloud-neutral where possible (e.g. `"7.0"` not `"REDIS_7_0"`)
3. The mapping from neutral value to provider value is in the controller — not in the user spec
4. Fields only appear at Layer 2 if the concept genuinely exists on all clouds; if it only exists on some, it belongs in Layer 3

**VpcPeering — Layer 2:**
```yaml
kind: VpcPeering
metadata:
  name: app-to-data
spec:
  remoteVpcId: "my-peer-network"
  # Layer 2 — neutral tuning
  deleteRemotePeering: true      # exists on AWS, GCP, Azure — same concept
  allowDnsResolution: true       # exists on AWS, GCP, Azure — controller maps to each
```

**RedisInstance — Layer 2:**
```yaml
kind: RedisInstance
metadata:
  name: session-cache
spec:
  memorySizeGb: 6
  engineVersion: "7.0"
  # Layer 2 — neutral tuning
  replicasPerShard: 1            # was: replicasPerPrimary on Azure, replicasPerShard elsewhere
                                 # controller maps "replicasPerShard" → Azure's field name
  parameters:                    # consistent field name — was: redisConfigs (GCP), parameters (AWS)
    maxmemory-policy: volatile-lru
```

**NfsVolume — Layer 2:**
```yaml
kind: NfsVolume
metadata:
  name: shared-storage
spec:
  capacity: "1Ti"
  # No additional Layer 2 fields for NfsVolume — performance and tier are genuinely provider-specific
```

At Layer 2, the user learns **neutral cloud concepts**. They do not need to know that GCP calls the field `redisVersion` with value `REDIS_7_0`, or that Azure calls replicas `replicasPerPrimary`. They write what they mean and the controller handles the translation. The API documentation at this layer reads like a product spec, not a cloud SDK reference.

---

### Layer 3 — Provider Escape Hatch

Optional typed sub-structs, one per cloud provider, for configuration that is genuinely provider-specific — either because the concept does not exist on other clouds, or because the value space is provider-native and cannot be meaningfully abstracted.

**Design rules for Layer 3:**
1. The sub-struct key matches the cloud name: `spec.aws`, `spec.gcp`, `spec.azure`
2. Exactly one sub-struct may be set (CRD validation: `maxProperties: 1` on the `spec` extension section or via CEL rule)
3. The controller validates that the populated sub-struct matches the cloud the cluster is on; if it does not match, the controller rejects with a clear error message
4. Values in Layer 3 use the provider's own vocabulary — users writing `spec.aws` are expected to know AWS
5. Layer 3 fields have full CRD schema validation per provider (typed struct, not `map[string]interface{}`) — unless the content is truly schemaless (see Pattern 6 below)

**VpcPeering — Layer 3:**
```yaml
kind: VpcPeering
metadata:
  name: app-to-data
spec:
  remoteVpcId: "vpc-0abc123"
  deleteRemotePeering: true
  # Layer 3 — AWS-specific escape hatch
  aws:
    remoteRegion: "us-west-2"          # cross-region — no GCP/Azure equivalent
    remoteAccountId: "123456789012"    # cross-account IAM — no GCP/Azure equivalent
    remoteRouteTableUpdateStrategy: "AUTO"  # route propagation — no equivalent elsewhere

# OR on a GCP cluster:
spec:
  remoteVpcId: "my-peer-network"
  deleteRemotePeering: true
  gcp:
    remoteProject: "my-peer-project"   # GCP project scope — no AWS/Azure concept
    importCustomRoutes: true           # GCP routing option — no equivalent elsewhere

# OR on an Azure cluster:
spec:
  remoteVpcId: "/subscriptions/.../virtualNetworks/peer-vnet"
  deleteRemotePeering: true
  azure:
    remoteTenant: "00000000-..."       # cross-tenant AAD — no AWS/GCP concept
    useRemoteGateway: false            # Azure gateway transit — no equivalent
```

**RedisInstance — Layer 3:**
```yaml
kind: RedisInstance
metadata:
  name: session-cache
spec:
  memorySizeGb: 6
  engineVersion: "7.0"
  aws:
    autoMinorVersionUpgrade: true       # AWS-only
    preferredMaintenanceWindow: "sat:15:00-sat:16:00"  # AWS-specific format
  # OR:
  gcp:
    maintenancePolicy:                  # GCP-specific structure
      dayOfWeek:
        day: "SATURDAY"
        startTime:
          hours: 15
  # OR:
  azure:
    minimumTlsVersion: "1.2"           # Azure-specific
    publicNetworkAccess: "Disabled"
```

At Layer 3, the user is **a cloud expert deliberately reaching for provider power**. The API makes this explicit — the sub-struct name (`aws:`, `gcp:`, `azure:`) signals "you are now in provider territory." The API documentation for Layer 3 can reference the provider's own docs directly.

---

## The Full API Shape

All three layers together, for VpcPeering:

```yaml
# ── COMPLETE API SHAPE: VpcPeering ──────────────────────────────────────────
kind: VpcPeering
metadata:
  name: app-to-data
spec:

  # ── Layer 1: Zero-config intent ─────────────────────────────────────────
  # required
  remoteVpcId: string

  # ── Layer 2: Neutral tuning ─────────────────────────────────────────────
  # optional — same name and semantics on all clouds
  deleteRemotePeering: bool    # default: false
  allowDnsResolution: bool     # default: true  — controller maps to each provider

  # ── Layer 3: Provider escape hatch ─────────────────────────────────────
  # optional — at most one may be set; must match current cloud
  # (CRD validation: maxProperties: 1 across aws/gcp/azure)
  aws:
    remoteRegion:                  string  # optional
    remoteAccountId:               string  # optional
    remoteRouteTableUpdateStrategy: string  # optional, enum: AUTO | MANUAL

  gcp:
    remoteProject:       string  # optional
    importCustomRoutes:  bool    # optional

  azure:
    remoteTenant:        string  # optional
    remotePeeringName:   string  # optional (GCP has this too — see note below)
    useRemoteGateway:    bool    # optional

# ── STATUS: normalised output — same fields on all clouds ───────────────────
status:
  state: string          # Ready / Creating / Error — same on all clouds
  peeringId: string      # cloud-assigned connection ID — normalised field name
  conditions: []         # standard Kubernetes condition list
```

---

## The Controller's Responsibilities

The user writes three layers. The controller has three corresponding jobs.

### Job 1 — Cloud Detection and Routing

The controller reads the Kyma Scope to determine the current cloud provider. This never touches the user spec.

```
Kyma Cluster Context
  └── scope.spec.cloudProvider = "aws" | "gcp" | "azure"
         │
         └──▶ Route to aws.VpcPeeringReconciler
                         gcp.VpcPeeringReconciler
                         azure.VpcPeeringReconciler
```

### Job 2 — Neutral Value Translation (Layer 2 fields)

For each Layer 2 field, the controller maps the neutral value to the provider's vocabulary:

| Neutral field | AWS | GCP | Azure |
|---|---|---|---|
| `engineVersion: "7.0"` | `engineVersion: "7.0"` (passthrough) | `redisVersion: "REDIS_7_0"` | `redisVersion: "7.0"` (passthrough) |
| `replicasPerShard: 1` | `replicasPerNodeGroup: 1` | `replicaCount: 1` | `replicasPerPrimary: 1` |
| `capacity: "1Ti"` | `capacityBytes: 1099511627776` | `capacityGb: 1024` | `diskSizeGB: 1024` |
| `allowDnsResolution: true` | `accepter.allowDnsResolutionFromRemoteVpc: true` | (handled via network config) | (handled via Azure private DNS) |

These mappings live in the controller as Go functions — type-safe, testable, and invisible to the user.

### Job 3 — Smart Defaults for Layer 1 Fields

When a Layer 1 field is minimal (e.g. `remoteVpcId` is the only thing the user provides), the controller fills in everything else the cloud API requires from cluster context:

| Cloud | What the controller infers | Source |
|---|---|---|
| AWS | `region` (requester) | from cluster metadata |
| AWS | `vpcId` (requester) | from Kyma Scope |
| AWS | `accepterRegion` | from `remoteVpcId` pattern or same as requester |
| GCP | `network` (self-link) | built from project + `remoteVpcId` |
| GCP | `project` | from Kyma Scope |
| Azure | VNet resource group | from Kyma Scope |

The user who writes only `remoteVpcId: "my-peer"` gets a correctly provisioned peering connection. The user who needs cross-account peering adds `spec.aws.remoteAccountId` — they know they need it because they know their architecture.

---

## Layer Validation Rules

The controller enforces these at admission time (via CRD CEL rules or a validating webhook):

### Rule 1 — Layer 3 cloud mismatch is rejected

```
IF spec.aws is set AND cluster.cloud != "aws" THEN
  Reject with: "spec.aws is only valid on AWS clusters. This cluster is GCP."
```

### Rule 2 — Only one Layer 3 sub-struct allowed

```
IF count(aws, gcp, azure) where spec.<cloud> is set > 1 THEN
  Reject with: "Only one provider extension may be set."
```

CRD level expression (CEL):
```yaml
# In CRD validation
x-kubernetes-validations:
  - rule: >
      [has(self.aws), has(self.gcp), has(self.azure)]
        .filter(x, x).size() <= 1
    message: "At most one provider extension (aws, gcp, azure) may be set."
```

### Rule 3 — Layer 1 required fields always present

Standard CRD `required` array enforcement. No CEL needed.

---

## How the Status is Always Normalised

Status is always cloud-neutral, regardless of how the resource was created. The controller maps provider-specific status back to the common vocabulary:

```yaml
status:
  # Always present, always cloud-neutral
  state: Ready               # mapped from AWS "active", GCP "ACTIVE", Azure "Connected"
  peeringId: "pcx-0abc123"   # mapped from AWS connection ID, GCP peering name, Azure peering ID
  conditions:
    - type: Ready
      status: "True"
      reason: PeeringActive
      message: "VPC peering connection is active."
```

A tool that reads `status.state` or `status.conditions` works on any cloud without modification.

---

## How the Six Kyma Patterns Map to the Three Layers

The six patterns from the Cloud Manager API principles spike each address one part of this model:

| Pattern | Which Layer | What It Governs |
|---|---|---|
| **P2 — Neutral Intent Resource** | Layer 1 | Every Layer 1 field must be a neutral concept (k8s Quantity, dotted version strings, not cloud enums) |
| **P3 — Consistent Field Naming** | Layers 1 + 2 | Same field name for the same concept across all clouds — enforced because there is only one CRD |
| **P1 — Common Base + Extensions** | All layers | Layer 1+2 = the common base; `spec.aws/gcp/azure` = the typed extensions |
| **P4 — Provider-Specific Resource** | Layer 3 | The escape hatch sub-struct for concepts that genuinely differ structurally |
| **P5 — Typed Sub-Struct** | Layer 3 | The `spec.aws/gcp/azure` sub-structs are **typed** (full CRD schema validation), not `map[string]interface{}` |
| **P6 — Unstructured Payload** | Layer 3 (special case) | For future resources like WafPolicy where provider rule content is schemaless — the sub-struct uses `runtime.RawExtension` instead of a typed struct |

---

## What the Industry Taught Us (and What We Are Not Copying)

### From Crossplane
**Take:** The principle that the user-facing schema is pure intent and the controller/composition does the translation. The user writes `engineVersion: "7.0"` and the controller maps to `"REDIS_7_0"` for GCP. This is Pattern 3 applied systematically.

**Leave:** The need for users to choose a `compositionSelector` or know which `Composition` to use. In Kyma Cloud Manager the cloud is detected from cluster context — the user never selects it.

### From kro
**Take:** The instinct to keep genuinely different things separate. When cloud concepts differ structurally (VpcPeering remote identity models), do not force them into a false union — put them in Layer 3 typed sub-structs.

**Leave:** The one-RGD-per-cloud model that gives each cloud a different Kind. Kyma's goal is **one Kind, any cloud** — the opposite of kro's default.

### From Kratix
**Take:** The pipeline model's ability to pass schemaless content through unchanged (Pattern 6). For future resources with large, provider-versioned payloads (WAF rules, complex policy documents), `runtime.RawExtension` in the Layer 3 sub-struct lets the user write provider-native JSON without CM needing to schema it.

**Leave:** The Promise/pipeline execution model itself. CM implements this in Go controllers, not container pipelines.

### From CAPI
**Take:** The status contract model — `status.ready` and `status.conditions` are the universal readiness signal regardless of what happened inside the provider reconciler.

**Leave:** The two-object model (user writes base + provider extension separately). CM generates the internal representation — the user creates one object.

### From Kubernetes StorageClass
**Take:** The `parameters: map[string]string` pass-through for provider-native config. This is the right model for Redis `parameters` (maxmemory-policy, etc.) — CM validates the envelope, the cloud validates the keys.

**Leave:** The lack of typed validation. CM's Layer 3 sub-structs are typed wherever the schema is stable; only the genuinely schemaless payloads use `runtime.RawExtension`.

### From Kubernetes PVC / Service / DNSEntry
**Take:** The core UX principle — `capacity: "10Gi"` means the same thing on every storage backend. `port: 80` means the same thing on every cloud load balancer. This is what Layer 1 must achieve for every CM resource.

---

## Concrete Examples Across All Three Layers

### VpcPeering — Same scenario, three levels of user sophistication

```yaml
# ── Level 1: New user, no cloud knowledge ────────────────────────────────────
kind: VpcPeering
metadata:
  name: app-to-data
spec:
  remoteVpcId: "data-network"
# Controller fills: cloud detection, local VPC ID, region, routing tables, DNS options


# ── Level 2: User who knows their architecture ───────────────────────────────
kind: VpcPeering
metadata:
  name: app-to-data
spec:
  remoteVpcId: "data-network"
  deleteRemotePeering: true      # "clean up the other side when I delete this"
  allowDnsResolution: false      # "I manage DNS myself"


# ── Level 3: AWS engineer doing cross-account peering ────────────────────────
kind: VpcPeering
metadata:
  name: app-to-data
spec:
  remoteVpcId: "vpc-0abc123def456789"
  deleteRemotePeering: true
  aws:
    remoteAccountId: "123456789012"     # I know this is cross-account
    remoteRegion: "us-west-2"           # I know the peer is in a different region
    remoteRouteTableUpdateStrategy: "AUTO"
```

All three manifests use the same `kind: VpcPeering`. A user can start at Level 1 and graduate to Level 3 over time — each step adds what they understand, never requires them to rewrite what they already wrote.

### RedisInstance — Same pattern

```yaml
# ── Level 1: New user ────────────────────────────────────────────────────────
kind: RedisInstance
metadata:
  name: session-cache
spec:
  memorySizeGb: 6
  engineVersion: "7.0"


# ── Level 2: User who cares about HA and config ──────────────────────────────
kind: RedisInstance
metadata:
  name: session-cache
spec:
  memorySizeGb: 6
  engineVersion: "7.0"
  replicasPerShard: 1
  parameters:
    maxmemory-policy: volatile-lru


# ── Level 3: GCP engineer with maintenance window requirements ───────────────
kind: RedisInstance
metadata:
  name: session-cache
spec:
  memorySizeGb: 6
  engineVersion: "7.0"
  replicasPerShard: 1
  parameters:
    maxmemory-policy: volatile-lru
  gcp:
    maintenancePolicy:
      dayOfWeek:
        day: "SATURDAY"
        startTime:
          hours: 15
    persistenceMode: "RDB"
```

### NfsVolume — Capacity unification

```yaml
# ── Level 1: Exactly the same on AWS, GCP, Azure, SAP ───────────────────────
kind: NfsVolume
metadata:
  name: shared-data
spec:
  capacity: "1Ti"     # controller converts: GCP → capacityGb: 1024
                      #                      SAP → capacityGb: 1024
                      #                      AWS → matches EFS elastic sizing
                      #                      Azure → diskSizeGB: 1024
```

---

## What "No Learning Required" Means in Practice

The test for whether a CM resource achieves the goal is this question:

> **Can a developer who has never used AWS, GCP, or Azure provision a working [RedisInstance / VpcPeering / NfsVolume] by following only the Kyma Cloud Manager docs — without once visiting AWS Console docs, GCP Cloud docs, or Azure Portal docs?**

If the answer is yes at Layer 1, the resource meets the standard. Layers 2 and 3 are for users who choose to go deeper — but going deeper is always optional, never required for basic use.

---

## Summary: The Design Contract

| Layer | What the user writes | What the user needs to know | Controller's job |
|---|---|---|---|
| **Layer 1** (Zero-config) | Minimal required fields only | The concept they want | Detect cloud, apply all defaults, call provider API |
| **Layer 2** (Neutral tuning) | Shared concept fields with neutral names | What the concept means (not which cloud calls it what) | Map neutral field name + value to provider vocabulary |
| **Layer 3** (Provider escape) | `spec.aws/gcp/azure` typed sub-struct | The provider's own API for that feature | Validate cloud match, pass through to provider API |
| **Status** (always) | Reads normalised output | `status.state` and `status.conditions` | Map provider-specific status to neutral output fields |

This is not a compromise between usability and power. It is **additive** — each layer adds capability without removing what lower layers gave you. A user at Layer 1 never feels blocked; a power user at Layer 3 never feels constrained.

---

## Sources and Inspirations

| Source | What it contributed |
|---|---|
| [Kyma Cloud Manager API Principles spike](https://github.com/kyma-project/cloud-manager/tree/spike-api-principles/docs/spike/api-principles/02-industry-patterns) | The six patterns framework; VpcPeering and RedisInstance field analysis; GcpSubnet exception rationale |
| [Kubernetes PersistentVolumeClaim](https://github.com/kubernetes/api/blob/master/core/v1/types.go) | `capacity: "10Gi"` as the canonical neutral-intent field |
| [Gardener DNSEntry](https://github.com/gardener/external-dns-management/blob/master/pkg/apis/dns/v1alpha1/dnsentry.go) | Runtime cloud routing from cluster context, not user-provided selector |
| [Kubernetes StorageClass](https://github.com/kubernetes/api/blob/master/storage/v1/types.go) | `parameters: map[string]string` for schemaless pass-through payload |
| [Crossplane Compositions](https://docs.crossplane.io/latest/composition/compositions/) | Neutral intent schema + controller-side value translation |
| [Cluster API ClusterClass](https://github.com/kubernetes-sigs/cluster-api/blob/main/api/v1beta1/cluster_types.go) | Status contract pattern (ready condition, failure domains) |
| [Gateway API GatewayClass](https://github.com/kubernetes-sigs/gateway-api/blob/main/apis/v1/gatewayclass_types.go) | `parametersRef` as the typed escape hatch for provider-specific extensions |
| [ACK](https://aws-controllers-k8s.github.io/community/) · [ASO](https://azure.github.io/azure-service-operator/) · [KCC](https://cloud.google.com/config-connector/docs/overview) | Provider-specific field vocabulary used in Layer 3 sub-structs |
| [Kubernetes API conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md) | Naming consistency rules (same concept = same name everywhere) |
