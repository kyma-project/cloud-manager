# Use Case 5: IP Allowlist and Blocklist

## User Story

"As a platform operator, I want to block known malicious IP addresses and ranges from accessing my application, while explicitly allowing trusted internal networks and partner IPs, to reduce attack surface and prevent abuse."

## Requirements

1. Block specific IP addresses or CIDR ranges (blocklist)
2. Allow only specific IP addresses or CIDR ranges (allowlist)
3. Combine IP rules with path conditions (e.g., block IPs only on /admin)
4. Priority ordering (allowlist before blocklist)
5. Support for both IPv4 and IPv6

## Real-World Scenarios

### Scenario A: Block Known Malicious IPs
```yaml
# Block IPs from threat intelligence feeds
# Block entire countries with high attack rates
blocklist:
  - 203.0.113.0/24      # Known botnet range
  - 198.51.100.42       # Specific attacker IP
  - 192.0.2.0/24        # Tor exit nodes
```

### Scenario B: Admin Panel IP Allowlist
```yaml
# Only allow admin access from office and VPN
path: /admin
allowlist:
  - 10.0.0.0/8          # Internal network
  - 172.16.0.0/12       # Office VPN
  - 203.0.113.50        # Admin home IP
```

### Scenario C: Partner API Access Control
```yaml
# Different partners have access to different APIs
/api/partner-a:
  allow: 198.51.100.0/24
/api/partner-b:
  allow: 203.0.113.0/24
/api/internal:
  allow: 10.0.0.0/8
```

### Scenario D: Defense in Depth
```yaml
# Priority order:
1. Allow internal network (10.0.0.0/8) everywhere
2. Block known malicious IPs everywhere
3. Allow specific partner IPs on /api/partner
4. Block all other IPs on /admin
5. Allow everyone else on public paths
```

## Portable API Design

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafPolicy
metadata:
  name: ip-access-control-policy
spec:
  managedRuleGroups:
    - type: CoreRuleSet
      action: block
    - type: SQLInjectionProtection
      action: block
  
  # Custom rules for IP-based access control
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
          cidr: "172.16.0.0/12"
      reason: "Office VPN always allowed"
    
    # Priority 100-199: Global blocklist
    - name: block-known-botnet
      priority: 100
      action: block
      conditions:
        sourceIP:
          cidr: "203.0.113.0/24"
      reason: "Known botnet range"
    
    - name: block-tor-exit-nodes
      priority: 110
      action: block
      conditions:
        sourceIP:
          cidr: "192.0.2.0/24"
      reason: "Tor exit nodes"
    
    - name: block-specific-attacker
      priority: 120
      action: block
      conditions:
        sourceIP:
          ip: "198.51.100.42"
      reason: "Specific attacker IP from incident response"
    
    # Priority 200-299: Path-specific allowlist
    - name: admin-only-from-trusted-ips
      priority: 200
      action: allow
      conditions:
        path:
          prefix: "/admin"
        sourceIP:
          anyOf:
            - cidr: "10.0.0.0/8"
            - cidr: "172.16.0.0/12"
            - ip: "203.0.113.50"
      reason: "Admin panel only accessible from trusted IPs"
    
    - name: partner-a-api-access
      priority: 210
      action: allow
      conditions:
        path:
          prefix: "/api/partner-a"
        sourceIP:
          cidr: "198.51.100.0/24"
      reason: "Partner A API access from their network"
    
    - name: partner-b-api-access
      priority: 220
      action: allow
      conditions:
        path:
          prefix: "/api/partner-b"
        sourceIP:
          cidr: "203.0.113.0/24"
      reason: "Partner B API access from their network"
    
    # Priority 300-399: Path-specific blocklist (deny all others on sensitive paths)
    - name: block-admin-from-public
      priority: 300
      action: block
      conditions:
        path:
          prefix: "/admin"
      reason: "Block all other IPs from accessing admin panel"
    
    - name: block-partner-api-from-public
      priority: 310
      action: block
      conditions:
        path:
          prefix: "/api/partner-"
      reason: "Block public access to partner APIs"
