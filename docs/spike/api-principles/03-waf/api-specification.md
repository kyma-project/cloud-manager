# WAF API Specification - Phased Approach

## Overview

This document defines Cloud Manager's two-level WAF API architecture, designed for **phased implementation**:

- **Phase 1 (Ship First)**: WafPolicy only - provider-specific passthrough
- **Phase 2 (Add Later)**: WafConfiguration - portable intent with translation

This phased approach allows Cloud Manager to ship quickly with provider-specific policies, then add a portable layer after validating cross-provider capabilities with real user feedback.

---

## Phase 1: WafPolicy (Provider-Specific Passthrough)

**Ship this first.** Simple, proven pattern from Terraform/Crossplane.

### WafPolicy Design

- **Single field**: `spec.data` - provider-specific JSON passthrough
- **No translation**: Remote reconciliation from KCP to SKR cloud APIs
- **Gated provisioning**: Not provisioned until referenced by AppLoadBalancer
- **Pre-deployed presets**: Cloud Manager ships standard policies that users reference

### WafPolicy API

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: WafPolicy
metadata:
  name: my-policy
spec:
  # ONLY field: provider-specific JSON passthrough
  data: |
    {
      "DefaultAction": {"Allow": {}},
      "Rules": [
        {
          "Name": "CustomRule",
          "Priority": 1,
          "Statement": {...},
          "Action": {"Block": {}}
        }
      ],
      "VisibilityConfig": {...}
    }

status:
  providerId: "arn:aws:wafv2:us-east-1:123456789012:regional/webacl/..."
  conditions:
    - type: Ready
      status: False
      reason: NotUsed
      message: "Not provisioned until referenced by AppLoadBalancer"
      observedGeneration: 1
```

### Pre-deployed WafPolicy Presets

Cloud Manager ships these standard WafPolicy resources. They exist in the cluster but are **not provisioned in cloud** until referenced.

```yaml
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
      # AWS gets AWS-specific JSON
      # Azure gets Azure-specific JSON
      # GCP gets GCP-specific JSON
    }
status:
  conditions:
    - type: Ready
      status: False
      reason: NotUsed
      observedGeneration: 1
```

**Shipped presets:**
- `owasp-basic` - Basic OWASP protection (block mode)
- `owasp-moderate` - Moderate OWASP protection (block mode)
- `owasp-strict` - Strict OWASP protection (block mode)
- `owasp-detection` - OWASP detection only (count mode)

### AppLoadBalancer Integration (Phase 1)

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: AppLoadBalancer
metadata:
  name: my-app
spec:
  backend:
    kind: Service
    name: my-service
    namespace: default
  frontend:
    - hosts: ["my-app.example.com"]
      port: {number: 443, protocol: HTTPS}
      tls: {secretRef: {name: my-tls-cert}}
  
  # Reference WafPolicy (preset or custom)
  policy:
    apiVersion: cloud-resources.kyma-project.io/v1beta1
    kind: WafPolicy
    name: owasp-moderate
```

### Phase 1 Usage Patterns

**Pattern 1: Use preset (simplest)**
```yaml
kind: AppLoadBalancer
spec:
  policy:
    kind: WafPolicy
    name: owasp-moderate
```

**Pattern 2: Custom policy (full control)**
```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: WafPolicy
metadata:
  name: my-custom-policy
spec:
  data: |
    {
      # Full provider-specific JSON
      # Complete control, no portable fields
    }
---
kind: AppLoadBalancer
spec:
  policy:
    kind: WafPolicy
    name: my-custom-policy
```

### Phase 1 Benefits

✅ **Ship immediately** - No need to analyze cross-provider differences first  
✅ **Simple architecture** - WafPolicy → KCP → Cloud (no translation layer)  
✅ **Proven pattern** - Terraform/Crossplane use provider-specific approach  
✅ **Full control** - Users can write any provider-specific configuration  
✅ **Gather feedback** - Learn what users actually need before building portable layer  

---

## Phase 2: WafConfiguration (Portable Intent with Translation)

**Add this later** after validating cross-provider capabilities and gathering user feedback.

### WafConfiguration Design

