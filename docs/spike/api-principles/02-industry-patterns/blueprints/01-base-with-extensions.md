# Pattern: Common Base with Provider-Specific Extensions

This is the foundational pattern for Cloud Manager CRD API design. It appears across multiple production Kubernetes projects under different names, but the core idea is always the same: **shared user intent lives in a common base; provider-specific configuration lives in typed extensions.**

---

## Where This Pattern Comes From

### Cluster API (CAPI)
**Repo:** [kubernetes-sigs/cluster-api](https://github.com/kubernetes-sigs/cluster-api)
**Key file:** [`api/v1beta1/cluster_types.go`](https://github.com/kubernetes-sigs/cluster-api/blob/main/api/core/v1beta1/cluster_types.go)

Base `Cluster` holds provider-agnostic fields. Provider-specific resources (`AWSCluster`, `GCPCluster`, `AzureCluster`) live in separate repos and are referenced via `spec.infrastructureRef`.

```yaml
# What the user writes — base (provider-agnostic)
kind: Cluster
spec:
  clusterNetwork:
    pods:
      cidrBlocks: ["192.168.0.0/16"]
  infrastructureRef:          # typed pointer to provider extension
    apiVersion: infrastructure.cluster.x-k8s.io/v1beta2
    kind: AWSCluster           # or GCPCluster, AzureCluster
    name: my-cluster

---
# Provider extension — user also creates this (CAPI's main cost)
kind: AWSCluster
spec:
  region: "us-east-1"
  network:
    vpc:
      cidrBlock: "10.0.0.0/16"
```

**Lesson for CM:** The base+ref split is correct. CM improves on CAPI by *generating* the provider object (KCP) from the user object (SKR) — the user only creates one resource, not two.

---

### Crossplane Composite Resources
**Repo:** [crossplane/crossplane](https://github.com/crossplane/crossplane)
**Key file:** [`apis/apiextensions/v1/composition_types.go`](https://github.com/crossplane/crossplane/blob/master/apis/apiextensions/v1/composition_types.go)

A `CompositeResourceDefinition` (XRD) defines the provider-agnostic schema (the Claim). A `Composition` maps Claim fields to provider-specific Managed Resource fields via patches and transforms — including value mapping (e.g. `tier: standard` → `cacheNodeType: cache.t3.small`).

```yaml
# User writes (Claim — neutral)
kind: RedisInstanceClaim
spec:
  tier: standard          # neutral concept
  engineVersion: "7.0"   # neutral concept
  memorySizeGb: 5

# Composition maps to AWS (internal, not user-facing)
# tier=standard → cacheNodeType=cache.t3.small
# engineVersion="7.0" → engineVersion="7.0" (passthrough for AWS)
# engineVersion="7.0" → redisVersion="REDIS_7_0" (transform for GCP)
```

**Lesson for CM:** The Composition principle — *user writes the concept, controller maps to provider vocabulary* — is exactly what CM should do for fields like `engineVersion` and `replicasPerShard`. The mapping is compiled Go in CM rather than YAML patches, which is strictly better (type-safe, testable).

---

### Kubernetes Gateway API
**Repo:** [kubernetes-sigs/gateway-api](https://github.com/kubernetes-sigs/gateway-api)
**Key file:** [`apis/v1/gatewayclass_types.go`](https://github.com/kubernetes-sigs/gateway-api/blob/main/apis/v1/gatewayclass_types.go)

`GatewayClass` carries the common structure. `spec.parametersRef` is the typed escape hatch pointing to a provider-specific config object. Standard fields (listeners, routes) are on the base; anything provider-specific is in the referenced object.

```yaml
kind: GatewayClass
spec:
  controllerName: "example.com/nginx-controller"
  parametersRef:              # optional typed pointer to provider extension
    group: "gateway.example.com"
    kind: "NginxGatewayConfig"
    name: "nginx-params"
```

**Lesson for CM:** The `parametersRef` pattern is the right model for cases where most users don't need provider-specific tuning, but expert users must be able to reach it. The reference is typed and optional — not a free-form annotation.
