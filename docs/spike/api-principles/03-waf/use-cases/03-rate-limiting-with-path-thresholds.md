# Use Case 4: Rate Limiting with Path-Based Thresholds

## User Story

"As a platform operator, I want to apply different rate limits to different paths - strict limits on expensive operations like /api/search and /api/export, but more permissive limits on cheap operations like /health and /api/users, to prevent abuse while maintaining good user experience."

## Requirements

1. Enable rate limiting for different paths
2. Apply different rate thresholds per path
3. Different actions per rate limit (block vs count)
4. Aggregate by source IP or other keys

## Real-World Scenarios

### Scenario A: Tiered API Rate Limits
- Expensive operations: 10 req/min
- Standard operations: 100 req/min  
- Health checks: No rate limit

### Scenario B: Authenticated vs Unauthenticated
- Unauthenticated: 10 req/min globally
- With API key: 100 req/min
- With JWT token: 1000 req/min

## Portable API Design

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafConfiguration
metadata:
  name: tiered-rate-limit-policy
spec:
  # Start from preset that includes managed rules
  basePolicyRef:
    name: owasp-moderate
  
  customRules:
    - name: health-check-no-rate-limit
      priority: 50
      action: allow
      conditions:
        path:
          exact: "/health"
      reason: "Health checks bypass rate limiting"
    
    - name: expensive-search-low-limit
      priority: 100
      action: block
      rateLimit:
        requestsPerMinute: 10
        aggregateKeyType: IP
      conditions:
        path:
          prefix: "/api/search"
      reason: "Search API limited to 10 req/min per IP"
    
    - name: standard-api-medium-limit
      priority: 200
      action: block
      rateLimit:
        requestsPerMinute: 100
        aggregateKeyType: IP
      conditions:
        path:
          prefix: "/api"
      reason: "Standard API limited to 100 req/min per IP"
```

## Provider Capability Check

### AWS WAFv2

**Status:** ✅ **Perfect Support**

AWS supports custom rate-based rules with:
- `RateBasedStatement` with configurable limits (per 5 minutes)
- `AggregateKeyType`: IP, FORWARDED_IP, CUSTOM_KEYS
- `ScopeDownStatement` for path/condition matching

**Complexity:** Low - Native support
**Note:** AWS limits are per 5 minutes, so multiply by 5

---

### Azure Application Gateway WAF

**Status:** ❌ **Not Supported**

Azure Application Gateway WAF does **NOT** support:
- Path-based rate limiting
- Conditional rate limiting
- Per-rule rate limiting

**Workarounds:**
1. Use Azure API Management (full rate limiting support)
2. Use Azure Front Door (better rate limiting than App Gateway)
3. Implement rate limiting at application level

**Complexity:** High - Cannot implement in WAF

**Critical:** Status must report "not supported" for Azure

---

### GCP Cloud Armor

**Status:** ✅ **Perfect Support**

GCP supports rate limiting with:
- `rate_based_ban` or `throttle` actions
- `rateLimitOptions` with flexible thresholds
- `enforceOnKey`: IP, HEADER, COOKIE, PATH, etc.
- Configurable ban duration (rate_based_ban only)

**Complexity:** Low - Excellent native support

---

## Cross-Provider Comparison

| Feature | AWS WAFv2 | Azure WAF | GCP Cloud Armor |
|---------|-----------|-----------|-----------------|
| **Path-based rate limiting** | ✅ Yes | ❌ No | ✅ Yes |
| **Custom thresholds** | ✅ Per rule | ❌ Global only | ✅ Per rule |
| **Time window** | 5 minutes | N/A | Configurable |
| **Aggregate keys** | IP, FORWARDED_IP, CUSTOM | N/A | IP, HEADER, COOKIE, etc. |
| **Condition-based** | ✅ ScopeDownStatement | ❌ No | ✅ CEL expressions |
| **Fidelity** | Perfect | ❌ Not implementable | Perfect |

---

## Design Decision Impact

### Recommendation: ⚠️ **Include with Azure Warning**

**Rationale:**
1. AWS and GCP have perfect support (2/3 providers)
2. Azure fundamentally cannot implement this in WAF
3. Common use case for API protection
4. Users must understand Azure limitation via status reporting

**API Design:**
```yaml
spec:
  customRules:
    - name: rate-limit-rule
      priority: 100
      action: block
      rateLimit:
        requestsPerMinute: 10
        aggregateKeyType: IP
        banDurationSeconds: 600    # GCP only
        exceedAction: deny429      # Response code
      conditions:
        path: {prefix: "/api/search"}
```

### Status Reporting

**AWS (Perfect):**
```yaml
status:
  appliedCustomRules:
    - name: search-rate-limit
      appliedStrategy: "native"
      message: "Using RateBasedStatement with limit 50 per 5 minutes"
```

**Azure (Not Supported):**
```yaml
status:
  conditions:
    - type: "RateLimitingNotSupported"
      status: "True"
      reason: "AzureWAFLimitation"
      message: "Azure Application Gateway WAF does not support path-based rate limiting. Use Azure API Management or Front Door instead."
```

**GCP (Perfect):**
```yaml
status:
  appliedCustomRules:
    - name: search-rate-limit
      appliedStrategy: "native"
      message: "Using rate_based_ban with threshold 10 per 60 seconds"
```

---

## Implementation Strategy

**AWS KCP Reconciler:**

**Azure KCP Reconciler:**

**GCP KCP Reconciler:**

---

## Validation Matrix

| Test Case | AWS | Azure | GCP | Expected Behavior |
|-----------|-----|-------|-----|-------------------|
| /api/search: 11 req/min | ✅ Blocked | ❌ Not supported | ✅ Blocked + banned | Rate limit enforced |
| /api/users: 101 req/min | ✅ Blocked | ❌ Not supported | ✅ Throttled | Higher limit enforced |
| /health: unlimited | ✅ Allowed | ✅ Allowed | ✅ Allowed | Bypass works |
| Multiple IPs under limit | ✅ Allowed | N/A | ✅ Allowed | Per-IP tracking |

---

## Conclusion

**Path-based rate limiting works on AWS and GCP, but NOT on Azure WAF.**

**Implementation Priority:**
1. ✅ AWS: Full support (convert req/min → req/5min)
2. ✅ GCP: Full support (use rate_based_ban or throttle)
3. ⚠️ Azure: Status warning only

This is valuable despite Azure limitation because AWS and GCP are the most capable WAF providers and rate limiting is critical for API protection.
