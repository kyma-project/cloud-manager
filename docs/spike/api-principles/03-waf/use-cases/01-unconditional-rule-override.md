# Use Case 1: Unconditional Rule Override in Managed Rule Group

## User Story

"As a platform operator, I want to enable OWASP CoreRuleSet in block mode globally, but change rule 942100 to count mode because it triggers false positives for our application."

## Requirements

1. Enable managed rule groups (e.g., OWASP CoreRuleSet)
2. Set default action to block
3. Override specific rule IDs to different action (e.g., count)
4. Override applies to ALL requests (no conditions)

## Portable API Design

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafConfiguration
metadata:
  name: production-policy
spec:
  managedRuleGroups:
    - type: CoreRuleSet
      action: block
    - type: SQLInjectionProtection
      action: block
  
  # Unconditional rule overrides
  ruleOverrides:
    - managedRuleGroup: CoreRuleSet
      ruleId: "942100"
      action: count
      reason: "False positives on legitimate GraphQL queries"
    - managedRuleGroup: CoreRuleSet
      ruleId: "920280"
      action: count
      reason: "Testing for false positives"
```

## Provider Capability Check

### AWS WAFv2

**Status:** ✅ **Perfect Support**

**How it works:**
- Uses `RuleActionOverrides` field in `ManagedRuleGroupStatement`
- Can override individual rule actions by rule name
- No conditions needed

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
          "Name": "AWSManagedRulesCommonRuleSet",
          "RuleActionOverrides": [
            {
              "Name": "SizeRestrictions_BODY",
              "ActionToUse": {"Count": {}}
            },
            {
              "Name": "GenericRFI_BODY",
              "ActionToUse": {"Count": {}}
            }
          ]
        }
      },
      "OverrideAction": {"None": {}},
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "CommonRuleSet"
      }
    },
    {
      "Name": "AWSManagedRulesSQLiRuleSet",
      "Priority": 1010,
      "Statement": {
        "ManagedRuleGroupStatement": {
          "VendorName": "AWS",
          "Name": "AWSManagedRulesSQLiRuleSet"
        }
      },
      "OverrideAction": {"None": {}},
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "SQLiRuleSet"
      }
    }
  ]
}
```

**Complexity:** Low - Direct 1:1 mapping

---

### Azure Application Gateway WAF

**Status:** ✅ **Perfect Support**

**How it works:**
- Uses `ruleGroupOverrides` in managed rule sets
- Can override rule actions at rule group → rule level
- Supports `action` field (Log, Block, Allow)

**Native Azure Translation:**
```json
{
  "location": "eastus",
  "properties": {
    "customRules": [],
    "managedRules": {
      "managedRuleSets": [
        {
          "ruleSetType": "Microsoft_DefaultRuleSet",
          "ruleSetVersion": "2.1",
          "ruleGroupOverrides": [
            {
              "ruleGroupName": "SQLI",
              "rules": [
                {
                  "ruleId": "942100",
                  "state": "Enabled",
                  "action": "Log"
                },
                {
                  "ruleId": "920280",
                  "state": "Enabled",
                  "action": "Log"
                }
              ]
            }
          ]
        },
        {
          "ruleSetType": "Microsoft_SQLInjectionRuleSet",
          "ruleSetVersion": "1.0"
        }
      ]
    },
    "policySettings": {
      "state": "Enabled",
      "mode": "Prevention"
    }
  }
}
```

**Complexity:** Low - Direct mapping via ruleGroupOverrides

