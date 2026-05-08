# AWS WebACL Nesting Depth Decision

## Executive Summary

We need to decide on the maximum nesting depth for logical operators (AND/OR/NOT) in AWS WebACL rules. This decision impacts CRD size, API complexity, and user experience.

**Key Question:** Given that label-based rule chaining solves all complex nesting scenarios, do we need more than `[AND/OR/NOT](LeafStatement)` - i.e., just 2 levels (Statement → LeafStatement)?

## Background

AWS WAF API supports **unlimited nesting** of logical operators. However, we must limit nesting depth due to Kubernetes CRD size constraints (1.5MB etcd limit).

### AWS Console Behavior

AWS Console supports **up to 2 levels of logical operators** (AND/OR at root, then statements with optional NOT checkbox):
- ✅ Supports: `AND(Statement1, Statement2)` with NOT checkbox per statement
- ✅ Supports: `OR(Statement1, Statement2)` with NOT checkbox per statement
- ✅ Supports: `AND(NOT(X), Y)` - NOT as checkbox on statement X
- ❌ Deeper nesting (e.g., `AND(OR(X, Y), NOT(Z))`) shows "View rule in JSON" (not editable in UI)

This means the console effectively supports 3-level nesting when NOT is applied via checkbox to leaf statements within AND/OR.

## Nesting Level Options

### Option 1: Current Implementation - 3 Levels (Statement → Statement1 → Statement2)

**Structure:**
```
Level 0: AND/OR/NOT/RateBased/ManagedRuleGroup/Leaf
Level 1: NOT/Leaf (no AND/OR)
Level 2: Leaf only
```

**Example:**
```yaml
statement:
  andStatement:           # Level 0
    statements:
      - notStatement:     # Level 1
          statement:
            byteMatch:    # Level 2 (leaf)
              searchString: "/admin"
              fieldToMatch:
                uriPath: {}
      - geoMatch:         # Level 1 (leaf)
          countryCodes: ["US"]
```

**Pros:**
- ✅ Allows moderately complex nesting
- ✅ Supports `AND(NOT(X), Y)` patterns directly

**Cons:**
- ❌ Large CRD: ~13k lines, 867KB
- ❌ Complex to understand and validate
- ❌ Customers will use label-based chaining anyway for real complexity
- ❌ Limited headroom for future features (logging, etc.)

**AWS Console:** Supports NOT checkbox within AND/OR (3-level nesting)

---

### Option 2: Recommended - 2 Levels (Statement → LeafStatement)

**Structure:**
```
Level 0: AND/OR/NOT/RateBased/ManagedRuleGroup/Leaf
Level 1: Leaf only
```

**Example:**
```yaml
statement:
  andStatement:      # Level 0
    statements:
      - byteMatch:   # Level 1 (leaf)
          searchString: "/api"
          fieldToMatch:
            uriPath: {}
      - geoMatch:    # Level 1 (leaf)
          countryCodes: ["US"]

# NOT pattern becomes:
statement:
  notStatement:      # Level 0
    statement:
      geoMatch:      # Level 1 (leaf)
        countryCodes: ["CN"]
```

**What's Lost:**
Cannot express `AND(NOT(X), Y)` or `OR(NOT(X), Y)` directly - must use label-based chaining (see workaround below).

**Pros:**
- ✅ Significantly smaller CRD: ~9-10k lines, ~600KB (30% reduction)
- ✅ Simpler API - easier to understand and validate
- ✅ Covers most common use cases
- ✅ More headroom for future features
- ✅ **Non-breaking to add Statement2 later if needed**

**Cons:**
- ❌ Cannot nest NOT inside AND/OR (must use label-based chaining)

---

## Label-Based Rule Chaining: The Solution to Complexity

**Key Insight:** Label-based chaining can express **ANY** complex logic, regardless of nesting depth.

### How It Works

Instead of nesting statements deeply, split logic across multiple rules:

1. **First rule(s)**: Match conditions and apply labels (use `count` action)
2. **Final rule**: Check labels and take action

### Example: Complex Nested Logic

