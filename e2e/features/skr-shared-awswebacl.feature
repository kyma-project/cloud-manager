Feature: AwsWebAcl feature

  @skr @aws @waf
  Scenario: AwsWebAcl with comprehensive statement types

    Given there is shared SKR with "AWS" provider

    And resource declaration:
      | Alias     | Kind       | ApiVersion                              | Name        | Namespace |
      | webacl    | AwsWebAcl  | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()} |           |

    # Create WebACL demonstrating all statement types
    When resource "webacl" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: AwsWebAcl
      spec:
        defaultAction:
          allow: {}
        description: "E2E comprehensive WebACL demonstrating all statement types"
        visibilityConfig:
          cloudWatchMetricsEnabled: true
          metricName: E2EComprehensiveWebACL
          sampledRequestsEnabled: true
        tokenDomains:
          - example.com
        captchaConfig:
          immunityTime: 3600
        challengeConfig:
          immunityTime: 7200
        customResponseBodies:
          block-page:
            contentType: TEXT_HTML
            content: "<html><body><h1>Access Denied</h1></body></html>"
        rules:
          # GeoMatch statement
          - name: block-high-risk-countries
            priority: 0
            action:
              block:
                customResponse:
                  responseCode: 403
                  customResponseBodyKey: block-page
            statements:
              - geoMatch:
                  countryCodes:
                    - "CN"
                    - "RU"
                    - "KP"
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: block-high-risk-countries
              sampledRequestsEnabled: true

          # RateBased statement
          - name: rate-limit-per-ip
            priority: 1
            action:
              block: {}
            statements:
              - rateBased:
                  limit: 2000
                  aggregateKeyType: IP
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: rate-limit-per-ip
              sampledRequestsEnabled: true

          # ByteMatch - query string
          - name: block-path-traversal
            priority: 2
            action:
              block: {}
            statements:
              - byteMatch:
                  searchString: "../"
                  positionalConstraint: CONTAINS
                  fieldToMatch:
                    queryString: {}
                  textTransformations:
                    - priority: 0
                      type: URL_DECODE
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: block-path-traversal
              sampledRequestsEnabled: true

          # ByteMatch - header with text transformations
          - name: block-malicious-user-agents
            priority: 3
            action:
              block: {}
            statements:
              - byteMatch:
                  searchString: "sqlmap"
                  positionalConstraint: CONTAINS
                  fieldToMatch:
                    singleHeader:
                      name: "user-agent"
                  textTransformations:
                    - priority: 0
                      type: LOWERCASE
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: block-malicious-user-agents
              sampledRequestsEnabled: true

          # ByteMatch - uri path
          - name: block-admin-path-access
            priority: 4
            action:
              block: {}
            statements:
              - byteMatch:
                  searchString: "/admin"
                  positionalConstraint: STARTS_WITH
                  fieldToMatch:
                    uriPath: {}
                  textTransformations:
                    - priority: 0
                      type: LOWERCASE
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: block-admin-path-access
              sampledRequestsEnabled: true

          # SizeConstraint statement
          - name: block-oversized-query
            priority: 5
            action:
              block: {}
            statements:
              - sizeConstraint:
                  comparisonOperator: GT
                  size: 8192
                  fieldToMatch:
                    queryString: {}
                  textTransformations:
                    - priority: 0
                      type: URL_DECODE
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: block-oversized-query
              sampledRequestsEnabled: true

          # SqliMatch statement
          - name: block-sql-injection
            priority: 6
            action:
              block: {}
            statements:
              - sqliMatch:
                  fieldToMatch:
                    queryString: {}
                  textTransformations:
                    - priority: 0
                      type: URL_DECODE
                    - priority: 1
                      type: HTML_ENTITY_DECODE
                  sensitivityLevel: HIGH
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: block-sql-injection
              sampledRequestsEnabled: true

          # XssMatch statement
          - name: block-xss-attacks
            priority: 7
            action:
              block: {}
            statements:
              - xssMatch:
                  fieldToMatch:
                    queryString: {}
                  textTransformations:
                    - priority: 0
                      type: URL_DECODE
                    - priority: 1
                      type: HTML_ENTITY_DECODE
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: block-xss-attacks
              sampledRequestsEnabled: true

          # RegexMatch statement
          - name: block-suspicious-patterns
            priority: 8
            action:
              block: {}
            statements:
              - regexMatch:
                  regexString: "^/(admin|root|config|backup)/.*$"
                  fieldToMatch:
                    uriPath: {}
                  textTransformations:
                    - priority: 0
                      type: LOWERCASE
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: block-suspicious-patterns
              sampledRequestsEnabled: true

          # AsnMatch statement
          - name: block-suspicious-asns
            priority: 9
            action:
              block: {}
            statements:
              - asnMatch:
                  autonomousSystemNumbers:
                    - 64496
                    - 64497
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: block-suspicious-asns
              sampledRequestsEnabled: true

          # ManagedRuleGroup - Common Rule Set with overrideAction
          - name: AWS-CommonRuleSet
            priority: 10
            overrideAction:
              count: {}
            statements:
              - managedRuleGroup:
                  vendorName: AWS
                  name: AWSManagedRulesCommonRuleSet
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: AWS-CommonRuleSet
              sampledRequestsEnabled: true

          # ManagedRuleGroup - SQLi with excluded rules
          - name: AWS-SQLi
            priority: 11
            overrideAction:
              none: {}
            statements:
              - managedRuleGroup:
                  vendorName: AWS
                  name: AWSManagedRulesSQLiRuleSet
                  excludedRules:
                    - name: SQLi_QUERYARGUMENTS
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: AWS-SQLi
              sampledRequestsEnabled: true

          # ManagedRuleGroup - Linux Rule Set with version
          - name: AWS-LinuxRuleSet
            priority: 12
            overrideAction:
              none: {}
            statements:
              - managedRuleGroup:
                  vendorName: AWS
                  name: AWSManagedRulesLinuxRuleSet
                  version: "Version_2.0"
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: AWS-LinuxRuleSet
              sampledRequestsEnabled: true

          # Captcha action
          - name: captcha-suspicious-traffic
            priority: 13
            action:
              captcha:
                customRequestHandling:
                  insertHeaders:
                    - name: X-Captcha-Required
                      value: "true"
            statements:
              - geoMatch:
                  countryCodes:
                    - "SY"
            captchaConfig:
              immunityTime: 300
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: captcha-suspicious-traffic
              sampledRequestsEnabled: true

          # Challenge action
          - name: challenge-high-rate
            priority: 14
            action:
              challenge: {}
            statements:
              - rateBased:
                  limit: 500
                  aggregateKeyType: IP
            challengeConfig:
              immunityTime: 600
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: challenge-high-rate
              sampledRequestsEnabled: true

          # Count action for monitoring
          - name: monitor-api-access
            priority: 15
            action:
              count: {}
            statements:
              - byteMatch:
                  searchString: "/api"
                  positionalConstraint: STARTS_WITH
                  fieldToMatch:
                    uriPath: {}
                  textTransformations:
                    - priority: 0
                      type: LOWERCASE
            ruleLabels:
              - name: "api-access"
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: monitor-api-access
              sampledRequestsEnabled: true

          # LabelMatch statement
          - name: block-labeled-threats
            priority: 16
            action:
              block: {}
            statements:
              - labelMatch:
                  key: "api-access"
                  scope: LABEL
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: block-labeled-threats
              sampledRequestsEnabled: true

          # AND operator - combining GeoMatch and ByteMatch
          - name: block-admin-from-high-risk
            priority: 17
            action:
              block: {}
            logicalOperator: AND
            statements:
              - geoMatch:
                  countryCodes:
                    - "CN"
                    - "RU"
              - byteMatch:
                  searchString: "/admin"
                  positionalConstraint: STARTS_WITH
                  fieldToMatch:
                    uriPath: {}
                  textTransformations:
                    - priority: 0
                      type: LOWERCASE
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: block-admin-from-high-risk
              sampledRequestsEnabled: true

          # OR operator - combining multiple conditions
          - name: block-threats-or-suspicious-asn
            priority: 18
            action:
              block: {}
            logicalOperator: OR
            statements:
              - byteMatch:
                  searchString: "sqlmap"
                  positionalConstraint: CONTAINS
                  fieldToMatch:
                    singleHeader:
                      name: "user-agent"
                  textTransformations:
                    - priority: 0
                      type: LOWERCASE
              - asnMatch:
                  autonomousSystemNumbers:
                    - 64496
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: block-threats-or-suspicious-asn
              sampledRequestsEnabled: true

          # NOT operator - allow all except blocked countries
          - name: allow-except-blocked-countries
            priority: 19
            action:
              allow: {}
            logicalOperator: NOT
            statements:
              - geoMatch:
                  countryCodes:
                    - "KP"
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: allow-except-blocked-countries
              sampledRequestsEnabled: true
      """

    Then eventually "webacl.status.state == 'Ready'" is ok, unless:
      | webacl.status.state == 'Error' |
      | #timeout=20m                   |

    And "findConditionTrue(webacl, 'Ready')" is ok
    And "webacl.status.arn" is ok
    And "webacl.status.capacity > 0" is ok

    # Clean up
    When resource "webacl" is deleted
    Then eventually resource "webacl" does not exist
