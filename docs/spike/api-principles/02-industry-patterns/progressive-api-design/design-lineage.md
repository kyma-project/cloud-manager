# Design Lineage — From Research to Proposal

> This document traces every major design decision in the Progressive API Design proposal
> back to the tools and projects that inspired it, explains what problem each source was
> solving, and is explicit about what we took, what we left behind, and why.
>
> It is written for anyone who wants to understand not just *what* we are proposing
> but *why* — where each idea came from and what evidence supports it.

---

## The Starting Question

Before any tool was examined, the goal was stated as:

> *Deploy the same object across all cloud providers without the need to learn anything
> to get started — but provide the capabilities to personalise their need if required.*

That sentence contains three distinct design requirements:

| Requirement | Design challenge |
|---|---|
| **Same object across all clouds** | One CRD per resource, not one per cloud |
| **No learning required to get started** | Sensible defaults that work; no cloud-specific fields required |
| **Personalise when needed** | Escape hatches that are accessible but not mandatory |

No single tool in the industry fully solves all three. Each tool makes a different trade-off. The research task was to find which tool solved which requirement best, and then combine those solutions.

---

## The Tools Examined and Their Trade-Offs

Before going decision-by-decision, here is the honest summary of where each tool excels and where it falls short against our three requirements:

| Tool | Same object across clouds | No learning required | Personalise when needed |
|---|---|---|---|
| **Crossplane** | ✅ One XRD (Claim) works on any cloud | ⚠️ User must choose a `compositionSelector` | ✅ Composition maps to full provider API |
| **kro** | ❌ One RGD per cloud, separate Kinds | ✅ Minimal schema per RGD | ✅ CEL expressions expose full provider fields |
| **Kratix** | ✅ One Promise CRD routes to any cloud | ✅ Single minimal spec | ✅ Pipeline script exposes anything |
| **CAPI** | ⚠️ One base Kind, but user creates two objects | ⚠️ User must know their provider | ✅ Typed provider resources |
| **ACK / ASO / KCC** | ❌ Per-cloud resource Kinds | ❌ User must know provider API | ✅ Full provider fidelity |
| **Kubernetes PVC** | ✅ One Kind, any storage backend | ✅ `storage: 10Gi` requires zero cloud knowledge | ⚠️ StorageClass `parameters` is untyped |
| **Gardener DNSEntry** | ✅ One Kind, any DNS provider | ✅ Zero cloud knowledge needed | ⚠️ Limited customisation surface |

The proposal does not copy any one of these. It takes the strongest idea from each.

---

## Decision 1 — One CRD per Resource Type, Not One per Cloud

**Design rule:** `kind: VpcPeering` works on AWS, GCP, and Azure. There is no `AwsVpcPeering`, `GcpVpcPeering`, or `AzureVpcPeering` at the user-facing layer.

### Where this came from

