Feature: AwsWebAcl feature

  @skr @aws @waf
  Scenario: AwsWebAcl with ManagedRuleGroup statements

    Given there is shared SKR with "AWS" provider

    And resource declaration:
      | Alias     | Kind       | ApiVersion                              | Name        | Namespace |
      | webacl    | AwsWebAcl  | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()} |           |

    # Create WebACL demonstrating ManagedRuleGroup with different configurations
    When resource "webacl" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: AwsWebAcl
      spec:
        defaultAction:
          allow: {}
        description: "E2E WebACL demonstrating AWS managed rule groups"
        visibilityConfig:
          cloudWatchMetricsEnabled: true
          metricName: E2EManagedRulesWebACL
          sampledRequestsEnabled: true
        customResponseBodies:
          block-page:
            contentType: TEXT_HTML
            content: "<html><body><h1>Access Denied</h1></body></html>"
        rules:
          # ManagedRuleGroup - Default override action (None)
          - name: AWS-CommonRuleSet
            priority: 0
            # overrideAction omitted - defaults to None (use managed group's actions)
            statements:
              - managedRuleGroup:
                  vendorName: AWS
                  name: AWSManagedRulesCommonRuleSet
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: AWS-CommonRuleSet
              sampledRequestsEnabled: true

          # ManagedRuleGroup - Explicit None override action
          - name: AWS-KnownBadInputs
            priority: 1
            overrideAction:
              none: {}
            statements:
              - managedRuleGroup:
                  vendorName: AWS
                  name: AWSManagedRulesKnownBadInputsRuleSet
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: AWS-KnownBadInputs
              sampledRequestsEnabled: true

          # ManagedRuleGroup - Count override (monitoring mode)
          - name: AWS-SQLi-Monitor
            priority: 2
            overrideAction:
              count: {}
            statements:
              - managedRuleGroup:
                  vendorName: AWS
                  name: AWSManagedRulesSQLiRuleSet
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: AWS-SQLi-Monitor
              sampledRequestsEnabled: true

          # ManagedRuleGroup - With excluded rules
          - name: AWS-SQLi-WithExclusions
            priority: 3
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
              metricName: AWS-SQLi-WithExclusions
              sampledRequestsEnabled: true

          # ManagedRuleGroup - With version specified
          - name: AWS-LinuxRuleSet
            priority: 4
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

          # ManagedRuleGroup - With rule action overrides
          - name: AWS-CommonRuleSet-CustomActions
            priority: 5
            overrideAction:
              none: {}
            statements:
              - managedRuleGroup:
                  vendorName: AWS
                  name: AWSManagedRulesCommonRuleSet
                  ruleActionOverrides:
                    - name: SizeRestrictions_BODY
                      actionToUse:
                        count: {}
                    - name: NoUserAgent_HEADER
                      actionToUse:
                        block:
                          customResponse:
                            responseCode: 403
                            customResponseBodyKey: block-page
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: AWS-CommonRuleSet-CustomActions
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
