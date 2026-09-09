# WAF

## Provider portable WAF API

- Intent portable resources `AppLoadBalancer`. The `WafPolicy` has portable schema, but not the content. 
- Limiting bind to single K8S Service in order to reconcile provider differences, non-blocking for user wanting multiservice coverage since they can create additional AppLoadBalancer for other services
- Backend low level Service, possibility to bind to higher level abstractions like Istio Gateway, Kubernetes Gateway
- Automatic derivation of configuration from the high level backend references (low level service, domains, certificates)
- Unstructured provider specific WAF policy configuration with local limited and remote (cloud) full validation
- Predeployed or documentation provided general "good enough to give you compliance pass unless you're picky" standard policies
- Advanced users, not satisfied with standard policies tap into the specific provider knowledge pool and without obstacles craft policy on their own (not supported by Kyma)

```yaml
apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: AppLoadBalancer
metadata:
  namespace: my-names
  generation: 34
spec:
  backend:
    # Supported: Service, Istio Gateway, Kubernetes Gateway
    # For K8S Service full frontend spec must be provided, for other types details can be derived
    # Regardless of the backend target, the Service must have nodePort in the status - ie be of the types NodePort or LoadBalancer
    apiVersion: v1
    kind: Service
    name: my-service
    # optional, defaults to AppLoadBalancer namespace
    namespace: some-namespace

  # optional, can be taken from the Service status
  healthCheck:
    path: /healthz
    nodePort: 32451
    # TODO: what are other parameters here that can be tweaked?
    timeout: 30s
    retryCount: 10s
    # TODO: find all the settings possible

  frontend:
    # HTTP port ------------------------------
    - hosts:
        - my-service.example.com
        - some-other-host.example.com
        # TODO: check if cloud frontend accepts wildcard '*'
      port:
        number: 80
        protocol: HTTP

    # HTTP port with redirect to HTTPS -------
    - hosts:
        - my-service.example.com
        - some-other-host.example.com
      port:
        number: 80
        protocol: HTTP
      httpsRedirect: true

    # HTTPS port ----------------------------
    # TODO: check if cloud frontend can be configured to have two different domains each with own certificate on the same port (ie 443)
    - hosts:
        - my-service.example.com
      port:
        number: 443
        protocol: HTTPS
      tls:
        secretRef:
          name: my-tls-secret
          namespace: some-names # optional, if not specified defaults to service namespace

    - hosts:
        - some-other-host.example.com
      port:
        number: 443
        protocol: HTTPS
      tls:
        secretRef:
          name: my-other-tls-secret
          namespace: some-names # optional, if not specified defaults to service namespace

  # required only for Azure, not supported for others
  # IpRange class must be azure/waf specific
  networkBinding:
    apiVersion: cloud-resources.kyma-project.io/v1alpha1
    kind: IpRange
    name: waf-dedicated

  # Optional, LoadBalancer can be created and functional without policy
  policy:
    apiVersion: cloud-resources.kyma-project.io/v1alpha1
    kind: WafPolicy
    name: my-policy

status:
  certificates:
    - source: some-names/my-other-tls-secret
      observedGeneration: 12 # first check, tells if secret spec has changed
      hash: alkfhkjafhkjahf # second check, even if secret has changed, the certificate in it may remain the same, this is the hard check
      providerId: asjhaljfhakjsfhdkj
      status: Processing
      message: Error creating certificate
      lastTransitionTime: 2026-01-01T00:00:00.0000Z
  observedGeneration: 34
  conditions:
    - type: Ready
      status: Unknown
      reason: Processsing

---

# WafPolicy
# Do not provision in cloud right away - gated creation only if referenced from some AppLoadBalancer. 
# We initally deploy stadnard policies that are documented as common base. 
# This means they will exist in the SKR but since not referenced form a AppLoadBalancer there's no need to provision them in the cloud.
# Deletion of WafPolicy checks if referenced on some AppLoadBalancer, refused if used, deleted only if not used
# There's no delete cascade from the AppLoadBalancer to the WafPolicy

---

apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafPolicy
metadata:
  name: my-azure-policy
spec:
  data: |
    {
        "customRules": [
            {
                "name": "gbaas",
                "priority": 53,
                "ruleType": "MatchRule",
                "action": "Block",
                "matchConditions": [
                    {
                        "matchVariables": [
                            {
                                "variableName": "RemoteAddr"
                            }
                        ],
                        "operator": "IPMatch",
                        "negationConditon": false,
                        "matchValues": [
                            "181.55.22.103",
                            "141.11.252.217",
                            "136.65.234.235"
                        ],
                        "transforms": []
                    }
                ],
                "skippedManagedRuleSets": [],
                "state": "Enabled"
            },
            {
                "name": "ratelimit",
                "priority": 67,
                "ruleType": "RateLimitRule",
                "rateLimitDuration": "OneMin",
                "action": "Block",
                "rateLimitThreshold": 100,
                "matchConditions": [
                    {
                        "matchVariables": [
                            {
                                "variableName": "RemoteAddr"
                            }
                        ],
                        "operator": "IPMatch",
                        "negationConditon": false,
                        "matchValues": [
                            "0.0.0.0/0"
                        ],
                        "transforms": []
                    }
                ],
                "groupByUserSession": [
                    {
                        "groupByVariables": [
                            {
                                "variableName": "ClientAddr"
                            }
                        ]
                    }
                ],
                "skippedManagedRuleSets": [],
                "state": "Enabled"
            }
        ],
        "policySettings": {
            "requestBodyCheck": true,
            "maxRequestBodySizeInKb": 128,
            "fileUploadLimitInMb": 100,
            "state": "Enabled",
            "mode": "Detection",
            "jsChallengeCookieExpirationInMins": 30,
            "requestBodyInspectLimitInKB": 128,
            "fileUploadEnforcement": true,
            "requestBodyEnforcement": true
        },
        "managedRules": {
            "managedRuleSets": [
                {
                    "ruleSetType": "Microsoft_DefaultRuleSet",
                    "ruleSetVersion": "2.2",
                    "ruleGroupOverrides": [
                        {
                            "ruleGroupName": "MS-ThreatIntel-CVEs",
                            "rules": [
                                {
                                    "ruleId": "99001015",
                                    "state": "Enabled",
                                    "action": "AnomalyScoring"
                                }
                            ]
                        },
                        {
                            "ruleGroupName": "MS-ThreatIntel-XSS",
                            "rules": [
                                {
                                    "ruleId": "99032002",
                                    "state": "Enabled",
                                    "action": "Block"
                                }
                            ]
                        }
                    ]
                },
                {
                    "ruleSetType": "Microsoft_BotManagerRuleSet",
                    "ruleSetVersion": "1.1",
                    "ruleGroupOverrides": []
                },
                {
                    "ruleSetType": "Microsoft_HTTPDDoSRuleSet",
                    "ruleSetVersion": "1.0",
                    "ruleGroupOverrides": []
                }
            ],
            "exclusions": [
                {
                    "matchVariable": "RequestHeaderValues",
                    "selectorMatchOperator": "Contains",
                    "selector": "My-Header",
                    "exclusionManagedRuleSets": [
                        {
                            "ruleSetType": "Microsoft_DefaultRuleSet",
                            "ruleSetVersion": "2.2",
                            "ruleGroups": [
                                {
                                    "ruleGroupName": "XSS",
                                    "rules": [
                                        {
                                            "ruleId": "941100"
                                        }
                                    ]
                                }
                            ]
                        }
                    ]
                }
            ]
        }
    }
status:
  providerId: akjshdlakdhjlakdjlk
---

apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafPolicy
metadata:
  name: my-gcp-policy
spec:
  # TODO: extend to showcase standard + custom rule
  data: |
    {
      "priority": 1000,
      "description": "OWASP CRS 4.22 - SQL injection",
      "action": "deny(403)",
      "preview": true,
      "match": {
        "expr": {
          "expression": "evaluatePreconfiguredWaf('sqli-v422-stable', {'sensitivity': 1})"
        }
      }
    }

---

apiVersion: cloud-resources.kyma-project.io/v1alpha1
kind: WafPolicy
metadata:
  name: my-aws-policy
  generation: 9
spec:
  data: |
    {
      "DefaultAction": {
        "Allow": {}
      },
      "Rules": [
        {
          "Name": "RateLimitRule",
          "Priority": 1,
          "Statement": {
            "RateBasedStatement": {
              "Limit": 100,
              "AggregateKeyType": "IP"
            }
          },
          "Action": {
            "Block": {}
          },
          "VisibilityConfig": {
            "SampledRequestsEnabled": true,
            "CloudWatchMetricsEnabled": true,
            "MetricName": "RateLimitRuleMetric"
          }
        }
      ],
      "VisibilityConfig": {
        "SampledRequestsEnabled": true,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "ExampleWebACLMetric"
      }
    }
status:
  observedGeneration: 9
  conditions:
    - type: Ready
      status: False
      reason: Ready
  providerId: aksjdhakjsdh
```

