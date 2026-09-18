# Rule Statement Granularity Comparison Across Providers

This document compares the rule composition and statement nesting capabilities across AWS WAFv2, Azure Application Gateway WAF, and GCP Cloud Armor, going beyond just managed rule groups to examine custom rule flexibility.

## Core Question

Does Azure and GCP offer the same granular rule composition as AWS WAFv2's `WebACL.Rules[].Statement[]` nested structure?

**Short Answer:** ❌ No - AWS has significantly more granular composition capabilities than Azure or GCP.

---

## AWS WAFv2 - Highly Granular Nested Statements

AWS WAFv2 provides **deep, recursive statement nesting** with complex boolean logic.

### Statement Types

AWS offers 20+ statement types that can be arbitrarily nested:

| Category | Statement Types |
|----------|----------------|
| **Match Statements** | `ByteMatchStatement`, `SqliMatchStatement`, `XssMatchStatement`, `SizeConstraintStatement`, `GeoMatchStatement`, `IPSetReferenceStatement`, `RegexMatchStatement`, `RegexPatternSetReferenceStatement` |
| **Logical Operators** | `AndStatement`, `OrStatement`, `NotStatement` |
| **Rate Limiting** | `RateBasedStatement` |
| **Managed Rules** | `ManagedRuleGroupStatement`, `RuleGroupReferenceStatement` |
| **Label Matching** | `LabelMatchStatement` |

### Nesting Depth

✅ **Unlimited nesting** - Statements can be nested as deeply as needed  
✅ **Arbitrary boolean logic** - AND, OR, NOT at any level  
✅ **Mixed statement types** - Combine any statement types in logical operators  

### AWS Statement Structure Example

```json
{
  "Statement": {
    "AndStatement": {
      "Statements": [
        {
          "OrStatement": {
            "Statements": [
              {
                "ByteMatchStatement": {
                  "SearchString": "/admin",
                  "FieldToMatch": {"UriPath": {}},
                  "PositionalConstraint": "STARTS_WITH"
                }
              },
              {
                "ByteMatchStatement": {
                  "SearchString": "/management",
                  "FieldToMatch": {"UriPath": {}},
                  "PositionalConstraint": "STARTS_WITH"
                }
              }
            ]
          }
        },
        {
          "NotStatement": {
            "Statement": {
              "IPSetReferenceStatement": {
                "Arn": "arn:aws:wafv2:region:account:regional/ipset/blocked-ips/..."
              }
            }
          }
        },
        {
          "OrStatement": {
            "Statements": [
              {
                "SizeConstraintStatement": {
                  "FieldToMatch": {"SingleHeader": {"Name": "authorization"}},
                  "ComparisonOperator": "GE",
                  "Size": 1
                }
              },
              {
                "AndStatement": {
                  "Statements": [
                    {
                      "ByteMatchStatement": {
                        "SearchString": "internal",
                        "FieldToMatch": {"SingleHeader": {"Name": "x-forwarded-for"}},
                        "PositionalConstraint": "CONTAINS"
                      }
                    },
                    {
                      "IPSetReferenceStatement": {
                        "Arn": "arn:aws:wafv2:region:account:regional/ipset/internal-network/..."
                      }
                    }
                  ]
                }
              }
            ]
          }
        }
      ]
    }
  }
}
```

**Translation:** `(path=/admin OR path=/management) AND NOT(ip IN blocked-list) AND (header:authorization EXISTS OR (header:x-forwarded-for CONTAINS "internal" AND ip IN internal-network))`

### AWS Capabilities Summary

✅ **Deep nesting** - 5+ levels deep without issue  
✅ **Boolean algebra** - Full AND/OR/NOT at any level  
✅ **Mixed conditions** - Combine path, IP, headers, body, geo, rate, labels  
✅ **Scope-down statements** - Apply complex conditions to managed rules  
✅ **Reusable logic** - Label-based chaining reduces duplication  

---

## Azure Application Gateway WAF - Limited Flat Structure

Azure WAF uses a **flat match condition model** with limited composition.

### Rule Structure

Azure rules have:
- **Single-level match conditions** - All conditions in a flat array
- **Implicit AND** - All conditions must match (no OR support)
- **No nesting** - Cannot nest logical operators
- **No NOT operator** - Only negation via `negationConditon: true` on individual conditions

