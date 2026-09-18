# Managed Rules Structure Comparison Across Providers

## Overview

This document compares how managed rule groups are structured and invoked across AWS WAFv2, Azure Application Gateway WAF, and GCP Cloud Armor. Understanding these structural differences is critical for designing a portable API that maps correctly to each provider's native structure.

## Key Finding

**AWS uses nested statements** - Managed rule groups are wrapped in Rule objects that contain Statements, while Azure and GCP use flat arrays of managed rule sets/rules.

## AWS WAFv2 Structure

### Hierarchy

```
WebACL
└── Rules[]                           # Array of Rule objects
    └── Rule
        ├── Name
        ├── Priority
        ├── Statement                  # ← Managed rule group goes here
        │   └── ManagedRuleGroupStatement
        │       ├── VendorName
        │       ├── Name
        │       ├── Version (optional)
        │       ├── ExcludedRules[] (optional)
        │       ├── RuleActionOverrides[] (optional)
        │       └── ScopeDownStatement (optional)
        ├── OverrideAction              # Count or None
        └── VisibilityConfig
```

### Example: AWS Managed Rule Group

```json
{
  "Name": "my-webacl",
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
          "Version": "Version_1.3",
          "ExcludedRules": [
            {"Name": "SizeRestrictions_BODY"}
          ],
          "RuleActionOverrides": [
            {
              "Name": "SQLi_QUERYARGUMENTS",
              "ActionToUse": {"Count": {}}
            }
          ],
          "ScopeDownStatement": {
            "ByteMatchStatement": {
              "SearchString": "/admin",
              "FieldToMatch": {"UriPath": {}},
              "PositionalConstraint": "STARTS_WITH"
            }
          }
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

### Key Characteristics

1. **Nested Structure**: Managed rule groups are wrapped in `Rules[].Statement`
2. **Explicit Priorities**: Each rule has a numeric priority (ascending order)
3. **OverrideAction vs Action**: Managed rules use `OverrideAction` (Count/None), custom rules use `Action`
4. **Flexibility**: Can mix managed and custom rules in same Rules[] array
5. **Scope-Down**: Conditions can narrow when managed rules apply (ScopeDownStatement)

---

## Azure Application Gateway WAF Structure

### Hierarchy

```
WebApplicationFirewallPolicy
└── managedRules
    ├── managedRuleSets[]             # Flat array of managed rule sets
    │   └── ManagedRuleSet
    │       ├── ruleSetType
    │       ├── ruleSetVersion
    │       └── ruleGroupOverrides[] (optional)
    │           └── RuleGroupOverride
    │               ├── ruleGroupName
    │               └── rules[]
    │                   └── RuleOverride
    │                       ├── ruleId
    │                       ├── state (Enabled/Disabled)
    │                       └── action (optional: Log, Block, Allow)
    └── exclusions[] (optional)
```

### Example: Azure Managed Rule Set

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
                  "ruleId": "942200",
                  "state": "Disabled"
                }
              ]
            }
          ]
        },
        {
          "ruleSetType": "Microsoft_BotManagerRuleSet",
          "ruleSetVersion": "1.0"
        }
      ],
      "exclusions": [
        {
          "matchVariable": "RequestCookieNames",
          "selectorMatchOperator": "StartsWith",
          "selector": "session",
          "exclusionManagedRuleSets": [
            {
              "ruleSetType": "Microsoft_DefaultRuleSet",
              "ruleSetVersion": "2.1",
              "ruleGroups": [
                {
                  "ruleGroupName": "SQLI",
                  "rules": [{"ruleId": "942100"}]
                }
              ]
            }
          ]
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

### Key Characteristics

1. **Flat Structure**: `managedRuleSets[]` is a simple array, not nested in rules
2. **No Priorities**: Order matters but no explicit priority numbers
3. **Rule Group Overrides**: Override actions at rule group → rule level
4. **Limited Conditional Logic**: No per-rule-set conditions; use customRules for that
5. **Global Exclusions**: Exclusions apply across rule sets, not per-rule-set

---

## GCP Cloud Armor Structure

### Hierarchy

```
SecurityPolicy
└── rules[]                           # Flat array of all rules (custom + preconfigured)
    └── Rule
        ├── priority
        ├── description
        ├── action (allow, deny, rate_based_ban, redirect, throttle)
        ├── preview (boolean)
        └── match
            └── expr
                └── expression        # CEL expression
                    # Can include evaluatePreconfiguredWaf()
