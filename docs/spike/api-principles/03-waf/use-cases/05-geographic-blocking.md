# Use Case 6: Geographic Blocking and Allowlisting

## User Story

"As a platform operator, I want to block or allow traffic based on geographic location (country/region), to comply with data residency requirements, reduce attack surface from high-risk regions, or restrict service availability to specific markets."

## Requirements

1. Block traffic from specific countries
2. Allow traffic only from specific countries
3. Combine geographic rules with path conditions
4. Support for country codes (ISO 3166-1 alpha-2)
5. Handle unknown/unresolvable geographic locations

## Real-World Scenarios

### Scenario A: GDPR Compliance - EU Only Service
```yaml
# SaaS product only available in EU
# Block all non-EU traffic
allowlist:
  - EU countries only
  - Block all others
reason: "GDPR compliance - service only for EU residents"
```

### Scenario B: High-Risk Region Blocking
```yaml
# Block traffic from countries with high attack rates
blocklist:
  - CN (China)
  - RU (Russia)
  - KP (North Korea)
reason: "99% of attacks originate from these regions"
```

### Scenario C: Market-Specific APIs
```yaml
# Different APIs for different markets
/api/us: Allow only US
/api/eu: Allow only EU countries
/api/apac: Allow only APAC countries
```

### Scenario D: Admin Panel Geographic Lock
```yaml
# Admin panel only from headquarters country
path: /admin
allow: US only
reason: "Admin operations only from US headquarters"
```

## Portable API Design

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafConfiguration
metadata:
  name: geographic-policy
spec:
  managedRuleGroups:
    - type: CoreRuleSet
      action: block
  
  customRules:
    # Scenario A: EU-only service
    - name: allow-eu-only
      priority: 10
      action: allow
      conditions:
        geographic:
          countries:
            - AT  # Austria
            - BE  # Belgium
            - DE  # Germany
            - FR  # France
            - IT  # Italy
            - NL  # Netherlands
            # ... all EU countries
      reason: "GDPR compliance - EU only service"
    
    - name: block-non-eu
      priority: 20
      action: block
      conditions: {}  # Match all
      reason: "Block all non-EU traffic"
    
    # Scenario B: Block high-risk regions
    - name: block-high-risk-countries
      priority: 100
      action: block
      conditions:
        geographic:
          countries:
            - CN  # China
            - RU  # Russia
            - KP  # North Korea
      reason: "Block traffic from high-risk regions"
    
    # Scenario C: Market-specific APIs
    - name: us-api-only-us-traffic
      priority: 200
      action: allow
      conditions:
        path:
          prefix: "/api/us"
        geographic:
          countries:
            - US
      reason: "US API only accessible from US"
    
    - name: block-us-api-from-non-us
      priority: 210
      action: block
      conditions:
        path:
          prefix: "/api/us"
      reason: "Block non-US traffic to US API"
    
    - name: eu-api-only-eu-traffic
      priority: 220
      action: allow
      conditions:
        path:
          prefix: "/api/eu"
        geographic:
          countries:
            - AT
            - BE
            - DE
            - FR
            # ... EU countries
      reason: "EU API only accessible from EU"
    
    # Scenario D: Admin geographic lock
    - name: admin-only-from-us
      priority: 300
      action: allow
      conditions:
        path:
          prefix: "/admin"
        geographic:
          countries:
            - US
      reason: "Admin panel only from US headquarters"
    
    - name: block-admin-from-non-us
      priority: 310
      action: block
      conditions:
        path:
          prefix: "/admin"
      reason: "Block non-US access to admin"
```

## Alternative API Design: Geographic Groups

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafConfiguration
metadata:
  name: geographic-policy-v2
spec:
  # Reusable geographic groups
  geographicGroups:
    - name: eu-countries
      countries:
        - AT
        - BE
        - DE
        - FR
        - IT
        # ... all EU countries
      description: "European Union member states"
    
    - name: high-risk-countries
      countries:
        - CN
        - RU
        - KP
      description: "Countries with high attack rates"
  
  customRules:
    - name: allow-eu-only
      priority: 10
      action: allow
      conditions:
        geographic:
          groupRef: eu-countries
    
    - name: block-high-risk
      priority: 100
      action: block
      conditions:
        geographic:
          groupRef: high-risk-countries
```

## Provider Capability Check

### AWS WAFv2

**Status:** ✅ **Perfect Support**

**How it works:**
- **GeoMatchStatement**: Native geographic matching
- **ISO 3166-1 alpha-2 country codes**
- **Supports all countries and territories**
- **Can combine with ForwardedIPConfig** for X-Forwarded-For header