### Azure Rule Structure

```json
{
  "customRules": [
    {
      "name": "ExampleRule",
      "priority": 10,
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
          "matchVariables": [{"variableName": "RemoteAddr"}],
          "operator": "IPMatch",
          "negationConditon": true,
          "matchValues": ["10.0.0.0/8"],
          "transforms": []
        },
        {
          "matchVariables": [{"variableName": "RequestHeaders", "selector": "Authorization"}],
          "operator": "Exists",
          "negationConditon": false,
          "matchValues": [],
          "transforms": []
        }
      ],
      "state": "Enabled"
    }
  ]
}
```

**Translation:** `path STARTS_WITH /admin AND ip NOT_IN 10.0.0.0/8 AND header:authorization EXISTS`

### Azure Limitations

❌ **No OR logic** - Cannot express `(A OR B) AND C`  
❌ **No statement nesting** - Flat condition list only  
❌ **No complex boolean expressions** - Only AND with per-condition NOT  
❌ **No scope-down on managed rules** - Cannot apply conditions to managed rule evaluation  
❌ **Limited composition** - Must create multiple rules for OR logic  

### Azure Workaround for OR Logic

To express `(path=/admin OR path=/api) AND ip=internal`, you need **TWO separate rules**:

```json
{
  "customRules": [
    {
      "name": "AdminFromInternal",
      "priority": 10,
      "action": "Allow",
      "matchConditions": [
        {"matchVariables": [{"variableName": "RequestUri"}], "operator": "BeginsWith", "matchValues": ["/admin"]},
        {"matchVariables": [{"variableName": "RemoteAddr"}], "operator": "IPMatch", "matchValues": ["10.0.0.0/8"]}
      ]
    },
    {
      "name": "APIFromInternal",
      "priority": 11,
      "action": "Allow",
      "matchConditions": [
        {"matchVariables": [{"variableName": "RequestUri"}], "operator": "BeginsWith", "matchValues": ["/api"]},
        {"matchVariables": [{"variableName": "RemoteAddr"}], "operator": "IPMatch", "matchValues": ["10.0.0.0/8"]}
      ]
    }
  ]
}
```

### Azure Match Variables

Azure supports these match variables:
- `RemoteAddr` - Client IP
- `RequestMethod` - HTTP method
- `RequestUri` - Full URI
- `QueryString` - Query string
- `PostArgs` - POST body arguments
- `RequestHeaders` - HTTP headers (with selector)
- `RequestCookieNames` - Cookie names
- `RequestCookieValues` - Cookie values (with selector)

### Azure Capabilities Summary

⚠️ **Flat structure** - No nesting beyond single-level conditions  
⚠️ **AND-only** - All conditions in array are AND-ed  
⚠️ **Per-condition negation** - `negationConditon` on each condition  
❌ **No OR operator** - Must create separate rules  
❌ **No deep composition** - Cannot nest logical operators  

---

## GCP Cloud Armor - CEL Expression Flexibility

GCP Cloud Armor uses **CEL (Common Expression Language)** which provides good flexibility within expression limits.

### Expression-Based Model

GCP rules use a single CEL expression per rule:
- **String-based boolean logic** - Express AND/OR/NOT in CEL syntax
- **Function calls** - Built-in functions like `inIpRange()`, `has()`, `matches()`
- **Expression composition** - Combine conditions with `&&`, `||`, `!` operators
- **No deep nesting** - Expression complexity limits

### GCP Rule Structure

```json
{
  "rules": [
    {
      "priority": 100,
      "description": "Complex condition example",
      "action": "deny(403)",
      "preview": false,
      "match": {
        "expr": {
          "expression": "(request.path.matches('/admin.*') || request.path.matches('/management.*')) && !inIpRange(origin.ip, '192.168.1.0/24') && (has(request.headers['authorization']) || (request.headers['x-forwarded-for'].contains('internal') && inIpRange(origin.ip, '10.0.0.0/8')))"
        }
      }
    }
  ]
}
```

**Translation:** Same as AWS example - `(path=/admin OR path=/management) AND NOT(ip IN 192.168.1.0/24) AND (header:authorization EXISTS OR (header:x-forwarded-for CONTAINS "internal" AND ip IN 10.0.0.0/8))`

### CEL Operators and Functions

