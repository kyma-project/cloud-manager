# Pattern 6: Portable Container with Unstructured Payload

**Origin:**
- Kubernetes `StorageClass` `parameters: map[string]string` · [k8s.io/api/storage/v1/types.go](https://github.com/kubernetes/api/blob/master/storage/v1/types.go)
- Kubernetes `ConfigMap` / `Secret` `data: map[string]string`

**What it solves:** When the resource envelope is provider-neutral (create, reference, lifecycle, status) but the payload content — rule sets, policies, scripts — is provider-native and content-rich. An unstructured field hands off payload validation to the provider while keeping the CM resource shape consistent and the common fields strongly typed.

---

## Reference example

**StorageClass — neutral envelope, opaque provider payload:**
```yaml
kind: StorageClass
provisioner: ebs.csi.aws.com
reclaimPolicy: Delete              # neutral — validated by k8s
allowVolumeExpansion: true         # neutral — validated by k8s
parameters:                        # opaque — provisioner validates, not k8s
  type: gp3
  iops: "3000"
  encrypted: "true"
  kms-key-id: "arn:aws:kms:..."
```

The envelope fields (`reclaimPolicy`, `allowVolumeExpansion`) are strongly typed and consistent across all provisioners. The `parameters` map is opaque — each provisioner defines its own keys.

---

## Cloud Manager — future WafPolicy (clean test case)

WAF rule sets are provider-native and content-rich. An AWS `AWSManagedRulesCommonRuleSet` has no direct structural equivalent on GCP Cloud Armor or Azure WAF Policy. Forcing a portable rule schema would require either a lowest-common-denominator set of toggles (losing fidelity) or a large union type (mostly empty for any given provider).

The right shape: strongly-typed portable fields for what is truly shared (target attachment, common toggles), and typed provider sub-structs for rule configuration where the content is provider-native.

**Proposed WafPolicy:**
```yaml
kind: WafPolicy
spec:
  # Portable envelope — strongly typed, validated by CM CRD
  targetRef:
    name: my-gateway              # what to protect (provider-neutral reference)
  rules:
    owaspTop10: true              # common toggle — all providers support this concept
    rateLimit:
      requestsPerMinute: 1000     # common toggle — all providers support this concept

  # Provider-specific payload — structure and content are provider-native
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

The `instance.*` sub-structs are typed (not raw `map[string]string`) — CM can validate structure while the cloud provider validates content. The portable `rules` fields are the intersection that all providers support and that users can set without provider knowledge.

---

## When to use this pattern

Use when:
- The resource envelope (lifecycle, target attachment, status) is provider-neutral
- The payload content (rule sets, policy documents, scripts) is provider-native with no meaningful cross-provider schema
- Abstracting the payload would require users to learn a CM DSL instead of the provider's own documented format
- The payload schema is large, complex, or provider-versioned (changes with provider API updates)

Do not use as a shortcut to avoid designing a proper base+extension schema. The escape hatch is for content that genuinely has no cross-provider representation, not for fields where the concept is shared and only the name differs.