- **Base + overrides pattern**: Start from a WafPolicy (preset or custom), apply changes on top
- **Always generates fresh WafPolicy**: WafConfiguration controller creates new WafPolicy with merged configuration
- **`customRules` are genuinely portable**: path/header/IP-based custom rules translate cleanly across all three providers
- **`ruleOverrides` are NOT portable**: they use provider-specific rule group names and rule IDs; GCP cannot override individual rules at all — any override degrades the entire ruleset to preview (count) mode, which is a fundamentally different behavior than what the user expressed

> ⚠️ **Design concern**: `ruleOverrides` breaks the portability promise of `WafConfiguration`. The field uses provider-specific names (`managedRuleGroup`, `ruleId`) and behaves differently per provider. This raises the question of whether `ruleOverrides` belongs in `WafConfiguration` at all, or whether it should live in `WafPolicy` where provider-specificity is explicit. See [Design Concerns](#design-concerns) below.

### WafConfiguration API

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafConfiguration
metadata:
  name: my-config
spec:
  # Reference base WafPolicy (pre-deployed or custom)
  basePolicyRef:
    name: owasp-moderate  # Can reference pre-deployed WafPolicy or custom one
  
  # --- Everything below overrides/extends the base ---
  
  # Optional: Override specific rules from base (unconditional)
  # ⚠️ NOT PORTABLE: managedRuleGroup and ruleId are provider-specific.
  # GCP does not support individual rule overrides — any entry here degrades
  # the ENTIRE matching ruleset to preview mode on GCP, not the specific rule.
  ruleOverrides:
    - managedRuleGroup: AWSManagedRulesCommonRuleSet  # AWS-specific (Azure: Microsoft_DefaultRuleSet, GCP: not supported at rule level)
      ruleId: "SizeRestrictions_BODY"  # AWS rule name (Azure: "942100", GCP: no equivalent)
      action: count     # Change to detection mode
      reason: "Testing for false positives on file uploads"
    
    - managedRuleGroup: AWSManagedRulesCommonRuleSet  # AWS-specific managed rule group
      ruleId: "GenericRFI_BODY"  # AWS rule name (Azure: "931130", GCP: no equivalent)
      action: allow  # Disable specific rule
      reason: "Known false positive on API endpoints"
  
  # Optional: Custom rules with conditions (universal support)
  customRules:
    - name: health-check-bypass
      priority: 10
      action: allow
      conditions:
        path:
          exact: "/health"
      reason: "Health check endpoint bypass all WAF rules"
    
    - name: internal-admin-no-rate-limit
      priority: 20
      action: allow
      conditions:
        path:
          prefix: "/admin"
        sourceIP:
          cidr: "10.0.0.0/8"
      reason: "Internal admin traffic bypasses rate limiting"
    
    - name: block-debug-header-external
      priority: 100
      action: block
      conditions:
        path:
          prefix: "/admin"
        header:
          name: "X-Debug"
          exists: true
        sourceIP:
          cidr: "10.0.0.0/8"
          negate: true
      reason: "Block debug header on admin paths from external IPs"
  
  # Future Phase 2.2 fields (example):
  # managedRuleGroups: [...]  # When portable mapping is validated
  # sizeLimits: {...}
  # geoBlocking: {...}

status:
  # Reference to generated WafPolicy
  generatedPolicy:
    apiVersion: cloud-resources.kyma-project.io/v1beta1
    kind: WafPolicy
    name: my-config-wafpolicy
  
  conditions:
    - type: Ready
      status: True
      reason: PolicyGenerated
      message: "Generated WafPolicy my-config-wafpolicy from preset owasp-moderate with 2 rule overrides, 3 custom rules"
      observedGeneration: 1
```

### How WafConfiguration Translation Works

**Simple merge pattern**: Base + WafConfiguration changes = Generated WafPolicy

```
1. Controller reads base WafPolicy
   - If basePolicyRef: reads specified policy (can be pre-deployed like owasp-moderate or custom)
   - If neither: starts with empty base

2. Parse base spec.data (provider-specific JSON)

3. Apply WafConfiguration changes on top:
   - ruleOverrides: Change specific rule actions (Phase 2.1)
   - customRules: Add new rules with conditions (Phase 2.1)
   - Future fields: sizeLimits, geoBlocking, etc. (Phase 2.2+)

4. Generate new WafPolicy with merged result:
   - Name: <wafconfig-name>-wafpolicy
   - spec.data: Merged provider-specific JSON
   - ownerReferences: Points to WafConfiguration

5. KCP reconciler provisions generated WafPolicy to cloud
```

**Controller logic**:

1. Load base WafPolicy (from basePolicyRef, which may point to pre-deployed policy like owasp-moderate)
2. Parse base spec.data (provider-specific JSON)
3. Apply ruleOverrides (tune individual rules - Phase 2.1)
4. Add customRules (new rules with conditions - Phase 2.1)
5. Generate new WafPolicy with merged result
6. Set owner references for lifecycle management

### AppLoadBalancer Integration (Phase 2)

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: AppLoadBalancer
metadata:
  name: my-app
spec:
  backend:
    kind: Service
    name: my-service
  
  # Can reference WafConfiguration OR WafPolicy
  policy:
    kind: WafConfiguration  # Portable intent
    name: my-config
---
# Phase 1 pattern still works (backward compatible)
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: AppLoadBalancer
spec:
  policy:
    kind: WafPolicy  # Provider-specific
    name: owasp-moderate
```

### Phase 2 Usage Patterns

**Pattern 1: Base policy with simple overrides (most common)**
```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafConfiguration
metadata:
  name: my-app-waf
spec:
  basePolicyRef:
    name: owasp-moderate
  
  ruleOverrides:
    - managedRuleGroup: AWSManagedRulesCommonRuleSet  # AWS (Azure: Microsoft_DefaultRuleSet, GCP: owasp-crs-v030301-id)
      ruleId: "SizeRestrictions_BODY"  # AWS rule name (Azure: "942100", GCP: not supported)
      action: count
  
  customRules:
    - name: health-check-bypass
      priority: 10
      action: allow
      conditions:
        path:
          exact: "/health"
```

**Pattern 2: Reference custom base policy**
```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafConfiguration
metadata:
  name: my-app-waf
spec:
  basePolicyRef:
    name: company-standard-waf
  
  ruleOverrides:
    - managedRuleGroup: AWSManagedRulesCommonRuleSet  # AWS (Azure: Microsoft_DefaultRuleSet, GCP: owasp-crs-v030301-id)
      ruleId: "GenericRFI_BODY"  # AWS rule name (Azure: "931130", GCP: not supported)
      action: count
```

**Pattern 3: Define from scratch (no base)**
```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafConfiguration
metadata:
  name: my-custom-waf
spec:
  # No basePolicyRef - define everything yourself
  
  # Phase 2.1+ fields:
  ruleOverrides: [...]        # Tune individual rules
  customRules: [...]          # Add custom protection
  
  # Phase 2.2+ fields:
  # sizeLimits: {...}
  # geoBlocking: {...}
```

### Validation Rules

**CEL validation** ensures correct usage:

```yaml
# ruleOverrides requires a base (basePolicyRef)
- rule: "!has(self.ruleOverrides) || has(self.basePolicyRef)"
  message: "ruleOverrides requires basePolicyRef to override"

# customRules priority must be unique
- rule: "self.customRules.all(r1, self.customRules.all(r2, r1.name == r2.name || r1.priority != r2.priority))"
  message: "customRules must have unique priorities"
```

### Phase 2 Benefits

✅ **Base policy + customization** - Start from proven base, tweak as needed  
✅ **Base policy upgrades** - When base policy updates, regenerate includes updates  
✅ **No duplication** - Don't copy full policy, just specify changes  
✅ **Backward compatible** - Phase 1 WafPolicy references keep working  
✅ **Universal custom rules** - `customRules` work perfectly on all providers  
⚠️ **`ruleOverrides` portability is a lie** - Field names are provider-specific; GCP degrades behavior silently to entire-ruleset preview mode  
⚠️ **`basePolicyRef` is provider-specific** - The base policy contains provider-specific JSON, so the "portable" layer is built on a provider-specific foundation

---

## Phased Implementation Comparison

| Aspect | Phase 1: WafPolicy Only | Phase 2: Add WafConfiguration |
|--------|-------------------------|-------------------------------|
| **What users write** | Provider-specific JSON | Base + portable overrides |
| **Customization** | Full JSON or use preset as-is | Preset + targeted overrides |
| **Learning curve** | Must learn AWS/Azure/GCP | Learn override patterns once |
| **Flexibility** | Full control | 80% use cases + full control fallback |
| **Ship timeline** | Immediate | After cross-provider analysis |
| **User feedback** | Gather what's needed | Deliver what's validated |
| **Risk** | Low (proven pattern) | Medium (new abstraction) |
| **Maintenance** | Low (passthrough) | Medium (translation logic) |
| **Preset upgrades** | Manual (fork preset) | Automatic (regenerate from preset) |

---

## Architecture Flow

### Phase 1 Architecture
```
AppLoadBalancer
    ↓ references
WafPolicy (preset or custom)
    ↓ KCP reconciles
Cloud WAF Resources
```

### Phase 2 Architecture
```
AppLoadBalancer
    ↓ references WafConfiguration OR WafPolicy
    
Option A: Via WafConfiguration
AppLoadBalancer
    ↓ references
WafConfiguration (base + overrides)
    ↓ controller generates
WafPolicy (merged result)
    ↓ KCP reconciles
Cloud WAF Resources

Option B: Via WafPolicy (Phase 1 pattern still works)
AppLoadBalancer
    ↓ references
WafPolicy (preset or custom)
    ↓ KCP reconciles
Cloud WAF Resources
```

---

## Recommended Approach

### Start with Phase 1

**Ship WafPolicy first** because:
1. ✅ Faster time to market
2. ✅ Gather real user feedback on what overrides matter
3. ✅ Validate AppLoadBalancer integration
4. ✅ No wasted work if portable layer isn't needed
5. ✅ Learn which cross-provider features to prioritize

### Add Phase 2 When Ready

**Add WafConfiguration later** when:
1. ✅ Cross-provider analysis is complete
2. ✅ User feedback shows need for preset customization
3. ✅ Translation patterns are validated
4. ✅ You know which override features to make portable

### Both Patterns Coexist

Phase 1 and Phase 2 users can use the same Cloud Manager:

```
Phase 1 Users              Phase 2 Users
    ↓                          ↓
  WafPolicy  ←————————→  WafConfiguration
    ↓                          ↓ (generates)
    ↓                      WafPolicy
    ↓                          ↓
    └────────→ KCP ←───────────┘
               ↓
          Cloud WAF
```

---

## Future Evolution (Phase 2.1+)

WafConfiguration can grow with new portable fields:

**Phase 2.1** — what was planned:
- `customRules` - Add new rules with conditions ✅ Genuinely portable (universal provider support)
- `ruleOverrides` - Unconditional changes to managed rules ⚠️ **Not portable** — see Design Concerns below

**Phase 2.2+** (all with provider gaps):
- `sizeLimits` - Request size constraints (AWS/Azure native, GCP limited)
- `geoBlocking` - Geographic restrictions (Azure does not support at WAF level)
- `timeBasedRules` - Time-based activation (no universal support)

**NOT PLANNED for portable API:**
- ❌ `managedRuleGroups` - Providers have completely different offerings and naming schemes
  - AWS: `AWSManagedRulesCommonRuleSet`, `AWSManagedRulesSQLiRuleSet`
  - Azure: `Microsoft_DefaultRuleSet` (bundles multiple protections)
  - GCP: `owasp-crs-v030301-id`, `sqli-v33-stable`
  - **Solution**: Phase 1 WafPolicy presets already include provider-specific managed rules

---

## Design Concerns

### `ruleOverrides` does not belong in WafConfiguration

The `ruleOverrides` field was included in `WafConfiguration` as a "portable" way to tune managed rules, but the analysis reveals it is not portable:

1. **Provider-specific names**: `managedRuleGroup` and `ruleId` use provider-specific identifiers. The same rule does not exist under the same name on AWS, Azure, and GCP. A `WafConfiguration` that works on AWS silently means something different (or nothing) on Azure/GCP.

2. **GCP cannot implement it**: GCP Cloud Armor has no per-rule override mechanism. Any `ruleOverrides` entry on GCP degrades the **entire** matched ruleset to `preview: true` (count mode) — a blunt, unintended side-effect that contradicts the user's expressed intent.

3. **It is provider-specific by design**: The `WafPolicy` layer already exists for provider-specific configuration. `ruleOverrides` in `WafConfiguration` is a provider-specific feature with a portable-looking wrapper.

**Options being considered:**

| Option | Trade-off |
|--------|-----------|
| Remove `ruleOverrides` from `WafConfiguration` | Honest — keeps `WafConfiguration` to what is actually portable (`customRules`); users who need rule tuning drop to `WafPolicy` |
| Keep `ruleOverrides` with explicit provider scope | `ruleOverrides` becomes `spec.aws.ruleOverrides` / `spec.azure.ruleOverrides` — but then `WafConfiguration` is just a multi-provider `WafPolicy` with extra steps |
| Drop `WafConfiguration` entirely | If `customRules` is the only genuinely portable feature, a simpler `customRules` field directly on `AppLoadBalancer` or `WafPolicy` may suffice |

**Current leaning**: `WafConfiguration` makes less and less sense as a portable abstraction. The only feature that justifies it is `customRules` (path/header/IP rules). Everything else either requires provider-specific names or has uneven provider support. The spike should inform whether a dedicated portable resource is worth the added complexity, or whether `customRules` on `WafPolicy` is sufficient.

---

## Migration Path

### Phase 1 → Phase 2

Users on Phase 1 have three options:

**Option 1: Keep using WafPolicy** (no migration needed)
```yaml
# Works forever - no breaking changes
policy:
  kind: WafPolicy
  name: owasp-moderate
```

**Option 2: Migrate to WafConfiguration with base policy reference**
```yaml
# Before (Phase 1)
policy:
  kind: WafPolicy
  name: owasp-moderate

# After (Phase 2)
---
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafConfiguration
metadata:
  name: my-config
spec:
  basePolicyRef:
    name: owasp-moderate
---
policy:
  kind: WafConfiguration
  name: my-config
```

**Option 3: Add overrides and custom rules to base policy**
```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafConfiguration
metadata:
  name: my-config
spec:
  basePolicyRef:
    name: owasp-moderate
  ruleOverrides:
    - managedRuleGroup: AWSManagedRulesCommonRuleSet  # AWS (Azure: Microsoft_DefaultRuleSet, GCP: owasp-crs-v030301-id)
      ruleId: "SizeRestrictions_BODY"  # AWS rule name (Azure: "942100", GCP: not supported)
      action: count
  customRules:
    - name: health-bypass
      priority: 10
      action: allow
      conditions:
        path:
          exact: "/health"
---
policy:
  kind: WafConfiguration
  name: my-config
```

---

## Benefits Summary

### Two-Level Architecture Benefits

**WafPolicy (Provider-Specific)**:
- Full control for experts
- Direct cloud API passthrough
- No translation overhead
- Proven pattern (Terraform/Crossplane)

**WafConfiguration (Portable Intent)**:
- Base policy starting point (pre-deployed or custom)
- Targeted overrides without full JSON
- Automatic base policy upgrades
- Future-proof extensibility

**Both Coexist**:
- Users choose their level
- No forced migration
- Learn incrementally

### Key Design Principles

1. ✅ **Honest about provider differences** - WafPolicy is explicitly provider-specific
2. ✅ **Value through curation** - Presets provide proven starting points
3. ✅ **Clear escape hatches** - Can always drop to WafPolicy for full control
4. ✅ **Phased delivery** - Ship fast (Phase 1), add value incrementally (Phase 2)
5. ✅ **User feedback driven** - Phase 2 design informed by Phase 1 usage
6. ✅ **Future-proof** - New fields add on top without breaking changes
7. ✅ **Simple mental model** - "Base + my changes = result"