```

## Alternative API Design: Dedicated IP Lists

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafPolicy
metadata:
  name: ip-access-control-policy-v2
spec:
  # Reusable IP sets (optional syntactic sugar)
  ipSets:
    - name: internal-network
      cidrs:
        - "10.0.0.0/8"
        - "172.16.0.0/12"
      description: "Corporate internal network"
    
    - name: known-threats
      cidrs:
        - "203.0.113.0/24"
        - "192.0.2.0/24"
      ips:
        - "198.51.100.42"
      description: "Known malicious IPs from threat intel"
    
    - name: partner-a-network
      cidrs:
        - "198.51.100.0/24"
      description: "Partner A network range"
  
  customRules:
    - name: allow-internal
      priority: 10
      action: allow
      conditions:
        sourceIP:
          ipSetRef: internal-network
    
    - name: block-threats
      priority: 100
      action: block
      conditions:
        sourceIP:
          ipSetRef: known-threats
    
    - name: admin-access-control
      priority: 200
      action: allow
      conditions:
        path: {prefix: "/admin"}
        sourceIP:
          ipSetRef: internal-network
```

## Provider Capability Check

### AWS WAFv2

**Status:** ✅ **Perfect Support**

**How it works:**
- **IPSet resources**: Create reusable IP sets, reference in rules
- **IPSetReferenceStatement**: Match against IP set
- **Supports IPv4 and IPv6**
- **Up to 10,000 IP addresses per IP set**
- **Can update IP sets independently from WebACL**

**Native AWS Translation:**
```json
{
  "IPSets": [
    {
      "Name": "InternalNetwork",
      "Scope": "REGIONAL",
      "IPAddressVersion": "IPV4",
      "Addresses": [
        "10.0.0.0/8",
        "172.16.0.0/12"
      ]
    },
    {
      "Name": "KnownThreats",
      "Scope": "REGIONAL",
      "IPAddressVersion": "IPV4",
      "Addresses": [
        "203.0.113.0/24",
        "192.0.2.0/24",
        "198.51.100.42/32"
      ]
    }
  ],
  "WebACL": {
    "Name": "ip-access-control-webacl",
    "Scope": "REGIONAL",
    "DefaultAction": {"Allow": {}},
    "Rules": [
      {
        "Name": "AllowInternalNetwork",
        "Priority": 10,
        "Action": {"Allow": {}},
        "Statement": {
          "IPSetReferenceStatement": {
            "Arn": "arn:aws:wafv2:region:account:regional/ipset/InternalNetwork/..."
          }
        },
        "VisibilityConfig": {
          "SampledRequestsEnabled": true,
          "CloudWatchMetricsEnabled": true,
          "MetricName": "AllowInternalNetwork"
        }
      },
      {
        "Name": "BlockKnownThreats",
        "Priority": 100,
        "Action": {"Block": {}},
        "Statement": {
          "IPSetReferenceStatement": {
            "Arn": "arn:aws:wafv2:region:account:regional/ipset/KnownThreats/..."
          }
        },
        "VisibilityConfig": {
          "SampledRequestsEnabled": true,
          "CloudWatchMetricsEnabled": true,
          "MetricName": "BlockKnownThreats"
        }
      },
      {
        "Name": "AdminAccessControl",
        "Priority": 200,
        "Action": {"Allow": {}},
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
                "IPSetReferenceStatement": {
                  "Arn": "arn:aws:wafv2:region:account:regional/ipset/InternalNetwork/..."
                }
              }
            ]
          }
        },
        "VisibilityConfig": {
          "SampledRequestsEnabled": true,
          "CloudWatchMetricsEnabled": true,
          "MetricName": "AdminAccessControl"
        }
      },
      {
        "Name": "BlockAdminFromPublic",
        "Priority": 300,
        "Action": {"Block": {}},
        "Statement": {
          "ByteMatchStatement": {
            "SearchString": "/admin",
            "FieldToMatch": {"UriPath": {}},
            "PositionalConstraint": "STARTS_WITH",
            "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
          }
        },
        "VisibilityConfig": {
          "SampledRequestsEnabled": true,
          "CloudWatchMetricsEnabled": true,
          "MetricName": "BlockAdminFromPublic"
        }
      }
    ]
  }
}
```

**Key Features:**
- Separate IPSet resources (lifecycle independent from WebACL)
- Can update IP sets without modifying WebACL
- Support for very large IP lists (10K addresses)
- IPv4 and IPv6 support

**Complexity:** Low - Excellent native support

---

### Azure Application Gateway WAF

**Status:** ✅ **Perfect Support**

**How it works:**
- **IPMatch operator** in custom rules
- **Inline IP addresses/CIDRs** in matchValues
- **Supports IPv4 and IPv6**
- **No separate IP set resource** (IPs inline in rules)