**Kubernetes PersistentVolumeClaim** is the canonical example.
[`kubernetes/api — core/v1/types.go`](https://github.com/kubernetes/api/blob/master/core/v1/types.go)

```yaml
# Works on EBS (AWS), Persistent Disk (GCP), Azure Disk, Ceph, NFS, local disk...
kind: PersistentVolumeClaim
spec:
  resources:
    requests:
      storage: 10Gi
```

Nobody writes `AwsPersistentVolumeClaim`. The storage backend is a runtime routing concern, not a user-authored kind name. This has been the Kubernetes standard since 2016.

**Kubernetes Service** is another example. `kind: Service` with `type: LoadBalancer` works on any cloud. Nobody writes `AwsLoadBalancerService`.

**Gardener DNSEntry** is the most direct analogy for our case.
[`gardener/external-dns-management`](https://github.com/gardener/external-dns-management/blob/master/pkg/apis/dns/v1alpha1/dnsentry.go)

```yaml
kind: DNSEntry
spec:
  dnsName: "my-service.example.com"
  targets: ["1.2.3.4"]
  # No "provider: route53" field. No "provider: cloud-dns" field.
  # Controller routes at runtime from DNSProvider.spec.domains.include
```

Gardener's insight: the provider is determined by matching the domain against registered `DNSProvider` objects, entirely at runtime. The user spec has no provider field at all.

### What we left behind

**kro** takes the opposite approach: one `ResourceGraphDefinition` generates one Kind per cloud. `VpcPeeringAWS`, `VpcPeeringGCP`, `VnetPeeringAzure` are separate types. This is intentional in kro — it trades cross-cloud portability for provider completeness.

We rejected this because it fails the first requirement: the user must know which cloud they are on to choose the right Kind. A user who deploys an application to multiple clusters cannot write one manifest.

**ACK, ASO, and KCC** also use per-cloud Kinds. These are the right choice when you want maximum provider fidelity with no abstraction — but that is the opposite of our goal.

---

## Decision 2 — Cloud Is Detected at Runtime, Never a User Field

**Design rule:** There is no `spec.cloud: aws`, `spec.provider: gcp`, or `compositionSelector` in any user-facing spec. The controller reads the cloud provider from Kyma Scope.

### Where this came from

**Gardener DNSEntry** (same source as Decision 1) makes this explicit: the `DNSProvider` is a cluster-level object. The `DNSEntry` carries no provider reference. Routing is purely a runtime concern based on domain matching.

**Kubernetes StorageClass** makes the same separation. The `PVC` does not say `provisioner: ebs.csi.aws.com`. That is in the `StorageClass`. The user picks a `storageClassName` (an abstract name), not a provisioner. The controller resolves the provisioner at runtime.

**Crossplane** is the counter-example that shows why this matters. In Crossplane, the user must set a `compositionSelector`:

```yaml
# Crossplane — user must know what compositionSelector to set
spec:
  compositionSelector:
    matchLabels:
      cloud: aws     # ← user must write this
```

This breaks "no learning required" — the user must know:
1. That compositionSelectors exist
2. What labels the platform team used
3. Which cloud they are on

Kyma already knows the cloud from the Kyma Scope. The controller can route without asking the user.

### What we left behind

The **Crossplane `compositionSelector`** pattern. It is elegant when users are making a deliberate choice between Compositions (e.g. "use the high-availability Composition, not the standard one"). But for cloud routing, it is unnecessary friction — the cluster can only be one cloud at a time.

---

## Decision 3 — Neutral Field Names and Values in the Base Spec

**Design rule:** Fields that represent the same concept on all clouds use the same name and a cloud-neutral value format everywhere. The controller translates to provider vocabulary internally.

### Where this came from

**Crossplane's Composition model** is the clearest demonstration of this principle in production.
[`crossplane/crossplane — apis/apiextensions/v1/composition_types.go`](https://github.com/crossplane/crossplane/blob/master/apis/apiextensions/v1/composition_types.go)

A Crossplane Composition maps a neutral Claim field to a provider Managed Resource field, including value transformation:

```yaml
# User writes once (Claim — neutral):
spec:
  engineVersion: "7.0"

# Composition maps to GCP internally:
# "7.0" → "REDIS_7_0"

# Composition maps to AWS internally:
# "7.0" → "7.0" (passthrough)
```

The user writes `engineVersion: "7.0"` on any cloud. The transformation is the Composition's (or in our case, the controller's) responsibility.

**Kubernetes API conventions** formalise this as a rule.
[`kubernetes/community — contributors/devel/sig-architecture/api-conventions.md`](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#naming-conventions)

> Fields representing the same concept use the same name across all resource types.

`spec.replicas` means replicas everywhere. `spec.selector` means selector everywhere. CM's `replicasPerShard` must mean replicas-per-shard everywhere — not `replicasPerPrimary` on Azure.

**Kyma's own API principles spike** (Patterns 2 and 3) documents the specific inconsistencies this principle would fix in Cloud Manager:

| Current inconsistency | After applying the principle |
|---|---|
| `redisVersion: "REDIS_7_0"` (GCP) vs `engineVersion: "7.0"` (AWS) | `engineVersion: "7.0"` everywhere |
| `replicasPerPrimary` (Azure) vs `replicasPerShard` (others) | `replicasPerShard` everywhere |
| `capacityGb: 1024` (GCP/SAP) vs `capacity: "1Ti"` (AWS/Azure) | `capacity: "1Ti"` everywhere |
| `redisConfigs` (GCP) vs `parameters` (AWS) vs `redisConfiguration` (Azure) | `parameters` everywhere |

### What we left behind

**ACK, ASO, and KCC** expose provider vocabulary directly in the spec. `ReplicationGroup.spec.engineVersion` on ACK, `redis.instances.redisVersion` via KCC, `RedisCacheEnterprise.spec.sku.name` on ASO. These are correct for users who know the provider. They are the wrong model for a user who should not need to know the provider.

---

## Decision 4 — The Three-Layer Structure

**Design rule:** The API has three additive layers — zero-config intent (Layer 1), neutral tuning (Layer 2), and typed provider escape hatch (Layer 3). Each layer adds capability; no layer removes what a lower layer provided.

### Where this came from

No single tool implements exactly this structure. It is a synthesis.

**Layer 1 (zero-config intent)** comes from **Kubernetes PVC, Service, and Ingress** as the model:

```yaml
# Kubernetes PVC — Layer 1 only, already complete
kind: PersistentVolumeClaim
spec:
  resources:
    requests:
      storage: 10Gi    # the only concept the user needs to express
```

The principle: what is the minimum the user must say to get a working resource? That is Layer 1.

**Layer 2 (neutral tuning)** comes from **Crossplane's XRD schema design** combined with Kyma's **Pattern 3 (Consistent Field Naming)**. Crossplane showed that you can expose a family of shared-concept fields at the user layer and map them to different provider fields internally. Kyma's spike showed exactly which Cloud Manager fields fall into this category.

**Layer 3 (provider escape hatch)** comes from three sources:

1. **Kubernetes StorageClass `parameters`** — the escape hatch for provisioner-specific configuration:
   ```yaml
   kind: StorageClass
   provisioner: ebs.csi.aws.com
   parameters:           # opaque to Kubernetes; provisioner validates
     type: gp3
     iops: "3000"
   ```
   Lesson: some configuration is correctly provider-native. Do not abstract it away — provide a typed, optional path to it.

2. **Gateway API `parametersRef`** — the typed optional pointer to provider-specific configuration:
   [`kubernetes-sigs/gateway-api — apis/v1/gatewayclass_types.go`](https://github.com/kubernetes-sigs/gateway-api/blob/main/apis/v1/gatewayclass_types.go)
   ```yaml
   kind: GatewayClass
   spec:
     parametersRef:    # typed pointer — optional, expert users only
       group: "gateway.example.com"
       kind: "NginxGatewayConfig"
   ```
   Lesson: the escape hatch should be typed (not free-form), optional, and clearly named to signal "this is provider territory."

3. **Kyma Cloud Manager KCP layer (Pattern 5)** — which already implements typed provider sub-structs correctly:
   ```yaml
   kind: RedisInstance   # KCP internal resource
   spec:
     instance:           # MaxProperties=1 — exactly one sub-struct
       gcp:
         redisVersion: "REDIS_7_0"
   ```
   Lesson: the discriminated-union typed sub-struct is already proven in CM at the KCP layer. The same structure should be applied at the SKR (user-facing) layer.

**The "additive" property** comes from observing what breaks user experience in tools that do not have it. In **CAPI**, if a user starts with a `Cluster` and needs provider-specific features, they must create a second object (`AWSCluster`). In **Crossplane**, if a user starts with a Claim and needs provider-specific fields, they must change to a different Composition. Our design avoids this: a user adds fields to the same object they already have.

### What we left behind

**kro's single-object approach** (one RGD = one Kind per cloud) is clean but trades the "same object on any cloud" property for simplicity of definition. We explicitly prioritised cross-cloud portability over simplicity of the RGD authoring experience.

**Crossplane's Composition pipeline** for value translation is YAML-based (patch-and-transform) and declarative. Our translation is Go code in the controller — type-safe, unit-testable, and without the indirection of a pipeline. The outcome is the same; the implementation is different.

---

## Decision 5 — Typed Sub-Structs for Layer 3 (Not `map[string]interface{}`)

**Design rule:** `spec.aws`, `spec.gcp`, `spec.azure` are typed Go structs with full CRD schema validation. Exception: genuinely schemaless provider content (like WAF rule sets) uses `runtime.RawExtension`.

### Where this came from

**Crossplane Managed Resources** (`spec.forProvider`) are the clearest production example of typed provider sub-structs.
[`crossplane-contrib/provider-aws — apis/cache/v1beta1/replicationgroup_types.go`](https://github.com/crossplane-contrib/provider-aws/blob/master/apis/cache/v1beta1/replicationgroup_types.go)

```go
// Crossplane provider-aws: typed sub-struct, not map[string]interface{}
type ReplicationGroupParameters struct {
    Region                  string  `json:"region"`
    CacheNodeType           string  `json:"cacheNodeType"`
    EngineVersion           *string `json:"engineVersion,omitempty"`
    ReplicasPerNodeGroup    *int64  `json:"replicasPerNodeGroup,omitempty"`
    AtRestEncryptionEnabled *bool   `json:"atRestEncryptionEnabled,omitempty"`
}
```

Every field is typed. The CRD validates them at admission time. A user who writes `cacheNodeType: 123` gets a validation error immediately, not a confusing cloud API error later.

**Kyma's KCP layer (Pattern 5)** already uses this correctly:

```go
// Existing CM code: api/cloud-control/v1beta1/redisinstance_types.go
type RedisInstanceSpec struct {
    Scope   ScopeRef   `json:"scope"`
    IpRange IpRangeRef `json:"ipRange"`
    Instance struct {
        Gcp      *RedisInstanceGcp      `json:"gcp,omitempty"`
        Aws      *RedisInstanceAws      `json:"aws,omitempty"`
        Azure    *RedisInstanceAzure    `json:"azure,omitempty"`
        Alicloud *RedisInstanceAlicloud `json:"alicloud,omitempty"`
    } `json:"instance"`
}
```

The SKR (user-facing) layer should apply the same discipline.

**Kubernetes StorageClass `parameters: map[string]string`** shows the alternative and its cost. The map is schemaless — Kubernetes cannot validate whether `type: gp3` is a valid EBS parameter or a typo. The provisioner validates at provision time, producing a slow feedback loop.

We use the typed struct for Layer 3 fields where the schema is stable (VpcPeering remote identity, Redis maintenance windows, NfsVolume tier and performance mode). We use `runtime.RawExtension` (from Kubernetes CRD tooling) only for future resources like `WafPolicy` where rule content is provider-versioned and changes independently of CM — as described in Kyma's Pattern 6.

### What we left behind

**Kratix's pipeline model** is the most powerful approach to schemaless content — the container receives raw YAML and can pass anything through. This is correct for Kratix because Kratix is a pipeline executor, not a typed controller. For CM's Go controllers, typed structs give compile-time safety that a shell script cannot.

---

## Decision 6 — Status Is Always Normalised

**Design rule:** `status.state`, `status.conditions`, and resource-specific fields like `status.primaryEndpoint` use the same field names and value formats on all clouds. The controller maps provider-specific status to the neutral vocabulary before writing it.

### Where this came from

**Cluster API's status contract** is the canonical Kubernetes example.
[`kubernetes-sigs/cluster-api — api/v1beta1/cluster_types.go`](https://github.com/kubernetes-sigs/cluster-api/blob/main/api/v1beta1/cluster_types.go)

CAPI requires every infrastructure provider (`AWSCluster`, `GCPCluster`, etc.) to populate a defined set of status fields — specifically `status.ready` and `status.failureDomains`. The base `Cluster` controller only proceeds when `infrastructureRef.status.ready = true`. This is a formal contract between the base and its providers.

Lesson: the status is a contract. Every provider must fulfil it. Tools that consume the status (monitoring dashboards, operators, CI pipelines) work on any cloud without modification.

**Kubernetes Conditions** (`status.conditions`) are the standard status reporting pattern across all Kubernetes resources. Using them means any tool that understands Kubernetes conditions (Argo CD, Flux, custom operators, `kubectl wait`) works with CM resources without modification.

**Kyma Cloud Manager already does this correctly in the KCP layer** (Pattern 3):

```yaml
# KCP RedisInstance status — normalised
status:
  memorySizeGb: 6            # neutral — same field on GCP, AWS, Azure, Alicloud
  replicaCount: 1            # neutral — same field on all providers
  primaryEndpoint: "10.0.0.5:6379"  # neutral — same field on all providers
```

The proposal applies the same discipline to the SKR layer and to resources that currently lack normalised status fields (NfsVolume has no `capacity` in status, for example).

### What we left behind

**ACK, ASO, and KCC** expose provider-specific status fields. ACK's `ReplicationGroup` status has `memberClusters`, `clusterEnabled`, `authTokenEnabled` — AWS-specific concepts. This is correct for provider-specific tools. It is not the right model for a status that must be readable without cloud knowledge.

---

## Decision 7 — Smart Defaults from Cluster Context

**Design rule:** Fields that the controller can infer from Kyma Scope (local VPC ID, region, account/project/tenant) are never required in the user spec. The user only provides what cannot be inferred.

### Where this came from

**Kubernetes itself** applies this principle universally:
- A `Pod` does not require `spec.nodeName` — the scheduler infers it
- A `Service` does not require the load balancer's region — it reads it from the node metadata
- A `PVC` does not require the storage class's provisioner endpoint — the controller discovers it

**Gardener DNSEntry** makes this most explicit for our case: the controller discovers which `DNSProvider` to use by matching the domain against registered providers at runtime. The `DNSEntry` user provides only the domain and IP — nothing about where DNS is hosted.

**Encore** (from the research) takes this furthest: a static analysis compiler reads application code and infers *all* infrastructure requirements, including IAM permissions, from the code's behaviour. The developer writes zero infrastructure configuration. We do not go this far, but the principle is the same — default to inference, only ask what cannot be inferred.

The Kyma-specific application: a cluster running on AWS in `us-east-1` knows its own VPC ID, region, and account ID. A `VpcPeering` that says `remoteVpcId: "data-network"` needs nothing else for the same-account, same-region case. The controller fills the rest.

### What we left behind

**Terraform** requires every field to be explicit — there are no smart defaults. `aws_instance` requires you to specify `ami`, `subnet_id`, `vpc_security_group_ids`. This is correct for Terraform's use case (reproducible, auditable IaC) but wrong for our use case (a developer who just needs something to work).

---

## The Synthesis

Each design decision traces to a specific source. Here is the complete lineage in one view:

| Design element | Primary source | Secondary source | Rejected alternative |
|---|---|---|---|
| One Kind per resource, not one per cloud | Kubernetes PVC / Service / DNSEntry | Crossplane XRD/Claim model | kro (one RGD per cloud) |
| Cloud detected from cluster context, not user field | Gardener DNSEntry | Kubernetes StorageClass / PVC | Crossplane `compositionSelector` |
| Neutral field names (same name for same concept) | Crossplane Composition + Kubernetes API conventions | Kyma spike Pattern 3 | ACK / ASO / KCC (provider vocabulary in spec) |
| Neutral field values (`"7.0"` not `"REDIS_7_0"`) | Crossplane value mapping in Compositions | Kyma spike Pattern 2 | GCP/SAP `capacityGb` integer leaking into spec |
| Three-layer additive structure | Synthesis: PVC (L1) + Crossplane (L2) + StorageClass/Gateway API (L3) | Kyma spike Patterns 1, 2, 3, 4, 5 | None — this structure is new |
| Typed sub-structs for Layer 3 | Crossplane `spec.forProvider` typed struct | Kyma KCP layer Pattern 5 | `map[string]interface{}` (untyped) |
| `runtime.RawExtension` for schemaless payloads | Kubernetes `x-kubernetes-preserve-unknown-fields` | Kyma spike Pattern 6 | Typed struct (too rigid for versioned payloads) |
| Status always normalised | CAPI status contract | Kyma KCP layer (already does this) | ACK/ASO/KCC provider-specific status |
| Smart defaults from Kyma Scope | Kubernetes scheduler / service controller | Gardener DNSEntry runtime routing | Terraform explicit-everything |
| Layer 3 cloud mismatch validation | CEL validation rules | CRD `maxProperties` | Runtime error from cloud API |

---

## What Each Tool Contributed in One Sentence

**Kubernetes PVC / Service / DNSEntry →** The single Kind that works on any backend is a proven, widely-adopted pattern — not a novel idea. We have 10 years of evidence it works.

**Crossplane →** Showed that neutral user fields + controller-side value translation is achievable at scale and that the Composition (mapping) layer can be completely invisible to the user.

**kro →** Showed that CEL-based resource graphs with typed schemas are clean for single-provider use cases, and confirmed — by contrast — that requiring one Kind per cloud fails the portability requirement.

**Kratix →** Showed that the pipeline model is the best fit for schemaless, arbitrary provider content (Pattern 6). Also confirmed that routing from a single CRD to any cloud via a pipeline is practical.

**CAPI →** Provided the status contract model and the typed provider resource pattern that Kyma's KCP layer already uses correctly.

**ACK / ASO / KCC →** Provided the exact field names and schemas for Layer 3 sub-structs. When a user sets `spec.aws`, they write AWS vocabulary — and ACK's CRD definitions are the authoritative source for what that vocabulary is.

**Gateway API →** Provided the `parametersRef` pattern: the escape hatch is typed, optional, and clearly named. Users who do not need it never see it.

**Kyma Cloud Manager API Principles Spike →** Provided the specific field-level analysis of which Cloud Manager fields are inconsistent today and what the corrected forms should look like. Without this, the proposal would be general principles without concrete CRD-level guidance.

---

## What We Explicitly Did Not Do

**We did not adopt Crossplane.** Crossplane is the closest tool to our goal, but it requires users to choose a `compositionSelector`, it stores state in Kubernetes etcd rather than deriving it from Kyma Scope, and it introduces a separate operator and provider binary dependency. We took its ideas without taking its architecture.

**We did not adopt kro.** kro's per-cloud RGD model is clean but produces separate Kinds per cloud. A user cannot write one manifest and deploy it anywhere.

**We did not adopt Kratix.** Kratix's pipeline model is more flexible than what we need. CM's controllers are Go programs with type safety — not shell scripts in containers. We took Pattern 6 (schemaless payload) from Kratix's conceptual model but not its execution model.

**We did not build a universal abstraction.** Resources where concepts differ structurally (VpcPeering remote identity models, NfsVolume performance tiers) are not forced into a false shared vocabulary. Layer 3 exists precisely to avoid that. The goal is *progressive disclosure*, not *forced lowest-common-denominator abstraction*.

---

## Sources

| Source | What it contributed |
|---|---|
| [Kubernetes PVC](https://github.com/kubernetes/api/blob/master/core/v1/types.go) | Single Kind, any backend; `capacity: "10Gi"` as neutral intent |
| [Kubernetes Service](https://github.com/kubernetes/api/blob/master/core/v1/types.go) | `type: LoadBalancer` works on any cloud |
| [Kubernetes API conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md) | Same concept = same field name rule |
| [Kubernetes StorageClass](https://github.com/kubernetes/api/blob/master/storage/v1/types.go) | `parameters` as typed escape hatch; `reclaimPolicy` as neutral envelope |
| [Gardener DNSEntry](https://github.com/gardener/external-dns-management/blob/master/pkg/apis/dns/v1alpha1/dnsentry.go) | Cloud routing from cluster context — no provider field in user spec |
| [Crossplane Composition](https://github.com/crossplane/crossplane/blob/master/apis/apiextensions/v1/composition_types.go) | Neutral schema + controller-side value translation |
| [Crossplane provider-aws forProvider](https://github.com/crossplane-contrib/provider-aws/blob/master/apis/cache/v1beta1/replicationgroup_types.go) | Typed sub-struct for provider-specific fields |
| [CAPI Cluster types](https://github.com/kubernetes-sigs/cluster-api/blob/main/api/v1beta1/cluster_types.go) | Status contract; typed infrastructure provider pattern |
| [Gateway API GatewayClass](https://github.com/kubernetes-sigs/gateway-api/blob/main/apis/v1/gatewayclass_types.go) | Typed optional `parametersRef` escape hatch |
| [kro ResourceGraphDefinition](https://kro.run/docs/concepts/rgd/overview/) | CEL-based resource graphs; per-cloud RGD as contrast case |
| [Kratix Promise model](https://docs.kratix.io/main/reference/promises/intro) | Pipeline-based routing; schemaless pass-through payload |
| [ACK ec2-controller](https://aws-controllers-k8s.github.io/community/) | AWS field vocabulary for Layer 3 `spec.aws` sub-struct |
| [ASO v2](https://azure.github.io/azure-service-operator/reference/network/) | Azure field vocabulary for Layer 3 `spec.azure` sub-struct |
| [Google Config Connector](https://cloud.google.com/config-connector/docs/reference/resource-docs/compute/computenetworkpeering) | GCP field vocabulary for Layer 3 `spec.gcp` sub-struct |
| [Kyma Cloud Manager API Principles spike](https://github.com/kyma-project/cloud-manager/tree/spike-api-principles/docs/spike/api-principles/02-industry-patterns) | Specific inconsistencies to fix; field-level before/after analysis; six-pattern framework |
