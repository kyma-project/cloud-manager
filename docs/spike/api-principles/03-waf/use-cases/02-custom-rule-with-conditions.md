# Use Case 3: Custom Rule with Path and Header Conditions

## User Story

"As a platform operator, I want to block requests to /admin paths that contain a specific suspicious header value (e.g., X-Debug: true from external networks), because this indicates an attempted exploit of a debug endpoint that should only be accessible internally."

## Requirements

1. Define custom blocking rule (not managed rule override)
2. Apply rule only to specific path prefix (e.g., /admin)
3. Match on header name and value
4. Optionally combine with source IP conditions
5. Block action when conditions match

## Real-World Threat Scenarios

### Scenario A: Debug Header Exploitation
```yaml
# Block X-Debug: true on /admin from public internet
# Legitimate: Internal tools use X-Debug on /admin from 10.0.0.0/8
# Threat: Attackers try X-Debug: true on /admin from public IPs
condition:
  path: {prefix: "/admin"}
  header:
    name: "X-Debug"
    value: "true"
  sourceIP:
    cidr: "0.0.0.0/0"
    negate: true  # NOT internal network
action: block
```

### Scenario B: Admin Panel Credential Stuffing
```yaml
# Block requests to /admin/login with X-Forwarded-For header
# Threat: Attackers use X-Forwarded-For to bypass rate limits
condition:
  path: {exact: "/admin/login"}
  header:
    name: "X-Forwarded-For"
    exists: true
  method: POST
action: block
```

### Scenario C: Scanner Detection
```yaml
# Block known scanner User-Agent on admin paths
# Threat: Automated scanners probing admin interfaces
condition:
  path: {prefix: "/admin"}
  header:
    name: "User-Agent"
    contains: "sqlmap"
action: block
```

## Portable API Design

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafPolicy
metadata:
  name: admin-protection-policy
spec:
  # Enable managed rules globally
  managedRuleGroups:
    - type: CoreRuleSet
      action: block
    - type: SQLInjectionProtection
      action: block
  
  # Custom rules for specific threats
  customRules:
    - name: block-debug-header-on-admin
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
          negate: true  # Block if NOT from internal network
      reason: "Debug header on admin paths only allowed from internal network"
    
    - name: block-forwarded-header-on-admin-login
      priority: 110
      action: block
      conditions:
        path:
          exact: "/admin/login"
        method: POST
        header:
          name: "X-Forwarded-For"
          exists: true
      reason: "Prevent rate limit bypass via X-Forwarded-For on admin login"
    
    - name: block-scanner-user-agent
      priority: 120
      action: block
      conditions:
        path:
          prefix: "/admin"
        header:
          name: "User-Agent"
          anyOf:
            - contains: "sqlmap"
            - contains: "nikto"
            - contains: "nmap"
      reason: "Block known security scanners on admin paths"
    
    - name: block-suspicious-content-type
      priority: 130
      action: block
      conditions:
        path:
          prefix: "/admin"
        method: POST
        header:
          name: "Content-Type"
          value: "application/x-www-form-urlencoded"
          negate: true  # Block if NOT standard form encoding
        header:
          name: "Content-Type"
          value: "application/json"
          negate: true  # AND NOT JSON
      reason: "Only allow standard content types on admin POST"
