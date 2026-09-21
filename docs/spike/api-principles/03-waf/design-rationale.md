# WAF API Design Discussion - Final Architecture

## The Core Insight

Cloud Manager should be **honest about provider differences** while **providing value through curation and lifecycle management**.

This document captures the architectural discussion that led to the two-level WAF API design.

---

## The Tension

**Terraform/Crossplane/Pulumi** → Provider-specific is the norm. They don't pretend AWS and Azure work the same way.

**Initial Cloud Manager exploration** → Trying to force "one API to rule them all" when the underlying providers are fundamentally different.

---

## The Solution: Two-Level Architecture

```
WafConfiguration (simple, portable intent + presets)
    ↓ SKR controller translates
WafPolicy (provider-specific JSON passthrough)
    ↓ KCP remote reconciler provisions to SKR's cloud account
Provider-specific cloud resources (AWS WAFv2, Azure Front Door WAF, GCP Cloud Armor)
```

This gives you:

1. **Simple for simple cases**: "I just want basic OWASP protection" → reference predefined WafPolicy (Phase 1)
2. **Base policy customization**: Use WafConfiguration with `basePolicyRef` + `ruleOverrides` + `customRules` (Phase 2)
3. **Portable and expressive**: Write WafConfiguration with portable `customRules` for path-based protection
4. **Honest about provider differences**: WafPolicy is explicitly provider-specific (`spec.data` field with raw JSON)
5. **Clear escape hatch**: Advanced users write WafPolicy directly

---

## Key Pattern

Looking at the requirements:
- `AppLoadBalancer` is **intent-based** (frontend/backend/health check)
- `WafConfiguration` is **portable configuration** with managed rule groups, classifications, conditions
- `WafPolicy` is **provider-specific** with unstructured JSON passthrough
- Resources are **composed** via reference, not inheritance

This is closer to **Kubernetes Gateway API** pattern:
- Gateway (portable intent)
- HTTPRoute (portable routing)
- GatewayClass (provider-specific implementation details)

---

## Resource Definitions

### WafConfiguration (Portable Intent)

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafConfiguration
metadata:
  name: my-config
spec:
  # Optional: Start from a base policy (pre-deployed or custom)
  basePolicyRef:
    name: owasp-moderate  # Can reference owasp-basic, owasp-moderate, owasp-strict, or custom policy
  
  # Simple rule overrides (unconditional)
  ruleOverrides:
    - managedRuleGroup: CoreRuleSet
      ruleId: "942100"
      action: count
      reason: "Testing for false positives"
  
  # Custom rules with conditions (universal support)
  customRules:
    - name: health-check-bypass
      priority: 10
      action: allow
      conditions:
        path:
          exact: "/health"
      reason: "Health check bypass"
    
    - name: block-debug-header-external
      priority: 100
      action: block
      conditions:
        path:
          prefix: "/admin"
        sourceIP:
          cidr: "10.0.0.0/8"
          negate: true
        header:
          name: "X-Debug"
          exists: true
      reason: "Block debug header on admin from external IPs"
  
  # Future portable fields:
  # managedRuleGroups: [...]  # When portable mapping validated
  # sizeLimits: {...}

status:
  # Controller generates WafPolicy
  generatedPolicy:
    apiVersion: cloud-resources.kyma-project.io/v1beta1
    kind: WafPolicy
    name: my-config-wafpolicy
  conditions:
    - type: Ready
      status: True
      reason: PolicyGenerated
      observedGeneration: 1
```

### WafPolicy (Provider-Specific Passthrough)

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: WafPolicy
metadata:
  name: my-advanced-policy
spec:
  # ONLY provider-specific JSON passthrough
  # No portable fields, no translation
  data: |
    {
      "DefaultAction": {"Allow": {}},
      "Rules": [
        {
          "Name": "CustomAdvancedRule",
          "Priority": 1,
          "Statement": {
            "AndStatement": {
              "Statements": [...]
            }
          },
          "Action": {"Block": {}}
        }
      ],
      "VisibilityConfig": {...}
    }

status:
  providerId: "arn:aws:wafv2:us-east-1:123456789012:regional/webacl/..."
  conditions:
    - type: Ready
      status: True
      reason: Available
      observedGeneration: 1
```

---

## AppLoadBalancer Integration

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: AppLoadBalancer
metadata:
  name: my-app
spec:
  backend:
    kind: Service
    name: my-service
  
  # Can reference either type
  policy:
    kind: WafConfiguration  # ← Most users (portable)
    name: my-config
