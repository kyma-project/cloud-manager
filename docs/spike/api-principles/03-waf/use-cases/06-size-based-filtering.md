# Use Case 7: Size-Based Request Filtering

## User Story

"As a platform operator, I want to limit request body sizes on specific paths to prevent resource exhaustion attacks, block large file uploads on sensitive endpoints, and apply different size limits for different API operations based on their legitimate use cases."

## Requirements

1. Limit request body size globally or per-path
2. Limit specific header sizes
3. Limit query string length
4. Limit URI path length
5. Different size thresholds for different paths/methods
6. Block or log oversized requests

## Real-World Scenarios

### Scenario A: Prevent Upload Bombs
```yaml
# Normal API endpoints: 100 KB max body
# File upload endpoint: 10 MB max body
# GraphQL endpoint: 1 MB max body
paths:
  - /api: 100 KB
  - /api/upload: 10 MB
  - /api/graphql: 1 MB
reason: "Prevent resource exhaustion from large payloads"
```

### Scenario B: Admin Panel Protection
```yaml
# Admin operations should be small commands
path: /admin
body_size: 10 KB
reason: "Admin commands are small; large bodies indicate attack"
```

### Scenario C: Query String Attack Prevention
```yaml
# Prevent SQL injection via long query strings
query_string_length: 2048
reason: "Legitimate queries are short; long strings indicate injection attempts"
```

### Scenario D: Header Size Limits
```yaml
# Prevent header-based DoS attacks
header_size: 8 KB per header
total_headers_size: 64 KB
reason: "Prevent slowloris and header smuggling attacks"
```

## Portable API Design

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafConfiguration
metadata:
  name: size-based-policy
spec:
  managedRuleGroups:
    - type: CoreRuleSet
      action: block
    - type: SQLInjectionProtection
      action: block
  
  # Global size limits (applied to all requests)
  sizeLimits:
    maxBodySizeKB: 128              # Default: 128 KB
    maxQueryStringLengthBytes: 2048  # Default: 2048 bytes
    maxUriLengthBytes: 8192          # Default: 8192 bytes
    maxSingleHeaderSizeKB: 8         # Default: 8 KB
    maxHeadersSizeKB: 64             # Default: 64 KB total
  
  # Custom rules for path-specific size limits
  customRules:
    # Health check: minimal payload
    - name: health-check-tiny-body
      priority: 50
      action: block
      conditions:
        path:
          exact: "/health"
        bodySize:
          maxKB: 1
          exceeds: true  # Block if exceeds
      reason: "Health checks should have no body"
    
    # Admin panel: small commands only
    - name: admin-small-body
      priority: 100
      action: block
      conditions:
        path:
          prefix: "/admin"
        bodySize:
          maxKB: 10
          exceeds: true
      reason: "Admin commands are small; large body indicates attack"
    
    # GraphQL: moderate size limit
    - name: graphql-moderate-body
      priority: 110
      action: block
      conditions:
        path:
          exact: "/api/graphql"
        method: POST
        bodySize:
          maxKB: 1024  # 1 MB
          exceeds: true
      reason: "GraphQL queries should be under 1 MB"
    
    # File upload: allow larger bodies
    - name: upload-large-body-allowed
      priority: 120
      action: allow
      conditions:
        path:
          exact: "/api/upload"
        method: POST
        bodySize:
          maxKB: 10240  # 10 MB
          exceeds: false  # Allow if within limit
      reason: "Upload endpoint allows up to 10 MB"
    
    # Query string attack prevention
    - name: block-long-query-strings
      priority: 200
      action: block
      conditions:
        queryStringLength:
          maxBytes: 2048
          exceeds: true
      reason: "Long query strings indicate injection attempts"
    
    # URI length attack prevention
    - name: block-long-uris
      priority: 210
      action: block
      conditions:
        uriLength:
          maxBytes: 8192
          exceeds: true
      reason: "Excessively long URIs indicate attack"
