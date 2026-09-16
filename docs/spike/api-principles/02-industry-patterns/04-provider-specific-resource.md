# Pattern 4: Provider-Specific Resource

When cloud concepts differ structurally across providers — not just in vocabulary — a shared resource either silently drops fidelity or produces a union type where most fields are irrelevant for any given provider. A dedicated resource per provider is the honest design.

---

## Where This Pattern Comes From

### AWS Controllers for Kubernetes (ACK)
**Repo:** [aws-controllers-k8s/elasticache-controller](https://github.com/aws-controllers-k8s/elasticache-controller)
**Key file:** [`apis/v1alpha1/replication_group.go`](https://github.com/aws-controllers-k8s/elasticache-controller/blob/main/apis/v1alpha1/replication_group.go)

CRD fields mirror the AWS API directly. No abstraction. Maximum completeness, zero portability. Users who know AWS write exactly the fields AWS expects.

```yaml
kind: ReplicationGroup           # AWS-specific kind
spec:
  replicationGroupID: my-redis
  cacheNodeType: cache.r6g.large     # AWS instance type
  numNodeGroups: 3                   # AWS term for shards
  replicasPerNodeGroup: 1
  automaticFailoverEnabled: true
  atRestEncryptionEnabled: true
```

**Lesson for CM:** When the underlying concepts differ structurally (not just in naming), provider-specific resources are correct. Forcing a union type makes `numNodeGroups` meaningful only for AWS, `shardCount` only for GCP — a worse API.

---

### Azure Service Operator (ASO)
**Repo:** [Azure/azure-service-operator](https://github.com/Azure/azure-service-operator)
**Example:** [`api/cache/v1api20230401`](https://github.com/Azure/azure-service-operator/tree/main/v2/api/cache/v1api20230401)

CRD mirrors the Azure ARM API. ARM resource IDs, SKU families, and Azure-specific networking constructs appear directly in spec.

```yaml
kind: RedisCacheEnterprise        # Azure-specific kind
spec:
  location: westeurope
  sku:
    name: EnterpriseFlash_F300    # Azure SKU — no AWS/GCP equivalent
    capacity: 3
  zones: ["1", "2", "3"]
  minimumTlsVersion: "1.2"
```

**Lesson for CM:** Azure VNet peering uses ARM resource IDs and AAD tenant IDs — concepts that do not exist on AWS or GCP. A shared `VpcPeering` with a `provider` field would just hide these in a union, not eliminate them.

---

## How Cloud Manager Uses This Pattern

**VpcPeering — correctly uses this pattern:**

The peering identity model, routing strategy, and remote VPC identification differ structurally across providers. All three resources are correct as-is.

```yaml
kind: AwsVpcPeering
spec:
  remoteVpcId: "vpc-0a1b2c3d"
  remoteRegion: "us-east-1"
  remoteAccountId: "123456789012"        # cross-account IAM identity — no GCP/Azure concept
  remoteRouteTableUpdateStrategy: "AUTO" # AWS route propagation — no equivalent elsewhere
  deleteRemotePeering: true

---
kind: GcpVpcPeering
spec:
  remoteVpc: "my-remote-vpc"
  remoteProject: "my-gcp-project"        # GCP project scope — no AWS/Azure concept
  remotePeeringName: "my-peering"
  importCustomRoutes: false              # GCP routing option — no equivalent elsewhere
  deleteRemotePeering: true

---
kind: AzureVpcPeering
spec:
  remoteVnet: "/subscriptions/.../virtualNetworks/my-vnet"  # ARM resource ID
  remotePeeringName: "my-peering"
  remoteTenant: "00000000-..."           # cross-tenant AAD — no AWS/GCP concept
  useRemoteGateway: false               # Azure gateway transit — no equivalent
  deleteRemotePeering: true
```

`deleteRemotePeering` and `remotePeeringName` (GCP + Azure) are the only shared concepts. The rest is structurally provider-specific. No changes needed.

---

## Applying to CM CRD Families

### VpcPeering — remotePeeringName validation consistency

The resources are correct as-is. One minor improvement: `remotePeeringName` exists on both GCP and Azure with different validation constraints (GCP: 1–63 chars, lowercase alphanumeric+hyphens; Azure: 1–80 chars, word characters and hyphens). The field name is already consistent — but the CEL validation rules could reference this shared origin in their error messages to make the difference explicit to users.

No structural change needed.

## When to Use This Pattern vs. Pattern 1

| Use Pattern 1 (base + extensions) | Use Pattern 4 (provider-specific resource) |
|-----------------------------------|---------------------------------------------|
| User decision is the same across providers | User decision is inherently provider-specific |
| Concepts are the same, only names differ | Concepts themselves differ structurally |
| e.g. capacity, version, auth, config map | e.g. peering identity model, routing strategy, network attachment type |
