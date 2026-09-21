# WAF Implementation Examples - Use Case Translations

This document shows complete working examples for key WAF use cases, demonstrating how portable WafConfiguration translates to provider-specific WafPolicy JSON for AWS, Azure, and GCP.

## Table of Contents

1. [Use Case 1: Unconditional Rule Override](#use-case-1-unconditional-rule-override)
2. [Use Case 2: Custom Rule with Conditions](#use-case-2-custom-rule-with-conditions)
3. [Use Case 3: IP Allowlist/Blocklist](#use-case-3-ip-allowlistblocklist)
4. [Translation Patterns Reference](#translation-patterns-reference)

---

## Use Case 1: Unconditional Rule Override

**Scenario:** Enable OWASP protection in block mode globally, but change specific rules to count mode because they trigger false positives.

### Portable WafConfiguration

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafConfiguration
metadata:
  name: production-tuned
spec:
  # Start from base policy with managed rules
  basePolicyRef:
    name: owasp-moderate
  
  # Tune managed rules for false positives
  ruleOverrides:
    - managedRuleGroup: AWSManagedRulesCommonRuleSet  # AWS-specific (Azure: Microsoft_DefaultRuleSet)
      ruleId: "SizeRestrictions_BODY"  # AWS rule name (Azure: "942100")
      action: count
      reason: "False positives on legitimate file uploads"
    
    - managedRuleGroup: AWSManagedRulesCommonRuleSet
      ruleId: "GenericRFI_BODY"  # AWS rule name (Azure: "931130")
      action: count
      reason: "False positives on API endpoints"
```

### AWS WAFv2 Translation

```json
{
  "Name": "production-tuned",
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
        "SampledRequestsEnabled": false,
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
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "SQLiRuleSet"
      }
    }
  ]
}
```

### Azure WAF Translation

```json
{
  "name": "production-tuned",
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
                }
              ]
            },
            {
              "ruleGroupName": "PROTOCOL",
              "rules": [
                {
                  "ruleId": "931130",
                  "state": "Enabled",
                  "action": "Log"
                }
              ]
            }
          ]
        },
        {
          "ruleSetType": "Microsoft_BotManagerRuleSet",
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

### GCP Cloud Armor Translation

```json
{
  "name": "production-tuned",
  "rules": [
    {
      "priority": 1000,
      "description": "OWASP Core Rule Set (degraded to preview mode)",
      "action": "deny(403)",
      "preview": true,
      "match": {
        "expr": {
          "expression": "evaluatePreconfiguredWaf('owasp-crs-v030301-id', {'sensitivity': 1})"
        }
      }
    },
    {
      "priority": 1010,
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
      "description": "Default rule",
      "action": "allow",
      "match": {
        "versionedExpr": "SRC_IPS_V1",
        "config": {"srcIpRanges": ["*"]}
      }
    }
  ]
}
```

**Note:** GCP doesn't support individual rule overrides, so entire `owasp-crs` ruleset goes to `preview: true` mode (degraded).

---

## Use Case 2: Custom Rule with Conditions

**Scenario:** Add custom rules with conditions for health checks, admin bypasses, and blocking debug headers from external IPs.

### Portable WafConfiguration

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafConfiguration
metadata:
  name: production-custom-rules
spec:
  basePolicyRef:
    name: owasp-moderate
  
  # Custom rules (portable!)
  customRules:
    - name: health-check-bypass
      priority: 10
      action: allow
      conditions:
        path:
          exact: "/health"
      reason: "Health checks bypass all WAF rules"
    
    - name: admin-internal-bypass
      priority: 20
      action: allow
      conditions:
        path:
          prefix: "/admin"
        sourceIP:
          cidr: "10.0.0.0/8"
      reason: "Internal admin traffic bypasses rate limiting"
    
    - name: block-debug-external
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
      reason: "Debug header only allowed from internal network"
```

### AWS WAFv2 Translation

```json
{
  "Name": "production-custom-rules",
  "Scope": "REGIONAL",
  "DefaultAction": {"Allow": {}},
  "Rules": [
    {
      "Name": "HealthCheckBypass",
      "Priority": 10,
      "Statement": {
        "ByteMatchStatement": {
          "SearchString": "/health",
          "FieldToMatch": {"UriPath": {}},
          "TextTransformations": [{"Priority": 0, "Type": "NONE"}],
          "PositionalConstraint": "EXACTLY"
        }
      },
      "Action": {"Allow": {}},
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "HealthCheckBypass"
      }
    },
    {
      "Name": "AdminInternalBypass",
      "Priority": 20,
      "Statement": {
        "AndStatement": {
          "Statements": [
            {
              "ByteMatchStatement": {
                "SearchString": "/admin",
                "FieldToMatch": {"UriPath": {}},
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}],
                "PositionalConstraint": "STARTS_WITH"
              }
            },
            {
              "IPSetReferenceStatement": {
                "Arn": "arn:aws:wafv2:us-east-1:123456789012:regional/ipset/internal-10-0-0-0-8/a1b2c3d4"
              }
            }
          ]
        }
      },
      "Action": {"Allow": {}},
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "AdminInternalBypass"
      }
    },
    {
      "Name": "BlockDebugExternal",
      "Priority": 100,
      "Statement": {
        "AndStatement": {
          "Statements": [
            {
              "ByteMatchStatement": {
                "SearchString": "/admin",
                "FieldToMatch": {"UriPath": {}},
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}],
                "PositionalConstraint": "STARTS_WITH"
              }
            },
            {
              "SizeConstraintStatement": {
                "FieldToMatch": {"SingleHeader": {"Name": "x-debug"}},
                "ComparisonOperator": "GE",
                "Size": 1,
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            },
            {
              "NotStatement": {
                "Statement": {
                  "IPSetReferenceStatement": {
                    "Arn": "arn:aws:wafv2:us-east-1:123456789012:regional/ipset/internal-10-0-0-0-8/a1b2c3d4"
                  }
                }
              }
            }
          ]
        }
      },
      "Action": {"Block": {}},
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "BlockDebugExternal"
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
      "OverrideAction": {"None": {}},
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "CommonRuleSet"
      }
    }
  ]
}
```

### Azure WAF Translation

```json
{
  "name": "production-custom-rules",
  "properties": {
    "customRules": [
      {
        "name": "HealthCheckBypass",
        "priority": 10,
        "ruleType": "MatchRule",
        "action": "Allow",
        "matchConditions": [
          {
            "matchVariables": [{"variableName": "RequestUri"}],
            "operator": "Equals",
            "negationConditon": false,
            "matchValues": ["/health"],
            "transforms": []
          }
        ],
        "state": "Enabled"
      },
      {
        "name": "AdminInternalBypass",
        "priority": 20,
        "ruleType": "MatchRule",
        "action": "Allow",
        "matchConditions": [
          {
            "matchVariables": [{"variableName": "RequestUri"}],
            "operator": "BeginsWith",
            "negationConditon": false,
            "matchValues": ["/admin"],
            "transforms": []
          },
          {
            "matchVariables": [{"variableName": "RemoteAddr"}],
            "operator": "IPMatch",
            "negationConditon": false,
            "matchValues": ["10.0.0.0/8"],
            "transforms": []
          }
        ],
        "state": "Enabled"
      },
      {
        "name": "BlockDebugExternal",
        "priority": 100,
        "ruleType": "MatchRule",
        "action": "Block",
        "matchConditions": [
          {
            "matchVariables": [{"variableName": "RequestUri"}],
            "operator": "BeginsWith",
            "negationConditon": false,
            "matchValues": ["/admin"],
            "transforms": []
          },
          {
            "matchVariables": [{"variableName": "RequestHeaders", "selector": "X-Debug"}],
            "operator": "Exists",
            "negationConditon": false,
            "matchValues": [],
            "transforms": []
          },
          {
            "matchVariables": [{"variableName": "RemoteAddr"}],
            "operator": "IPMatch",
            "negationConditon": true,
            "matchValues": ["10.0.0.0/8"],
            "transforms": []
          }
        ],
        "state": "Enabled"
      }
    ],
    "managedRules": {
      "managedRuleSets": [
        {
          "ruleSetType": "Microsoft_DefaultRuleSet",
          "ruleSetVersion": "2.1"
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

### GCP Cloud Armor Translation

```json
{
  "name": "production-custom-rules",
  "rules": [
    {
      "priority": 10,
      "description": "Health checks bypass all WAF rules",
      "action": "allow",
      "preview": false,
      "match": {
        "expr": {
          "expression": "request.path == '/health'"
        }
      }
    },
    {
      "priority": 20,
      "description": "Internal admin traffic bypasses rate limiting",
      "action": "allow",
      "preview": false,
      "match": {
        "expr": {
          "expression": "request.path.matches('/admin.*') && inIpRange(origin.ip, '10.0.0.0/8')"
        }
      }
    },
    {
      "priority": 100,
      "description": "Debug header only allowed from internal network",
      "action": "deny(403)",
      "preview": false,
      "match": {
        "expr": {
          "expression": "request.path.matches('/admin.*') && has(request.headers['x-debug']) && !inIpRange(origin.ip, '10.0.0.0/8')"
        }
      }
    },
    {
      "priority": 1000,
      "description": "OWASP Core Rule Set",
      "action": "deny(403)",
      "preview": false,
      "match": {
        "expr": {
          "expression": "evaluatePreconfiguredWaf('owasp-crs-v030301-id', {'sensitivity': 1})"
        }
      }
    },
    {
      "priority": 2147483647,
      "description": "Default rule",
      "action": "allow",
      "match": {
        "versionedExpr": "SRC_IPS_V1",
        "config": {"srcIpRanges": ["*"]}
      }
    }
  ]
}
```

---

## Use Case 3: IP Allowlist/Blocklist

**Scenario:** Allow internal network and office VPN, block known bad actors, apply rate limits to everyone else.

### Portable WafConfiguration

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafConfiguration
metadata:
  name: ip-access-control
spec:
  basePolicyRef:
    name: owasp-moderate
  
  customRules:
    # Priority 1-99: Global allowlist (highest priority)
    - name: allow-internal-network
      priority: 10
      action: allow
      conditions:
        sourceIP:
          cidr: "10.0.0.0/8"
      reason: "Internal network always allowed"
    
    - name: allow-office-vpn
      priority: 20
      action: allow
      conditions:
        sourceIP:
          cidr: "203.0.113.0/24"
      reason: "Office VPN range"
    
    # Priority 100-199: Blocklist
    - name: block-bad-actor-1
      priority: 100
      action: block
      conditions:
        sourceIP:
          exact: "198.51.100.42"
      reason: "Known bad actor"
    
    - name: block-bad-actor-subnet
      priority: 110
      action: block
      conditions:
        sourceIP:
          cidr: "192.0.2.0/24"
      reason: "Malicious subnet"
```

### AWS WAFv2 Translation

```json
{
  "Name": "ip-access-control",
  "DefaultAction": {"Allow": {}},
  "Rules": [
    {
      "Name": "AllowInternalNetwork",
      "Priority": 10,
      "Statement": {
        "IPSetReferenceStatement": {
          "Arn": "arn:aws:wafv2:us-east-1:123456789012:regional/ipset/internal-10-0-0-0-8/abc123"
        }
      },
      "Action": {"Allow": {}},
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "AllowInternal"
      }
    },
    {
      "Name": "AllowOfficeVPN",
      "Priority": 20,
      "Statement": {
        "IPSetReferenceStatement": {
          "Arn": "arn:aws:wafv2:us-east-1:123456789012:regional/ipset/office-vpn-203-0-113-0-24/def456"
        }
      },
      "Action": {"Allow": {}},
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "AllowOfficeVPN"
      }
    },
    {
      "Name": "BlockBadActor1",
      "Priority": 100,
      "Statement": {
        "IPSetReferenceStatement": {
          "Arn": "arn:aws:wafv2:us-east-1:123456789012:regional/ipset/bad-actor-198-51-100-42/ghi789"
        }
      },
      "Action": {"Block": {}},
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "BlockBadActor1"
      }
    },
    {
      "Name": "BlockBadActorSubnet",
      "Priority": 110,
      "Statement": {
        "IPSetReferenceStatement": {
          "Arn": "arn:aws:wafv2:us-east-1:123456789012:regional/ipset/bad-subnet-192-0-2-0-24/jkl012"
        }
      },
      "Action": {"Block": {}},
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "BlockBadSubnet"
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
      "OverrideAction": {"None": {}},
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "CommonRuleSet"
      }
    }
  ]
}
```

**Note:** AWS requires pre-created IPSet resources for each IP/CIDR.

### Azure WAF Translation

```json
{
  "name": "ip-access-control",
  "properties": {
    "customRules": [
      {
        "name": "AllowInternalNetwork",
        "priority": 10,
        "ruleType": "MatchRule",
        "action": "Allow",
        "matchConditions": [
          {
            "matchVariables": [{"variableName": "RemoteAddr"}],
            "operator": "IPMatch",
            "negationConditon": false,
            "matchValues": ["10.0.0.0/8"],
            "transforms": []
          }
        ],
        "state": "Enabled"
      },
      {
        "name": "AllowOfficeVPN",
        "priority": 20,
        "ruleType": "MatchRule",
        "action": "Allow",
        "matchConditions": [
          {
            "matchVariables": [{"variableName": "RemoteAddr"}],
            "operator": "IPMatch",
            "negationConditon": false,
            "matchValues": ["203.0.113.0/24"],
            "transforms": []
          }
        ],
        "state": "Enabled"
      },
      {
        "name": "BlockBadActor1",
        "priority": 100,
        "ruleType": "MatchRule",
        "action": "Block",
        "matchConditions": [
          {
            "matchVariables": [{"variableName": "RemoteAddr"}],
            "operator": "IPMatch",
            "negationConditon": false,
            "matchValues": ["198.51.100.42/32"],
            "transforms": []
          }
        ],
        "state": "Enabled"
      },
      {
        "name": "BlockBadActorSubnet",
        "priority": 110,
        "ruleType": "MatchRule",
        "action": "Block",
        "matchConditions": [
          {
            "matchVariables": [{"variableName": "RemoteAddr"}],
            "operator": "IPMatch",
            "negationConditon": false,
            "matchValues": ["192.0.2.0/24"],
            "transforms": []
          }
        ],
        "state": "Enabled"
      }
    ],
    "managedRules": {
      "managedRuleSets": [
        {
          "ruleSetType": "Microsoft_DefaultRuleSet",
          "ruleSetVersion": "2.1"
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

**Note:** Azure supports IP CIDRs directly in matchValues.

### GCP Cloud Armor Translation

```json
{
  "name": "ip-access-control",
  "rules": [
    {
      "priority": 10,
      "description": "Internal network always allowed",
      "action": "allow",
      "preview": false,
      "match": {
        "expr": {
          "expression": "inIpRange(origin.ip, '10.0.0.0/8')"
        }
      }
    },
    {
      "priority": 20,
      "description": "Office VPN range",
      "action": "allow",
      "preview": false,
      "match": {
        "expr": {
          "expression": "inIpRange(origin.ip, '203.0.113.0/24')"
        }
      }
    },
    {
      "priority": 100,
      "description": "Known bad actor",
      "action": "deny(403)",
      "preview": false,
      "match": {
        "expr": {
          "expression": "origin.ip == '198.51.100.42'"
        }
      }
    },
    {
      "priority": 110,
      "description": "Malicious subnet",
      "action": "deny(403)",
      "preview": false,
      "match": {
        "expr": {
          "expression": "inIpRange(origin.ip, '192.0.2.0/24')"
        }
      }
    },
    {
      "priority": 1000,
      "description": "OWASP Core Rule Set",
      "action": "deny(403)",
      "preview": false,
      "match": {
        "expr": {
          "expression": "evaluatePreconfiguredWaf('owasp-crs-v030301-id', {'sensitivity': 1})"
        }
      }
    },
    {
      "priority": 2147483647,
      "description": "Default rule",
      "action": "allow",
      "match": {
        "versionedExpr": "SRC_IPS_V1",
        "config": {"srcIpRanges": ["*"]}
      }
    }
  ]
}
```

**Note:** GCP uses CEL expressions `inIpRange()` for CIDRs and `origin.ip ==` for exact IPs.

---

## Translation Patterns Reference

### Condition Types

| Portable Condition | AWS | Azure | GCP |
|-------------------|-----|-------|-----|
| **Path exact** | `ByteMatchStatement EXACTLY` | `Equals` | `request.path == '/health'` |
| **Path prefix** | `ByteMatchStatement STARTS_WITH` | `BeginsWith` | `request.path.matches('/admin.*')` |
| **Path regex** | `RegexMatchStatement` | `Regex` | `request.path.matches('regex')` |
| **Header exists** | `SizeConstraintStatement >= 1` | `Exists` | `has(request.headers['name'])` |
| **Header value** | `ByteMatchStatement EXACTLY` | `Equals` | `request.headers['name'] == 'value'` |
| **IP CIDR** | `IPSetReferenceStatement` | `IPMatch` | `inIpRange(origin.ip, 'cidr')` |
| **IP exact** | `IPSetReferenceStatement` | `IPMatch` with /32 | `origin.ip == 'x.x.x.x'` |
| **Negate** | `NotStatement` | `negationConditon: true` | `!condition` |
| **AND logic** | `AndStatement` | Array of `matchConditions` | `&&` in CEL |

### Rule Override Mechanisms

| Provider | Mechanism | Granularity | Example |
|----------|-----------|-------------|---------|
| **AWS** | `RuleActionOverrides` | Per rule ID | `{"Name": "SizeRestrictions_BODY", "ActionToUse": {"Count": {}}}` |
| **Azure** | `ruleGroupOverrides` | Per rule ID | `{"ruleId": "942100", "action": "Log"}` |
| **GCP** | `preview: true` | Entire ruleset | `"preview": true` on preconfigured WAF rule |

---

## Summary

This document provides **complete, working examples** for the three main Phase 2 use cases:

1. **Unconditional Rule Override** - Tune managed rules for false positives
2. **Custom Rules with Conditions** - Add portable rules for health checks, admin bypasses, and conditional blocking
3. **IP Allowlist/Blocklist** - IP-based access control

All examples show:
- ✅ Portable WafConfiguration input (what users write)
- ✅ AWS WAFv2 translation (complete JSON)
- ✅ Azure Application Gateway WAF translation (complete JSON)
- ✅ GCP Cloud Armor translation (complete JSON)
- ✅ Provider-specific notes and limitations
