# Investigate API Design Principles for Cloud Manager Resource Modeling

> https://github.tools.sap/kyma/backlog/issues/10259 with [local copy](./spike-request.md)

 
# 1) Current model

Detailed analysis of the current CloudManager CRDs with possible separation into portable intent and provider specific resources can be found [here](./01-current-model/overview.md). The remodeling is forcefully fitted to accommodate the request, but the validity is questionable since union of the portable intent resource with provider specific resource produces the originally remodeled shape. The change is cosmetic illusion of portability introducing more CRDs. Not even the benefit of Kyma provided reasonable defaults is possible since provider specific fields carry instance specific fields like unique name or capacity.

## IpRange & GcpSubet ✅

From the UX perspective the provider-specific resources is not the right choice.

No shape differences. GcpSubnet can collapse into IpRange with addition on the `type` field indicating a predefined enumerable set of possible values, without additional configuration.

Carry significant network reconfiguration pre-requisites that are hidden from the user, and their outcomes are the foundation for all other network bound resources (NFS, Redis, WAF, DNS...). Problem is lack of user exposed address space configuration. For some providers their reconciliation is extremely complex, constrained and cardinality sensitive. Brings additional entanglement of unrelated functionality leveraged with all networking functional debt. Deeper grooming required!

[Details](./01-current-model/overview.md#iprange-and-gcpsubnet)

## VpcPeering ❓

Vpc peering provider-specific resources maybe the right choice.

- provider specific shape carries fields with unique constraints and thus can not serve as global Kyma provided best practices shared resource
- risk of a need to define own custom format for remote vpc network for cases when provider doesn't already have a standard format

[Details](./01-current-model/overview.md#vpcpeering)


## NFS ❌

NFS provider-specific resources is NOT the right choice.

- provider specific shape carries the capacity field and can not serve as global Kyma provided best practices shared resource

[Details](./01-current-model/overview.md#nfsvolume)


## Redis ❌

Redis provider-specific resources is NOT the right choice.

- provider specific shape carries the capacity field and can not serve as global Kyma provided best practices shared resource

[Details](./01-current-model/overview.md#redis)

## Open questions

- undefined behavior of feature flags and control of existence and reconciler start on unsupported SKRs  
- unsubstantiated transition from N → N+1 shapes
- doesn't change much the mental model since union of portable intent resource and provider specific config evaluates into a shape very similar to the original

---

# 2) Industry patterns

- [Established infrastructure provisioning tools](./02-industry-patterns/infrastructure.md)
- [General API blueprints](./02-industry-patterns/blueprints/README.md)
- [Progressive API design](./02-industry-patterns/progressive-api-design/README.md)

## CloudOrchestrator Highlight

Similarly to CloudManager - Unifying API access across different cloud providers is not the focus of CloudOrchestrator. They learned that this is not the main concern of many their stakeholders. Either, you have no config options or you have all config options, first is unusable, second is unmanageable. Experiments where done, without much success (from checking with consumers). They are focused on basic building blocks and leave abstractions on top to be built by their customers.


---

# 3) WAF test case

- [WAF test case](./03-waf/README.md)

---

# 4) Principles

- [CloudManager API Principles](./04-api-principles/README.md)
