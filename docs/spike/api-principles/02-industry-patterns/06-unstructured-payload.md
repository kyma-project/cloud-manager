# Pattern 6: Portable Container with Unstructured Payload

When the resource envelope is provider-neutral (lifecycle, status, common references) but the payload content — rule sets, policies, scripts — is provider-native and content-rich with no meaningful cross-provider schema. The envelope is strongly typed; the payload is handed to the provider for validation.

---

## Where This Pattern Comes From

### Kubernetes StorageClass — parameters
**Repo:** [kubernetes/api](https://github.com/kubernetes/api)
**Key file:** [`storage/v1/types.go`](https://github.com/kubernetes/api/blob/master/storage/v1/types.go)

The `StorageClass` envelope fields (`reclaimPolicy`, `allowVolumeExpansion`, `volumeBindingMode`) are strongly typed and consistent across all provisioners. The `parameters` field is a `map[string]string` — opaque to Kubernetes, validated by the provisioner.

```yaml
kind: StorageClass
provisioner: ebs.csi.aws.com
reclaimPolicy: Delete              # strongly typed — same field on all provisioners
allowVolumeExpansion: true         # strongly typed — same field on all provisioners
parameters:                        # opaque map — provisioner validates, not k8s
  type: gp3
  iops: "3000"
  encrypted: "true"
  kms-key-id: "arn:aws:kms:..."

---
kind: StorageClass
provisioner: disk.csi.azure.com
reclaimPolicy: Delete              # same field, same semantics
parameters:                        # different keys — Azure-specific
  skuName: Premium_LRS
  cachingMode: ReadOnly
```

**Lesson for CM:** Strongly type what is universal. Use an opaque or typed escape hatch for what the provider owns. Do not invent a CM abstraction for content that already has a well-documented provider-native format.

---

### Kubernetes ConfigMap / Secret
**Repo:** [kubernetes/api](https://github.com/kubernetes/api)
**Key file:** [`core/v1/types.go`](https://github.com/kubernetes/api/blob/master/core/v1/types.go#L5671)

`ConfigMap` and `Secret` have a neutral envelope (name, namespace, labels) and an opaque `data: map[string]string` payload. The content is arbitrary — the consumer validates it, not Kubernetes.

```yaml
kind: ConfigMap
metadata:
  name: my-config               # neutral envelope
data:                           # opaque — any key-value content
  redis.conf: |
    maxmemory-policy volatile-lru
    activedefrag yes
```

**Lesson for CM:** When the payload is consumer-specific and content-rich (e.g. WAF rule sets, firewall policies), an opaque `data` field with consumer-side validation is the right escape valve over forcing a false abstraction.

---

## How Cloud Manager Uses This Pattern

CM does not currently use this pattern explicitly. The closest analog is the Redis `parameters: map[string]string` field on AWS and GCP — a freeform config map where the controller passes the keys through to the provider API without validating individual key names.

```yaml
kind: AwsRedisInstance
spec:
  parameters:                   # freeform — ElastiCache validates keys, not CM
    maxmemory-policy: volatile-lru
    activedefrag: "yes"
    lazyfree-lazy-eviction: "yes"
```

---

## Applying to CM CRD Families

### Future WafPolicy — rule configuration

WAF rule sets are provider-native and content-rich. An AWS `AWSManagedRulesCommonRuleSet` has no direct structural equivalent on GCP Cloud Armor or Azure WAF Policy. Forcing a portable rule schema would require a lowest-common-denominator set of toggles (losing fidelity) or a large union (mostly empty for any given provider).

The right shape: strongly typed portable fields for what is truly shared; typed provider sub-structs for rule configuration where the content is provider-native.

```yaml
kind: WafPolicy
spec:
  # Portable envelope — strongly typed, validated by CM CRD
  targetRef:
    name: my-gateway              # what to protect — provider-neutral
  rules:
    owaspTop10: true              # toggle — all providers support this concept
    rateLimit:
      requestsPerMinute: 1000     # toggle — all providers support this concept

  # Provider-specific payload — structure is provider-native
  instance:
    aws:
      managedRuleGroups:
        - vendorName: AWS
          name: AWSManagedRulesCommonRuleSet
          excludedRules: []
    gcp:
      preconfiguredWafRules:
        - id: "sqli-v33-stable"
        - id: "xss-v33-stable"
    azure:
      managedRuleSets:
        - ruleSetType: OWASP
          ruleSetVersion: "3.2"
          ruleGroupOverrides: []
    alicloud:
      defenseTemplateId: "waf-template-001"
```

The `instance.*` sub-structs are typed — CM validates structure, the cloud provider validates content. The portable `rules` fields are the intersection that all providers support without provider knowledge.

---

## When to Use This Pattern

Use when:
- The resource lifecycle and status contract are provider-neutral
- The payload (rule sets, policy documents, scripts) is provider-native with no meaningful cross-provider representation
- Abstracting the payload would require users to learn a CM DSL instead of the provider's own documented format
- The payload schema is large, complex, or versioned independently of CM

Do not use as a shortcut to avoid designing a proper base+extension schema. This pattern is for content that genuinely has no cross-provider representation — not for fields where the concept is shared and only the name differs (see Pattern 1).