**Native Azure Translation:**
```json
{
  "location": "eastus",
  "properties": {
    "customRules": [
      {
        "name": "AllowInternalNetwork",
        "priority": 10,
        "ruleType": "MatchRule",
        "action": "Allow",
        "matchConditions": [
          {
            "matchVariables": [
              {"variableName": "RemoteAddr"}
            ],
            "operator": "IPMatch",
            "matchValues": [
              "10.0.0.0/8",
              "172.16.0.0/12"
            ],
            "negationConditon": false
          }
        ]
      },
      {
        "name": "BlockKnownThreats",
        "priority": 100,
        "ruleType": "MatchRule",
        "action": "Block",
        "matchConditions": [
          {
            "matchVariables": [
              {"variableName": "RemoteAddr"}
            ],
            "operator": "IPMatch",
            "matchValues": [
              "203.0.113.0/24",
              "192.0.2.0/24",
              "198.51.100.42/32"
            ],
            "negationConditon": false
          }
        ]
      },
      {
        "name": "AdminAccessControl",
        "priority": 200,
        "ruleType": "MatchRule",
        "action": "Allow",
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
              {"variableName": "RemoteAddr"}
            ],
            "operator": "IPMatch",
            "matchValues": [
              "10.0.0.0/8",
              "172.16.0.0/12",
              "203.0.113.50/32"
            ],
            "negationConditon": false
          }
        ]
      },
      {
        "name": "BlockAdminFromPublic",
        "priority": 300,
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
    },
    "policySettings": {
      "state": "Enabled",
      "mode": "Prevention"
    }
  }
}
```

**Key Features:**
- Simple IPMatch operator
- Multiple IPs/CIDRs in single matchCondition (OR logic)
- Can combine with path, header, etc. (AND logic)
- IPv4 and IPv6 support

**Limitation:**
- No separate IP set resource (IPs duplicated across rules)
- Updating IP lists requires updating policy

**Complexity:** Low - Simple and direct

---

### GCP Cloud Armor

**Status:** ✅ **Perfect Support**

**How it works:**
- **CEL expressions**: `inIpRange(origin.ip, 'CIDR')` or `origin.ip == 'IP'`
- **Multiple IPs via OR**: `inIpRange(origin.ip, 'CIDR1') || inIpRange(origin.ip, 'CIDR2')`
- **Supports IPv4 and IPv6**
- **Preconfigured IP lists**: Regional IP lists available

**Native GCP Translation:**
```json
{
  "name": "ip-access-control-security-policy",
  "rules": [
    {
      "priority": 10,
      "description": "Allow internal network",
      "action": "allow",
      "match": {
        "expr": {
          "expression": "inIpRange(origin.ip, '10.0.0.0/8') || inIpRange(origin.ip, '172.16.0.0/12')"
        }
      }
    },
    {
      "priority": 100,
      "description": "Block known threats",
      "action": "deny(403)",
      "match": {
        "expr": {
          "expression": "inIpRange(origin.ip, '203.0.113.0/24') || inIpRange(origin.ip, '192.0.2.0/24') || origin.ip == '198.51.100.42'"
        }
      }
    },
    {
      "priority": 200,
      "description": "Admin access control - only internal network",
      "action": "allow",
      "match": {
        "expr": {
          "expression": "request.path.matches('/admin.*') && (inIpRange(origin.ip, '10.0.0.0/8') || inIpRange(origin.ip, '172.16.0.0/12') || origin.ip == '203.0.113.50')"
        }
      }
    },
    {
      "priority": 300,
      "description": "Block admin from public",
      "action": "deny(403)",
      "match": {
        "expr": {
          "expression": "request.path.matches('/admin.*')"
        }
      }
    },
    {
      "priority": 1000,
      "description": "OWASP ModSecurity CRS",
      "action": "deny(403)",
      "match": {
        "expr": {
          "expression": "evaluatePreconfiguredWaf('owasp-crs-v030301-id', {'sensitivity': 1})"
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

**Key Features:**
- Very flexible CEL expressions
- Easy to combine multiple IP ranges with `||`
- Can reference preconfigured regional IP lists
- Clean integration with other conditions

**Complexity:** Low - CEL is very expressive

---

## Cross-Provider Comparison

| Feature | AWS WAFv2 | Azure WAF | GCP Cloud Armor |
|---------|-----------|-----------|-----------------|
| **IP allowlist/blocklist** | ✅ Perfect | ✅ Perfect | ✅ Perfect |
| **CIDR support** | ✅ Yes | ✅ Yes | ✅ Yes |
| **Individual IP support** | ✅ Yes | ✅ Yes | ✅ Yes |
| **IPv4 support** | ✅ Yes | ✅ Yes | ✅ Yes |
| **IPv6 support** | ✅ Yes | ✅ Yes | ✅ Yes |
| **Reusable IP sets** | ✅ IPSet resource | ❌ Inline only | ❌ Inline only |
| **Max IPs per rule** | 10,000 (IPSet) | ~100 (inline) | ~100 (CEL length) |
| **Path + IP combination** | ✅ AndStatement | ✅ Multiple matchConditions | ✅ CEL && |
| **Update without policy change** | ✅ Update IPSet | ❌ Update policy | ❌ Update policy |
| **Complexity** | Low | Low | Low |
| **Fidelity** | Perfect | Perfect | Perfect |

---

## Design Decision Impact

### API Design - Approach A: Inline IPs (Simpler)

```yaml
spec:
  customRules:
    - name: allow-internal
      priority: 10
      action: allow
      conditions:
        sourceIP:
          anyOf:
            - cidr: "10.0.0.0/8"
            - cidr: "172.16.0.0/12"
            - ip: "203.0.113.50"