## AppLoadBalancer & WafPolicy conditions

- Not used
  ```yaml
  type: Ready
  status: False
  reason: NotUsed
  ```
- Processing
  ```yaml
  type: Ready
  status: Unknown
  reason: Processing
  ```
- Ready
  ```yaml
  type: Ready
  status: True
  reason: Ready
  ```
- Deleting (optionally use Processing instead)
  ```yaml
  type: Ready
  status: Unknown
  reason: Deleting
  ```
- Error
  ```yaml
  type: Ready
  status: False
  reason: Error
  message: What ever the provider said. Also tell here that AppLoadBalancer referenced policy has error so users know where to look
  ```


## AppLoadBalancer secrets status

- Processing
  ```yaml
  certificates:
    - source: some-names/my-other-tls-secret
      observedGeneration: 12
      hash: alkfhkjafhkjahf
      status: Processing
      lastTransitionTime: 2026-01-01T00:00:00.0000Z
  ```
- Ready
  ```yaml
  certificates:
    - source: some-names/my-other-tls-secret
      observedGeneration: 12 # first check, tells if secret spec has changed
      hash: alkfhkjafhkjahf # second check, even if secret has changed, the certificate in it may remain the same, this is the hard check
      providerId: asjhaljfhakjsfhdkj
      status: Ready
      lastTransitionTime: 2026-01-01T00:00:00.0000Z
  ```
- Deleting
  ```yaml
  certificates:
    - source: some-names/my-other-tls-secret
      observedGeneration: 12
      hash: alkfhkjafhkjahf
      status: Deleting # because it was removed from the spec, ie the whole host section with domain and certificate removed
      lastTransitionTime: 2026-01-01T00:00:00.0000Z
  ```
- Error
  ```yaml
  certificates:
    - source: some-names/my-other-tls-secret
      observedGeneration: 12
      hash: alkfhkjafhkjahf
      status: Error
      message: The error we get from the cloud provider
      lastTransitionTime: 2026-01-01T00:00:00.0000Z
  ```
 