---
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: AppLoadBalancer
metadata:
  name: my-advanced-app
spec:
  backend:
    kind: Service
    name: my-service
  
  policy:
    kind: WafPolicy  # ← Expert users (provider-specific)
    name: my-advanced-policy
```

---

## The Three Usage Patterns

### Pattern 1: Use WafPolicy as-is (Phase 1 - Simplest)

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: AppLoadBalancer
spec:
  policy:
    kind: WafPolicy
    name: owasp-moderate  # Pre-deployed WafPolicy
```

### Pattern 2: WafConfiguration with Base Policy + Overrides (Phase 2 - Most Common) ⭐

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafConfiguration
metadata:
  name: my-config
spec:
  basePolicyRef:
    name: owasp-moderate  # Reference pre-deployed WafPolicy
  
  # Add rule overrides (unconditional)
  ruleOverrides:
    - managedRuleGroup: CoreRuleSet
      ruleId: "942100"
      action: count
  
  # Add custom rules with conditions
  customRules:
    - name: health-check-bypass
      priority: 10
      action: allow
      conditions:
        path:
          exact: "/health"
---
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: AppLoadBalancer
spec:
  policy:
    kind: WafConfiguration
    name: my-config
```

### Pattern 3: Custom WafPolicy (Advanced - Full Control)

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: WafPolicy
metadata:
  name: my-advanced
spec:
  data: |
    {
      # Full provider-specific JSON
      # Complete control, no translation
    }
---
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: AppLoadBalancer
spec:
  policy:
    kind: WafPolicy
    name: my-advanced
```

---

## What We Ship

### 1. Pre-deployed WafPolicy Presets

```yaml
# Cloud Manager ships these WafPolicy resources by default
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: WafPolicy
metadata:
  name: owasp-basic
  labels:
    cloud-resources.kyma-project.io/preset: "true"
  finalizers:
    - cloud-resources.kyma-project.io/preset-protection
spec:
  data: |
    {
      # Provider-specific JSON for OWASP basic protection
      # AWS gets AWS-specific, Azure gets Azure-specific, GCP gets GCP-specific
    }
```

**Presets shipped:**
- `owasp-basic` - Basic OWASP protection (block mode)
- `owasp-moderate` - Moderate OWASP protection (block mode)
- `owasp-strict` - Strict OWASP protection (block mode)
- `owasp-detection` - OWASP detection only (count mode)

These exist in the cluster but:
- **Not provisioned in cloud** until referenced by `AppLoadBalancer` (Phase 1) or `WafConfiguration` (Phase 2)
- Protected from deletion (finalizer)
- Cannot be modified (validation webhook)

### 2. WafConfiguration Controller (Phase 2)

The controller:
1. Reads `WafConfiguration` spec
2. Detects provider (from cluster context)
3. Loads base WafPolicy if `spec.basePolicyRef` is set
4. Parses base WafPolicy `spec.data` (provider-specific JSON)
5. Applies `ruleOverrides` (unconditional changes to managed rules)
6. Translates `customRules` to provider-specific format:
   - AWS: Custom rules with `Action` and `AndStatement`
   - Azure: `customRules` array with `matchConditions`
   - GCP: CEL expressions in `rules` array
7. Creates/updates generated `WafPolicy` with `spec.data` containing merged JSON
8. Sets owner reference so lifecycle is tied

### 3. KCP Remote Reconcilers

KCP reconcilers watch `WafPolicy` resources in SKR clusters:
- Extract `spec.data` (provider-specific JSON)
- Validate against provider API
- Provision cloud WAF resources in SKR's cloud account
- Update `status.providerId` and conditions

---

## Translation Examples

### WafConfiguration → WafPolicy (AWS Example)

**Input (WafConfiguration):**
```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafConfiguration
metadata:
  name: my-config
spec:
  basePolicyRef:
    name: owasp-moderate
  customRules:
    - name: health-check-bypass
      priority: 10
      action: allow
      conditions:
        path:
          exact: "/health"
```

