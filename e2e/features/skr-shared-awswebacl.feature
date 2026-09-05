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

  @skr @aws @waf
  Scenario: Deploy httpbin application and protect it with AWS WAF

    Given there is shared SKR with "AWS" provider

    And resource declaration:
      | Alias           | Kind          | ApiVersion                              | Name                    | Namespace |
      | webacl          | AwsWebAcl     | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}             | default   |
      | sa              | ServiceAccount| v1                                      | ${webacl.metadata.name} | default   |
      | service         | Service       | v1                                      | ${webacl.metadata.name} | default   |
      | deployment      | Deployment    | apps/v1                                 | ${webacl.metadata.name} | default   |
      | ingressclass    | IngressClass  | networking.k8s.io/v1                    | alb                     |           |
      | ingress         | Ingress       | networking.k8s.io/v1                    | ${webacl.metadata.name} | default   |


    # Step 1: Create AWS WebACL with security rules
    When resource "webacl" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: AwsWebAcl
      spec:
        defaultAction:
          allow: {}

        description: "Common threat protection for httpbin"

        visibilityConfig:
          cloudWatchMetricsEnabled: true
          metricName: httpbin-waf-metrics
          sampledRequestsEnabled: true

        rules:
          - name: known-bad-inputs
            priority: 0
            overrideAction:
              none: {}
            statements:
              - managedRuleGroup:
                  vendorName: AWS
                  name: AWSManagedRulesKnownBadInputsRuleSet
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: known-bad-inputs
              sampledRequestsEnabled: true

          - name: common-rules
            priority: 1
            overrideAction:
              none: {}
            statements:
              - managedRuleGroup:
                  vendorName: AWS
                  name: AWSManagedRulesCommonRuleSet
            visibilityConfig:
              cloudWatchMetricsEnabled: true
              metricName: common-rules
              sampledRequestsEnabled: true
      """

    Then eventually "webacl.status.state == 'Ready'" is ok, unless:
      | webacl.status.state == 'Error' |
      | #timeout=3m                    |

    And "findConditionTrue(webacl, 'Ready')" is ok
    And "webacl.status.arn" is ok
    And "webacl.status.capacity > 0" is ok

    # Step 2: Deploy httpbin application
    When resource "sa" is created:
      """
      apiVersion: v1
      kind: ServiceAccount
      """

    And resource "service" is created:
      """
      apiVersion: v1
      kind: Service
      metadata:
        labels:
          app: ${webacl.metadata.name}
          service: ${webacl.metadata.name}
      spec:
        type: NodePort
        ports:
        - name: http
          port: 8000
          targetPort: 80
        selector:
          app: ${webacl.metadata.name}
      """

    And resource "deployment" is created:
      """
      apiVersion: apps/v1
      kind: Deployment
      spec:
        replicas: 1
        selector:
          matchLabels:
            app: ${webacl.metadata.name}
            version: v1
        template:
          metadata:
            labels:
              app: ${webacl.metadata.name}
              version: v1
          spec:
            serviceAccountName: ${webacl.metadata.name}
            containers:
            - image: docker.io/kennethreitz/httpbin
              imagePullPolicy: IfNotPresent
              name: httpbin
              ports:
              - containerPort: 80
      """

    Then eventually "deployment.status.availableReplicas > 0" is ok, unless:
      | #timeout=2m |

    # Step 3: Create IngressClass for ALB
    When resource "ingressclass" is created:
      """
      apiVersion: networking.k8s.io/v1
      kind: IngressClass
      metadata:
        name: alb
      spec:
        controller: ingress.k8s.aws/alb
      """

    # Step 4: Expose application via ALB
    When resource "ingress" is created:
      """
      apiVersion: networking.k8s.io/v1
      kind: Ingress
      metadata:
        annotations:
          alb.ingress.kubernetes.io/scheme: internet-facing
          alb.ingress.kubernetes.io/target-type: instance
          alb.ingress.kubernetes.io/listen-ports: '[{"HTTP":80}]'
          alb.ingress.kubernetes.io/wafv2-acl-arn: "${webacl.status.arn}"
      spec:
        ingressClassName: alb
        rules:
        - http:
            paths:
            - path: /
              pathType: Prefix
              backend:
                service:
                  name: ${webacl.metadata.name}
                  port:
                    number: 8000
      """

    Then eventually "ingress.status.loadBalancer.ingress[0].hostname != ''" is ok, unless:
      | #timeout=5m |

    # Step 5 & 6: Test with WAF protection -  XSS is blocked, normal requests work
    And HTTP operation succeeds:
      | Url              | http://${ingress.status.loadBalancer.ingress[0].hostname}/base64/SFRUUEJJTiBpcyBhd2Vzb21l?q=%3Cscript%3Ealert%281%29%3C%2Fscript%3E |
      | ExpectedOutput   | 403 Forbidden                                                                                                                       |
      | Retry            | 10                                                                                                                                  |

    And HTTP operation succeeds:
      | Url            | http://${ingress.status.loadBalancer.ingress[0].hostname}/base64/SFRUUEJJTiBpcyBhd2Vzb21l |
      | ExpectedOutput | HTTPBIN is awesome                                                                        |

    # Clean up

    When resource "ingress" is deleted
    Then eventually resource "ingress" does not exist

    When resource "webacl" is deleted
    Then eventually resource "webacl" does not exist

    When resource "ingressclass" is deleted
    Then eventually resource "ingressclass" does not exist

    When resource "deployment" is deleted
    Then eventually resource "deployment" does not exist

    When resource "service" is deleted
    Then eventually resource "service" does not exist

    When resource "sa" is deleted
    Then eventually resource "sa" does not exist
