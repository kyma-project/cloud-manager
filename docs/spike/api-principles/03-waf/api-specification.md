# WAF API Specification

## Overview

This document defines Cloud Manager's WAF API. The stable foundation is `WafPolicy` — a single SKR resource that provisions a WAF policy on any supported cloud provider. The policy content is supplied inline as `spec.data`, a free-form JSON object that the controller passes through to the provider without schema enforcement.

A higher-level `WafConfiguration` resource (intent-based, generates `WafPolicy`) was explored but is **deferred**. The fragment-merge model it depends on does not have a clean cross-provider solution. See [waf-configuration-design.md](waf-configuration-design.md) for the full design and the reasons it stalls.

The design follows the [progressive API design model](../02-industry-patterns/progressive-api-design/progressive-api-design.md): `WafPolicy` is the Layer 3 unstructured payload (P6) resource — provider-native JSON passthrough, controller-validated only, no typed sub-struct. See [design-rationale.md](design-rationale.md) for the full layer mapping.

---

## WafPolicy — Provider Policy Passthrough

### Design

- Carries the provider-specific WAF policy JSON **inline** in `spec.data`
- Provisioned in cloud as soon as it is created and successfully reconciled — independently of whether `AppLoadBalancer` references it
- `ownerReference` is nil when created directly by the user; set to the generating resource if created programmatically

### WafPolicy API

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: WafPolicy
metadata:
  name: my-policy
  namespace: my-namespace
spec:
  data: |
    {
      "Name": "my-web-acl",
      "DefaultAction": { "Allow": {} },
      "Rules": [
        {
          "Name": "AWSManagedRulesCommonRuleSet",
          "Priority": 1,
          "OverrideAction": { "None": {} },
          "Statement": {
            "ManagedRuleGroupStatement": {
              "VendorName": "AWS",
              "Name": "AWSManagedRulesCommonRuleSet"
            }
          },
          "VisibilityConfig": {
            "SampledRequestsEnabled": true,
            "CloudWatchMetricsEnabled": true,
            "MetricName": "AWSManagedRulesCommonRuleSet"
          }
        }
      ],
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "my-web-acl"
      }
    }

status:
  providerId: "arn:aws:wafv2:us-east-1:123456789012:regional/webacl/..."
  conditions:
    - type: Ready
      status: True
      reason: Available
      observedGeneration: 1
```

### Validation

`WafPolicy.spec.data` is a free-form JSON object (`runtime.RawExtension`). The API server does not validate its contents at admission time. All structural validation is controller-side — the controller parses `spec.data`, checks provider-specific constraints, and reports violations via `Ready=False` status conditions with descriptive messages. Users will not receive an immediate rejection on `kubectl apply`; they will see the condition on the resource status.

### Status

`WafPolicy` does not expose a top-level `state: Ready | Creating | Error` field. The overall state is derived from the `conditions` array — specifically the `Ready` condition's `reason` field. This is intentional: a redundant top-level state field would need to be kept in sync with the conditions and adds no information not already present in `conditions.reason`.

This is a deliberate departure from the standard status contract in the [progressive API design model](../02-industry-patterns/progressive-api-design/progressive-api-design.md), which includes `status.state` as a normalised output field. For `WafPolicy`, `conditions` carries the full state signal; a parallel `status.state` field would duplicate it without adding observability value.

### AppLoadBalancer Integration

`AppLoadBalancer.spec.policyRef` always resolves to a `WafPolicy`. Reference it by name:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: AppLoadBalancer
spec:
  policyRef:
    name: my-policy
```

---

## Architecture Flow

```
[SKR cluster]

WafPolicy (spec.data: complete provider-native WAF policy JSON)
    ↓
AppLoadBalancer (spec.policyRef: WafPolicy)
    ↓
Cloud WAF resources (AWS WAFv2 / Azure Front Door WAF / GCP Cloud Armor)
```

`WafPolicy` is the single contract `AppLoadBalancer` references — always. There is no higher-level resource currently specified.

**Progressive model coverage:** The [progressive API design model](../02-industry-patterns/progressive-api-design/progressive-api-design.md) expects a Layer 1 zero-config intent resource as the primary user entry point. For WAF, that resource is `WafConfiguration` — explored but deferred. As a result, the current API starts at **Layer 3 unstructured payload (P6)**: every user must supply a complete provider-native WAF policy JSON in `spec.data`. There is no zero-config path yet. This is a deliberate gap, not an oversight — see [waf-configuration-design.md](waf-configuration-design.md) for why Layer 1 is not yet implementable.

---

## Design Decisions

See [design-rationale.md](design-rationale.md) for the reasoning behind `spec.data` (inline) over `spec.configMapRef` and over typed provider sub-structs (`spec.aws/gcp/azure`).

### WafConfiguration is not a cloud Kind

`WafPolicy` fulfils the "one Kind, any cloud" goal: it is the single Kind that provisions a WAF resource on AWS, Azure, or GCP. There are no `AwsWafPolicy`, `GcpWafPolicy`, or `AzureWafPolicy` variants.

The deferred `WafConfiguration` is not a cloud-provisioning Kind — it would never call a cloud provider API. Cloud provisioning always flows through `WafPolicy`. Users who need a higher-level abstraction above `WafPolicy` will do so via a future resolved design; the single-CRD principle is not affected.

---

## Use Cases

| Use case | Resources user creates | `AppLoadBalancer.spec.policyRef` | Controller behaviour |
|---|---|---|---|
| **Full control** | `WafPolicy` with `spec.data` (inline policy JSON) | `name: my-policy` | No generation — user owns everything |
| **Intent-based** | `WafConfiguration` (deferred) | — | See [waf-configuration-design.md](waf-configuration-design.md) |
