# Investigate API Design Principles for Cloud Manager Resource Modeling

> https://github.tools.sap/kyma/backlog/issues/10259 with [local copy](./spike-request.md)

# 1) Current model

Detailed analysis of the current CloudManager CRDs with possible separation into portable intent and provider specific resources can be found [here](./01-current-model/overview.md). The remodeling is forcefully fitted to accommodate the request, but the validity is questionable since union of the portable intent resource with provider specific resource produces the originally remodeled shape. The change is cosmetic illusion of portability introducing more CRDs. The only benefit can be found in Kyma provided reasonable defaults, but only some resource (NFS & Redis).

## IpRange & GcpSubet

No shape differences. GcpSubnet can collapse into IpRange with addition on the `type` field indicating a predefined enumerable set of possible values, without additional configuration.

Carry significant network reconfiguration pre-requisites that are hidden from the user, and their outcomes are the foundation for all other network bound resources (NFS, Redis, WAF, DNS...). Problem is lack of user exposed address space configuration. For some providers their reconciliation is extremely complex, constrained and cardinality sensitive. Brings additional entanglement of unrelated functionality leveraged with all networking functional debt. Deeper grooming required!

[Details](./01-current-model/overview.md#iprange-and-gcpsubnet)

## VpcPeering

If risk of custom remote VPC identifier for providers not having standard format is accepted, then a non-empty portable intent resource `VpcPeering` can be defined with separate per provider `VpcPeeringConfig`.

**⚠️ Risk** providers not having a specific standard resource identifier where we must define **own custom format**. Otherwise, the remote VPC identifier moves to the provider specific configuration and portable intent resource remains empty, which doesn't make much sense.

[Details](./01-current-model/overview.md#vpcpeering)


## NFS

Common fields can stay in portable intent resource `NfsVolume`, with rest moved to provider specific `NfsConfig`.

[Details](./01-current-model/overview.md#nfsvolume)


## Redis

Common fields can stay in portable intent resource `Redis`, with rest moved to provider specific `RedisConfig`, `RedisClusterConfig`, `ManagedRedisConfig`.

[Details](./01-current-model/overview.md#redis)

## Open questions

- undefined behavior of feature flags and control of existence and reconciler start on unsupported SKRs  
- unsubstantiated transition from N → N+1 shapes
- doesn't change much the mental model since union of portable intent resource and provider specific config evaluates into a shape very similar to the original

---

# 2) Industry patterns

- [overview](./02-industry-patterns/blueprints/overview.md)
- [Pattern 1: Common base + provider-specific extensions](./02-industry-patterns/blueprints/01-base-with-extensions.md)
- [Pattern 2: Neutral intent resource](./02-industry-patterns/blueprints/02-neutral-intent-resource.md)
- [Pattern 3: Consistent field naming for shared concepts](./02-industry-patterns/blueprints/03-normalize-status.md)
- [Pattern 4: Provider-specific resource](./02-industry-patterns/blueprints/04-provider-specific-resource.md)
- [Pattern 5: Typed provider sub-struct](./02-industry-patterns/blueprints/05-typed-provider-sub-struct.md)
- [Pattern 6: Portable container + unstructured payload](./02-industry-patterns/blueprints/06-unstructured-payload.md)
