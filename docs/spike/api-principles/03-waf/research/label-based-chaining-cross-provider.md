# Label-Based Chaining Support Across Providers

This document analyzes label-based rule chaining support across AWS WAFv2, Azure Application Gateway WAF, and GCP Cloud Armor for WafPolicy conditional overrides.

## What is Label-Based Chaining?

Label-based chaining is an advanced WAF pattern where:
1. Rules add **labels** to requests that match certain conditions
2. Subsequent rules match **based on those labels** rather than re-evaluating complex conditions
3. This reduces nesting depth, improves readability, and enables reusable logic

### Benefits:
- **Reduced complexity**: Avoids deeply nested `AndStatement`/`OrStatement` structures
- **Better performance**: Conditions evaluated once, label checked multiple times
- **Reusability**: Multiple rules can reference the same label
- **Clarity**: Intent expressed through semantic label names

### Example Pattern:
```
Rule 1: IF request matches conditions → ADD label "trusted-source"
Rule 2: IF label "trusted-source" exists → ALLOW (bypass rate limiting)
Rule 3: IF label "trusted-source" exists → ALLOW (bypass SQLi checks)
```

---

## Provider Support Matrix

| Feature | AWS WAFv2 | Azure Application Gateway WAF | GCP Cloud Armor |
|---------|-----------|-------------------------------|-----------------|
| **Add Labels** | ✅ Yes (`LabelMatchStatement`) | ❌ No | ❌ No |
| **Match on Labels** | ✅ Yes (`Label` namespace) | ❌ No | ❌ No |
| **Label Propagation** | ✅ Yes (within WebACL) | ❌ N/A | ❌ N/A |
| **Label Namespaces** | ✅ Yes (prefix-based) | ❌ N/A | ❌ N/A |
| **Multiple Labels** | ✅ Yes (unlimited) | ❌ N/A | ❌ N/A |
| **Label-Based Actions** | ✅ Yes (Allow/Block/Count) | ❌ N/A | ❌ N/A |
| **Workaround Available** | ✅ N/A | ⚠️ Custom rules | ⚠️ Priority ordering |

**Legend:**
- ✅ **Native support** - First-class feature
- ⚠️ **Workaround** - Can be simulated with other features
- ❌ **No support** - Feature not available

---

## AWS WAFv2 - Full Native Support

AWS WAFv2 has **comprehensive label-based chaining support** introduced specifically to reduce nesting depth and improve rule organization.

### Key Features:

1. **LabelMatchStatement**: Add labels to requests in rules with `Count` or `Captcha` override action
2. **Label matching**: Subsequent rules can match on presence of labels
3. **Label namespaces**: Organize labels with prefixes (e.g., `awswaf:managed:aws:bot-control:`)
4. **Label scope**: Labels persist for the duration of the WebACL evaluation

### AWS Syntax:

```json
{
  "Rules": [
    {
      "Name": "AddTrustedSourceLabel",
      "Priority": 10,
      "Statement": {
        "AndStatement": {
          "Statements": [
            {
              "ByteMatchStatement": {
                "SearchString": "/admin",
                "FieldToMatch": {"UriPath": {}},
                "PositionalConstraint": "STARTS_WITH"
              }
            },
            {
              "IPSetReferenceStatement": {
                "Arn": "arn:aws:wafv2:region:account:regional/ipset/internal-network/..."
              }
            },
            {
              "SizeConstraintStatement": {
                "FieldToMatch": {"SingleHeader": {"Name": "authorization"}},
                "ComparisonOperator": "GE",
                "Size": 1
              }
            }
          ]
        }
      },
      "Action": {
        "Count": {}
      },
      "RuleLabels": [
        {
          "Name": "myapp:trusted-source"
        },
        {
          "Name": "myapp:skip-rate-limit"
        },
        {
          "Name": "myapp:skip-sqli-check"
        }
      ],
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "TrustedSourceLabeling"
      }
    },
    {
      "Name": "TrustedSourceBypassRateLimit",
      "Priority": 900,
      "Statement": {
        "NotStatement": {
          "Statement": {
            "LabelMatchStatement": {
              "Scope": "LABEL",
              "Key": "myapp:skip-rate-limit"
            }
          }
        }
      },
      "Action": {
        "Block": {}
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "RateLimitWithBypass"
      }
    },
    {
      "Name": "SQLiRuleSetWithBypass",
      "Priority": 1010,
      "Statement": {
        "ManagedRuleGroupStatement": {
          "VendorName": "AWS",
          "Name": "AWSManagedRulesSQLiRuleSet",
          "ScopeDownStatement": {
            "NotStatement": {
              "Statement": {
                "LabelMatchStatement": {
                  "Scope": "LABEL",
                  "Key": "myapp:skip-sqli-check"
                }
              }
            }
          }
        }
      },
      "OverrideAction": {"None": {}},
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "SQLiWithBypass"
      }
    },
    {
      "Name": "CoreRuleSetWithBypass",
      "Priority": 1020,
      "Statement": {
        "ManagedRuleGroupStatement": {
          "VendorName": "AWS",
          "Name": "AWSManagedRulesCommonRuleSet",
          "ScopeDownStatement": {
            "NotStatement": {
              "Statement": {
                "LabelMatchStatement": {
                  "Scope": "LABEL",
                  "Key": "myapp:trusted-source"
                }
              }
            }
          }
        }
      },
      "OverrideAction": {"None": {}},
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "CoreRuleSetWithBypass"
      }
    }
  ]
}
```

### AWS Label Namespaces:

Labels can use namespaces for organization:
- `awswaf:managed:aws:*` - Reserved for AWS-managed rules
- `myapp:*` - Custom application labels
- `security:*` - Security-related labels
- `bypass:*` - Bypass/exemption labels

### AWS Label Best Practices:

1. **Evaluate conditions once**: Add label with `Count` action (doesn't block/allow)
2. **Use semantic names**: `trusted-source`, `authenticated-user`, `internal-network`
3. **Namespace your labels**: Prevent conflicts with managed rule labels
4. **Match negatively**: Use `NotStatement` around `LabelMatchStatement` to apply rules when label is absent
5. **Priority ordering**: Label-adding rules must have lower priority numbers than label-matching rules

---

## Azure Application Gateway WAF - No Native Support

Azure Application Gateway WAF **does not support label-based chaining**. However, similar patterns can be achieved through workarounds.

### Why Not Supported:

Azure WAF uses a simpler model:
1. Custom rules (priority 1-100) evaluated sequentially
2. Managed rules (priority > 100) evaluated after custom rules
3. First matching action terminates evaluation

There is no concept of:
- Adding metadata/labels to requests during evaluation
- Referencing previous rule results in subsequent rules

### Workaround: Duplicate Conditions in Custom Rules

Instead of labels, you must **duplicate the condition logic** in each custom rule that needs it.

```json
{
  "customRules": [
    {
      "name": "TrustedSourceBypassRateLimit",
      "priority": 10,
      "ruleType": "RateLimitRule",
      "rateLimitThreshold": 10000,
      "action": "Block",
      "matchConditions": [
        {
          "matchVariables": [{"variableName": "RequestUri"}],
          "operator": "BeginsWith",
          "negationConditon": true,
          "matchValues": ["/admin"]
        },
        {
          "matchVariables": [{"variableName": "RemoteAddr"}],
          "operator": "IPMatch",
          "negationConditon": true,
          "matchValues": ["10.0.0.0/8"]
        },
        {
          "matchVariables": [{"variableName": "RequestHeaders", "selector": "Authorization"}],
          "operator": "Exists",
          "negationConditon": true
        }
      ],
      "state": "Enabled"
    },
    {
      "name": "TrustedSourceBypassManagedRules",
      "priority": 20,
      "ruleType": "MatchRule",
      "action": "Allow",
      "matchConditions": [
        {
          "matchVariables": [{"variableName": "RequestUri"}],
          "operator": "BeginsWith",
          "matchValues": ["/admin"]
        },
        {
          "matchVariables": [{"variableName": "RemoteAddr"}],
          "operator": "IPMatch",
          "matchValues": ["10.0.0.0/8"]
        },
        {
          "matchVariables": [{"variableName": "RequestHeaders", "selector": "Authorization"}],
          "operator": "Exists"
        }
      ],
      "state": "Enabled"
    }
  ]
}
```

### Azure Workaround Limitations:

❌ **Code duplication**: Same condition logic repeated in multiple rules  
❌ **Performance**: Conditions evaluated multiple times  
❌ **Maintainability**: Changes require updating all rules  
❌ **No reusability**: Cannot reference "trusted source" concept across rules  

### Azure Alternative: IP Allow Lists

For IP-based trust, Azure supports a simpler pattern:
- Use custom allow rule with high priority for trusted IPs
- Terminates evaluation before managed rules run

---

## GCP Cloud Armor - No Native Support (But CEL Helps)

GCP Cloud Armor **does not support labels**, but CEL expressions provide some flexibility.

### Why Not Supported:

GCP Cloud Armor evaluates rules independently:
1. Rules evaluated in ascending priority order
2. First matching rule action is applied
3. No state carried between rule evaluations
4. No concept of request-scoped metadata

### Workaround: CEL Expression Reuse (Limited)

While you can't add labels, CEL expressions can be composed to avoid some duplication:

```json
{
  "rules": [
    {
      "priority": 10,
      "description": "Trusted source bypass - rate limiting",
      "action": "allow",
      "match": {
        "expr": {
          "expression": "request.path.matches('/admin.*') && inIpRange(origin.ip, '10.0.0.0/8') && has(request.headers['authorization'])"
        }
      }
    },
    {
      "priority": 1000,
      "description": "SQL injection protection with bypass",
      "action": "deny(403)",
      "match": {
        "expr": {
          "expression": "!(request.path.matches('/admin.*') && inIpRange(origin.ip, '10.0.0.0/8') && has(request.headers['authorization'])) && evaluatePreconfiguredWaf('sqli-v33-stable')"
        }
      }
    },
    {
      "priority": 1010,
      "description": "XSS protection with bypass",
      "action": "deny(403)",
      "match": {
        "expr": {
          "expression": "!(request.path.matches('/admin.*') && inIpRange(origin.ip, '10.0.0.0/8') && has(request.headers['authorization'])) && evaluatePreconfiguredWaf('xss-v33-stable')"
        }
      }
    },
    {
      "priority": 2000,
      "description": "Rate limiting with bypass",
      "action": "rate_based_ban",
      "match": {
        "expr": {
          "expression": "!(request.path.matches('/admin.*') && inIpRange(origin.ip, '10.0.0.0/8') && has(request.headers['authorization']))"
        }
      },
      "rateLimitOptions": {
        "conformAction": "allow",
        "exceedAction": "deny(429)",
        "enforceOnKey": "IP",
        "rateLimitThreshold": {
          "count": 100,
          "intervalSec": 60
        }
      }
    }
  ]
}
```

### GCP Workaround Limitations:

❌ **Expression duplication**: Complex conditions repeated in multiple rules  
❌ **Performance**: Same condition evaluated multiple times  
❌ **Expression length limits**: CEL has max expression length  
❌ **No semantic names**: Can't reference "trusted-source" concept by name  

### GCP Alternative: Priority-Based Bypasses

The cleaner GCP pattern is priority-based allow rules:

```json
{
  "rules": [
    {
      "priority": 10,
      "description": "Trusted sources - bypass all security checks",
      "action": "allow",
      "match": {
        "expr": {
          "expression": "request.path.matches('/admin.*') && inIpRange(origin.ip, '10.0.0.0/8') && has(request.headers['authorization'])"
        }
      }
    },
    {
      "priority": 1000,
      "description": "SQL injection protection",
      "action": "deny(403)",
      "match": {
        "expr": {
          "expression": "evaluatePreconfiguredWaf('sqli-v33-stable')"
        }
      }
    }
  ]
}
```

This is simpler but **terminates all evaluation** for trusted sources (less flexible).

---

## Portable WafPolicy API Design for Label Chaining

Given that only AWS supports native label chaining, the portable API should:

1. **Option 1: AWS-Only Feature** - Add label chaining as AWS-specific capability
2. **Option 2: Abstract Pattern** - Design portable abstraction that degrades gracefully
3. **Option 3: Skip Labels** - Don't expose labels in portable API

### Recommendation: Option 2 - Abstract Pattern

Introduce a **portable "classification" concept** that maps to:
- AWS: Native labels
- Azure: Duplicated conditions (Cloud Manager generates)
- GCP: Expression composition or priority bypass

### Proposed Portable API:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafPolicy
metadata:
  name: label-chaining-example
spec:
  managedRuleGroups:
    - type: CoreRuleSet
      action: block
    - type: SQLInjectionProtection
      action: block
    - type: RateLimit
      action: block
      requestsPerMinute: 100
  
  # Define classifications that can be reused
  classifications:
    - name: trusted-source
      conditions:
        path:
          prefix: "/admin"
        sourceIP:
          cidr: "10.0.0.0/8"
        header:
          name: "Authorization"
          exists: true
      description: "Authenticated admin from internal network"
    
    - name: api-consumer
      conditions:
        path:
          prefix: "/api"
        header:
          name: "X-API-Key"
          exists: true
      description: "API request with valid key"
  
  # Reference classifications in overrides
  conditionalOverrides:
    # When classification matches, apply overrides
    - classification: trusted-source
      overrides:
        - managedRuleGroup: CoreRuleSet
          ruleId: "*"
          action: allow
        - managedRuleGroup: SQLInjectionProtection
          ruleId: "*"
          action: allow
        - managedRuleGroup: RateLimit
          action: allow
      reason: "Trusted sources bypass all checks"
    
    - classification: api-consumer
      overrides:
        - managedRuleGroup: RateLimit
          action: block
          requestsPerMinute: 1000
      reason: "API consumers get higher rate limit"
```

### Translation Strategy:

**AWS WAFv2:**
```json
{
  "Rules": [
    {
      "Name": "ClassifyTrustedSource",
      "Priority": 10,
      "Statement": { /* conditions from classification */ },
      "Action": {"Count": {}},
      "RuleLabels": [{"Name": "myapp:trusted-source"}]
    },
    {
      "Name": "CoreRuleSet",
      "Priority": 1000,
      "Statement": {
        "ManagedRuleGroupStatement": {
          "VendorName": "AWS",
          "Name": "AWSManagedRulesCommonRuleSet",
          "ScopeDownStatement": {
            "NotStatement": {
              "Statement": {
                "LabelMatchStatement": {
                  "Scope": "LABEL",
                  "Key": "myapp:trusted-source"
                }
              }
            }
          }
        }
      }
    }
  ]
}
```

**Azure WAF:**
```json
{
  "customRules": [
    {
      "name": "TrustedSourceBypass",
      "priority": 10,
      "action": "Allow",
      "matchConditions": [ /* conditions duplicated */ ]
    }
  ]
}
```

**GCP Cloud Armor:**
```json
{
  "rules": [
    {
      "priority": 10,
      "action": "allow",
      "match": {
        "expr": {
          "expression": "request.path.matches('/admin.*') && inIpRange(origin.ip, '10.0.0.0/8') && has(request.headers['authorization'])"
        }
      }
    }
  ]
}
```

---

## Benefits of Classification-Based Approach

### For Users:
✅ **DRY principle**: Define condition once, reference multiple times  
✅ **Semantic naming**: "trusted-source" is clearer than repeated conditions  
✅ **Easier maintenance**: Update classification definition in one place  
✅ **Provider agnostic**: Same YAML works across providers  

### For Cloud Manager:
✅ **AWS optimization**: Leverage native labels for best performance  
✅ **Azure compatibility**: Generate duplicated conditions automatically  
✅ **GCP compatibility**: Generate priority-based or expression-based rules  
✅ **Clear translation**: Explicit mapping from portable to provider-specific  

---

## Implementation Complexity

| Provider | Implementation Effort | Notes |
|----------|----------------------|-------|
| **AWS WAFv2** | 🟢 Low | Direct mapping to native labels |
| **Azure WAF** | 🟡 Medium | Must duplicate conditions in multiple custom rules |
| **GCP Cloud Armor** | 🟡 Medium | Must compose CEL expressions or use priority bypass |

---

## Example: Full Portable Policy with Classifications

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafPolicy
metadata:
  name: comprehensive-with-classifications
spec:
  managedRuleGroups:
    - type: CoreRuleSet
      action: block
    - type: SQLInjectionProtection
      action: block
    - type: CrossSiteScripting
      action: block
    - type: RateLimit
      action: block
      requestsPerMinute: 100
  
  # Reusable classifications
  classifications:
    - name: internal-network
      conditions:
        sourceIP:
          cidr: "10.0.0.0/8"
      description: "Requests from internal corporate network"
    
    - name: authenticated-admin
      conditions:
        path:
          prefix: "/admin"
        header:
          name: "Authorization"
          regex: "^Bearer [A-Za-z0-9\\-._~+/]+=*$"
      description: "Authenticated admin with valid JWT"
    
    - name: trusted-source
      conditions:
        # Combines multiple classifications
        anyOf:
          - classification: internal-network
          - classification: authenticated-admin
      description: "Either internal network OR authenticated admin"
    
    - name: api-key-holder
      conditions:
        path:
          prefix: "/api"
        header:
          name: "X-API-Key"
          exists: true
      description: "API request with valid key header"
    
    - name: webhook-callback
      conditions:
        path:
          prefix: "/webhooks/"
        method: POST
        header:
          name: "X-Webhook-Signature"
          exists: true
      description: "Authenticated webhook POST"
  
  # Use classifications in overrides
  conditionalOverrides:
    - classification: trusted-source
      overrides:
        - managedRuleGroup: CoreRuleSet
          ruleId: "*"
          action: count
        - managedRuleGroup: SQLInjectionProtection
          ruleId: "*"
          action: count
        - managedRuleGroup: RateLimit
          action: allow
      reason: "Trusted sources get detection mode + no rate limiting"
    
    - classification: api-key-holder
      overrides:
        - managedRuleGroup: RateLimit
          action: block
          requestsPerMinute: 1000
      reason: "API key holders get 10x higher rate limit"
    
    - classification: webhook-callback
      overrides:
        - managedRuleGroup: RateLimit
          action: allow
        - managedRuleGroup: CrossSiteScripting
          ruleId: "*"
          action: count
      reason: "Webhooks bypass rate limit, XSS in detection mode"

status:
  appliedClassifications:
    - name: trusted-source
      ruleCount: 3
      affectedManagedRuleGroups: ["CoreRuleSet", "SQLInjectionProtection", "RateLimit"]
      implementationStrategy: "aws-native-labels"  # or "azure-custom-rules", "gcp-priority-bypass"
    
    - name: api-key-holder
      ruleCount: 1
      affectedManagedRuleGroups: ["RateLimit"]
      implementationStrategy: "aws-native-labels"
```

---

## Recommendation Summary

### For Cloud Manager WafPolicy:

1. ✅ **Implement classifications API** - Provides portable abstraction over label chaining
2. ✅ **AWS: Use native labels** - Best performance and AWS-recommended pattern
3. ✅ **Azure: Generate duplicated custom rules** - Acceptable workaround
4. ✅ **GCP: Use priority-based allow rules** - Simpler than expression duplication
5. ✅ **Document strategy per provider** - Make translation approach visible in status

### Benefits:

- Users write DRY, maintainable policies
- AWS gets optimal implementation
- Azure/GCP get functional (if less optimal) implementation
- Provider differences abstracted away
- Future providers can add native support without breaking changes

### Trade-offs:

- Azure: Generates more custom rules (hits 100-rule limit faster)
- GCP: Priority-based bypass means all-or-nothing (less granular than AWS)
- Complexity: Cloud Manager translation logic is more complex

---

## Conclusion

**Label-based chaining is AWS-specific**, but the pattern can be abstracted into a portable "classifications" API that:
- Maps to native AWS labels (optimal)
- Generates workarounds for Azure (duplicated conditions)
- Uses priority bypass for GCP (simpler pattern)

This allows users to express the **intent** (reusable conditions) in a provider-agnostic way while Cloud Manager handles provider-specific translation.
