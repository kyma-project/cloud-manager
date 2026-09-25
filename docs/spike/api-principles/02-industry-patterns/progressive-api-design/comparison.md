# Crossplane v2 vs kro vs Kratix — Comparison Guide

> A decision-making guide for product managers, architects, and engineering leads evaluating Kubernetes-native infrastructure abstraction tools.

---

## The One-Sentence Summary of Each Tool

| Tool | In One Sentence |
|---|---|
| **Crossplane v2** | A production-ready, CNCF-graduated platform that lets you define cloud-neutral APIs (XRDs) and map them to real cloud resources through declarative Compositions — continuously reconciled via Kubernetes controllers. |
| **kro** | A simpler, single-object alternative to Crossplane Compositions, combining the API schema and resource templates in one YAML file with CEL expressions and auto-inferred dependency ordering. |
| **Kratix** | A meta-framework that turns any infrastructure tooling (Terraform, Helm, scripts) into self-service developer APIs via Promises — with first-class multi-cluster routing and a pipeline-container model. |

---

## What Problem Are They All Solving?

All three tools tackle the same root problem: **cloud infrastructure is too complex and too cloud-specific for developers to manage themselves**.

Without any of these tools:

- A developer who needs a network peering connection must learn AWS/GCP/Azure APIs
- They must wait for a platform engineer with cloud access to help them
- Cloud resources are provisioned inconsistently — different scripts, different standards, different lifecycle management per team

With any of these tools:

- The platform team builds a simplified, company-specific API once
- Developers submit a clean YAML form ("I need VPC peering between these two networks")
- The tool handles all cloud-specific complexity behind the scenes
- Infrastructure is continuously reconciled — drift is detected and corrected automatically

---

## The Analogy

Think of each tool as a different approach to running a restaurant:

| | Crossplane | kro | Kratix |
|---|---|---|---|
| **The "menu"** | XRD (the schema developers fill in) | RGD schema | Promise API (spec.api CRD) |
| **The "recipe"** | Composition (declarative YAML template with patch steps) | RGD resources (CEL expressions mapping fields) | Pipeline script (shell/Go/Python that runs in a container) |
| **The "kitchen"** | Crossplane provider plugins (talk directly to cloud APIs) | ACK / ASO / KCC (already installed, kro orchestrates them) | ACK / ASO / KCC on Destination clusters (Kratix routes to them) |
| **The "waiter"** | Developer submits an XR | Developer submits an instance of the generated Kind | Developer submits a Promise request |
| **The "franchise model"** | One menu item (XRD), different kitchens per cloud (Compositions) | Different menu items per cloud (VpcPeeringAWS, VpcPeeringGCP…) | One menu item, kitchen decides based on the order |

---

## Side-by-Side Feature Comparison

### Architectural Model

| Dimension | Crossplane v2 | kro | Kratix |
|---|---|---|---|
| **Core abstraction object** | XRD (schema) + Composition (impl) — two separate objects | ResourceGraphDefinition (RGD) — one object combining schema + resources | Promise — CRD API + Workflows (pipeline containers) |
| **Developer-facing Kind** | One XR Kind per domain (e.g., `XVPCPeering`) — works for all clouds via `compositionSelector` | One Kind per cloud (e.g., `VpcPeeringAWS`, `VpcPeeringGCP`) | One Kind for all clouds (e.g., `VpcPeering`) with `spec.cloud` field |
| **Translation mechanism** | Declarative patch-and-transform pipeline steps in YAML | CEL expressions mapping schema fields → resource fields | Imperative pipeline script (reads request, writes output YAML) |
| **Dependency ordering** | Explicit via pipeline step order + `readinessChecks` | Implicit — inferred automatically from CEL cross-references | Script-level — sequential execution in the bash/Go script |
| **Status write-back to developer** | ✅ Native — `ToCompositeFieldPath` patches write cloud values back to XR status | ✅ Native — CEL expressions in `schema.status` evaluated from resource status | ⚠️ Limited — requires additional tooling or custom scripting |
| **Multi-cluster routing** | ⚠️ Via provider-kubernetes (possible but indirect) | ❌ Not native — all resources on one cluster | ✅ First-class — Destinations with label-based routing |
| **Wraps existing tools** | ❌ No — replaces Terraform/Ansible for provisioning | ❌ No — orchestrates Kubernetes CRDs only | ✅ Yes — pipeline can call Terraform, Helm, scripts, anything |

---

### Cloud Provider Layer