**What we want to express:**
```
Block if: (path = "/api" AND NOT from US) OR (path = "/admin" AND from China)
```

**Option A: Deep Nesting (would require 3+ levels)**
```yaml
# ❌ This would require deep nesting we don't support:
statement:
  orStatement:
    statements:
      - andStatement:
          statements:
            - byteMatch: {searchString: "/api"}
            - notStatement: {statement: {geoMatch: {countryCodes: ["US"]}}}
      - andStatement:
          statements:
            - byteMatch: {searchString: "/admin"}
            - geoMatch: {countryCodes: ["CN"]}
```

**Option B: Label-Based Chaining (works with ANY nesting level)**
```yaml
rules:
  # Rule 1: Label requests not from US
  - name: label-not-us
    priority: 10
    action:
      count: {}
    statement:
      notStatement:
        statement:
          geoMatch:
            countryCodes: ["US"]
    ruleLabels:
      - name: "geo:not-us"

  # Rule 2: Label /api requests not from US
  - name: label-api-not-us
    priority: 11
    action:
      count: {}
    statement:
      andStatement:
        statements:
          - byteMatch:
              searchString: "/api"
              fieldToMatch:
                uriPath: {}
          - labelMatch:
              key: "geo:not-us"
              scope: "LABEL"
    ruleLabels:
      - name: "threat:api-not-us"

  # Rule 3: Label /admin requests from China
  - name: label-admin-china
    priority: 12
    action:
      count: {}
    statement:
      andStatement:
        statements:
          - byteMatch:
              searchString: "/admin"
              fieldToMatch:
                uriPath: {}
          - geoMatch:
              countryCodes: ["CN"]
    ruleLabels:
      - name: "threat:admin-china"

  # Rule 4: Block if either label matches (OR logic)
  - name: block-labeled-threats
    priority: 20
    action:
      block: {}
    statement:
      orStatement:
        statements:
          - labelMatch:
              key: "threat:api-not-us"
              scope: "LABEL"
          - labelMatch:
              key: "threat:admin-china"
              scope: "LABEL"
```

### Benefits of Label-Based Chaining

1. **Unlimited Complexity**: Can express any logic, regardless of nesting depth
2. **More Readable**: Each rule has a clear purpose and name
3. **Easier to Debug**: See which rules matched in AWS logs
4. **More Flexible**: Can reuse labels across multiple rules
5. **Better Performance**: AWS evaluates rules sequentially (no performance penalty)
6. **Industry Standard**: This is how complex WAF logic is structured in production

### When Label-Based Chaining is Required

With **2-level nesting**, label-based chaining is needed for:
- ❌ `AND(NOT(X), Y)` - NOT inside AND
- ❌ `OR(NOT(X), Y)` - NOT inside OR
- ✅ Simple patterns work: `AND(X, Y)`, `OR(X, Y)`, `NOT(X)`

With **3-level nesting** (current), it's needed for:
- ❌ `AND(OR(X, Y), NOT(Z))` - mixing operators at depth
- ❌ Any pattern deeper than 3 levels

**With any nesting depth < unlimited**, complex patterns eventually require label-based chaining.

---

## CRD Size Impact

| Nesting Levels | CRD Lines | CRD Size | Change from 3-level |
|----------------|-----------|----------|---------------------|
| 3 levels (current) | ~13,000 | 867KB | baseline |
| **2 levels** | **~9-10k** | **~600KB** | **-30%** |

**Note:** Kubernetes etcd limit is 1.5MB. Even at 867KB we're safe, but reducing gives more headroom for future features (logging configuration, advanced managed rule configs, etc.).

---

## Reversibility Analysis

**Adding nesting levels later:**
- ✅ **Non-breaking change** - existing rules continue to work
- ✅ New rules can use deeper nesting
- ✅ No customer impact

**Removing nesting levels later:**
- ❌ **Breaking change** - existing rules with deep nesting would fail
- ❌ Requires customer migration
- ❌ Impossible after going live

**Conclusion:** Start conservative, expand if needed.

---

## Recommendation