```

## Provider Capability Check

### AWS WAFv2

**Status:** ✅ **Perfect Support**

**How it works:**
- Custom rules use `Action` (not `OverrideAction` like managed rules)
- Multiple conditions combined with `AndStatement`
- Supports all condition types (path, header, method, IP)
- `NotStatement` for negation

**Native AWS Translation:**
```json
{
  "Name": "admin-protection-webacl",
  "Scope": "REGIONAL",
  "DefaultAction": {"Allow": {}},
  "Rules": [
    {
      "Name": "BlockDebugHeaderOnAdmin",
      "Priority": 100,
      "Action": {"Block": {}},
      "Statement": {
        "AndStatement": {
          "Statements": [
            {
              "ByteMatchStatement": {
                "SearchString": "/admin",
                "FieldToMatch": {"UriPath": {}},
                "PositionalConstraint": "STARTS_WITH",
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            },
            {
              "ByteMatchStatement": {
                "SearchString": "X-Debug",
                "FieldToMatch": {
                  "SingleHeader": {"Name": "x-debug"}
                },
                "PositionalConstraint": "EXACTLY",
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            },
            {
              "NotStatement": {
                "Statement": {
                  "IPSetReferenceStatement": {
                    "Arn": "arn:aws:wafv2:region:account:regional/ipset/internal-network/..."
                  }
                }
              }
            }
          ]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "BlockDebugHeaderOnAdmin"
      }
    },
    {
      "Name": "BlockForwardedHeaderOnAdminLogin",
      "Priority": 110,
      "Action": {"Block": {}},
      "Statement": {
        "AndStatement": {
          "Statements": [
            {
              "ByteMatchStatement": {
                "SearchString": "/admin/login",
                "FieldToMatch": {"UriPath": {}},
                "PositionalConstraint": "EXACTLY",
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            },
            {
              "ByteMatchStatement": {
                "SearchString": "POST",
                "FieldToMatch": {"Method": {}},
                "PositionalConstraint": "EXACTLY",
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            },
            {
              "ByteMatchStatement": {
                "SearchString": "X-Forwarded-For",
                "FieldToMatch": {
                  "SingleHeader": {"Name": "x-forwarded-for"}
                },
                "PositionalConstraint": "EXACTLY",
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            }
          ]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "BlockForwardedHeaderOnAdminLogin"
      }
    },
    {
      "Name": "BlockScannerUserAgent",
      "Priority": 120,
      "Action": {"Block": {}},
      "Statement": {
        "AndStatement": {
          "Statements": [
            {
              "ByteMatchStatement": {
                "SearchString": "/admin",
                "FieldToMatch": {"UriPath": {}},
                "PositionalConstraint": "STARTS_WITH",
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            },
            {
              "OrStatement": {
                "Statements": [
                  {
                    "ByteMatchStatement": {
                      "SearchString": "sqlmap",
                      "FieldToMatch": {
                        "SingleHeader": {"Name": "user-agent"}
                      },
                      "PositionalConstraint": "CONTAINS",
                      "TextTransformations": [{"Priority": 0, "Type": "LOWERCASE"}]
                    }
                  },
                  {
                    "ByteMatchStatement": {
                      "SearchString": "nikto",
                      "FieldToMatch": {
                        "SingleHeader": {"Name": "user-agent"}
                      },
                      "PositionalConstraint": "CONTAINS",
                      "TextTransformations": [{"Priority": 0, "Type": "LOWERCASE"}]
                    }
                  },
                  {
                    "ByteMatchStatement": {
                      "SearchString": "nmap",
                      "FieldToMatch": {
                        "SingleHeader": {"Name": "user-agent"}
                      },
                      "PositionalConstraint": "CONTAINS",
                      "TextTransformations": [{"Priority": 0, "Type": "LOWERCASE"}]
                    }
                  }
                ]
              }
            }
          ]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "BlockScannerUserAgent"
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
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "CommonRuleSet"
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
- Custom rules defined in `customRules` array
- Each custom rule has `matchConditions` (all conditions are AND)
- Supports path, method, header, IP conditions
- `negationConditon: true` for negation

**Native Azure Translation:**
```json
{
  "location": "eastus",
  "properties": {
    "customRules": [
      {
        "name": "BlockDebugHeaderOnAdmin",
        "priority": 100,
        "ruleType": "MatchRule",
        "action": "Block",
        "matchConditions": [
          {
            "matchVariables": [
              {"variableName": "RequestUri"}
            ],
            "operator": "BeginsWith",
            "matchValues": ["/admin"],
            "negationConditon": false
          },
          {
            "matchVariables": [
              {"variableName": "RequestHeaders"}
            ],
            "selector": "X-Debug",
            "operator": "Contains",
            "matchValues": [""],
            "negationConditon": false
          },
          {
            "matchVariables": [
              {"variableName": "RemoteAddr"}
            ],
            "operator": "IPMatch",
            "matchValues": ["10.0.0.0/8"],
            "negationConditon": true
          }
        ]
      },
      {
        "name": "BlockForwardedHeaderOnAdminLogin",
        "priority": 110,
        "ruleType": "MatchRule",
        "action": "Block",
        "matchConditions": [
          {
            "matchVariables": [
              {"variableName": "RequestUri"}
            ],
            "operator": "Equal",
            "matchValues": ["/admin/login"],
            "negationConditon": false
          },
          {
            "matchVariables": [
              {"variableName": "RequestMethod"}
            ],
            "operator": "Equal",
            "matchValues": ["POST"],
            "negationConditon": false
          },
          {
            "matchVariables": [
              {"variableName": "RequestHeaders"}
            ],
            "selector": "X-Forwarded-For",
            "operator": "Contains",
            "matchValues": [""],
            "negationConditon": false
          }
        ]
      },
      {
        "name": "BlockScannerUserAgent",
        "priority": 120,
        "ruleType": "MatchRule",
        "action": "Block",
        "matchConditions": [
          {
            "matchVariables": [
              {"variableName": "RequestUri"}
            ],
            "operator": "BeginsWith",
            "matchValues": ["/admin"],
            "negationConditon": false
          },
          {
            "matchVariables": [
              {"variableName": "RequestHeaders"}
            ],
            "selector": "User-Agent",
            "operator": "Contains",
            "matchValues": ["sqlmap", "nikto", "nmap"],
            "negationConditon": false
          }
        ]
      }
    ],
    "managedRules": {
      "managedRuleSets": [
        {
          "ruleSetType": "Microsoft_DefaultRuleSet",
          "ruleSetVersion": "2.1"
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

**Complexity:** Low - Direct mapping to customRules

**Azure Quirk:**
- Header exists check: Use `operator: "Contains"` with `matchValues: [""]`
- Multiple match values in single condition treated as OR (good for User-Agent list)

---

### GCP Cloud Armor

**Status:** ✅ **Perfect Support**

**How it works:**
- Custom rules use CEL (Common Expression Language) expressions
- CEL supports rich boolean logic: `&&`, `||`, `!`, `has()`, `matches()`, `contains()`
- Very flexible and expressive
- Lower priority numbers evaluated first

**Native GCP Translation:**
```json
{
  "name": "admin-protection-security-policy",
  "rules": [
    {
      "priority": 100,
      "description": "Block debug header on admin paths from external IPs",
      "action": "deny(403)",
      "preview": false,
      "match": {
        "expr": {
          "expression": "request.path.matches('/admin.*') && has(request.headers['x-debug']) && !inIpRange(origin.ip, '10.0.0.0/8')"
        }
      }
    },
    {
      "priority": 110,
      "description": "Block X-Forwarded-For header on admin login POST",
      "action": "deny(403)",
      "preview": false,
      "match": {
        "expr": {
          "expression": "request.path == '/admin/login' && request.method == 'POST' && has(request.headers['x-forwarded-for'])"
        }
      }
    },
    {
      "priority": 120,
      "description": "Block known scanner user agents on admin paths",
      "action": "deny(403)",
      "preview": false,
      "match": {
        "expr": {
          "expression": "request.path.matches('/admin.*') && (request.headers['user-agent'].lower().contains('sqlmap') || request.headers['user-agent'].lower().contains('nikto') || request.headers['user-agent'].lower().contains('nmap'))"
        }
      }
    },
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

**Complexity:** Low - CEL expressions are very expressive

**GCP Advantages:**
- Most flexible condition syntax via CEL
- Built-in string functions: `contains()`, `matches()`, `lower()`, `startsWith()`
- Clean OR logic: `|| ` operator
- Negation: `!` operator or `!has()`

---

## Cross-Provider Comparison

| Feature | AWS WAFv2 | Azure WAF | GCP Cloud Armor |
|---------|-----------|-----------|-----------------|
| **Custom rules with conditions** | ✅ Perfect | ✅ Perfect | ✅ Perfect |
| **Path matching** | ✅ Exact, Prefix, Regex | ✅ Exact, Prefix, Contains | ✅ CEL: ==, matches(), startsWith() |
| **Header matching** | ✅ Exact, Contains, Exists | ✅ Exact, Contains, Exists | ✅ CEL: ==, contains(), has() |
| **Method matching** | ✅ Yes | ✅ Yes | ✅ CEL: request.method |
| **IP matching** | ✅ CIDR, negation | ✅ CIDR, negation | ✅ CEL: inIpRange(), negation |
| **Boolean logic** | ✅ Nested AND/OR/NOT | ⚠️ AND only (flat) | ✅ CEL: &&, \|\|, ! |
| **Multiple values OR** | ✅ OrStatement | ✅ Multiple matchValues | ✅ CEL: \|\| |
| **Complexity** | Low | Low | Low |
| **Fidelity** | Perfect | Perfect | Perfect |

## Design Decision Impact

### API Design

✅ **Include `customRules` field** - this works great across all providers:

```yaml
spec:
  customRules:
    - name: block-suspicious-admin-access
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
      reason: "Block debug header on admin from public internet"
```

**Key Features:**
- `name`: Human-readable identifier
- `priority`: Evaluation order (lower = higher priority)
- `action`: block, allow, count
- `conditions`: Path, headers, IP, method matching
- `reason`: Explanation for audit/documentation

**Rationale:**
- ✅ All 3 providers have perfect support
- ✅ Critical security feature (custom threat protection)
- ✅ Common use case (protect admin interfaces, block scanners)
- ✅ Clean portable API maps well to all providers

---

### Relationship to Other Features

**Custom Rules vs Rule Overrides:**

| Feature | Purpose | Portability | When to Use |
|---------|---------|-------------|-------------|
| **customRules** | Define new blocking/allowing rules | ✅ Perfect (all providers) | Path-specific protection, scanner blocking, custom threats, conditional bypasses |
| **ruleOverrides** | Change managed rule actions globally | ✅ Great (AWS/Azure perfect, GCP partial) | Tune managed rules for false positives |

**Example combining both:**
```yaml
spec:
  # Managed rules for baseline protection
  managedRuleGroups:
    - type: CoreRuleSet
      action: block
  
  # Tune managed rules for false positives (unconditional)
  ruleOverrides:
    - managedRuleGroup: CoreRuleSet
      ruleId: "942100"
      action: count
      reason: "Known false positive on GraphQL queries"
  
  # Custom rules for specific threats and conditional bypasses
  customRules:
    - name: health-check-bypass
      priority: 10
      action: allow
      conditions:
        path:
          exact: "/health"
      reason: "Bypass WAF for health checks"
    
    - name: block-debug-on-admin-external
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
      reason: "Block debug header on admin from external IPs"
    
    - name: internal-admin-bypass
      priority: 20
      action: allow
      conditions:
        path:
          prefix: "/admin"
        sourceIP:
          cidr: "10.0.0.0/8"
      reason: "Allow internal admin traffic"
```

---

### Implementation Notes

**AWS KCP Reconciler**:
- Custom rules use `Action` (not `OverrideAction` like managed rules)
- Conditions combined with `AndStatement` or `OrStatement`
- Supports nested boolean logic with `NotStatement`

**Azure KCP Reconciler**:
- Custom rules defined in `customRules` array
- Each rule has `matchConditions` (all conditions are AND)
- Use `negationConditon: true` for negation

**GCP KCP Reconciler**:
- Custom rules use CEL (Common Expression Language) expressions
- CEL supports rich boolean logic: `&&`, `||`, `!`, `has()`, `matches()`, `contains()`
- Very flexible and expressive

---

## Validation Matrix

| Test Case | AWS | Azure | GCP | Expected Behavior |
|-----------|-----|-------|-----|-------------------|
| /admin + X-Debug header from public IP | ✅ Block | ✅ Block | ✅ Block | All providers block |
| /admin + X-Debug header from 10.0.0.0/8 | ✅ Allow | ✅ Allow | ✅ Allow | Internal network allowed |
| /api + X-Debug header from public IP | ✅ Allow | ✅ Allow | ✅ Allow | Path doesn't match |
| /admin/login POST with X-Forwarded-For | ✅ Block | ✅ Block | ✅ Block | All providers block |
| User-Agent contains "sqlmap" on /admin | ✅ Block | ✅ Block | ✅ Block | Scanner blocked |
| User-Agent contains "Chrome" on /admin | ✅ Allow | ✅ Allow | ✅ Allow | Legitimate browser allowed |

---

## Conclusion

**Custom rules with path and header conditions work perfectly across all providers.**

This is the **most portable** feature of the WAF API:
- ✅ Use Case 1 (Unconditional rule override): AWS + Azure perfect, GCP partial
- ✅ Use Case 3 (Custom rules): **All providers perfect** ⭐

**Recommendation:** ✅ **Prioritize implementing `customRules` first**

**Implementation Order:**
1. **Phase 2.1:** `ruleOverrides` + `customRules` (universal or great support)
2. **Phase 2.2:** `managedRuleGroups` (portable managed rule selection - when validated)
3. **Phase 2.3:** Advanced features (sizeLimits, geoBlocking, etc.)

This use case validates that path + header + IP conditions work great for custom rules across all providers!