All three tools rely on a **provider layer** that translates Kubernetes objects into real cloud API calls. They use different providers:

| Cloud | Crossplane | kro | Kratix |
|---|---|---|---|
| **AWS** | `upbound/provider-aws-*` (Upbound marketplace) | ACK — `ec2.services.k8s.aws` | ACK (on Destination cluster) |
| **GCP** | `upbound/provider-gcp-*` (Upbound marketplace) | KCC — `compute.cnrm.cloud.google.com` | KCC (on Destination cluster) |
| **Azure** | `upbound/provider-azure-*` (Upbound marketplace) | ASO v2 — `network.azure.com` | ASO v2 (on Destination cluster) |

> **Key difference:** Crossplane ships its own provider plugins (Upbound-maintained). kro and Kratix rely on the cloud vendors' own Kubernetes controllers (ACK/ASO/KCC), which have higher API fidelity since the vendors maintain them directly.

---

### VPC Peering: What Each Provider Creates

| Cloud | Resource | Crossplane | kro | Kratix |
|---|---|---|---|---|
| **AWS** | Peering connection | `VPCPeeringConnection` + `VPCPeeringConnectionAccepter` (Upbound) | `VPCPeeringConnection` with `acceptRequest: true` (ACK) | `VPCPeeringConnection` with `acceptRequest: true` (ACK) |
| **AWS** | Route tables | Separate Crossplane resources | Separate `Route` ACK resources | Separate `Route` ACK resources (if CIDRs provided) |
| **GCP** | Network peering | 2× `NetworkPeering` (Upbound) | 2× `ComputeNetworkPeering` (KCC) | 2× `ComputeNetworkPeering` (KCC) |
| **Azure** | VNet peering | 2× `VirtualNetworkPeering` (Upbound) | 2× `VirtualNetworksVirtualNetworkPeering` (ASO) | 2× `VirtualNetworksVirtualNetworkPeering` (ASO) |

---

### Maturity and Project Status

| Dimension | Crossplane v2 | kro | Kratix |
|---|---|---|---|
| **Status** | CNCF **Graduated** (Nov 2025) | **Alpha** — Kubernetes SIG Cloud Provider | Active OSS — maintained by Syntasso |
| **Latest version** | v2.4.0 (Aug 2026) | v0.9.3 (Jul 2026) | Active (last commit Aug 2026) |
| **Production readiness** | ✅ Production-grade | ⚠️ Evaluate carefully — alpha API | ⚠️ Used in production, but no formal versioning |
| **Backing** | CNCF + Upbound (commercial sponsor) | AWS + Azure + Google Cloud (multi-vendor SIG) | Syntasso (commercial + OSS) |
| **Community size** | Large (thousands of adopters) | Growing (early adopter phase) | Smaller but dedicated |
| **Commercial support** | Upbound (upbound.io) | AWS EKS Capabilities integration | Syntasso SKE (enterprise offering) |

---

### Developer Experience

| Dimension | Crossplane v2 | kro | Kratix |
|---|---|---|---|
| **What the developer writes** | One XR YAML, change `compositionSelector.matchLabels.cloud` to switch clouds | Different Kind per cloud (`VpcPeeringAWS`, `VpcPeeringGCP`, etc.) | One `VpcPeering` Kind with `spec.cloud: aws\|gcp\|azure` |
| **Cloud knowledge required** | None — XRD shields all cloud details | None — RGD schema shields all cloud details | None — Promise API shields all cloud details |
| **Self-service** | ✅ Yes — apply and wait | ✅ Yes — apply and wait | ✅ Yes — apply and wait |
| **Visibility into what was created** | ✅ Status fields written back to XR | ✅ Status CEL expressions show composed resource state | ⚠️ Limited status — requires custom scripting |

---

### Platform Engineer Experience

| Dimension | Crossplane v2 | kro | Kratix |
|---|---|---|---|
| **What they write** | XRD (schema) + one Composition per cloud | One RGD per cloud | One Promise (CRD + pipeline containers) |
| **Language** | YAML + `function-patch-and-transform` DSL (or KCL / Python Functions) | YAML + CEL expressions | Shell / Go / Python pipeline scripts |
| **Learning curve** | High — XRD + Composition + Provider concepts | Medium — simpler than Crossplane, CEL is learnable | Low to Medium — if comfortable with shell scripting |
| **Can wrap existing tools** | ❌ No | ❌ No | ✅ Yes — call Terraform, Helm, scripts in the pipeline |
| **Separation of schema and implementation** | ✅ Explicit — XRD and Composition are separate objects | ⚠️ Implicit — both in one RGD (less formal separation) | ✅ Explicit — `spec.api` (CRD) separate from `spec.workflows` |
| **Testing** | Crossplane testing framework | kubectl apply + observe | Local pipeline testing with mock mounts |