| Operator/Function | Description |
|-------------------|-------------|
| `&&` | Logical AND |
| `\|\|` | Logical OR |
| `!` | Logical NOT |
| `==`, `!=` | Equality |
| `<`, `<=`, `>`, `>=` | Comparison |
| `matches()` | Regex matching |
| `contains()` | Substring matching |
| `startsWith()` | Prefix matching |
| `endsWith()` | Suffix matching |
| `has()` | Key existence check |
| `inIpRange()` | IP range check |
| `size()` | Collection size |
| `lower()`, `upper()` | Case conversion |

### CEL Expression Examples

```javascript
// Simple AND
request.path == '/admin' && origin.ip == '203.0.113.1'

// OR logic
request.path.matches('/admin.*') || request.path.matches('/api.*')

// NOT logic
!inIpRange(origin.ip, '10.0.0.0/8')

// Complex nested logic
(request.method == 'POST' && request.path.startsWith('/api/')) && 
(has(request.headers['x-api-key']) || inIpRange(origin.ip, '10.0.0.0/8'))

// Multiple conditions with precedence
request.path == '/admin' && 
!inIpRange(origin.ip, '192.168.1.0/24') && 
(has(request.headers['authorization']) || request.headers['x-internal-token'] == 'secret')

// Combining preconfigured WAF evaluation
request.path.matches('/api/.*') && 
!has(request.headers['x-skip-waf']) && 
evaluatePreconfiguredWaf('sqli-v33-stable', {'sensitivity': 1})
```

### GCP CEL Request Fields

```javascript
// Request properties
request.method          // GET, POST, etc.
request.path            // /path/to/resource
request.query           // Query string
request.scheme          // http or https
request.headers         // Map of headers
request.headers['name'] // Specific header value

// Origin properties
origin.ip               // Client IP address
origin.region_code      // ISO region code (geo)

// Other
evaluatePreconfiguredWaf('rule-name', {'sensitivity': N})
```

### GCP Limitations

⚠️ **Expression length limit** - Max ~2048 characters per expression  
⚠️ **No statement reuse** - Cannot reference sub-expressions  
⚠️ **Complexity limits** - Very deep nesting may hit evaluation limits  
⚠️ **String-based** - Harder to validate than structured JSON  
✅ **Full boolean logic** - AND, OR, NOT supported  
✅ **Function composition** - Rich set of built-in functions  

### GCP Capabilities Summary

✅ **Boolean operators** - Full AND/OR/NOT support  
✅ **Expression composition** - Combine conditions freely  
✅ **Rich functions** - String, IP, header, geo functions  
⚠️ **Single expression** - All logic in one string (not nested statements)  
⚠️ **Length limits** - Complex logic may hit character limits  
❌ **No structured nesting** - Expression-based, not statement-based  

---

## Granularity Comparison Table

| Feature | AWS WAFv2 | Azure WAF | GCP Cloud Armor |
|---------|-----------|-----------|-----------------|
| **Nested Statements** | ✅ Yes (unlimited depth) | ❌ No (flat only) | ⚠️ Via expression composition |
| **AND Operator** | ✅ Yes (`AndStatement`) | ✅ Yes (implicit in array) | ✅ Yes (`&&`) |
| **OR Operator** | ✅ Yes (`OrStatement`) | ❌ No (need multiple rules) | ✅ Yes (`\|\|`) |
| **NOT Operator** | ✅ Yes (`NotStatement`) | ⚠️ Per-condition only | ✅ Yes (`!`) |
| **Arbitrary Nesting** | ✅ Yes | ❌ No | ⚠️ Limited by expression length |
| **Mixed Condition Types** | ✅ Yes | ✅ Yes (in flat array) | ✅ Yes |
| **Structured Composition** | ✅ Yes (JSON structure) | ⚠️ Limited (flat array) | ❌ No (string expression) |
| **Reusable Sub-Expressions** | ✅ Yes (via labels) | ❌ No | ❌ No |
| **Scope-Down Managed Rules** | ✅ Yes | ❌ No | ⚠️ Via expression in match |
| **Max Nesting Depth** | Unlimited | 1 level | ~10 levels (expression limit) |
| **Validation** | ✅ Structured (API validates) | ✅ Structured (API validates) | ⚠️ String (parsed at runtime) |