**Decision Required: Choose between Option 1 (3 levels) or Option 2 (2 levels)**

### Option 1 (3 levels) - Current Implementation

**Advantages:**
1. **Supports `AND(NOT(X), Y)`**: Common pattern works directly without label-based chaining
2. **Matches console capability**: Console also supports NOT within AND/OR
3. **Less customer friction**: Intuitive nesting for common patterns

**Disadvantages:**
1. **30% larger CRD**: 867KB vs 600KB
2. **More complex API**: Harder to understand and validate
3. **Less headroom**: Future features (logging, etc.) add to already large CRD

### Option 2 (2 levels) - Simpler Alternative

**Advantages:**
1. **30% Smaller CRD**: More headroom for future features (600KB vs 867KB)
2. **Simpler API**: Easier to understand, document, and validate
3. **Non-Breaking Future**: Can add Statement2 later if customer demand justifies it
4. **Forces Best Practice**: Complex logic uses label-based chaining (industry standard)

**Disadvantages:**
1. **Requires label-based chaining**: Even simple `AND(NOT(X), Y)` needs multiple rules
2. **Customer friction**: Less intuitive than console for this pattern

### When Would We Add Statement2 (upgrade from 2 to 3 levels)?

Only if customers demonstrate:
- Frequent need for `AND(NOT(X), Y)` patterns
- Resistance to label-based chaining
- CRD size impact is acceptable (<100KB growth)

---

## Questions for Decision

1. **Do customers need `AND(NOT(X), Y)` patterns badly enough to justify 30% larger CRD?**
   - If yes: Keep 3 levels (Option 1)
   - If no: Reduce to 2 levels (Option 2) - simpler API, smaller CRD

2. **Are we comfortable recommending label-based chaining as the standard approach for complex logic?**
   - If yes: Reduce to 2 levels (forces best practice)
   - If no: Keep 3 levels (allows more inline complexity)

3. **Do we want maximum headroom for future features (logging, etc.)?**
   - If yes: Reduce to 2 levels (saves ~250KB)
   - If no: Keep 3 levels

---

## Decision Timeline

**Urgency:** Must decide **before going live** - removing nesting levels later is a breaking change.

**Recommendation:** Choose 2 levels now, add Statement2 later only if customer demand justifies it.

---

## Appendix: Real-World Use Case Analysis

### Common Patterns (work with 2 levels)

1. **Path-based filtering**: `AND(path="/api", geo=US)` ✅
2. **Method + path**: `AND(method=POST, path="/login")` ✅
3. **Country blocking**: `NOT(geo=CN)` ✅
4. **Rate limit by path**: `AND(path="/api", rate>1000)` ✅
5. **Multiple conditions**: `OR(path="/api", path="/v2")` ✅

### Advanced Patterns (need labels with 2 levels)

1. **Exclude paths from geo-block**: `AND(NOT(path="/public"), geo=CN)` ❌
   - Label solution: Label non-public paths, then geo-block labeled requests
2. **Complex threat detection**: `OR(AND(bot, NOT(verified)), AND(suspicious, country=high-risk))` ❌
   - Label solution: Multiple labeling rules, final rule checks label combinations

### Managed Rule Group Scoping

**Note:** `scopeDownStatement` is limited to **leaf statements only** (no logical operators) regardless of main rule nesting level.

**Reason:** Complex scoping should use label-based filtering for efficiency.

**Example:**
```yaml
rules:
  # Label complex condition
  - name: label-api-production
    action: { count: {} }
    statement:
      andStatement:  # ✅ Full nesting available here
        statements:
          - byteMatch: { searchString: "/api", ... }
          - notStatement: { statement: { byteMatch: { searchString: "/api/internal", ... } } }
    ruleLabels:
      - name: "scope:api-production"

  # Scope managed rule using simple label
  - name: AWS-CommonRuleSet
    statement:
      managedRuleGroup:
        name: AWSManagedRulesCommonRuleSet
        scopeDownStatement:
          labelMatch:  # ✅ Leaf statement only
            key: "scope:api-production"
            scope: "LABEL"
```