---

### Operations

| Dimension | Crossplane v2 | kro | Kratix |
|---|---|---|---|
| **Drift correction** | ✅ Automatic — reconciles every 30–60s | ✅ Automatic — kro reconciles composed K8s objects | ✅ Depends on ACK/ASO/KCC on Destination (they reconcile) |
| **State store** | Kubernetes etcd | Kubernetes etcd | External State Store (Git repo or S3/GCS bucket) |
| **GitOps integration** | ✅ Works natively with Flux/Argo CD (XRs are K8s manifests) | ✅ Works natively (RGD instances are K8s manifests) | ✅ Built-in — Kratix writes to GitOps store; Flux/Argo syncs |
| **Multi-cluster** | ⚠️ Via provider-kubernetes | ❌ Not native | ✅ First-class via Destinations |
| **Secrets management** | Provider credentials via Kubernetes Secrets | Via ACK/ASO/KCC credential model | Pipeline env vars referencing Kubernetes Secrets |

---

## The VPC Peering Example — Developer's View

One of the most revealing comparisons is what a developer actually writes for each tool.

### Crossplane v2 — One Kind, `compositionSelector` picks the cloud

```yaml
apiVersion: network.platform.example.io/v1alpha1
kind: XVPCPeering
metadata:
  name: app-to-data-peering
  namespace: my-team
spec:
  compositionSelector:
    matchLabels:
      cloud: aws          # ← change to "gcp" or "azure" for other clouds
  parameters:
    cloud: aws
    localVpcId: "vpc-0a1b2c3d4e5f67890"
    peerVpcId:  "vpc-0f9e8d7c6b5a43210"
    localRegion: "us-east-1"
    peerRegion:  "us-west-2"
    allowDnsResolution: true
```

### kro — Separate Kind per cloud

```yaml
# AWS
apiVersion: kro.run/v1alpha1
kind: VpcPeeringAWS          # ← different Kind for each cloud
metadata:
  name: app-to-data-peering
  namespace: platform
spec:
  requesterVpcId: vpc-0123456789abcdef0
  accepterVpcId:  vpc-0fedcba9876543210
  accepterRegion: us-west-2
  requesterCidr:  10.0.0.0/16
  accepterCidr:   10.1.0.0/16
  requesterRouteTableId: rtb-0aaa111bbb222
  accepterRouteTableId:  rtb-0ccc333ddd444
```

```yaml
# GCP — different Kind, different fields
apiVersion: kro.run/v1alpha1
kind: VpcPeeringGCP           # ← different Kind
metadata:
  name: app-to-data-peering
  namespace: config-connector
spec:
  localNetworkName: app-vpc
  peerNetworkName:  data-vpc
```

### Kratix — One Kind, `spec.cloud` routes

```yaml
# Same Kind for all clouds — just change spec.cloud
apiVersion: networking.platform.example.io/v1alpha1
kind: VpcPeering              # ← same Kind for all clouds
metadata:
  name: app-to-data-peering
  namespace: my-team
spec:
  cloud: aws                  # ← change to "gcp" or "azure"
  localVpcId: "vpc-0123456789abcdef0"
  peerVpcId:  "vpc-0fedcba9876543210"
  peerRegion: "us-west-2"
```

---

## How Each Tool Handles the Three Clouds Internally

### The Translation Chain

```
Crossplane:
  Developer's XVPCPeering
    │ compositionSelector: cloud=aws
    ▼
  Composition: xvpcpeering-aws
    │ patch-and-transform pipeline
    ▼
  Managed Resources (Upbound provider):
    VPCPeeringConnection (ec2.aws.upbound.io/v1beta1)
    VPCPeeringConnectionAccepter (ec2.aws.upbound.io/v1beta1)
    │
    ▼ provider-aws-ec2 gRPC call
    AWS EC2 API
```

```
kro:
  Developer's VpcPeeringAWS
    │ kro reconciles RGD instance
    ▼
  RGD: aws-vpc-peering (CEL expressions map fields)
    │ creates K8s objects:
    ▼
  ACK Managed Resources:
    VPCPeeringConnection (ec2.services.k8s.aws/v1alpha1)
    Route x2 (ec2.services.k8s.aws/v1alpha1)
    │
    ▼ ACK EC2 controller
    AWS EC2 API
```

