Feature: WafPolicy feature

  @skr @aws @waf
  Scenario: WafPolicy with ManagedRuleGroup statements

    Given there is shared SKR with "AWS" provider

    And resource declaration:
      | Alias     | Kind                | ApiVersion                              | Name                 | Namespace |
      | webacl    | WafPolicy           | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}          |           |

    # Create WebACL demonstrating ManagedRuleGroup with different configurations
    When resource "webacl" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: WafPolicy
      spec:
        data: |
          {
            "DefaultAction": {
              "Allow": {}
            },
            "Description": "E2E WebACL demonstrating AWS managed rule groups",
            "VisibilityConfig": {
              "CloudWatchMetricsEnabled": true,
              "MetricName": "E2EManagedRulesWebACL",
              "SampledRequestsEnabled": true
            },
            "CustomResponseBodies": {
              "block-page": {
                "ContentType": "TEXT_HTML",
                "Content": "<html><body><h1>Access Denied</h1></body></html>"
              }
            },
            "Rules": [
              {
                "Name": "AWS-CommonRuleSet",
                "Priority": 0,
                "Statement": {
                  "ManagedRuleGroupStatement": {
                    "VendorName": "AWS",
                    "Name": "AWSManagedRulesCommonRuleSet"
                  }
                },
                "OverrideAction": {
                  "None": {}
                },
                "VisibilityConfig": {
                  "CloudWatchMetricsEnabled": true,
                  "MetricName": "AWS-CommonRuleSet",
                  "SampledRequestsEnabled": true
                }
              },
              {
                "Name": "AWS-KnownBadInputs",
                "Priority": 1,
                "Statement": {
                  "ManagedRuleGroupStatement": {
                    "VendorName": "AWS",
                    "Name": "AWSManagedRulesKnownBadInputsRuleSet"
                  }
                },
                "OverrideAction": {
                  "None": {}
                },
                "VisibilityConfig": {
                  "CloudWatchMetricsEnabled": true,
                  "MetricName": "AWS-KnownBadInputs",
                  "SampledRequestsEnabled": true
                }
              },
              {
                "Name": "AWS-SQLi-Monitor",
                "Priority": 2,
                "Statement": {
                  "ManagedRuleGroupStatement": {
                    "VendorName": "AWS",
                    "Name": "AWSManagedRulesSQLiRuleSet"
                  }
                },
                "OverrideAction": {
                  "Count": {}
                },
                "VisibilityConfig": {
                  "CloudWatchMetricsEnabled": true,
                  "MetricName": "AWS-SQLi-Monitor",
                  "SampledRequestsEnabled": true
                }
              },
              {
                "Name": "AWS-SQLi-WithExclusions",
                "Priority": 3,
                "Statement": {
                  "ManagedRuleGroupStatement": {
                    "VendorName": "AWS",
                    "Name": "AWSManagedRulesSQLiRuleSet",
                    "ExcludedRules": [
                      {
                        "Name": "SQLi_QUERYARGUMENTS"
                      }
                    ]
                  }
                },
                "OverrideAction": {
                  "None": {}
                },
                "VisibilityConfig": {
                  "CloudWatchMetricsEnabled": true,
                  "MetricName": "AWS-SQLi-WithExclusions",
                  "SampledRequestsEnabled": true
                }
              },
              {
                "Name": "AWS-LinuxRuleSet",
                "Priority": 4,
                "Statement": {
                  "ManagedRuleGroupStatement": {
                    "VendorName": "AWS",
                    "Name": "AWSManagedRulesLinuxRuleSet",
                    "Version": "Version_2.0"
                  }
                },
                "OverrideAction": {
                  "None": {}
                },
                "VisibilityConfig": {
                  "CloudWatchMetricsEnabled": true,
                  "MetricName": "AWS-LinuxRuleSet",
                  "SampledRequestsEnabled": true
                }
              },
              {
                "Name": "AWS-CommonRuleSet-CustomActions",
                "Priority": 5,
                "Statement": {
                  "ManagedRuleGroupStatement": {
                    "VendorName": "AWS",
                    "Name": "AWSManagedRulesCommonRuleSet",
                    "RuleActionOverrides": [
                      {
                        "Name": "SizeRestrictions_BODY",
                        "ActionToUse": {
                          "Count": {}
                        }
                      },
                      {
                        "Name": "NoUserAgent_HEADER",
                        "ActionToUse": {
                          "Block": {
                            "CustomResponse": {
                              "ResponseCode": 403,
                              "CustomResponseBodyKey": "block-page"
                            }
                          }
                        }
                      }
                    ]
                  }
                },
                "OverrideAction": {
                  "None": {}
                },
                "VisibilityConfig": {
                  "CloudWatchMetricsEnabled": true,
                  "MetricName": "AWS-CommonRuleSet-CustomActions",
                  "SampledRequestsEnabled": true
                }
              }
            ]
          }
      """

    # Then debug wait "webacl"

    Then eventually "webacl.status.state == 'Ready'" is ok, unless:
      | webacl.status.state == 'Error' |
      | #timeout=20m                   |

    And "findConditionTrue(webacl, 'Ready')" is ok
    And "webacl.status.providerId" is ok

    # Clean up
    When resource "webacl" is deleted
    Then eventually resource "webacl" does not exist

  @skr @aws @waf @debug
  Scenario: Deploy httpbin application and protect it with AWS WAF

    Given there is shared SKR with "AWS" provider

    And resource declaration:
      | Alias           | Kind               | ApiVersion                              | Name                       | Namespace |
      | webacl          | WafPolicy          | cloud-resources.kyma-project.io/v1beta1 | e2e-${scenarioId}          | default   |
      | sa              | ServiceAccount     | v1                                      | e2e-${scenarioId}          | default   |
      | service         | Service            | v1                                      | e2e-${scenarioId}          | default   |
      | deployment      | Deployment         | apps/v1                                 | e2e-${scenarioId}          | default   |
      | ingressclass    | IngressClass       | networking.k8s.io/v1                    | alb                        |           |
      | ingress         | Ingress            | networking.k8s.io/v1                    | e2e-${scenarioId}          | default   |

    # Step 1: Create WAF Policy with security rules
    When resource "webacl" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: WafPolicy
      spec:
        data: |
          {
            "DefaultAction": {
              "Allow": {}
            },
            "Description": "Common threat protection for httpbin",
            "VisibilityConfig": {
              "CloudWatchMetricsEnabled": true,
              "MetricName": "httpbin-waf-metrics",
              "SampledRequestsEnabled": true
            },
            "Rules": [
              {
                "Name": "known-bad-inputs",
                "Priority": 0,
                "Statement": {
                  "ManagedRuleGroupStatement": {
                    "VendorName": "AWS",
                    "Name": "AWSManagedRulesKnownBadInputsRuleSet"
                  }
                },
                "OverrideAction": {
                  "None": {}
                },
                "VisibilityConfig": {
                  "CloudWatchMetricsEnabled": true,
                  "MetricName": "known-bad-inputs",
                  "SampledRequestsEnabled": true
                }
              },
              {
                "Name": "common-rules",
                "Priority": 1,
                "Statement": {
                  "ManagedRuleGroupStatement": {
                    "VendorName": "AWS",
                    "Name": "AWSManagedRulesCommonRuleSet"
                  }
                },
                "OverrideAction": {
                  "None": {}
                },
                "VisibilityConfig": {
                  "CloudWatchMetricsEnabled": true,
                  "MetricName": "common-rules",
                  "SampledRequestsEnabled": true
                }
              }
            ]
          }
      """

    Then eventually "webacl.status.state == 'Ready'" is ok, unless:
      | webacl.status.state == 'Error' |
      | #timeout=3m                    |

    And "findConditionTrue(webacl, 'Ready')" is ok
    And "webacl.status.providerId" is ok

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
          alb.ingress.kubernetes.io/wafv2-acl-arn: "${webacl.status.providerId}"
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
      | #timeout=20m |

    # Then debug wait "ingress"

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