**Mapping Notes:**
- `action: count` → `action: "Log"` (logs but doesn't block)
- `action: block` → `action: "Block"`
- `action: allow` → `action: "Allow"` or `state: "Disabled"`

---

### GCP Cloud Armor

**Status:** ⚠️ **Partial Support - Rule Set Level Only**

**How it works:**
- GCP Cloud Armor preconfigured WAF rules have limited granularity
- Cannot override individual rule IDs within a rule set
- Can only affect entire rule set via `preview: true` (equivalent to count mode)
- Uses sensitivity levels (0-4) to tune rule sets, not individual rules

**Best-Effort Translation:**
```json
{
  "name": "production-security-policy",
  "rules": [
    {
      "priority": 1000,
      "description": "OWASP ModSecurity CRS - Count mode for entire rule set",
      "action": "deny(403)",
      "preview": true,
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

**Complexity:** Medium - Degrades to rule-set level

**Degradation Strategy:**
- If ANY rule in a managed rule group has `action: count` override → entire rule set in `preview: true`
- Cannot selectively override individual rules
- Status should report: "GCP Cloud Armor does not support individual rule overrides; entire CoreRuleSet in preview mode"

---

## Cross-Provider Comparison

| Feature | AWS WAFv2 | Azure WAF | GCP Cloud Armor |
|---------|-----------|-----------|-----------------|
| **Individual rule override** | ✅ Yes | ✅ Yes | ❌ No |
| **Granularity** | Rule ID level | Rule ID level | Rule set level only |
| **Override actions** | Count, Allow, Block | Log, Allow, Block, Disabled | Preview (entire rule set) |
| **Complexity** | Low | Low | Medium |
| **Fidelity** | Perfect | Perfect | Degraded |

## Design Decision Impact

### API Design

✅ **Include `ruleOverrides` field** in portable API:

```yaml
spec:
  ruleOverrides:
    - managedRuleGroup: CoreRuleSet
      ruleId: "942100"
      action: count
      reason: "False positives on GraphQL queries"
```

**Rationale:**
- AWS and Azure support this perfectly (2 out of 3 providers)
- GCP can degrade gracefully to rule-set level
- Common use case: tuning managed rules for false positives

### Status Reporting

Report actual applied behavior in status:

```yaml
status:
  appliedRuleOverrides:
    - managedRuleGroup: CoreRuleSet
      ruleId: "942100"
      action: count
      appliedStrategy: "native"  # AWS, Azure
      
    # GCP would report:
    - managedRuleGroup: CoreRuleSet
      ruleId: "942100"
      action: count
      appliedStrategy: "degraded-ruleset-preview"
      message: "GCP Cloud Armor does not support individual rule overrides; entire CoreRuleSet in preview mode"
```

### Implementation Notes

**AWS KCP Reconciler**: Uses `RuleActionOverrides` within `ManagedRuleGroupStatement` to override individual rule actions.

**Azure KCP Reconciler**: Uses `ruleGroupOverrides` with `ManagedRuleOverride` to change specific rule behavior.

**GCP KCP Reconciler**: Falls back to rule-set level preview mode (no individual rule override support).

## Validation Matrix

| Test Case | AWS | Azure | GCP | Expected Behavior |
|-----------|-----|-------|-----|-------------------|
| Override single rule to count | ✅ Rule level | ✅ Rule level | ⚠️ Rule set preview | Rule 942100 → count (AWS/Azure), entire set → preview (GCP) |
| Override multiple rules to count | ✅ Multiple rules | ✅ Multiple rules | ⚠️ Rule set preview | All specified rules → count (AWS/Azure), entire set → preview (GCP) |
| Override rule to allow | ✅ Rule disabled | ✅ Rule disabled | ⚠️ Rule set preview | Rule bypassed (AWS/Azure), entire set → preview (GCP) |
| No overrides | ✅ Default action | ✅ Default action | ✅ Default action | All rules follow managedRuleGroup action |

## Conclusion

**Recommendation:** ✅ **Implement `ruleOverrides` field**

**Justification:**
1. **High portability:** AWS and Azure have perfect support (2/3 providers)
2. **Common use case:** Tuning managed rules for false positives is a real operational need
3. **Graceful degradation:** GCP can fall back to rule-set level preview mode
4. **Clear status reporting:** Users understand what was actually applied per provider

**Next Steps:**
1. Add `ruleOverrides` to WafPolicy CRD spec
2. Implement AWS reconciler with RuleActionOverrides
3. Implement Azure reconciler with ruleGroupOverrides
4. Implement GCP reconciler with preview mode fallback
5. Add status reporting for applied overrides and degradation warnings