**Output (Generated WafPolicy):**
```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: WafPolicy
metadata:
  name: my-config-wafpolicy
  ownerReferences:
    - apiVersion: cloud-resources.kyma-project.io/v1alpha1
      kind: WafConfiguration
      name: my-config
spec:
  data: |
    {
      "Name": "my-config",
      "DefaultAction": {"Allow": {}},
      "Rules": [
        {
          "Name": "HealthCheckBypass",
          "Priority": 10,
          "Action": {"Allow": {}},
          "Statement": {
            "ByteMatchStatement": {
              "SearchString": "/health",
              "FieldToMatch": {"UriPath": {}},
              "PositionalConstraint": "EXACTLY"
            }
          },
          "VisibilityConfig": {
            "SampledRequestsEnabled": false,
            "CloudWatchMetricsEnabled": true,
            "MetricName": "HealthCheckBypass"
          }
        },
        {
          "Name": "AWSManagedRulesCommonRuleSet",
          "Priority": 1000,
          "Statement": {
            "ManagedRuleGroupStatement": {
              "VendorName": "AWS",
              "Name": "AWSManagedRulesCommonRuleSet"
            }
          },
          "OverrideAction": {"None": {}}
        }
      ]
    }
```

---

## Key Decisions

✅ **WafConfiguration for portable intent** - Most users  
✅ **WafPolicy for provider-specific passthrough** - Expert users  
✅ **SKR controller translates WafConfiguration → WafPolicy**  
✅ **KCP remote reconcilers provision WafPolicy → Cloud resources in SKR's account**  
✅ **Pre-deployed base policies** (not provisioned until used)  
✅ **AppLoadBalancer can reference either** kind  
✅ **No `spec.data` in WafConfiguration** - keeps it clean  
✅ **No portable fields in WafPolicy** - keeps it simple  
✅ **customRules (not conditionalOverrides)** - universal provider support  

---

## Open Questions for Implementation

1. **Preset storage format**: 
   - Ship as WafPolicy resources (one per provider per level)?
   - Or controller code with provider detection logic?
   - **Recommendation**: Ship as WafPolicy resources with labels (inspectable, upgradeable)

2. **Version management**:
   - How do we version base policies (owasp-moderate-v1, owasp-moderate-v2)?
   - Immutable once created? Auto-upgrade with Cloud Manager releases?
   - **Recommendation**: Immutable policies, version in name, document upgrade path

3. **Translation logic location**:
   - In WafConfiguration controller? (tight coupling)
   - Separate translator library? (reusable)
   - **Recommendation**: Separate package under `pkg/skr/wafconfiguration/translator/`

4. **Generated WafPolicy ownership**:
   - Can users modify generated WafPolicy?
   - **Recommendation**: No - WafPolicy is owned by WafConfiguration, re-generated on each reconcile

---

## Why This Design Works

1. **Honest about complexity**: We don't pretend AWS and Azure are the same
2. **Provides value**: Curated presets, portable intent, lifecycle management
3. **Clear separation**: WafConfiguration (what) vs WafPolicy (how)
4. **Escape hatches**: Users can always drop to WafPolicy for full control
5. **Discoverable**: Presets are visible resources, not hidden magic
6. **Incremental**: Start simple (preset), grow to portable (WafConfiguration), escape to expert (WafPolicy)

---

## What This Means for the Spike

The spike deliverable should document:

1. **The pattern**: Two-level architecture (WafConfiguration → WafPolicy → Cloud)
2. **When to use it**: Use WafConfiguration for most Cloud Manager resources with portable intent
3. **When NOT to use it**: Simple passthrough resources (like IpRange) don't need translation layer
4. **Trade-offs**: More complexity (two resources) but better separation of concerns

This pattern applies beyond WAF:
- **Certificate**: Could have CertificateRequest (portable) → Certificate (provider-specific)
- **RedisInstance**: Could have RedisConfiguration (portable) → RedisInstance (provider-specific)
- **VpcPeering**: Could have VpcPeeringConfiguration → VpcPeering

But not everything needs it:
- **IpRange**: Already provider-neutral, no translation needed
- **FileShare**: Simple enough to be provider-specific directly

---

## Summary

Cloud Manager adds value through:
1. **Standardized lifecycle management** (Ready/Error conditions, observedGeneration)
2. **Common integration points** (`AppLoadBalancer` → `WafConfiguration`/`WafPolicy` reference)
3. **Curated base policies** for simple cases (owasp-basic, owasp-moderate, owasp-strict)
4. **Portable intent** via `WafConfiguration` (managedRuleGroups, classifications, conditions)
5. **Clear escape hatches** via `WafPolicy` for full provider control
6. **Honesty about provider differences** (no fake portability)

This is:
- More honest than Crossplane's "CompositeResourceDefinition" magic
- More opinionated than raw Terraform (we provide presets and translation)
- More flexible than "one API to rule them all" (clear escape hatches)
