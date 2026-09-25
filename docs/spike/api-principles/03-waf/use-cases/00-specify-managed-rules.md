# Use Case 0: Specify Managed Rule Groups

## User Story

"As a platform operator, I want to enable OWASP protection on my application load balancer by specifying which managed rule groups to activate, so that I get baseline WAF protection without writing custom rules."

## Requirements

1. Select managed rule groups from provider's catalog
2. Set action for each group (block, count, allow)
3. No custom rule authoring required
4. Works on AWS, Azure, GCP (provider-specific)

## Real-World Scenarios

### Scenario A: Basic OWASP Protection (AWS)
```yaml
# AWS-specific managed rule group names
managedRuleGroups:
  - name: AWSManagedRulesCommonRuleSet
    action: block
```

### Scenario B: Detection Mode (Azure)
```yaml
# Azure-specific managed rule set names
managedRuleSets:
  - ruleSetType: Microsoft_DefaultRuleSet
    ruleSetVersion: "2.1"
  - ruleSetType: Microsoft_BotManagerRuleSet
    ruleSetVersion: "1.0"
# mode: Detection (global)
```

### Scenario C: Production Multi-Layer Defense (GCP)
```yaml
# GCP-specific preconfigured WAF identifiers
preconfiguredWafRules:
  - owasp-crs-v030301-id
  - sqli-v33-stable
  - xss-v33-stable
  - cve-canary
```

## Portable API Design

**NOT IMPLEMENTED** - Managed rule groups are provider-specific and will NOT be made portable.

**Reason:** Provider naming and bundling differences are too significant:

| Portable Concept | AWS | Azure | GCP |
|------------------|-----|-------|-----|
| OWASP Core Rules | `AWSManagedRulesCommonRuleSet` | `Microsoft_DefaultRuleSet` (includes SQL, XSS) | `owasp-crs-v030301-id` |
| SQL Injection | `AWSManagedRulesSQLiRuleSet` | (Included in DefaultRuleSet) | `sqli-v33-stable` |
| XSS Protection | (Included in CommonRuleSet) | (Included in DefaultRuleSet) | `xss-v33-stable` |
| Bot Protection | `AWSManagedRulesBotControlRuleSet` | `Microsoft_BotManagerRuleSet` | `cve-canary` |

**Solution:** Use Phase 1 WafPolicy with provider-specific `spec.data` JSON.

## Provider Capability Check

### AWS WAFv2

**Status:** ✅ **Perfect Support**

**How it works:**
- AWS provides managed rule groups from AWS and third-party vendors
- Each rule group is a `ManagedRuleGroupStatement`
- Common rule groups: `AWSManagedRulesCommonRuleSet`, `AWSManagedRulesSQLiRuleSet`, `AWSManagedRulesBotControlRuleSet`

**Native AWS Translation:**
```json
{
  "Name": "production-webacl",
  "Scope": "REGIONAL",
  "DefaultAction": {"Allow": {}},
  "Rules": [
    {
      "Name": "AWSManagedRulesCommonRuleSet",
      "Priority": 1000,
      "Statement": {
        "ManagedRuleGroupStatement": {
          "VendorName": "AWS",
          "Name": "AWSManagedRulesCommonRuleSet"
        }
      },
      "OverrideAction": {"None": {}},
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "CommonRuleSet"
      }
    },
    {
      "Name": "AWSManagedRulesSQLiRuleSet",
      "Priority": 1100,
      "Statement": {
        "ManagedRuleGroupStatement": {
          "VendorName": "AWS",
          "Name": "AWSManagedRulesSQLiRuleSet"
        }
      },
      "OverrideAction": {"None": {}},
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "SQLiRuleSet"
      }
    },
    {
      "Name": "AWSManagedRulesBotControlRuleSet",
      "Priority": 1200,
      "Statement": {
        "ManagedRuleGroupStatement": {
          "VendorName": "AWS",
          "Name": "AWSManagedRulesBotControlRuleSet"
        }
      },
      "OverrideAction": {"None": {}},
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "BotControl"
      }
    }
  ]
}
```

**Complexity:** Low - Direct mapping

**Key Insight:**
- Each managed rule group becomes a separate rule in the WebACL
- `OverrideAction: None` means use the rule group's default action
- Priority determines evaluation order

---

### Azure Application Gateway WAF

**Status:** ✅ **Perfect Support**