```
Kratix:
  Developer's VpcPeering { cloud: aws }
    │ Kratix triggers resource.configure pipeline
    ▼
  Pipeline container runs scripts/pipeline.sh:
    Reads /kratix/input/object.yaml
    Writes /kratix/output/vpc-peering.yaml (ACK CRDs)
    Writes /kratix/metadata/destination-selectors.yaml { cloud: aws }
    │
    ▼ Kratix scheduler
  State Store (S3/Git) → Flux/ArgoCD on AWS cluster
    │
    ▼
  ACK Managed Resources applied to cluster
    │
    ▼ ACK EC2 controller
    AWS EC2 API
```

---

## Decision Framework

Use this table to identify the right tool for your situation:

| If your primary need is… | Best choice | Why |
|---|---|---|
| Production-grade, CNCF-certified multi-cloud IaC | **Crossplane v2** | Graduated project, widest provider ecosystem, continuous reconciliation |
| Simplest Composition model, no two-object split | **kro** | One RGD combines schema + resources; CEL is lightweight |
| Wrapping existing Terraform / Helm / scripts | **Kratix** | Pipeline containers can call any tool without rewriting it |
| First-class multi-cluster routing | **Kratix** | Destinations model is purpose-built for this |
| Highest cloud API fidelity (AWS/GCP/Azure-maintained) | **kro or Kratix** | Both use ACK/ASO/KCC (first-party); Crossplane uses community Upbound providers |
| One developer Kind shared across all clouds | **Crossplane** or **Kratix** | Crossplane: compositionSelector; Kratix: spec.cloud field |
| GitOps-native delivery of infrastructure | **All three** | All work with Flux/Argo CD |
| Most active community and commercial support | **Crossplane** | CNCF Graduated + Upbound commercial backing |
| Exploring / early adopter stage | **kro** | Simplest to start, multi-vendor backing |

---

## Honest Limitations of Each Tool

### Crossplane v2
- **Steep learning curve** — XRD + Composition + Providers + Functions is a lot to absorb initially
- **YAML verbosity** — Compositions can be very long (100–200+ lines for a single domain)
- **Upbound provider fidelity** — community-maintained providers occasionally lag behind cloud API changes vs. first-party ACK/ASO/KCC
- **No multi-cluster routing** built in — requires additional tooling (provider-kubernetes + ArgoCD/Flux)

### kro
- **Alpha status** — API may change; not recommended for production without extensive testing
- **No cross-cloud Kind** — developers must know which cloud Kind to use (`VpcPeeringAWS` vs `VpcPeeringGCP`)
- **No built-in multi-cluster routing** — all resources land on one cluster
- **Requires ACK/ASO/KCC pre-installed** — adds operational overhead before kro itself can be used

