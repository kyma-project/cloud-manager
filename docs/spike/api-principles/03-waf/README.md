# Cloud Manager WAF API Specification

## Overview

This directory contains the complete WAF (Web Application Firewall) API specification for Cloud Manager, designed as a **two-level architecture** with **phased implementation**.

---

## Quick Start

### For Users (Phase 1)

Use pre-deployed WafPolicy presets:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: AppLoadBalancer
spec:
  backend: {kind: Service, name: my-service, namespace: default}
  policy:
    kind: WafPolicy
    name: owasp-moderate
```

### For Developers (Phase 2)

Customize presets with WafConfiguration:

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
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: AppLoadBalancer
spec:
  policy:
    kind: WafConfiguration
    name: my-config
```

---

## Architecture

### Two-Level Design

```
Phase 1: WafPolicy (Provider-Specific)
  - spec.data: JSON passthrough
  - Remote reconciliation from KCP to SKR
  - Pre-deployed presets

Phase 2: WafConfiguration (Portable Intent)
  - Base + overrides pattern
  - SKR controller translates to WafPolicy
  - Portable across AWS/Azure/GCP
```

### Phased Implementation

**Phase 1** (Ship first): Simple, proven, fast to market  
**Phase 2** (Add later): Portable, informed by Phase 1 feedback

Both coexist. No forced migration.

---

## Documentation Files

### Core Specification

**[api-specification.md](api-specification.md)** - THE complete API specification  
- Phase 1: WafPolicy API
- Phase 2: WafConfiguration API
- Pre-deployed presets
- Usage patterns
- Implementation checklist

**[design-rationale.md](design-rationale.md)** - Why this design  
- Two-level architecture explanation
- Phased approach reasoning
- Comparison to Terraform/Crossplane
- Key design decisions

**[implementation-examples.md](implementation-examples.md)** - Complete working examples  
- 15 comprehensive condition scenarios
- Full AWS WAFv2 JSON
- Full Azure Application Gateway WAF JSON
- Full GCP Cloud Armor JSON
- Translation notes

---

### Supporting Analysis

**[research/](research/)** - Cross-provider analysis  
- Boolean logic capabilities
- Label chaining support
- Condition type support matrix
- Managed rule structures
- See [research/README.md](research/README.md)

**[use-cases/](use-cases/)** - Provider capability validation  
- 8 real-world use cases
- Provider support matrix (AWS, Azure, GCP)
- Feasibility for Phase 2 features
- See [use-cases/README.md](use-cases/README.md)

---

## Key Concepts

### WafPolicy (Phase 1)

**Provider-specific passthrough:**
- Single field: `spec.data` (provider JSON)
- No translation
- Remote reconciliation from KCP to SKR
- Full control for experts

**Pre-deployed presets:**
- `owasp-basic`, `owasp-moderate`, `owasp-strict`, `owasp-detection`
- Not provisioned until referenced
- Protected from modification/deletion

---

### WafConfiguration (Phase 2)

**Base + overrides pattern:**
- Start from preset or custom WafPolicy
- Apply portable overrides on top
- Controller generates new WafPolicy
- Supports future extensions

**Key features:**
- `preset` or `basePolicyRef` - Starting point
- `ruleOverrides` - Change specific rule actions (unconditional)
- `customRules` - Add new rules with conditions (universal support)

---

## Reading Guide

### I want to...

**...understand the API design**  
→ Start with [api-specification.md](api-specification.md)

**...understand why we chose this approach**  
→ Read [design-rationale.md](design-rationale.md)

**...see complete working examples**  
→ Browse [implementation-examples.md](implementation-examples.md)

**...understand provider differences**  
→ Explore [research/](research/)

**...validate a specific use case**  
→ Check [use-cases/](use-cases/)

**...implement Phase 1**  
→ Follow checklist in [api-specification.md](api-specification.md#phase-1-implementation-checklist)

**...plan Phase 2**  
→ Review prerequisites in [api-specification.md](api-specification.md#phase-2-implementation-prerequisites)

---

## Design Principles

1. ✅ **Honest about provider differences** - WafPolicy is explicitly provider-specific
2. ✅ **Value through curation** - Presets provide proven starting points
3. ✅ **Clear escape hatches** - Can always use WafPolicy for full control
4. ✅ **Phased delivery** - Ship fast (Phase 1), add value incrementally (Phase 2)
5. ✅ **User feedback driven** - Phase 2 design informed by Phase 1 usage
6. ✅ **Future-proof** - New fields add on top without breaking changes
7. ✅ **Simple mental model** - "Base + my changes = result"

---

## Status

**Phase 1**: Ready to implement  
**Phase 2**: Fully designed, validated, ready to implement after Phase 1 feedback

---

## Related Documentation

- Spike request: [../spike-request.md](../spike-request.md)
- API principles: [../README.md](../README.md) (if exists)
- AppLoadBalancer integration: See main Cloud Manager docs

---

## Questions?

- **API usage**: See [api-specification.md](api-specification.md)
- **Provider differences**: See [research/](research/)
- **Specific use case**: See [use-cases/](use-cases/)
- **Implementation**: Follow checklists in api-specification.md