**How it works:**
- Azure provides managed rule sets based on OWASP CRS
- Each rule set is specified in `managedRuleSets` array
- Common rule sets: `Microsoft_DefaultRuleSet`, `Microsoft_BotManagerRuleSet`

**Native Azure Translation:**
```json
{
  "location": "eastus",
  "properties": {
    "managedRules": {
      "managedRuleSets": [
        {
          "ruleSetType": "Microsoft_DefaultRuleSet",
          "ruleSetVersion": "2.1"
        },
        {
          "ruleSetType": "Microsoft_BotManagerRuleSet",
          "ruleSetVersion": "1.0"
        }
      ]
    },
    "policySettings": {
      "state": "Enabled",
      "mode": "Prevention",
      "requestBodyCheck": true,
      "maxRequestBodySizeInKb": 128
    }
  }
}
```

**Complexity:** Low - Flat array structure

**Azure Note:**
- Mode applies to all rule sets: `Prevention` (block) or `Detection` (count)
- No per-rule-set action control (global mode only)
- Version pinning available for stability

---

### GCP Cloud Armor

**Status:** ✅ **Perfect Support**

**How it works:**
- GCP provides preconfigured WAF rules based on OWASP ModSecurity CRS
- Each rule group is a separate rule with `evaluatePreconfiguredWaf()`
- Common rule sets: `owasp-crs-v030301-id`, `sqli-v33-stable`, `xss-v33-stable`

**Native GCP Translation:**
```json
{
  "name": "production-security-policy",
  "rules": [
    {
      "priority": 1000,
      "description": "OWASP ModSecurity CRS",
      "action": "deny(403)",
      "preview": false,
      "match": {
        "expr": {
          "expression": "evaluatePreconfiguredWaf('owasp-crs-v030301-id', {'sensitivity': 1})"
        }
      }
    },
    {
      "priority": 1100,
      "description": "SQL injection protection",
      "action": "deny(403)",
      "preview": false,
      "match": {
        "expr": {
          "expression": "evaluatePreconfiguredWaf('sqli-v33-stable', {'sensitivity': 1})"
        }
      }
    },
    {
      "priority": 1200,
      "description": "XSS protection",
      "action": "deny(403)",
      "preview": false,
      "match": {
        "expr": {
          "expression": "evaluatePreconfiguredWaf('xss-v33-stable', {'sensitivity': 1})"
        }
      }
    },
    {
      "priority": 2147483647,
      "description": "Default rule - allow",
      "action": "allow",
      "match": {
        "expr": {
          "expression": "true"
        }
      }
    }
  ]
}
```

**Complexity:** Low - Rule per managed rule group

**GCP Advantages:**
- Per-rule action control (deny vs preview)
- Sensitivity levels (0-4) for tuning
- CEL expressions for advanced composition

---

## Cross-Provider Comparison

| Feature | AWS WAFv2 | Azure WAF | GCP Cloud Armor |
|---------|-----------|-----------|-----------------|
| **Managed rule groups** | ✅ Yes | ✅ Yes | ✅ Yes |
| **Per-group action** | ✅ Yes | ⚠️ Global mode only | ✅ Yes |
| **Count mode (detection)** | ✅ Per group | ✅ Global | ✅ Per rule |
| **Block mode** | ✅ Per group | ✅ Global | ✅ Per rule |
| **Rule group versioning** | ⚠️ Auto-update | ✅ Version pinning | ⚠️ Auto-update |
| **Naming consistency** | ❌ AWS-specific | ❌ Azure-specific | ❌ GCP-specific |
| **Bundling consistency** | ❌ Different | ❌ Different | ❌ Different |

## Provider-Specific Names

**AWS managed rule group names:**
- `AWSManagedRulesCommonRuleSet` - OWASP core rules
- `AWSManagedRulesSQLiRuleSet` - SQL injection rules
- `AWSManagedRulesKnownBadInputsRuleSet` - Known bad inputs
- `AWSManagedRulesBotControlRuleSet` - Bot protection

**Azure managed rule set types:**
- `Microsoft_DefaultRuleSet` - Includes OWASP core, SQL injection, XSS (bundled)
- `Microsoft_BotManagerRuleSet` - Bot protection

**GCP preconfigured WAF identifiers:**
- `owasp-crs-v030301-id` - OWASP ModSecurity Core Rule Set
- `sqli-v33-stable` - SQL injection stable ruleset
- `xss-v33-stable` - XSS stable ruleset
- `cve-canary` - Bot and scanner protection

