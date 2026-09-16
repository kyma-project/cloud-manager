# Pattern 6: Portable Container with Unstructured Payload

When the resource envelope is provider-neutral (lifecycle, status, common references) but the payload content — rule sets, policies, scripts — is provider-native, content-rich, and has no meaningful cross-provider schema. The envelope is strongly typed; the payload is passed through to the provider for validation.

This differs from Pattern 5 (typed provider sub-struct): in Pattern 5 the provider extension is a *typed* struct with CRD schema validation. In Pattern 6 the payload is *unstructured* — CM validates the envelope, the cloud provider validates the content.

---

## Where This Pattern Comes From

### Kubernetes StorageClass — parameters
**Repo:** [kubernetes/api](https://github.com/kubernetes/api)
**Key file:** [`storage/v1/types.go`](https://github.com/kubernetes/api/blob/master/storage/v1/types.go)

The `StorageClass` envelope fields (`reclaimPolicy`, `allowVolumeExpansion`, `volumeBindingMode`) are strongly typed and consistent across all provisioners. The `parameters` field is a `map[string]string` — opaque to Kubernetes, validated entirely by the provisioner.

```yaml
kind: StorageClass
provisioner: ebs.csi.aws.com
reclaimPolicy: Delete              # strongly typed — same field on all provisioners
allowVolumeExpansion: true         # strongly typed — same field on all provisioners
parameters:                        # opaque map[string]string — provisioner validates
  type: gp3
  iops: "3000"
  encrypted: "true"
  kms-key-id: "arn:aws:kms:..."

---
kind: StorageClass
provisioner: disk.csi.azure.com
reclaimPolicy: Delete              # same field, same semantics
parameters:                        # different keys — Azure-specific, no CRD schema
  skuName: Premium_LRS
  cachingMode: ReadOnly
```

**Lesson for CM:** When provider-specific content is large, versioned independently, or has no cross-provider schema, pass it through as `map[string]string` with the provider validating keys. Do not invent a CM abstraction that users must learn instead of the provider's own documented format.

---

### Kubernetes CRD — x-kubernetes-preserve-unknown-fields
**Repo:** [kubernetes/kubernetes](https://github.com/kubernetes/kubernetes)
**Key file:** [`staging/src/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1/types_jsonschema.go`](https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1/types_jsonschema.go)
**Kubebuilder marker docs:** [`book.kubebuilder.io — markers/crd-validation`](https://book.kubebuilder.io/reference/markers/crd-validation)

For structured but provider-versioned payloads that cannot be captured in a static CRD schema (e.g. policy documents, rule sets with nested objects), `x-kubernetes-preserve-unknown-fields: true` allows arbitrary structure at a given path while keeping the rest of the spec typed.

```go
// In CRD type definition:
// +kubebuilder:pruning:PreserveUnknownFields
// +kubebuilder:validation:Schemaless
RuleConfig *runtime.RawExtension `json:"ruleConfig,omitempty"`
```

```yaml
kind: WafPolicy
spec:
  targetRef:
    name: my-gateway
  ruleConfig:                      # schemaless — any structure, provider validates
    managed:
      - vendorName: AWS
        name: AWSManagedRulesCommonRuleSet
    custom:
      - name: block-sqli
        action: Block
        statement:
          sqliMatchStatement:
            fieldToMatch:
              body: {}
```

**Lesson for CM:** `runtime.RawExtension` (schemaless) is the right escape hatch when the payload is structured but provider-versioned and cannot be represented as a static CRD schema without frequent breaking changes.

---

## How Cloud Manager Uses This Pattern

CM does not currently use this pattern explicitly. The closest analog is the Redis `parameters: map[string]string` on AWS and GCP — a freeform config map where the controller passes keys through to the provider API without validating individual key names at the CRD level.

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

WAF rule sets are provider-native, content-rich, and provider-versioned. AWS managed rule group schemas, GCP preconfigured rule IDs, and Azure managed rule set versions change independently of CM releases. Encoding them as typed CRD fields would require CM schema updates every time a provider adds or renames a rule.

The right shape: strongly typed envelope for what is universal (`targetRef`, common toggles); unstructured `ruleConfig` per provider for rule content.

```yaml
kind: WafPolicy
spec:
  # Strongly typed envelope — validated by CM CRD
  # Note: targetRef (what to protect) is omitted here — the attachment model
  # differs structurally per provider and requires a separate design decision.
  rules:
    owaspTop10: true              # toggle — all providers support this concept
    rateLimit:
      requestsPerMinute: 1000

  # Unstructured per-provider rule payload — provider validates content
  # +kubebuilder:pruning:PreserveUnknownFields
  providerConfig:
    aws:                          # schemaless — AWS WAF JSON passed through
      managedRuleGroups:
        - vendorName: AWS
          name: AWSManagedRulesCommonRuleSet
    gcp:                          # schemaless — Cloud Armor JSON passed through
      preconfiguredWafRules:
        - id: "sqli-v33-stable"
    azure:                        # schemaless — Azure WAF JSON passed through
      managedRuleSets:
        - ruleSetType: OWASP
          ruleSetVersion: "3.2"
```

---

## When to Use This Pattern

Use when:
- The payload is provider-native with no meaningful cross-provider schema
- The payload schema is versioned independently of CM (provider adds/renames fields without CM involvement)
- Abstracting the payload would require users to learn a CM DSL instead of the provider's own documented format

Use Pattern 5 (typed sub-struct) instead when:
- The provider extension fields are stable and enumerable
- Full CRD schema validation per provider is needed
- The fields belong in spec as user decisions (not opaque pass-through content)