**Native AWS Translation:**
```json
{
  "Name": "geographic-webacl",
  "Scope": "REGIONAL",
  "DefaultAction": {"Allow": {}},
  "Rules": [
    {
      "Name": "AllowEUOnly",
      "Priority": 10,
      "Action": {"Allow": {}},
      "Statement": {
        "GeoMatchStatement": {
          "CountryCodes": [
            "AT",
            "BE",
            "DE",
            "FR",
            "IT",
            "NL"
          ]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "AllowEUOnly"
      }
    },
    {
      "Name": "BlockNonEU",
      "Priority": 20,
      "Action": {"Block": {}},
      "Statement": {
        "NotStatement": {
          "Statement": {
            "GeoMatchStatement": {
              "CountryCodes": ["AT", "BE", "DE", "FR", "IT", "NL"]
            }
          }
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "BlockNonEU"
      }
    },
    {
      "Name": "BlockHighRiskCountries",
      "Priority": 100,
      "Action": {"Block": {}},
      "Statement": {
        "GeoMatchStatement": {
          "CountryCodes": ["CN", "RU", "KP"]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "BlockHighRiskCountries"
      }
    },
    {
      "Name": "USApiOnlyUS",
      "Priority": 200,
      "Action": {"Allow": {}},
      "Statement": {
        "AndStatement": {
          "Statements": [
            {
              "ByteMatchStatement": {
                "SearchString": "/api/us",
                "FieldToMatch": {"UriPath": {}},
                "PositionalConstraint": "STARTS_WITH",
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            },
            {
              "GeoMatchStatement": {
                "CountryCodes": ["US"]
              }
            }
          ]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "USApiOnlyUS"
      }
    }
  ]
}
```

**Key Features:**
- Native `GeoMatchStatement`
- Support for `ForwardedIPConfig` (CloudFront, ALB)
- Can negate with `NotStatement`
- Combine with other conditions via `AndStatement`

**Complexity:** Low - Excellent native support

---

### Azure Application Gateway WAF

**Status:** ❌ **Not Supported**

**How it works:**
- **Azure Application Gateway WAF v2 does NOT support geographic matching**
- **Azure Front Door WAF has geo-filtering** (different product)

