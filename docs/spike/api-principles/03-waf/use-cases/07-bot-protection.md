# Use Case 8: Bot Protection and Management

## User Story

"As a platform operator, I want to detect and mitigate malicious bot traffic (scrapers, DDoS bots, credential stuffing bots) while allowing legitimate bots (search engines, monitoring tools, verified partners), and apply different protection levels to different paths based on sensitivity."

## Requirements

1. Enable bot detection and mitigation
2. Allow verified good bots (Googlebot, Bingbot, etc.)
3. Block known bad bots (scrapers, attack tools)
4. Challenge suspicious bot-like behavior
5. Different bot protection levels per path
6. Custom bot detection rules (User-Agent, behavior patterns)

## Real-World Scenarios

### Scenario A: E-Commerce Site Protection
```yaml
# Product pages: Allow search engine bots, block scrapers
# Checkout: Block all bots
# API: Allow verified partner bots with API keys
paths:
  - /products: Allow Google/Bing, block others
  - /checkout: Block all bots
  - /api: Allow bots with valid API key header
```

### Scenario B: Content Site with Ads
```yaml
# Articles: Allow search engines (SEO), block ad fraud bots
# Admin: Block all bots
level:
  - articles: Low protection (SEO-friendly)
  - admin: Maximum protection
```

### Scenario C: API Rate Limiting by Bot Type
```yaml
# Verified bots: Higher rate limits
# Unknown bots: Strict rate limits
# Malicious bots: Blocked
bot_type:
  - verified: 1000 req/min
  - unknown: 10 req/min
  - malicious: blocked
```

### Scenario D: Credential Stuffing Prevention
```yaml
# Login endpoint: Maximum bot protection
# Challenge suspicious login patterns
path: /login
protection: Maximum
actions: CAPTCHA challenge on bot detection
```

## Portable API Design

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafConfiguration
metadata:
  name: bot-protection-policy
spec:
  # Start from preset that includes managed rules
  basePolicyRef:
    name: owasp-moderate
  
  # Custom bot detection rules
  customRules:
    # Allow search engines globally
    - name: allow-search-engine-bots
      priority: 10
      action: allow
      conditions:
        header:
          name: "User-Agent"
          anyOf:
            - contains: "Googlebot"
            - contains: "Bingbot"
            - contains: "Slackbot"
      reason: "Allow verified search engine bots"
    
    # Block known bad bots
    - name: block-scraper-bots
      priority: 20
      action: block
      conditions:
        header:
          name: "User-Agent"
          anyOf:
            - contains: "scrapy"
            - contains: "python-requests"
            - contains: "curl"
            - exact: "PostmanRuntime"
      reason: "Block known scraper tools"
    
    # Block headless browsers (potential bots)
    - name: block-headless-browsers
      priority: 30
      action: block
      conditions:
        header:
          name: "User-Agent"
          anyOf:
            - contains: "HeadlessChrome"
            - contains: "PhantomJS"
            - contains: "Selenium"
      reason: "Block headless browser automation"
    
    # Checkout: Block all bots (strict)
    - name: checkout-no-bots
      priority: 100
      action: block
      conditions:
        path:
          prefix: "/checkout"
        botScore:
          min: 1  # Any bot indication
      reason: "Checkout requires human interaction"
    
    # Login: Maximum bot protection with challenge
    - name: login-bot-challenge
      priority: 110
      action: challenge  # CAPTCHA or JS challenge
      conditions:
        path:
          exact: "/login"
        method: POST
        botScore:
          min: 50  # Medium bot likelihood
      reason: "Challenge suspicious login attempts"
    
    # API: Allow bots with valid API key
    - name: api-verified-bots-allowed
      priority: 200
      action: allow
      conditions:
        path:
          prefix: "/api"
        header:
          name: "X-API-Key"
          exists: true
        botScore:
          min: 1
      reason: "API bots allowed with valid API key"
    
    # Product pages: Allow low bot scores (search engines)
    - name: products-allow-low-bot-score
      priority: 300
      action: allow
      conditions:
        path:
          prefix: "/products"
        botScore:
          max: 30  # Low bot likelihood (likely good bots)
      reason: "Allow search engine indexing of products"
    
    # Block high bot scores on sensitive paths
    - name: admin-block-bots
      priority: 400
      action: block
      conditions:
        path:
          prefix: "/admin"
        botScore:
          min: 50  # High bot likelihood
      reason: "Admin panel blocks automated access"