```

**Pros:**
- Simple, works everywhere
- No new concepts

**Cons:**
- IP duplication across rules
- Updating IPs requires policy update

### API Design - Approach B: Reusable IP Sets (Better for large lists)

```yaml
spec:
  ipSets:
    - name: internal-network
      cidrs: ["10.0.0.0/8", "172.16.0.0/12"]
      ips: ["203.0.113.50"]
  
  customRules:
    - name: allow-internal
      priority: 10
      action: allow
      conditions:
        sourceIP:
          ipSetRef: internal-network
```

**Pros:**
- DRY principle (define once, reference many times)
- AWS can map to native IPSet resources
- Easier to update large lists

**Cons:**
- Azure/GCP must expand inline (no native IP set concept)

### Recommendation: ✅ **Support Both Approaches**

```yaml
spec:
  # Optional: Reusable IP sets
  ipSets:
    - name: internal-network
      cidrs: ["10.0.0.0/8"]
      description: "Corporate network"
  
  customRules:
    # Inline approach
    - name: block-specific-threat
      priority: 100
      action: block
      conditions:
        sourceIP:
          ip: "198.51.100.42"
    
    # IPSet reference approach
    - name: allow-internal
      priority: 10
      action: allow
      conditions:
        sourceIP:
          ipSetRef: internal-network
```

**Rationale:**
- Inline IPs: Simple, works great for small lists (1-10 IPs)
- IPSet references: Better for large lists (100+ IPs), threat intel feeds
- AWS can optimize with native IPSet resources
- Azure/GCP expand IPSet refs to inline IPs

---

## Implementation Strategy

### AWS KCP Reconciler


### Azure KCP Reconciler


### GCP KCP Reconciler


---

## Validation Matrix

| Test Case | AWS | Azure | GCP | Expected Behavior |
|-----------|-----|-------|-----|-------------------|
| Request from 10.0.0.5 | ✅ Allowed (internal) | ✅ Allowed | ✅ Allowed | Internal network allowed globally |
| Request from 203.0.113.50 to /admin | ✅ Allowed | ✅ Allowed | ✅ Allowed | Trusted admin IP |
| Request from 198.51.100.42 | ✅ Blocked | ✅ Blocked | ✅ Blocked | Known threat blocked |
| Request from 203.0.113.50 to /api | ✅ Allowed | ✅ Allowed | ✅ Allowed | Public path allowed |
| Request from public IP to /admin | ✅ Blocked | ✅ Blocked | ✅ Blocked | Admin protected |
| Request from 198.51.100.0/24 (Partner A) to /api/partner-a | ✅ Allowed | ✅ Allowed | ✅ Allowed | Partner access granted |

---

## Conclusion

**IP allowlist and blocklist work perfectly across all three providers.** This is one of the most universally supported WAF features.

**Recommendation:** ✅ **High Priority - Implement Early**

**Key Features to Support:**
1. ✅ Inline IP/CIDR matching (all providers)
2. ✅ Reusable IP sets (AWS native, expand for Azure/GCP)
3. ✅ Combine IP + path conditions (all providers)
4. ✅ IPv4 and IPv6 support (all providers)
5. ✅ Priority-based evaluation (allowlist before blocklist)

**Implementation Priority:**
- **Phase 1:** Inline IP matching in customRules
- **Phase 2:** Reusable ipSets with references
- **Phase 3:** Integration with external threat intelligence feeds

This is a foundational security feature that should be prioritized alongside managed rule groups and custom rules!
