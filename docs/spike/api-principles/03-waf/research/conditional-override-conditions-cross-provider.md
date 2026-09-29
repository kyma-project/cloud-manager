# Conditional Override Conditions - Cross-Provider Support

This document provides a comprehensive cross-provider comparison table showing which condition types are supported for conditional overrides in WafPolicy across AWS WAFv2, Azure Application Gateway WAF, and GCP Cloud Armor.

## Condition Types Support Matrix

| Condition Type | Portable API | AWS WAFv2 | Azure Application Gateway WAF | GCP Cloud Armor | Notes |
|----------------|--------------|-----------|------------------------------|-----------------|-------|
| **Path - Exact Match** | `path.exact` | ✅ Yes | ✅ Yes | ✅ Yes | All providers support exact path matching |
| **Path - Prefix Match** | `path.prefix` | ✅ Yes | ✅ Yes | ✅ Yes | All providers support path prefix matching |
| **Path - Regex Match** | `path.regex` | ✅ Yes | ✅ Yes | ✅ Yes | All providers support regex path matching |
| **HTTP Method** | `method` | ✅ Yes | ✅ Yes | ✅ Yes | All standard methods (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS) |
| **Header - Exists** | `header.exists` | ✅ Yes | ✅ Yes | ✅ Yes | Check if header is present |
| **Header - Exact Value** | `header.value` | ✅ Yes | ✅ Yes | ✅ Yes | Check header exact value |
| **Header - Contains** | `header.contains` | ✅ Yes | ✅ Yes | ✅ Yes | Check if header contains substring |
| **Header - Regex** | `header.regex` | ✅ Yes | ✅ Yes | ✅ Yes | Match header value against regex |
| **Query Parameter - Exists** | `queryParam.exists` | ✅ Yes | ✅ Yes | ✅ Yes | Check if query parameter is present |
| **Query Parameter - Value** | `queryParam.value` | ✅ Yes | ✅ Yes | ✅ Yes | Check query parameter exact value |
| **Source IP - CIDR** | `sourceIP.cidr` | ✅ Yes | ✅ Yes | ✅ Yes | Match source IP against CIDR range |
| **Source IP - Exact** | `sourceIP.exact` | ✅ Yes | ✅ Yes | ✅ Yes | Match exact source IP address |
| **Cookie - Exists** | `cookie.exists` | ⚠️ Workaround | ✅ Yes | ✅ Yes | AWS uses header matching on `Cookie` header |
| **Cookie - Value** | `cookie.value` | ⚠️ Workaround | ✅ Yes | ✅ Yes | AWS uses header matching on `Cookie` header |
| **Request Body** | `body.contains` | ✅ Yes | ⚠️ Limited | ⚠️ Limited | AWS supports body inspection; Azure/GCP have limitations |
| **Request Size** | `requestSize` | ✅ Yes | ✅ Yes | ⚠️ Limited | Check total request size |
| **Geographic Location** | `geo.country` | ✅ Yes | ❌ No | ✅ Yes | AWS and GCP support geo-blocking; Azure requires custom rules |
| **Combined Conditions (AND)** | Multiple conditions in one block | ✅ Yes | ✅ Yes | ✅ Yes | All conditions in a block are AND-ed |
| **Combined Conditions (OR)** | Multiple condition blocks | ⚠️ Complex | ⚠️ Complex | ✅ Yes | GCP has best native OR support via CEL expressions |

## Legend

- ✅ **Yes**: Native support, straightforward translation
- ⚠️ **Workaround**: Supported but requires workaround or has limitations
- ❌ **No**: Not supported, condition will be ignored or requires alternative approach

## Detailed Provider Implementation Notes

### AWS WAFv2

**Native Support:**
- Path matching: `ByteMatchStatement` with `UriPath` field
- HTTP method: `ByteMatchStatement` with `Method` field
- Headers: `ByteMatchStatement` with `SingleHeader` field
- Query parameters: `ByteMatchStatement` with `QueryString` field
- Source IP: `IPSetReferenceStatement` (requires creating IPSet)
- Request size: `SizeConstraintStatement`
- Geographic location: `GeoMatchStatement`

**Workarounds:**
- **Cookies**: No native cookie matching; use header matching on `Cookie` header with regex
- **Body inspection**: Supported via `Body` field in `ByteMatchStatement` but limited to first 8KB
- **OR conditions**: Requires multiple rules with `OrStatement` wrapper or separate rule priorities

**AWS-Specific Syntax:**
```json
{
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
            "Arn": "arn:aws:wafv2:region:account:regional/ipset/..."
          }
        }
      ]
    }
  }
}
```

### Azure Application Gateway WAF

**Native Support:**
- Path matching: `matchVariable: RequestUri` with operators `BeginsWith`, `EndsWith`, `Contains`, `Equals`
- HTTP method: `matchVariable: RequestMethod`
- Headers: `matchVariable: RequestHeaders` with selector for header name
- Query parameters: Custom rules can match query string patterns
- Source IP: `matchVariable: RemoteAddr` with `IPMatch` operator
- Cookies: `matchVariable: RequestCookieNames` or `RequestCookieValues`
- Request size: Custom rules with size constraints