---

## Complex Logic Examples

### Example: `(A OR B) AND (C OR D) AND NOT E`

**AWS WAFv2** - Native nested structure:
```json
{
  "Statement": {
    "AndStatement": {
      "Statements": [
        {
          "OrStatement": {
            "Statements": [
              {"ByteMatchStatement": {"SearchString": "/admin", "FieldToMatch": {"UriPath": {}}, "PositionalConstraint": "STARTS_WITH"}},
              {"ByteMatchStatement": {"SearchString": "/management", "FieldToMatch": {"UriPath": {}}, "PositionalConstraint": "STARTS_WITH"}}
            ]
          }
        },
        {
          "OrStatement": {
            "Statements": [
              {"SizeConstraintStatement": {"FieldToMatch": {"SingleHeader": {"Name": "authorization"}}, "ComparisonOperator": "GE", "Size": 1}},
              {"IPSetReferenceStatement": {"Arn": "arn:aws:wafv2:region:account:regional/ipset/internal-network/..."}}
            ]
          }
        },
        {
          "NotStatement": {
            "Statement": {
              "IPSetReferenceStatement": {"Arn": "arn:aws:wafv2:region:account:regional/ipset/blocked-ips/..."}
            }
          }
        }
      ]
    }
  }
}
```

**Azure WAF** - Requires 4 separate rules (cross product):
```json
{
  "customRules": [
    {
      "name": "AdminWithAuthNotBlocked",
      "priority": 10,
      "action": "Allow",
      "matchConditions": [
        {"matchVariables": [{"variableName": "RequestUri"}], "operator": "BeginsWith", "matchValues": ["/admin"]},
        {"matchVariables": [{"variableName": "RequestHeaders", "selector": "Authorization"}], "operator": "Exists"},
        {"matchVariables": [{"variableName": "RemoteAddr"}], "operator": "IPMatch", "negationConditon": true, "matchValues": ["blocked-ips"]}
      ]
    },
    {
      "name": "AdminFromInternalNotBlocked",
      "priority": 11,
      "action": "Allow",
      "matchConditions": [
        {"matchVariables": [{"variableName": "RequestUri"}], "operator": "BeginsWith", "matchValues": ["/admin"]},
        {"matchVariables": [{"variableName": "RemoteAddr"}], "operator": "IPMatch", "matchValues": ["10.0.0.0/8"]},
        {"matchVariables": [{"variableName": "RemoteAddr"}], "operator": "IPMatch", "negationConditon": true, "matchValues": ["blocked-ips"]}
      ]
    },
    {
      "name": "ManagementWithAuthNotBlocked",
      "priority": 12,
      "action": "Allow",
      "matchConditions": [
        {"matchVariables": [{"variableName": "RequestUri"}], "operator": "BeginsWith", "matchValues": ["/management"]},
        {"matchVariables": [{"variableName": "RequestHeaders", "selector": "Authorization"}], "operator": "Exists"},
        {"matchVariables": [{"variableName": "RemoteAddr"}], "operator": "IPMatch", "negationConditon": true, "matchValues": ["blocked-ips"]}
      ]
    },
    {
      "name": "ManagementFromInternalNotBlocked",
      "priority": 13,
      "action": "Allow",
      "matchConditions": [
        {"matchVariables": [{"variableName": "RequestUri"}], "operator": "BeginsWith", "matchValues": ["/management"]},
        {"matchVariables": [{"variableName": "RemoteAddr"}], "operator": "IPMatch", "matchValues": ["10.0.0.0/8"]},
        {"matchVariables": [{"variableName": "RemoteAddr"}], "operator": "IPMatch", "negationConditon": true, "matchValues": ["blocked-ips"]}
      ]
    }
  ]
}
```

**GCP Cloud Armor** - Single CEL expression:
```json
{
  "rules": [
    {
      "priority": 100,
      "action": "allow",
      "match": {
        "expr": {
          "expression": "(request.path.matches('/admin.*') || request.path.matches('/management.*')) && (has(request.headers['authorization']) || inIpRange(origin.ip, '10.0.0.0/8')) && !inIpRange(origin.ip, '192.168.1.0/24')"
        }
      }
    }
  ]
}
```

