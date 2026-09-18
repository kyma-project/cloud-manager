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
  observedGeneration: 1
  conditions:
    - type: Ready
      status: False
      reason: NotUsed
      message: "Not provisioned until referenced by AppLoadBalancer"
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
- **Portable extensions**: Add `ruleOverrides`, `customRules`, and future fields in provider-neutral way
- **Future-proof**: New fields (customRules, sizeLimits, geoBlocking) can be added without breaking existing configs

### WafConfiguration API

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafConfiguration
metadata:
  name: my-config
spec:
  # Option 1: Start from named preset (simple)
  preset: owasp-moderate
  
  # OR Option 2: Start from any WafPolicy (flexible)
  basePolicyRef:
    name: my-company-standard
  
  # OR Option 3: No base - define everything yourself (advanced)
  # If neither preset nor basePolicyRef is specified, start from scratch
  
  # --- Everything below overrides/extends the base ---
  
  # Optional: Override specific rules from base (unconditional)
  ruleOverrides:
    - managedRuleGroup: CoreRuleSet
      ruleId: "942100"  # Provider-specific rule ID
      action: count     # Change to detection mode
      reason: "Testing for false positives on GraphQL API"
    
    - managedRuleGroup: SQLInjectionProtection
      ruleId: "942200"
      action: allow  # Disable specific rule
      reason: "Known false positive on search endpoint"
  
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
  
  observedGeneration: 1
  conditions:
    - type: Ready
      status: True
      reason: PolicyGenerated
      message: "Generated WafPolicy my-config-wafpolicy from preset owasp-moderate with 2 rule overrides, 3 custom rules"
```

### How WafConfiguration Translation Works

**Simple merge pattern**: Base + WafConfiguration changes = Generated WafPolicy

```
1. Controller reads base WafPolicy
   - If preset: "owasp-moderate" → reads kyma-system/owasp-moderate
   - If basePolicyRef: reads specified policy
   - If neither: starts with empty base

2. Parse base spec.data (provider-specific JSON)

3. Apply WafConfiguration changes on top:
   - ruleOverrides: Change specific rule actions (unconditional)
   - customRules: Add new rules with conditions
   - Future fields: sizeLimits, managedRuleGroups, etc.

4. Generate new WafPolicy with merged result:
   - Name: <wafconfig-name>-wafpolicy
   - spec.data: Merged provider-specific JSON
   - ownerReferences: Points to WafConfiguration

5. KCP reconciler provisions generated WafPolicy to cloud
```

**Controller logic**:

1. Load base WafPolicy (from preset or basePolicyRef)
2. Parse base spec.data (provider-specific JSON)
3. Apply ruleOverrides (unconditional changes to managed rules)
4. Add customRules (new rules with conditions)
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

**Pattern 1: Preset with simple overrides (most common)**
```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafConfiguration
metadata:
  name: my-app-waf
spec:
  preset: owasp-moderate
  
  ruleOverrides:
    - managedRuleGroup: CoreRuleSet
      ruleId: "942100"
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
    - managedRuleGroup: CoreRuleSet
      ruleId: "942100"
      action: count
```

**Pattern 3: Define from scratch (no base)**
```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafConfiguration
metadata:
  name: my-custom-waf
spec:
  # No preset, no basePolicyRef - define everything
  
  # Future Phase 2.1+ fields when managedRuleGroups are portable:
  # managedRuleGroups:
  #   - type: CoreRuleSet
  #     action: block
  #   - type: SQLInjectionProtection
  #     action: block
  
  customRules:
    - name: custom-protection
      priority: 100
      action: block
      conditions:
        path:
          prefix: "/api"
```

### Validation Rules

**CEL validation** ensures correct usage:

```yaml
# Can't have both preset and basePolicyRef
- rule: "!(has(self.preset) && has(self.basePolicyRef))"
  message: "Cannot specify both preset and basePolicyRef"

