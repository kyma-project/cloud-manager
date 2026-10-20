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