### Comparison:
- **AWS**: 1 rule, clean nested structure, easy to read/maintain
- **Azure**: 4 rules (2×2 cross product), duplicated conditions, hard to maintain
- **GCP**: 1 rule, compact but string-based, harder to validate

---

## Impact on Portable WafPolicy API

Given the granularity differences, the portable API should:

### Option 1: Limit to Common Denominator (Azure's Flat Model)

Only support flat AND conditions in portable API:

```yaml
conditionalOverrides:
  - condition:
      path: {prefix: "/admin"}
      sourceIP: {cidr: "10.0.0.0/8"}
      header: {name: "Authorization", exists: true}
    overrides: [...]
```

**Pros:**
- ✅ Works on all providers without workarounds
- ✅ Simple, predictable translation
- ✅ Easy to validate

**Cons:**
- ❌ Cannot express OR logic
- ❌ Underutilizes AWS/GCP capabilities
- ❌ Users need multiple override blocks for OR conditions

### Option 2: Support Rich Logic, Generate Multiple Rules for Azure

Allow complex boolean logic in portable API:

```yaml
conditionalOverrides:
  - condition:
      anyOf:  # OR
        - path: {prefix: "/admin"}
        - path: {prefix: "/management"}
      allOf:  # AND
        - anyOf:  # OR
            - header: {name: "Authorization", exists: true}
            - sourceIP: {cidr: "10.0.0.0/8"}
        - not:  # NOT
            sourceIP: {cidr: "192.168.1.0/24"}
    overrides: [...]
```

**Translation:**
- AWS: Single nested statement
- Azure: Generate 4 rules (cross product)
- GCP: Single CEL expression

**Pros:**
- ✅ Full expressiveness
- ✅ Leverages AWS/GCP capabilities
- ✅ Users write intent once

**Cons:**
- ❌ Azure hits 100 custom rule limit faster
- ❌ Complex Cloud Manager translation logic
- ❌ Azure performance impact (more rules to evaluate)

### Option 3: Hybrid - Flat with Limited OR Support

Support flat AND conditions + limited OR via arrays:

```yaml
conditionalOverrides:
  - condition:
      path:
        anyOf: ["/admin", "/management"]  # OR within single field
      header: {name: "Authorization", exists: true}
      sourceIP: {cidr: "10.0.0.0/8"}
    overrides: [...]
```

**Translation:**
- AWS: Optimal nested statement
- Azure: 2 rules (one per path)
- GCP: CEL expression with OR

**Pros:**
- ✅ Handles most common OR cases
- ⚠️ Limited Azure rule explosion
- ✅ Simpler than full boolean logic

**Cons:**
- ⚠️ Cannot express arbitrary OR/AND nesting
- ⚠️ Still generates multiple Azure rules

---

## Recommendations

### For Cloud Manager WafPolicy:

1. **Start with Option 1 (Flat AND model)** for initial release
   - Covers 80% of use cases
   - Works well on all providers
   - Simple to implement and validate

2. **Add Option 3 (Limited OR via arrays)** in next iteration
   - Handles common OR scenarios
   - Manageable Azure rule explosion
   - Still maps cleanly to AWS/GCP

3. **Consider Option 2 (Full boolean logic)** for advanced use cases
   - Mark as experimental/advanced
   - Document Azure limitations (rule count)
   - Provide estimation tool (predicts # Azure rules)

### Azure-Specific Guidance:

When generating Azure rules for OR logic:
- Warn users if cross product exceeds 20 rules
- Suggest refactoring into separate WafPolicy resources
- Document 100 custom rule limit clearly

### GCP-Specific Guidance:

When generating CEL expressions:
- Validate expression length < 2000 chars
- Fail with helpful error if too complex
- Suggest breaking into multiple rules

---

## Conclusion

**AWS WAFv2 offers significantly more granular rule composition** than Azure or GCP:

- **AWS**: Unlimited nested statements with full boolean logic
- **Azure**: Flat AND-only conditions (OR requires multiple rules)
- **GCP**: Expression-based with full boolean operators but length limits

For a portable WafPolicy API, we must either:
1. **Limit to Azure's capabilities** (simple, but restrictive)
2. **Support rich logic and generate workarounds** (complex, but powerful)
3. **Hybrid approach** (balance expressiveness with complexity)

**Recommendation: Start simple (flat AND), evolve toward hybrid approach** as user needs become clearer.