## Design Decision Impact

### API Design

❌ **DO NOT include `managedRuleGroups` in portable WafConfiguration** - provider differences are too significant:

**Why not portable:**
1. **Naming inconsistency:** AWS "CommonRuleSet" vs Azure "DefaultRuleSet" vs GCP "owasp-crs-v030301-id"
2. **Bundling differences:** Azure bundles SQL+XSS in DefaultRuleSet, AWS/GCP separate them
3. **Version management:** GCP uses explicit versions, AWS/Azure auto-update
4. **False abstraction:** A portable name hides real provider differences and creates confusion

**Correct approach:**
- **Phase 1 WafPolicy presets** already include provider-specific managed rules in `spec.data`
- Users can create custom WafPolicy with provider-specific JSON for full control
- **Phase 2 WafConfiguration** focuses on portable features: `ruleOverrides` and `customRules`

---

### Implementation Notes

**Phase 1: WafPolicy with spec.data (provider-specific JSON)**

Users specify managed rule groups using provider-native JSON:

**AWS example:**
```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: WafPolicy
metadata:
  name: my-aws-policy
spec:
  data: |
    {
      "Rules": [
        {
          "Name": "AWSManagedRulesCommonRuleSet",
          "Priority": 1000,
          "Statement": {
            "ManagedRuleGroupStatement": {
              "VendorName": "AWS",
              "Name": "AWSManagedRulesCommonRuleSet"
            }
          }
        }
      ]
    }
```

**Azure example:**
```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: WafPolicy
metadata:
  name: my-azure-policy
spec:
  data: |
    {
      "managedRules": {
        "managedRuleSets": [
          {
            "ruleSetType": "Microsoft_DefaultRuleSet",
            "ruleSetVersion": "2.1"
          }
        ]
      }
    }
```

**GCP example:**
```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: WafPolicy
metadata:
  name: my-gcp-policy
spec:
  data: |
    {
      "rules": [
        {
          "priority": 1000,
          "action": "deny(403)",
          "match": {
            "expr": {
              "expression": "evaluatePreconfiguredWaf('owasp-crs-v030301-id')"
            }
          }
        }
      ]
    }
```

---

## Validation Matrix

| Test Case | AWS | Azure | GCP | Expected Behavior |
|-----------|-----|-------|-----|-------------------|
| OWASP core rules in block mode | ✅ Block | ✅ Block | ✅ Block | All providers block |
| Multiple rule groups in block mode | ✅ All block | ✅ All block | ✅ All block | All providers apply all groups |
| Mixed block + count actions | ✅ Per-group | ⚠️ Global mode | ✅ Per-group | Azure uses first action globally |
| Detection mode (count/preview) | ✅ Count | ✅ Detection | ✅ Preview | All providers log without blocking |

---

## Recommendation

**Decision:** ⚠️ **DEFERRED - Too complex for portable API**

**Rationale:**
1. **Provider naming differences:** AWS "CommonRuleSet" vs Azure "DefaultRuleSet" vs GCP "owasp-crs-v030301-id"
2. **Structural differences:** Azure includes SQL injection in DefaultRuleSet, AWS separates it
3. **Version management:** GCP has explicit versions, AWS/Azure auto-update
4. **False portability:** Abstraction would hide real provider differences

**Alternative Approach:**
- **Phase 1:** WafPolicy presets (owasp-moderate, owasp-strict) already include provider-specific managed rules in `spec.data`
- **Phase 2.1:** Users customize via `ruleOverrides` (unconditional) and `customRules` (with conditions)
- **Future:** If strong user demand, revisit portable `managedRuleGroups` with clear provider mapping table

**For now:**
Users who need different managed rule groups should:
1. Use WafPolicy directly with provider-specific JSON (full control)
2. Or start from preset and use `ruleOverrides` to tune behavior

**This use case remains valuable** for validating provider capabilities, even though we defer the portable API.

---

## Conclusion

**Specifying managed rule groups is foundational, but too complex for a portable abstraction right now.**

Phase 1 WafPolicy presets already solve this with provider-specific JSON. Phase 2.1 focuses on the features that ARE portable: `ruleOverrides` and `customRules`.

**Marked as Use Case 00** because it's conceptually first, but **deferred from Phase 2 implementation** due to mapping complexity.