**Workarounds:**
1. **Use Azure Front Door** instead of Application Gateway (has geo-filtering rules)
2. **Use Azure Firewall** in front of Application Gateway (has geo-based rules)
3. **Implement at application level** with IP geolocation libraries

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
    }
  }
}
```

**Complexity:** High - **Cannot implement geographic rules in Application Gateway WAF**

**Critical Limitation:**
- ❌ No geographic matching capability
- ⚠️ Must use different Azure product (Front Door) or architecture

---

### GCP Cloud Armor

**Status:** ✅ **Perfect Support**

**How it works:**
- **CEL expression**: `origin.region_code` variable
- **ISO 3166-1 alpha-2 country codes**
- **Supports all countries**
- **Very flexible with CEL operators**

**Native GCP Translation:**
```json
{
  "name": "geographic-security-policy",
  "rules": [
    {
      "priority": 10,
      "description": "Allow EU countries only",
      "action": "allow",
      "match": {
        "expr": {
          "expression": "origin.region_code in ['AT', 'BE', 'DE', 'FR', 'IT', 'NL']"
        }
      }
    },
    {
      "priority": 20,
      "description": "Block non-EU traffic",
      "action": "deny(403)",
      "match": {
        "expr": {
          "expression": "!(origin.region_code in ['AT', 'BE', 'DE', 'FR', 'IT', 'NL'])"
        }
      }
    },
    {
      "priority": 100,
      "description": "Block high-risk countries",
      "action": "deny(403)",
      "match": {
        "expr": {
          "expression": "origin.region_code in ['CN', 'RU', 'KP']"
        }
      }
    },
    {
      "priority": 200,
      "description": "US API only from US",
      "action": "allow",
      "match": {
        "expr": {
          "expression": "request.path.matches('/api/us.*') && origin.region_code == 'US'"
        }
      }
    },
    {
      "priority": 210,
      "description": "Block non-US access to US API",
      "action": "deny(403)",
      "match": {
        "expr": {
          "expression": "request.path.matches('/api/us.*') && origin.region_code != 'US'"
        }
      }
    },
    {
      "priority": 300,
      "description": "Admin only from US",
      "action": "allow",
      "match": {
        "expr": {
          "expression": "request.path.matches('/admin.*') && origin.region_code == 'US'"
        }
      }
    },
    {
      "priority": 310,
      "description": "Block non-US access to admin",
      "action": "deny(403)",
      "match": {
        "expr": {
          "expression": "request.path.matches('/admin.*')"
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

**Key Features:**
- `origin.region_code` CEL variable
- Clean `in` operator for multiple countries
- Easy negation with `!`
- Combine with path/IP conditions naturally

**Complexity:** Low - CEL is very expressive

---

## Cross-Provider Comparison

| Feature | AWS WAFv2 | Azure WAF | GCP Cloud Armor |
|---------|-----------|-----------|-----------------|
| **Geographic matching** | ✅ Perfect | ❌ Not supported | ✅ Perfect |
| **Country codes** | ISO 3166-1 alpha-2 | N/A | ISO 3166-1 alpha-2 |
| **Multiple countries** | ✅ Array in GeoMatchStatement | N/A | ✅ CEL `in` operator |
| **Negation (NOT country)** | ✅ NotStatement | N/A | ✅ CEL `!` or `!=` |
| **Path + geo combination** | ✅ AndStatement | N/A | ✅ CEL `&&` |
| **Forwarded IP support** | ✅ ForwardedIPConfig | N/A | ✅ Automatic |
| **Complexity** | Low | High (not supported) | Low |
| **Fidelity** | Perfect | ❌ Cannot implement | Perfect |

---

## Design Decision Impact

### Recommendation: ⚠️ **Include with Clear Azure Limitation**

**Rationale:**
1. **AWS and GCP have perfect support** (2 out of 3 providers)
2. **Azure Application Gateway WAF fundamentally does not support geographic rules**
3. **Common compliance use case** (GDPR, data residency, export controls)
4. **Users must understand Azure requires different architecture** (Front Door, not App Gateway)

**API Design:**
```yaml
spec:
  customRules:
    - name: allow-eu-only
      priority: 10
      action: allow
      conditions:
        geographic:
          countries:
            - DE
            - FR
            - IT
          # OR use groupRef
          # groupRef: eu-countries
      reason: "GDPR compliance"
```

### Status Reporting

**AWS (Perfect):**
```yaml
status:
  appliedCustomRules:
    - name: allow-eu-only
      appliedStrategy: "native"
      message: "Using GeoMatchStatement with countries: DE, FR, IT"
```

**Azure (Not Supported):**
```yaml
status:
  conditions:
    - type: "GeographicRulesNotSupported"
      status: "True"
      reason: "AzureApplicationGatewayWAFLimitation"
      message: "Azure Application Gateway WAF does not support geographic filtering. For geographic rules, use Azure Front Door instead of Application Gateway, or implement geo-filtering at application level."
  
  appliedCustomRules:
    - name: allow-eu-only
      appliedStrategy: "not-supported"
      message: "Geographic rules not supported on Azure Application Gateway WAF. Use Azure Front Door for geographic filtering."
```

**GCP (Perfect):**
```yaml
status:
  appliedCustomRules:
    - name: allow-eu-only
      appliedStrategy: "native"
      message: "Using CEL expression: origin.region_code in ['DE', 'FR', 'IT']"
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
| Request from DE (Germany) | ✅ Allowed (EU) | ❌ Not supported | ✅ Allowed | EU traffic allowed |
| Request from US to /api/eu | ✅ Blocked | ❌ Not supported | ✅ Blocked | Non-EU to EU API blocked |
| Request from CN (China) | ✅ Blocked | ❌ Not supported | ✅ Blocked | High-risk country blocked |
| Request from US to /admin | ✅ Allowed | ❌ Not supported | ✅ Allowed | US admin access allowed |
| Request from FR to /admin | ✅ Blocked | ❌ Not supported | ✅ Blocked | Non-US admin blocked |

---

## Conclusion

**Geographic filtering works on AWS and GCP, but NOT on Azure Application Gateway WAF.**

**Recommendation:** ⚠️ **Include with Clear Azure Limitation**

**Required Documentation:**
```yaml
# ⚠️ Azure Application Gateway WAF Limitation
# 
# Geographic filtering is NOT supported by Azure Application Gateway WAF.
# 
# For geographic rules on Azure, you must use:
# 1. Azure Front Door (has geo-filtering support)
# 2. Azure Firewall (network-level geo-blocking)
# 3. Application-level geolocation libraries
#
# WafPolicy with geographic rules will report "not supported" on Azure
# Application Gateway. Consider using Azure Front Door instead.
```

**Implementation Priority:**
1. ✅ AWS: Full support (GeoMatchStatement)
2. ✅ GCP: Full support (CEL origin.region_code)
3. ⚠️ Azure: Status warning only, do not create rules

**Real-World Use Cases:**
- GDPR compliance (EU-only services)
- Export control compliance (block sanctioned countries)
- Attack surface reduction (block high-risk regions)
- Market-specific service availability
- Data residency requirements

This is a critical compliance feature despite Azure limitation, because geographic restrictions are often legal requirements, not optional preferences.