**Limitations:**
- **Conditional overrides**: Azure doesn't have native conditional rule overrides; Cloud Manager must implement via custom rules with higher priority
- **Body inspection**: Limited to managed rule evaluation; custom rules cannot inspect body content
- **Regex support**: Limited regex support compared to AWS/GCP

**Azure-Specific Syntax:**
```json
{
  "customRules": [
    {
      "name": "ConditionalAllow",
      "priority": 10,
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
        }
      ]
    }
  ]
}
```

### GCP Cloud Armor

**Native Support:**
- Path matching: CEL expression `request.path.matches()`
- HTTP method: CEL expression `request.method`
- Headers: CEL expression `request.headers['header-name']`
- Query parameters: CEL expression `request.query`
- Source IP: CEL expression `origin.ip` with `inIpRange()` function
- Geographic location: CEL expression `origin.region_code`
- Cookies: CEL expression `request.headers['cookie']` (requires parsing)
- Body inspection: Limited; not directly accessible in CEL expressions

**Advantages:**
- **CEL expressions**: Most flexible condition matching via Common Expression Language
- **OR conditions**: Native OR support via `||` operator in CEL
- **Complex logic**: Can express complex boolean logic in single expression

**GCP-Specific Syntax:**
```json
{
  "rules": [
    {
      "priority": 1000,
      "action": "allow",
      "match": {
        "expr": {
          "expression": "request.path.matches('/admin.*') && inIpRange(origin.ip, '10.0.0.0/8') && evaluatePreconfiguredWaf('sqli-v33-stable')"
        }
      }
    }
  ]
}
```

## Condition Matching Operators

### Path Matching

| Operator | AWS WAFv2 | Azure WAF | GCP Cloud Armor |
|----------|-----------|-----------|-----------------|
| Exact match | `EXACTLY` | `Equals` | `request.path == '/exact'` |
| Prefix match | `STARTS_WITH` | `BeginsWith` | `request.path.matches('/prefix.*')` |
| Suffix match | `ENDS_WITH` | `EndsWith` | `request.path.matches('.*\\.suffix$')` |
| Contains | `CONTAINS` | `Contains` | `request.path.matches('.*substring.*')` |
| Regex | `RegexMatchStatement` | Limited | `request.path.matches('regex')` |

### Header Matching

| Operator | AWS WAFv2 | Azure WAF | GCP Cloud Armor |
|----------|-----------|-----------|-----------------|
| Exists | `SizeConstraintStatement >= 1` | `Exists` operator | `has(request.headers['name'])` |
| Exact value | `ByteMatchStatement` with `EXACTLY` | `Equals` | `request.headers['name'] == 'value'` |
| Contains | `ByteMatchStatement` with `CONTAINS` | `Contains` | `request.headers['name'].contains('value')` |
| Regex | `RegexMatchStatement` | Limited | `request.headers['name'].matches('regex')` |
| Case insensitive | `TextTransformations: LOWERCASE` | `transforms: []` | `.lower()` in CEL |

### IP Address Matching

| Operator | AWS WAFv2 | Azure WAF | GCP Cloud Armor |
|----------|-----------|-----------|-----------------|
| CIDR range | `IPSetReferenceStatement` | `IPMatch` | `inIpRange(origin.ip, 'cidr')` |
| Exact IP | `IPSetReferenceStatement` (single IP) | `IPMatch` | `origin.ip == 'x.x.x.x'` |
| Multiple ranges | IPSet with multiple CIDRs | Multiple `matchValues` | Multiple `inIpRange()` with `\|\|` |
| IPv6 | ✅ Supported | ✅ Supported | ✅ Supported |

## Implementation Strategy by Provider

### AWS WAFv2 Translation Strategy

Cloud Manager translates portable conditional overrides to AWS using:

1. **Scope-Down Statements**: Apply conditions before managed rule evaluation
2. **RuleActionOverrides**: Override specific rule actions within scope
3. **Priority-Based Rules**: Higher-priority allow rules for bypass scenarios
4. **IPSet Resources**: Create and manage IPSet resources for IP-based conditions

### Azure Application Gateway WAF Translation Strategy

Cloud Manager translates portable conditional overrides to Azure using:

1. **Custom Rules**: Create high-priority custom rules that allow/block before managed rules
2. **Rule Group Overrides**: Override managed rule actions within specific rule groups
3. **Exclusions**: Use Azure exclusion mechanism for scoped rule exemptions
4. **Priority Ordering**: Carefully order custom rules (priority 1-100) before managed rules

**Important**: Azure doesn't support true "conditional overrides" of managed rules. Cloud Manager implements this by creating allow rules that bypass managed rules for matching conditions.

### GCP Cloud Armor Translation Strategy

Cloud Manager translates portable conditional overrides to GCP using:

1. **CEL Expressions**: Combine conditions with `evaluatePreconfiguredWaf()` in single expression
2. **Priority-Based Rules**: Lower-priority numbers (e.g., 900) evaluated before higher numbers (1000+)
3. **Preview Mode**: Use `preview: true` for count/detection mode
4. **Expression Composition**: Build complex AND/OR logic in CEL expressions

## Condition Priority and Evaluation Order

### AWS WAFv2

- Rules evaluated in ascending priority order (10 → 1000 → 2000)
- First matching rule action is applied
- Default action applies if no rules match
- Scope-down statements narrow rule evaluation scope

### Azure Application Gateway WAF

- Custom rules evaluated first (priority 1-100)
- Managed rules evaluated after custom rules (priority > 100)
- First matching `Allow` or `Block` action terminates evaluation
- Lower priority number = higher precedence

### GCP Cloud Armor

- Rules evaluated in ascending priority order (10 → 1000 → 2000)
- First matching rule action is applied
- Multiple rules can match if actions are compatible (preview mode)
- Default rule (priority 2147483647) evaluated last

## Limitations and Workarounds

### Common Limitations Across All Providers

1. **Request body inspection**: Limited to first few KB (AWS: 8KB, others vary)
2. **Cookie parsing**: No native cookie parsing; requires header regex matching
3. **Complex OR logic**: Requires multiple rules or complex expressions
4. **Dynamic content**: Cannot match based on response content or application state

### Provider-Specific Limitations

**AWS WAFv2:**
- IPSets must be pre-created and referenced by ARN
- Maximum 1,500 WCU (Web ACL Capacity Units) per WebACL
- Regex patterns consume higher WCU
- No native cookie matching

**Azure Application Gateway WAF:**
- No native conditional overrides; requires workaround via custom rules
- Limited regex support compared to AWS/GCP
- Custom rule limit: 100 per policy
- Exclusions are global or per-rule-group, not per-condition

**GCP Cloud Armor:**
- CEL expressions have complexity limits
- Limited preconfigured WAF rules compared to AWS managed rules
- Body inspection not available in CEL expressions
- Rate limiting uses separate adaptive protection feature

## Recommendations for Portable API

1. **Use simple conditions when possible**: Path prefix + method + header existence are universally supported
2. **Avoid body inspection**: Limited and inconsistent across providers
3. **Prefer CIDR ranges over exact IPs**: More efficient and better supported
4. **Test conditions across providers**: Provider-specific quirks may affect behavior
5. **Document provider limitations**: Make users aware of translation differences
6. **Use multiple simple conditions over complex OR logic**: Better portability

## Example: Same Condition Across All Providers

### Portable API
```yaml
conditionalOverrides:
  - condition:
      path:
        prefix: "/admin"
      header:
        name: "Authorization"
        exists: true
      sourceIP:
        cidr: "10.0.0.0/8"
    overrides:
      - managedRuleGroup: CoreRuleSet
        ruleId: "*"
        action: allow
```

### AWS WAFv2 Translation
```json
{
  "Statement": {
    "AndStatement": {
      "Statements": [
        {"ByteMatchStatement": {"SearchString": "/admin", "FieldToMatch": {"UriPath": {}}, "PositionalConstraint": "STARTS_WITH"}},
        {"SizeConstraintStatement": {"FieldToMatch": {"SingleHeader": {"Name": "authorization"}}, "ComparisonOperator": "GE", "Size": 1}},
        {"IPSetReferenceStatement": {"Arn": "arn:aws:wafv2:region:account:regional/ipset/internal-network/..."}}
      ]
    }
  },
  "Action": {"Allow": {}}
}
```

### Azure WAF Translation
```json
{
  "customRules": [{
    "name": "AdminInternalAllow",
    "priority": 10,
    "ruleType": "MatchRule",
    "action": "Allow",
    "matchConditions": [
      {"matchVariables": [{"variableName": "RequestUri"}], "operator": "BeginsWith", "matchValues": ["/admin"]},
      {"matchVariables": [{"variableName": "RequestHeaders", "selector": "Authorization"}], "operator": "Exists"},
      {"matchVariables": [{"variableName": "RemoteAddr"}], "operator": "IPMatch", "matchValues": ["10.0.0.0/8"]}
    ]
  }]
}
```

### GCP Cloud Armor Translation
```json
{
  "rules": [{
    "priority": 10,
    "action": "allow",
    "match": {
      "expr": {
        "expression": "request.path.matches('/admin.*') && has(request.headers['authorization']) && inIpRange(origin.ip, '10.0.0.0/8')"
      }
    }
  }]
}
```

## Conclusion

While all three providers support the core condition types (path, method, headers, source IP), the implementation details vary significantly:

- **AWS WAFv2**: Most comprehensive managed rules, requires IPSet resources, best for AWS-native deployments
- **Azure Application Gateway WAF**: Good integration with Azure ecosystem, requires custom rules for conditional overrides
- **GCP Cloud Armor**: Most flexible with CEL expressions, best for complex boolean logic

Cloud Manager successfully abstracts these differences through the portable `conditionalOverrides` API, allowing users to write once and deploy across all providers.