```

## Provider Capability Check

### AWS WAFv2

**Status:** ✅ **Excellent Support**

**How it works:**
- **AWS Managed Rules Bot Control**: `AWSManagedRulesBotControlRuleSet`
- **Bot detection levels**: Common, Targeted, Targeted_Extra
- **Bot categories**: CategorySearchEngine, CategoryMonitoring, CategoryAdvertising, CategorySocialMedia, CategoryScrapingFramework, etc.
- **Bot verification**: Verifies search engine bots (Googlebot, Bingbot)
- **Challenge actions**: CAPTCHA, Count, Block
- **Bot score**: 0-100 (0 = human, 100 = bot)

**Native AWS Translation:**
```json
{
  "Name": "bot-protection-webacl",
  "Scope": "REGIONAL",
  "DefaultAction": {"Allow": {}},
  "Rules": [
    {
      "Name": "AllowSearchEngineBots",
      "Priority": 10,
      "Action": {"Allow": {}},
      "Statement": {
        "OrStatement": {
          "Statements": [
            {
              "ByteMatchStatement": {
                "SearchString": "Googlebot",
                "FieldToMatch": {"SingleHeader": {"Name": "user-agent"}},
                "PositionalConstraint": "CONTAINS",
                "TextTransformations": [{"Priority": 0, "Type": "LOWERCASE"}]
              }
            },
            {
              "ByteMatchStatement": {
                "SearchString": "bingbot",
                "FieldToMatch": {"SingleHeader": {"Name": "user-agent"}},
                "PositionalConstraint": "CONTAINS",
                "TextTransformations": [{"Priority": 0, "Type": "LOWERCASE"}]
              }
            }
          ]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "AllowSearchEngineBots"
      }
    },
    {
      "Name": "BlockScraperBots",
      "Priority": 20,
      "Action": {"Block": {}},
      "Statement": {
        "OrStatement": {
          "Statements": [
            {
              "ByteMatchStatement": {
                "SearchString": "scrapy",
                "FieldToMatch": {"SingleHeader": {"Name": "user-agent"}},
                "PositionalConstraint": "CONTAINS",
                "TextTransformations": [{"Priority": 0, "Type": "LOWERCASE"}]
              }
            },
            {
              "ByteMatchStatement": {
                "SearchString": "python-requests",
                "FieldToMatch": {"SingleHeader": {"Name": "user-agent"}},
                "PositionalConstraint": "CONTAINS",
                "TextTransformations": [{"Priority": 0, "Type": "LOWERCASE"}]
              }
            }
          ]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "BlockScraperBots"
      }
    },
    {
      "Name": "AWSManagedRulesBotControlRuleSet",
      "Priority": 100,
      "Statement": {
        "ManagedRuleGroupStatement": {
          "VendorName": "AWS",
          "Name": "AWSManagedRulesBotControlRuleSet",
          "ManagedRuleGroupConfigs": [
            {
              "AWSManagedRulesBotControlRuleSet": {
                "InspectionLevel": "TARGETED"
              }
            }
          ],
          "RuleActionOverrides": [
            {
              "Name": "CategorySearchEngine",
              "ActionToUse": {"Count": {}}
            }
          ]
        }
      },
      "OverrideAction": {"None": {}},
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "BotControlRuleSet"
      }
    },
    {
      "Name": "LoginBotChallenge",
      "Priority": 110,
      "Action": {
        "Captcha": {
          "CustomRequestHandling": {
            "InsertHeaders": [
              {
                "Name": "x-bot-challenge",
                "Value": "login"
              }
            ]
          }
        }
      },
      "Statement": {
        "AndStatement": {
          "Statements": [
            {
              "ByteMatchStatement": {
                "SearchString": "/login",
                "FieldToMatch": {"UriPath": {}},
                "PositionalConstraint": "EXACTLY",
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            }
          ]
        }
      },
      "CaptchaConfig": {
        "ImmunityTimeProperty": {
          "ImmunityTime": 300
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "LoginBotChallenge"
      }
    }
  ],
  "CaptchaConfig": {
    "ImmunityTimeProperty": {
      "ImmunityTime": 300
    }
  }
}
```

**Key Features:**
- AWS Bot Control managed rule group (Common/Targeted/Targeted_Extra levels)
- Bot verification for major search engines
- CAPTCHA and Challenge actions
- Bot categories and labels
- RuleActionOverrides for specific bot categories

**Complexity:** Low - Excellent managed bot protection

**Cost:** ⚠️ Bot Control is a paid add-on (~$10/month + $1/million requests)

---

### Azure Application Gateway WAF

**Status:** ⚠️ **Limited Support - Bot Manager Rule Set**

**How it works:**
- **Microsoft Bot Manager Rule Set**: Basic bot detection
- **Rule group overrides**: Can disable/enable specific bot rules
- **No bot score or challenge actions**
- **No verified bot allowlisting**
- **Limited compared to AWS**

**Native Azure Translation:**
```json
{
  "location": "eastus",
  "properties": {
    "customRules": [
      {
        "name": "AllowSearchEngineBots",
        "priority": 10,
        "ruleType": "MatchRule",
        "action": "Allow",
        "matchConditions": [
          {
            "matchVariables": [
              {"variableName": "RequestHeaders"}
            ],
            "selector": "User-Agent",
            "operator": "Contains",
            "matchValues": ["Googlebot", "Bingbot"],
            "negationConditon": false
          }
        ]
      },
      {
        "name": "BlockScraperBots",
        "priority": 20,
        "ruleType": "MatchRule",
        "action": "Block",
        "matchConditions": [
          {
            "matchVariables": [
              {"variableName": "RequestHeaders"}
            ],
            "selector": "User-Agent",
            "operator": "Contains",
            "matchValues": ["scrapy", "python-requests", "curl"],
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

**Key Limitations:**
- ❌ No bot score or bot likelihood
- ❌ No CAPTCHA or challenge actions
- ❌ No verified bot list
- ✅ Basic bot detection via Bot Manager rule set
- ⚠️ Must use custom rules for User-Agent blocking

**Complexity:** Medium - Limited bot protection capabilities

---

### GCP Cloud Armor

**Status:** ✅ **Good Support - Adaptive Protection & reCAPTCHA**

**How it works:**
- **Preconfigured bot defense**: `evaluatePreconfiguredWaf('bot-defense')`
- **reCAPTCHA Enterprise integration**: Challenge suspicious requests
- **Adaptive Protection**: ML-based anomaly detection
- **Rate limiting per bot type**
- **Session affinity** for bot tracking

**Native GCP Translation:**
```json
{
  "name": "bot-protection-security-policy",
  "adaptiveProtectionConfig": {
    "layer7DdosDefenseConfig": {
      "enable": true,
      "ruleVisibility": "STANDARD"
    }
  },
  "rules": [
    {
      "priority": 10,
      "description": "Allow search engine bots",
      "action": "allow",
      "match": {
        "expr": {
          "expression": "request.headers['user-agent'].lower().contains('googlebot') || request.headers['user-agent'].lower().contains('bingbot')"
        }
      }
    },
    {
      "priority": 20,
      "description": "Block scraper bots",
      "action": "deny(403)",
      "match": {
        "expr": {
          "expression": "request.headers['user-agent'].lower().contains('scrapy') || request.headers['user-agent'].lower().contains('python-requests')"
        }
      }
    },
    {
      "priority": 100,
      "description": "reCAPTCHA challenge for login",
      "action": "redirect",
      "match": {
        "expr": {
          "expression": "request.path == '/login' && request.method == 'POST'"
        }
      },
      "redirectOptions": {
        "type": "GOOGLE_RECAPTCHA"
      },
      "recaptchaOptionsPath": {
        "recaptchaOptions": {
          "action": "login",
          "sessionTokenSiteKey": "projects/PROJECT_ID/keys/KEY_ID"
        }
      }
    },
    {
      "priority": 1000,
      "description": "Preconfigured bot defense",
      "action": "deny(403)",
      "match": {
        "expr": {
          "expression": "evaluatePreconfiguredWaf('bot-defense', {'sensitivity': 1})"
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
  }
}
```

**Key Features:**
- Preconfigured bot defense WAF rule
- reCAPTCHA Enterprise integration (challenge action)
- Adaptive Protection (ML-based)
- Custom User-Agent blocking via CEL
- Rate limiting by bot patterns

**Complexity:** Medium - Good bot protection with reCAPTCHA integration

**Cost:** ⚠️ reCAPTCHA Enterprise is paid (~$1/1000 assessments)

---

## Cross-Provider Comparison

| Feature | AWS WAFv2 | Azure WAF | GCP Cloud Armor |
|---------|-----------|-----------|-----------------|
| **Managed bot protection** | ✅ Bot Control (excellent) | ⚠️ Bot Manager (basic) | ✅ Bot defense (good) |
| **Bot score/likelihood** | ✅ 0-100 score | ❌ No | ⚠️ ML-based (not exposed) |
| **Verified bot allowlist** | ✅ Yes | ❌ No | ⚠️ Manual via CEL |
| **Bot categories** | ✅ 12+ categories | ❌ No | ❌ No |
| **Challenge actions** | ✅ CAPTCHA, JS challenge | ❌ No | ✅ reCAPTCHA Enterprise |
| **Custom User-Agent rules** | ✅ Yes | ✅ Yes | ✅ Yes (CEL) |
| **Per-path bot policies** | ✅ ScopeDownStatement | ⚠️ Custom rules only | ✅ CEL conditions |
| **Bot verification** | ✅ Native | ❌ Manual | ❌ Manual |
| **Adaptive/ML protection** | ⚠️ Via Fraud Control | ❌ No | ✅ Adaptive Protection |
| **Cost** | Paid add-on (~$10/mo) | Included | reCAPTCHA paid |
| **Complexity** | Low | Medium | Medium |
| **Fidelity** | Excellent | Basic | Good |

---

## Design Decision Impact

### Recommendation: ✅ **Include with Provider Differences**

**Rationale:**
1. **AWS has best bot protection** - Bot Control with categories, verification, CAPTCHA
2. **Azure has basic bot protection** - Bot Manager rule set, manual User-Agent rules
3. **GCP has good bot protection** - Preconfigured bot defense + reCAPTCHA Enterprise
4. **Critical security need** - Bot attacks (credential stuffing, scraping, DDoS) are common

**API Design:**

```yaml
spec:
  # Start from preset that includes managed rules
  basePolicyRef:
    name: owasp-moderate
      botProtection:
        level: standard  # minimal | standard | strict
        allowVerifiedBots: true
        verifiedBots: [GoogleBot, BingBot]
        blockCategories: [scraper, attack_tool]
  
  customRules:
    - name: allow-good-bots
      priority: 10
      action: allow
      conditions:
        header:
          name: "User-Agent"
          contains: "Googlebot"
    
    - name: login-bot-challenge
      priority: 100
      action: challenge  # CAPTCHA
      conditions:
        path: {exact: "/login"}
        method: POST
```

### Status Reporting

**AWS (Excellent):**
```yaml
status:
  appliedManagedRuleGroups:
    - name: AWSManagedRulesBotControlRuleSet
      appliedStrategy: "native"
      message: "Using AWS Bot Control Rule Set (TARGETED level) with verified bot allowlist"
```

**Azure (Basic):**
```yaml
status:
  conditions:
    - type: "BotProtectionLimited"
      status: "True"
      reason: "AzureBasicBotManager"
      message: "Azure Bot Manager provides basic bot detection. For advanced bot protection (bot score, verified bots, CAPTCHA), consider Azure Front Door Premium or application-level bot management."
  
  appliedManagedRuleGroups:
    - name: Microsoft_BotManagerRuleSet
      appliedStrategy: "basic"
      message: "Using Microsoft Bot Manager Rule Set (basic detection only)"
```

**GCP (Good):**
```yaml
status:
  appliedManagedRuleGroups:
    - name: cve-canary
      appliedStrategy: "native"
      message: "Using preconfigured bot-defense WAF rule + reCAPTCHA Enterprise challenges"
```

---

## Implementation Priority

### Phase 1: Custom User-Agent Rules (Universal)
```yaml
customRules:
  - name: allow-search-bots
    action: allow
    conditions:
      header: {name: "User-Agent", contains: "Googlebot"}
  
  - name: block-scraper-bots
    action: block
    conditions:
      header: {name: "User-Agent", contains: "scrapy"}
```
**Works on:** AWS ✅, Azure ✅, GCP ✅

### Phase 2: Managed Bot Protection
```yaml
  # Start from preset that includes managed rules
  basePolicyRef:
    name: owasp-moderate
```
**Works on:** AWS ✅ (excellent), Azure ⚠️ (basic), GCP ✅ (good)

### Phase 3: Challenge Actions
```yaml
customRules:
  - name: login-challenge
    action: challenge
    conditions:
      path: {exact: "/login"}
```
**Works on:** AWS ✅ (CAPTCHA), Azure ❌ (not supported), GCP ✅ (reCAPTCHA)

---

## Conclusion

**Bot protection has varying levels of support:**
- ✅ **AWS**: Excellent - Bot Control with categories, verification, CAPTCHA (paid add-on)
- ⚠️ **Azure**: Basic - Bot Manager rule set, no advanced features
- ✅ **GCP**: Good - Preconfigured bot defense + reCAPTCHA Enterprise

**Recommendation:** ✅ **High Priority - Implement with Provider Tiers**

**Implementation Strategy:**
1. **Phase 1:** Custom User-Agent rules (universal support)
2. **Phase 2:** Managed bot protection rule group (different capabilities per provider)
3. **Phase 3:** Challenge actions for sensitive paths (AWS/GCP only)

**Real-World Use Cases:**
- Credential stuffing prevention on login
- E-commerce scraper blocking
- Search engine bot allowlisting (SEO)
- DDoS bot mitigation
- API bot rate limiting

Bot protection is critical for modern web applications and should be included despite varying provider capabilities!