```

## Provider Capability Check

### AWS WAFv2

**Status:** ✅ **Excellent Support**

**How it works:**
- **SizeConstraintStatement**: Match based on size of request components
- **Field types**: BODY, QUERY_STRING, URI_PATH, HEADER, SINGLE_HEADER, ALL_QUERY_ARGUMENTS
- **Comparison operators**: EQ, NE, LE, LT, GE, GT
- **Text transformations**: NONE, COMPRESS_WHITE_SPACE, HTML_ENTITY_DECODE, LOWERCASE, CMD_LINE, URL_DECODE

**Native AWS Translation:**
```json
{
  "Name": "size-based-webacl",
  "Scope": "REGIONAL",
  "DefaultAction": {"Allow": {}},
  "Rules": [
    {
      "Name": "HealthCheckNoBody",
      "Priority": 50,
      "Action": {"Block": {}},
      "Statement": {
        "AndStatement": {
          "Statements": [
            {
              "ByteMatchStatement": {
                "SearchString": "/health",
                "FieldToMatch": {"UriPath": {}},
                "PositionalConstraint": "EXACTLY",
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            },
            {
              "SizeConstraintStatement": {
                "FieldToMatch": {"Body": {}},
                "ComparisonOperator": "GT",
                "Size": 1024,
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            }
          ]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "HealthCheckNoBody"
      }
    },
    {
      "Name": "AdminSmallBody",
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
              "SizeConstraintStatement": {
                "FieldToMatch": {"Body": {}},
                "ComparisonOperator": "GT",
                "Size": 10240,
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            }
          ]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "AdminSmallBody"
      }
    },
    {
      "Name": "GraphQLModerateBody",
      "Priority": 110,
      "Action": {"Block": {}},
      "Statement": {
        "AndStatement": {
          "Statements": [
            {
              "ByteMatchStatement": {
                "SearchString": "/api/graphql",
                "FieldToMatch": {"UriPath": {}},
                "PositionalConstraint": "EXACTLY",
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            },
            {
              "SizeConstraintStatement": {
                "FieldToMatch": {"Body": {}},
                "ComparisonOperator": "GT",
                "Size": 1048576,
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            }
          ]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "GraphQLModerateBody"
      }
    },
    {
      "Name": "BlockLongQueryStrings",
      "Priority": 200,
      "Action": {"Block": {}},
      "Statement": {
        "SizeConstraintStatement": {
          "FieldToMatch": {"QueryString": {}},
          "ComparisonOperator": "GT",
          "Size": 2048,
          "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "BlockLongQueryStrings"
      }
    },
    {
      "Name": "BlockLongURIs",
      "Priority": 210,
      "Action": {"Block": {}},
      "Statement": {
        "SizeConstraintStatement": {
          "FieldToMatch": {"UriPath": {}},
          "ComparisonOperator": "GT",
          "Size": 8192,
          "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "BlockLongURIs"
      }
    }
  ]
}
```

**Key Features:**
- `SizeConstraintStatement` with multiple field types
- Comparison operators: GT, LT, GE, LE, EQ, NE
- Can check BODY, QUERY_STRING, URI_PATH, HEADER sizes
- Combine with path conditions via AndStatement

**Complexity:** Low - Excellent native support

---

### Azure Application Gateway WAF

**Status:** ⚠️ **Limited Support - Policy-Level Only**

**How it works:**
- **Global size limits** at WAF policy level (not per-rule)
- `maxRequestBodySizeInKb` (default 128 KB, max 128 KB for WAF v2)
- `requestBodyCheck` (enable/disable body inspection)
- **No per-path or per-rule size constraints**

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
          "ruleSetVersion": "2.1"
        }
      ]
    },
    "policySettings": {
      "state": "Enabled",
      "mode": "Prevention",
      "requestBodyCheck": true,
      "maxRequestBodySizeInKb": 128,
      "fileUploadLimitInMb": 100
    }
  }
}
```

**Key Limitations:**
- ❌ **No per-path body size limits** (only global)
- ❌ **No query string length constraints**
- ❌ **No URI length constraints**
- ❌ **No header size constraints** in custom rules
- ✅ Only global `maxRequestBodySizeInKb` setting

**Complexity:** High - Cannot implement path-specific size rules

**Workaround:**
- Use Azure API Management for per-API size policies
- Implement size validation at application level

---

### GCP Cloud Armor

**Status:** ⚠️ **Partial Support - Limited Size Checks**

**How it works:**
- **No direct size constraint matching in CEL expressions**
- Body size limits configured at backend service level (not in Cloud Armor)
- Can use `request.path.size()` and `request.query.size()` in CEL but not body size
- **No built-in body size checking in security policy rules**

**Native GCP Translation:**
```json
{
  "name": "size-based-security-policy",
  "rules": [
    {
      "priority": 200,
      "description": "Block long query strings",
      "action": "deny(413)",
      "match": {
        "expr": {
          "expression": "size(request.query) > 2048"
        }
      }
    },
    {
      "priority": 210,
      "description": "Block long URIs",
      "action": "deny(413)",
      "match": {
        "expr": {
          "expression": "size(request.path) > 8192"
        }
      }
    },
    {
      "priority": 2147483647,
      "description": "Default allow",
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

**Key Limitations:**
- ❌ **Cannot check request body size** in Cloud Armor rules
- ✅ Can check query string size via `size(request.query)`
- ✅ Can check URI path size via `size(request.path)`
- ❌ Cannot check header sizes
- ⚠️ Body size limits configured at backend service, not in security policy

**Complexity:** Medium - Limited to query/URI size checks

**Backend Service Configuration (separate from Cloud Armor):**
```yaml
# Backend service configuration (not Cloud Armor)
maxStreamDuration: 60s
# Body size limits configured here, not in security policy
```

---

## Cross-Provider Comparison

| Feature | AWS WAFv2 | Azure WAF | GCP Cloud Armor |
|---------|-----------|-----------|-----------------|
| **Request body size check** | ✅ Per-rule | ⚠️ Global only | ❌ Backend service only |
| **Query string size check** | ✅ Per-rule | ❌ No | ✅ CEL `size(request.query)` |
| **URI path size check** | ✅ Per-rule | ❌ No | ✅ CEL `size(request.path)` |
| **Header size check** | ✅ Per-rule | ❌ No | ⚠️ Limited |
| **Per-path size limits** | ✅ AndStatement | ❌ Global only | ⚠️ Query/URI only |
| **Comparison operators** | GT, LT, GE, LE, EQ, NE | N/A | `>`, `<`, `>=`, `<=`, `==` |
| **Global size policy** | ✅ Per-rule | ✅ Policy level | ⚠️ Backend service |
| **Complexity** | Low | High (limited) | Medium |
| **Fidelity** | Perfect | Poor | Partial |

---

## Design Decision Impact

### Recommendation: ⚠️ **Include with Provider Limitations**

**Rationale:**
1. **AWS has excellent support** - full per-rule size constraints
2. **Azure has poor support** - only global body size limit
3. **GCP has partial support** - query/URI size checks, but not body
4. **Common security need** - resource exhaustion prevention

**API Design:**

```yaml
spec:
  # Global size limits (supported on Azure, ignored on AWS/GCP in favor of per-rule)
  sizeLimits:
    maxBodySizeKB: 128
    maxQueryStringLengthBytes: 2048
    maxUriLengthBytes: 8192
  
  # Per-rule size constraints (AWS excellent, GCP partial, Azure not supported)
  customRules:
    - name: admin-small-body
      priority: 100
      action: block
      conditions:
        path: {prefix: "/admin"}
        bodySize:
          maxKB: 10
          exceeds: true
      reason: "Admin commands should be small"
```

### Status Reporting

**AWS (Perfect):**
```yaml
status:
  appliedCustomRules:
    - name: admin-small-body
      appliedStrategy: "native"
      message: "Using SizeConstraintStatement: body > 10240 bytes on /admin"
```

**Azure (Limited):**
```yaml
status:
  conditions:
    - type: "PerPathSizeLimitsNotSupported"
      status: "True"
      reason: "AzureWAFLimitation"
      message: "Azure Application Gateway WAF only supports global maxRequestBodySizeInKb (128 KB). Per-path size constraints cannot be implemented."
  
  appliedSizeLimits:
    maxBodySizeKB: 128  # Global only
```

**GCP (Partial):**
```yaml
status:
  conditions:
    - type: "BodySizeCheckNotSupported"
      status: "True"
      reason: "GCPCloudArmorLimitation"
      message: "GCP Cloud Armor cannot check request body size in security policy rules. Body size limits must be configured at backend service level. Query string and URI size checks are supported."
  
  appliedCustomRules:
    - name: block-long-query-strings
      appliedStrategy: "native"
      message: "Using CEL: size(request.query) > 2048"
```

---

## Implementation Strategy

### AWS KCP Reconciler


### Azure KCP Reconciler


### GCP KCP Reconciler


---

## Validation Matrix

| Test Case | AWS | Azure | GCP | Expected Behavior |
|-----------|-----|-------|-----|-------------------|
| POST /health with 2 KB body | ✅ Blocked | ⚠️ Allowed (no per-path) | ❌ Not checked | Health check should have no body |
| POST /admin with 20 KB body | ✅ Blocked | ⚠️ Allowed (global 128KB) | ❌ Not checked | Admin body too large |
| POST /api/graphql with 2 MB body | ✅ Blocked | ✅ Blocked (global 128KB) | ❌ Not checked | GraphQL body too large |
| GET /api with 3 KB query string | ✅ Blocked | ❌ Not checked | ✅ Blocked | Query string too long |
| GET /somepath with 10 KB URI | ✅ Blocked | ❌ Not checked | ✅ Blocked | URI too long |
| POST /api/upload with 5 MB body | ✅ Allowed (per-path rule) | ✅ Blocked (global limit) | ❌ Not checked | Upload endpoint allows large body |

---

## Conclusion

**Size-based filtering has mixed support across providers:**
- ✅ **AWS**: Excellent - full per-rule size constraints for body, query, URI, headers
- ⚠️ **Azure**: Poor - only global body size limit, no per-path constraints
- ⚠️ **GCP**: Partial - query/URI size checks, but no body size in security policy

**Recommendation:** ⚠️ **Include with Clear Limitations**

**Implementation Priority:**
1. ✅ AWS: Full support (SizeConstraintStatement)
2. ⚠️ Azure: Global size limit only, warn about per-path limitation
3. ⚠️ GCP: Query/URI size checks only, warn about body size limitation

**Real-World Use Cases:**
- Resource exhaustion prevention
- Upload bomb protection
- Query injection attack mitigation
- Buffer overflow prevention
- DoS attack mitigation

**Key Insight:** Size-based filtering is a critical security control, but Azure and GCP have significant limitations. Users targeting Azure should consider Azure API Management for per-API size policies, and GCP users should configure body size limits at backend service level.
