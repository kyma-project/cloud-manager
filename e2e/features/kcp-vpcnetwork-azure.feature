Feature: VpcNetwork Azure

  @kcp @azure @vpcnetwork
  Scenario: VpcNetwork Azure is created and deleted

    Given current cluster is "kcp"

    Given resource declaration:
      | Alias     | Kind                  | ApiVersion                              | Name               | Namespace |
      | vpc       | VpcNetwork            | cloud-control.kyma-project.io/v1beta1   | e2e-${id()}        |           |

    Given Subscription "subscription" exists for "Azure" provider

    When resource "vpc" is created:
      """
      apiVersion: cloud-control.kyma-project.io/v1beta1
      kind: VpcNetwork
      spec:
        subscription: ${subscription.metadata.name}
        cidrBlocks:
          - "10.250.0.0/16"
        region: westeurope
      """

    Then eventually "findConditionTrue(vpc, 'Ready')" is ok, unless:
      | findConditionFalse(vpc, 'Ready') |

    When resource "vpc" is deleted

    Then eventually resource "vpc" does not exist