### Kratix
- **No CNCF graduation** — smaller community, less formal release process
- **Script maintenance** — pipeline scripts need testing and versioning just like any code
- **Limited status write-back** — the developer's request CR doesn't automatically get status fields updated from the cloud (unlike Crossplane's `ToCompositeFieldPath`)
- **More moving parts** — State Store + GitOps agent + Destination + Pipeline container adds operational complexity

---

## When to Use Them Together

These tools are not mutually exclusive. Common production patterns combine them:

### Pattern 1: Crossplane + Flux/Argo CD (Most Common)
```
Git → Flux/Argo CD → applies Crossplane XRDs/XRs → Crossplane → cloud resources
```
Flux/Argo handles GitOps delivery; Crossplane handles cloud resource lifecycle.

### Pattern 2: kro + ACK/ASO/KCC
```
Developer applies VpcPeeringAWS → kro orchestrates ACK CRDs → ACK → AWS API
```
kro adds the composition layer ACK is missing.

### Pattern 3: Kratix wrapping Crossplane
```
Developer applies VpcPeering via Kratix Promise
  → Kratix pipeline generates Crossplane XR manifests
  → Routes to a cluster running Crossplane
  → Crossplane provisions cloud resources
```
Kratix provides the multi-cluster routing and developer API; Crossplane provides the cloud provisioning engine.

---

## Summary Matrix

| Criterion | Crossplane v2 | kro | Kratix |
|---|---|---|---|
| Maturity | ⭐⭐⭐⭐⭐ CNCF Graduated | ⭐⭐ Alpha | ⭐⭐⭐ Active OSS |
| Multi-cloud abstraction | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ (per-cloud Kinds) | ⭐⭐⭐⭐⭐ |
| Declarative / YAML-native | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐ (pipeline scripts) |
| Wraps existing tooling | ⭐ | ⭐ | ⭐⭐⭐⭐⭐ |
| Multi-cluster routing | ⭐⭐ | ⭐ | ⭐⭐⭐⭐⭐ |
| Learning curve (lower = better) | ⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| Status write-back | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐ |
| Cloud API fidelity | ⭐⭐⭐ (Upbound) | ⭐⭐⭐⭐⭐ (ACK/ASO/KCC) | ⭐⭐⭐⭐⭐ (ACK/ASO/KCC) |
| Community & support | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ (growing) | ⭐⭐⭐ |
| Production readiness | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ |

---

## Alignment with Kyma Cloud Manager API Principles

The [Kyma Cloud Manager spike branch](https://github.com/kyma-project/cloud-manager/tree/spike-api-principles/docs/spike/api-principles/02-industry-patterns) documents six API design patterns derived from production Kubernetes projects (CAPI, Crossplane, Gateway API, ACK, ASO, StorageClass). Quick summary:

| Pattern | Crossplane v2 | kro | Kratix |
|---|---|---|---|
| P1 — Common Base + Provider Extensions | ✅ XRD = base; Composition = translation | 🟡 Per-RGD; convention only | 🟡 Full CRD control; convention only |
| P2 — Neutral Intent Resource | ✅ Core design philosophy | 🟡 Within one RGD; no single cross-cloud Kind | ✅ Single Promise CRD + pipeline |
| P3 — Consistent Field Naming | ✅ Enforced (one XRD) | 🟡 Convention only | ✅ Enforced (one Promise CRD) |
| P4 — Provider-Specific Resource | 🟡 Possible (separate XRDs); fights the grain | ✅ Native (one RGD per cloud) | 🟡 Separate Promises; clean trade-off |
| P5 — Typed Provider Sub-Struct | 🟡 Native in Managed Resources; manual in XRDs | 🟡 Possible; all provider CRDs must be installed | 🟡 Via `maxProperties: 1` in Promise CRD |
| P6 — Unstructured Payload | 🟡 Composition Functions required | 🔴 Incompatible with CEL type-checking | ✅ Fully native (pipeline passes raw YAML) |

**The headline finding:** Crossplane and Kratix are the strongest fits for Patterns 2–3 (neutral intent, consistent naming) because they expose a single schema. kro is the strongest fit for Pattern 4 (provider-specific resource) because separate-RGD-per-cloud is its natural model. Kratix uniquely handles Pattern 6 (unstructured payloads) — kro cannot. Crossplane explicitly appears as a reference implementation in Patterns 1, 3, and 5 of the CM document.

---

## Sources

- Crossplane v2.4.0: [docs.crossplane.io/latest/whats-new](https://docs.crossplane.io/latest/whats-new/)
- Crossplane 2.0 announcement: [blog.crossplane.io/announcing-crossplane-2-0](https://blog.crossplane.io/announcing-crossplane-2-0/)
- kro overview: [kro.run](https://kro.run)
- kro + ACK EKS example: [kro.run/examples/ack-eks-cluster](https://kro.run/examples/ack-eks-cluster/)
- kro + ACK AWS blog (Jan 2026): [aws.amazon.com/blogs/containers/simplify-kubernetes-cluster-management-using-ack-kro-and-amazon-eks](https://aws.amazon.com/blogs/containers/simplify-kubernetes-cluster-management-using-ack-kro-and-amazon-eks/)
- Kratix Promise reference: [docs.kratix.io/main/reference/promises/intro](https://docs.kratix.io/main/reference/promises/intro)
- Kratix Workflows: [docs.kratix.io/main/reference/workflows](https://docs.kratix.io/main/reference/workflows)
- Kratix Destinations: [docs.kratix.io/main/reference/destinations/multidestination-management](https://docs.kratix.io/main/reference/destinations/multidestination-management)
- ACK community: [aws-controllers-k8s.github.io/community](https://aws-controllers-k8s.github.io/community/)
- ASO v2: [azure.github.io/azure-service-operator](https://azure.github.io/azure-service-operator/)
- Google Config Connector: [cloud.google.com/config-connector/docs/overview](https://cloud.google.com/config-connector/docs/overview)
- Upbound Marketplace (Crossplane providers): [marketplace.upbound.io](https://marketplace.upbound.io)