# ruleOverrides requires a base (preset or basePolicyRef)
- rule: "!has(self.ruleOverrides) || has(self.preset) || has(self.basePolicyRef)"
  message: "ruleOverrides requires preset or basePolicyRef to override"

# customRules priority must be unique
- rule: "self.customRules.all(r1, self.customRules.all(r2, r1.name == r2.name || r1.priority != r2.priority))"
  message: "customRules must have unique priorities"
```

### Phase 2 Benefits

✅ **Preset + customization** - Start from proven base, tweak as needed  
✅ **Preset upgrades** - When preset updates, regenerate includes updates  
✅ **No duplication** - Don't copy full preset, just specify changes  
✅ **Future-proof** - New fields add on top without breaking existing configs  
✅ **Backward compatible** - Phase 1 WafPolicy references keep working  
✅ **Flexible base** - Can use preset, custom policy, or start from scratch  
✅ **Clear semantics** - "Base + my changes = result"  
✅ **Universal custom rules** - customRules work perfectly on all providers  

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

## Phase 1 Implementation Checklist

### Deliverables

- [ ] **WafPolicy CRD** with `spec.data` field only
- [ ] **Pre-deployed preset WafPolicy resources** (owasp-basic, owasp-moderate, owasp-strict, owasp-detection)
- [ ] **Preset protection** (finalizers, validation webhooks)
- [ ] **AppLoadBalancer.spec.policy** field supporting WafPolicy reference
- [ ] **KCP reconcilers** for AWS/Azure/GCP (provision WafPolicy to cloud)
- [ ] **Gated provisioning** (don't provision until referenced)
- [ ] **Status tracking** (Ready condition, providerId)

### Documentation

- [ ] User guide: Using preset WafPolicy
- [ ] User guide: Writing custom WafPolicy
- [ ] Provider-specific JSON examples (AWS/Azure/GCP)
- [ ] Preset upgrade guide (how to fork and customize)

---

## Phase 2 Implementation Prerequisites

Before adding WafConfiguration:

- [ ] **Phase 1 deployed and validated** in production
- [ ] **User feedback collected** on what customizations are needed
- [ ] **Cross-provider analysis complete** for common override patterns
- [ ] **Translation strategy validated** for at least 2 providers
- [ ] **Override patterns documented** (ruleOverrides, customRules)

### Phase 2 Deliverables

- [ ] **WafConfiguration CRD** with preset/basePolicyRef and override fields
- [ ] **WafConfiguration controller** (translation + WafPolicy generation)
- [ ] **CEL validation** (preset XOR basePolicyRef, etc.)
- [ ] **AppLoadBalancer update** to support WafConfiguration reference
- [ ] **Migration guide** from Phase 1 to Phase 2

### Phase 2 Documentation

- [ ] User guide: Using preset with overrides
- [ ] User guide: Portable override patterns
- [ ] Translation reference (how overrides map to providers)
- [ ] Migration examples (Phase 1 → Phase 2)

---

## Future Evolution (Phase 2.2+)

WafConfiguration can grow with new portable fields:

**Phase 2.2**:
- `managedRuleGroups` - Portable managed rule group selection (when pattern is validated)
- `sizeLimits` - Request size constraints

**Phase 2.3**:
- `geoBlocking` - Geographic restrictions
- `timeBasedRules` - Time-based activation
- Advanced rate limiting with custom keys

All added fields follow the same pattern: **Base + additions = Generated WafPolicy**

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

**Option 2: Migrate to WafConfiguration with preset**
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
  preset: owasp-moderate
---
policy:
  kind: WafConfiguration
  name: my-config
```

**Option 3: Add overrides and custom rules to preset**
```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafConfiguration
metadata:
  name: my-config
spec:
  preset: owasp-moderate
  ruleOverrides:
    - managedRuleGroup: CoreRuleSet
      ruleId: "942100"
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
- Preset-based starting point
- Targeted overrides without full JSON
- Automatic preset upgrades
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
