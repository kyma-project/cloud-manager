# Pattern 5: Typed Provider Sub-Struct

When the resource envelope is shared (lifecycle, status contract, common references) but some configuration is genuinely provider-specific, a typed `spec.instance.<provider>` sub-struct keeps the resource tidy — users only see fields relevant to their provider — while CRD-level `MaxProperties=1` enforces exactly one sub-struct is set.

This pattern describes the *structural mechanism* for housing provider-specific fields. Pattern 1 describes *which fields* belong in the base vs. the extension. Pattern 3 describes *naming consistency* within extensions. These are complementary, not overlapping.

---

## Where This Pattern Comes From

### Crossplane Provider — forProvider typed sub-struct
**Repo:** [crossplane-contrib/provider-aws](https://github.com/crossplane-contrib/provider-aws)
**Key file:** [`apis/cache/v1beta1/replicationgroup_types.go`](https://github.com/crossplane-contrib/provider-aws/blob/master/apis/cache/v1beta1/replicationgroup_types.go)

Each Crossplane provider's Managed Resource has a `spec.forProvider` sub-struct that is typed to the provider's API. The sub-struct has full CRD schema validation — not a `map[string]interface{}`. The controller reads exactly these fields and calls the provider API.

```yaml
# AWS Managed Resource — typed sub-struct, not unstructured
kind: ReplicationGroup
spec:
  forProvider:                       # typed struct — AwsReplicationGroupParameters
    region: us-east-1
    cacheNodeType: cache.t3.small
    engineVersion: "7.0"
    numNodeGroups: 3
    replicasPerNodeGroup: 1
    atRestEncryptionEnabled: true
  providerConfigRef:
    name: aws-provider
```

**Lesson for CM:** `spec.forProvider` in Crossplane is exactly `spec.instance.aws` in CM's KCP layer. The pattern is already applied correctly in CM at the KCP layer. The key property: it is *typed*, not a free-form map, giving full validation per provider.

---

### Cluster API — Infrastructure Provider Contract
**Repo:** [kubernetes-sigs/cluster-api](https://github.com/kubernetes-sigs/cluster-api)
**Key file:** [`api/v1beta1/cluster_types.go`](https://github.com/kubernetes-sigs/cluster-api/blob/main/api/v1beta1/cluster_types.go)

The provider object (`AWSCluster`, `GCPCluster`) is a fully typed CRD — not a `map[string]interface{}`. The base `Cluster` references it by type name. The contract between base and provider is enforced through a defined set of status fields the provider must populate.

```yaml
kind: AWSCluster                   # typed struct — AWSClusterSpec
spec:
  region: "us-east-1"              # AWS-specific field
  network:
    vpc:
      cidrBlock: "10.0.0.0/16"

status:
  ready: true                      # contract field — CAPI checks this before proceeding
  failureDomains:
    us-east-1a:
      controlPlane: true
```

**Lesson for CM:** The status contract is what makes the base resource portable. CM's KCP layer enforces a similar contract — the provider action pipeline must set `Ready` condition before the SKR reconciler proceeds.

---

## How Cloud Manager Uses This Pattern

**KCP RedisInstance — already correct:**

```yaml
kind: RedisInstance                # cloud-control API group
spec:
  scope:
    name: my-scope                 # base — same on all providers
  ipRange:
    name: my-iprange               # base — same on all providers
  instance:                        # exactly one sub-struct (MaxProperties=1)
    gcp:
      memorySizeGb: 6
      tier: "STANDARD_HA"
      redisVersion: "REDIS_7_0"
    # aws: / azure: / alicloud: mutually exclusive — enforced by CEL rule

status:
  memorySizeGb: 6                  # contract output — same field on all providers
  primaryEndpoint: "10.0.0.5:6379" # contract output
```

No change needed at the KCP layer. The sub-struct mechanism is correct.

---

## Applying to CM CRD Families

### KCP NfsInstance — provider sub-struct consistency

KCP `RedisInstance` and `RedisCluster` use typed provider sub-structs correctly (`spec.instance.gcp`, `spec.instance.aws`, etc.) with `MaxProperties=1`. KCP `NfsInstance` should follow the same structure. This makes the internal contract explicit and applies the same pattern uniformly across all KCP resource types.

```yaml
# KCP NfsInstance — should mirror RedisInstance's structure
kind: NfsInstance
spec:
  scope:
    name: my-scope
  ipRange:
    name: my-iprange
  instance:                        # typed discriminated union, MaxProperties=1
    gcp:
      tier: "BASIC_HDD"
      fileShareName: "vol1"
      capacityGb: 1024             # KCP-internal: integer GiB, matches Filestore API
                                   # SKR layer converts from k8s Quantity (Pattern 2)
    # aws: / azure: / alicloud: / openStack: mutually exclusive
```

This is a KCP-internal change. SKR users always write `capacity: "1Ti"` (k8s Quantity) on the SKR resource; the controller converts to `capacityGb` when creating the KCP object.