```

### Example: GCP Preconfigured WAF Rules

```json
{
  "name": "my-security-policy",
  "rules": [
    {
      "priority": 1000,
      "description": "OWASP CRS - SQL injection protection",
      "action": "deny(403)",
      "preview": false,
      "match": {
        "expr": {
          "expression": "evaluatePreconfiguredWaf('sqli-v33-stable', {'sensitivity': 1})"
        }
      }
    },
    {
      "priority": 1100,
      "description": "OWASP CRS - XSS protection",
      "action": "deny(403)",
      "preview": false,
      "match": {
        "expr": {
          "expression": "evaluatePreconfiguredWaf('xss-v33-stable', {'sensitivity': 1})"
        }
      }
    },
    {
      "priority": 900,
      "description": "Allow /admin paths - bypass SQLi check",
      "action": "allow",
      "preview": false,
      "match": {
        "expr": {
          "expression": "request.path.matches('/admin.*') && evaluatePreconfiguredWaf('sqli-v33-stable')"
        }
      }
    },
    {
      "priority": 2147483647,
      "description": "Default rule",
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

### Key Characteristics

1. **Flat Structure**: All rules (custom + preconfigured WAF) in single `rules[]` array
2. **CEL Expressions**: Preconfigured WAF invoked via `evaluatePreconfiguredWaf()` in CEL
3. **Priority-Based**: Lower numbers evaluated first
4. **No Rule Groups**: Each preconfigured WAF evaluation is a separate rule
5. **Conditions in CEL**: Combine conditions with `&&`, `||` operators in CEL expression

---

## Structural Comparison Table

| Aspect | AWS WAFv2 | Azure WAF | GCP Cloud Armor |
|--------|-----------|-----------|-----------------|
| **Top-level container** | `Rules[]` | `managedRules.managedRuleSets[]` | `rules[]` |
| **Managed rules location** | Nested in `Rules[].Statement` | Flat in `managedRuleSets[]` | Flat in `rules[]` via CEL |
| **Nesting depth** | 2 levels (Rule → Statement) | 1 level (managedRuleSets) | 1 level (rules) |
| **Priority mechanism** | Numeric `Priority` field | Array order | Numeric `priority` field |
| **Mixing custom + managed** | Yes, in same `Rules[]` | Separate: `customRules[]` + `managedRules` | Yes, in same `rules[]` |
| **Conditional application** | `ScopeDownStatement` on managed rule | Separate `customRules` | CEL `&&` with conditions |
| **Action override** | `RuleActionOverrides[]` | `ruleGroupOverrides[].rules[]` | Not supported (sensitivity only) |
| **Exclusions** | `ExcludedRules[]` | Global `exclusions[]` | Not applicable |

---

## Impact on Portable API Design

### Challenge: Mapping Portable managedRuleGroups to AWS Nested Structure

**Our portable API:**
```yaml
spec:
  managedRuleGroups:
    - type: CoreRuleSet
      action: block
    - type: SQLInjectionProtection
      action: count
```

**Must translate to AWS nested structure:**
```json
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
      },
      "OverrideAction": {"None": {}}
    }
  ]
}
```

### Solution: KCP Reconciler Translation

**AWS KCP Reconciler must:**

1. **Wrap each managedRuleGroup in a Rule object:**
   
2. **Handle conditional overrides via ScopeDownStatement:**
   
3. **Translate rule overrides to RuleActionOverrides:**
   
### Solution: Azure Translation (Flat Structure)

**Azure KCP Reconciler must:**

1. **Create managedRuleSets array directly:**
   
2. **Handle conditional overrides via customRules:**
   
### Solution: GCP Translation (CEL Expressions)

**GCP KCP Reconciler must:**

1. **Create rule per preconfigured WAF evaluation:**
   
2. **Handle conditional overrides via CEL &&:**
   
---

## Design Implications

### 1. Portable API Should Abstract Structure

**Good: Decision-focused, structure-agnostic**
```yaml
managedRuleGroups:
  - type: CoreRuleSet
    action: block
  - type: SQLInjectionProtection
    action: count
```

**Bad: AWS-specific nested structure**
```yaml
rules:  # ← Leaks AWS concept of "Rules"
  - statement:
      managedRuleGroup:
        type: CoreRuleSet
```

### 2. Reconcilers Handle Structure Translation

- **AWS reconciler** wraps in Rules[].Statement
- **Azure reconciler** creates flat managedRuleSets[]
- **GCP reconciler** creates flat rules[] with CEL

### 3. Classifications Must Account for Structure

**Our classifications API:**
```yaml
classifications:
  - name: trusted-source
    conditions: {...}
```

**Translates to:**
- **AWS**: ScopeDownStatement on managed rule statement
- **Azure**: Custom rules evaluated before managed rules
- **GCP**: CEL expression combined with evaluatePreconfiguredWaf()

---

## Validation: Current API Design

Our current portable API design **correctly abstracts the structural differences**:

✅ **No AWS-specific nesting** - `managedRuleGroups` is flat array
✅ **No provider structure leakage** - Users don't see "Rules" or "Statement"
✅ **KCP reconcilers handle mapping** - Each provider translates to native structure
✅ **Conditional overrides portable** - Don't assume AWS ScopeDownStatement
✅ **Classifications abstract conditions** - Don't assume AWS label chaining

### Example: How It All Works

**User writes portable YAML:**
```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafPolicy
metadata:
  name: my-policy
spec:
  managedRuleGroups:
    - type: CoreRuleSet
      action: block
  
  classifications:
    - name: admin
      conditions:
        path: {prefix: "/admin"}
  
  conditionalOverrides:
    - classification: admin
      overrides:
        - managedRuleGroup: CoreRuleSet
          ruleId: "*"
          action: allow
```

**AWS reconciler translates to:**
```json
{
  "Rules": [
    {
      "Name": "ClassifyAdmin",
      "Priority": 10,
      "Statement": {"ByteMatchStatement": {...}},
      "Action": {"Count": {}},
      "RuleLabels": [{"Name": "admin"}]
    },
    {
      "Name": "CoreRuleSetWithAdminBypass",
      "Priority": 1000,
      "Statement": {
        "ManagedRuleGroupStatement": {
          "VendorName": "AWS",
          "Name": "AWSManagedRulesCommonRuleSet",
          "ScopeDownStatement": {
            "NotStatement": {
              "Statement": {
                "LabelMatchStatement": {
                  "Key": "admin"
                }
              }
            }
          }
        }
      },
      "OverrideAction": {"None": {}}
    }
  ]
}
```

**Azure reconciler translates to:**
```json
{
  "customRules": [
    {
      "name": "AdminBypass",
      "priority": 10,
      "action": "Allow",
      "matchConditions": [
        {
          "matchVariables": [{"variableName": "RequestUri"}],
          "operator": "BeginsWith",
          "matchValues": ["/admin"]
        }
      ]
    }
  ],
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

**GCP reconciler translates to:**
```json
{
  "rules": [
    {
      "priority": 900,
      "action": "allow",
      "match": {
        "expr": {
          "expression": "request.path.matches('/admin.*')"
        }
      }
    },
    {
      "priority": 1000,
      "action": "deny(403)",
      "match": {
        "expr": {
          "expression": "evaluatePreconfiguredWaf('owasp-crs-v030301-id', {'sensitivity': 1})"
        }
      }
    }
  ]
}
```

---

## Recommendations

### 1. Keep Portable API Flat

✅ **Current design is correct:**
```yaml
managedRuleGroups:
  - type: CoreRuleSet
```

❌ **Don't introduce AWS nesting:**
```yaml
rules:
  - statement:
      managedRuleGroup:
        type: CoreRuleSet
```

### 2. Document Structure Translation

Add to implementation guide:
- How AWS wraps in Rules[].Statement
- How Azure uses flat managedRuleSets[]
- How GCP uses CEL expressions in rules[]

### 3. Priority Assignment Strategy

**AWS & GCP:** Assign explicit priorities
- Classifications: 10-900
- Managed rules: 1000-1999
- Rate limits: 2000+

**Azure:** Order matters
- Custom rules first (priority 1-100)
- Managed rules after (no explicit priority)

### 4. Testing

Ensure reconciler tests verify:
- Correct nesting structure per provider
- Priority assignment
- Mixing custom and managed rules
- Conditional override translation

---

## Conclusion

**Key Insight:** AWS's nested structure (Rules → Statement → ManagedRuleGroupStatement) is fundamentally different from Azure's flat managedRuleSets[] and GCP's CEL-based rules[]. Our portable API correctly abstracts this by using a flat `managedRuleGroups` array, with KCP reconcilers responsible for translating to each provider's native structure.

**Current API Design Status:** ✅ **Correctly handles structural differences**

The portable API does not leak AWS's nesting, Azure's separation of custom/managed, or GCP's CEL expressions. Each KCP reconciler translates the portable intent to the provider's native structure.
