# Pattern 4: Provider-Specific Resource

**Origin:**
- AWS Controllers for Kubernetes (ACK) · [aws-controllers-k8s/community](https://github.com/aws-controllers-k8s/community)
- Azure Service Operator (ASO) · [Azure/azure-service-operator](https://github.com/Azure/azure-service-operator)

**What it solves:** When cloud concepts differ structurally across providers — not just in vocabulary — forcing a shared resource either silently drops fidelity or produces a union type where most fields are irrelevant for any given provider. A dedicated resource per provider is the honest design.

---

## Reference examples

**ACK — CRD mirrors AWS API directly:**
```yaml
kind: ElasticacheReplicationGroup    # AWS-specific kind, no abstraction
spec:
  replicationGroupID: my-redis
  cacheNodeType: cache.r6g.large     # AWS instance type
  numNodeGroups: 3                   # AWS term for shards
  replicasPerNodeGroup: 1
  automaticFailoverEnabled: true
```

**ASO — Azure resource modeled directly:**
```yaml
kind: RedisCacheEnterprise            # Azure-specific kind
spec:
  location: westeurope
  sku:
    name: EnterpriseFlash_F300
    capacity: 3
  zones: ["1", "2", "3"]
```

---

## Cloud Manager — VpcPeering (correctly uses this pattern)

The peering identity model, routing strategy, and remote VPC identification differ structurally across providers. A shared `VpcPeering` resource would produce mostly-empty fields on every provider.

```yaml
kind: AwsVpcPeering
spec:
  remoteVpcId: "vpc-0a1b2c3d"
  remoteRegion: "us-east-1"
  remoteAccountId: "123456789012"          # cross-account identity — no GCP/Azure concept
  remoteRouteTableUpdateStrategy: "AUTO"   # AWS route propagation — no equivalent elsewhere
  deleteRemotePeering: true

---
kind: GcpVpcPeering
spec:
  remoteVpc: "my-remote-vpc"
  remoteProject: "my-gcp-project"          # GCP project scope — no AWS/Azure concept
  importCustomRoutes: false                 # GCP-specific routing option
  deleteRemotePeering: true

---
kind: AzureVpcPeering
spec:
  remoteVnet: "/subscriptions/.../virtualNetworks/my-vnet"  # ARM resource ID
  remoteTenant: "00000000-..."             # cross-tenant AAD — no AWS/GCP concept
  useRemoteGateway: false                  # Azure gateway transit — no equivalent
  deleteRemotePeering: true
```

**No change needed.** `deleteRemotePeering` is the only field shared across all three and is already present on all. A hypothetical unified `VpcPeering` with a `provider` discriminator would make `remoteAccountId`, `remoteProject`, and `remoteTenant` all present in the schema with two out of three always empty — a strictly worse API.

---

## When to use this pattern vs. Pattern 1 (base+extensions)

| Use Pattern 1 (base+extensions) | Use Pattern 4 (provider-specific resource) |
|----------------------------------|---------------------------------------------|
| User decision is the same across providers | User decision is inherently provider-specific |
| Field names differ but concepts are identical | Concepts themselves differ structurally |
| e.g. capacity, version, auth, config map | e.g. peering identity model, routing strategy, network attachment type |
