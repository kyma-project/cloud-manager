Feature: VpcNetwork AWS

  Scenario: VpcNetwork AWS is created and deleted

    Given current cluster is "kcp"

    Given resource declaration:
      | Alias     | Kind                  | ApiVersion                              | Name               | Namespace |
      | vpc       | VpcNetwork            | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}        |           |

    Given Subscription "subscription" exists for "AWS" provider

    When resource "vpc" is created:
      """
      apiVersion: cloud-control.kyma-project.io/v1beta1
      kind: VpcNetwork
      spec:
        subscription: {{pera.name}}
        cidrBlocks:
          - "10.250.0.0/16"
         region: us-east-1
      """

    Then eventually "findConditionTrue(vpc, 'Ready')" is ok, unless:
      | findConditionFalse(vol, 'Ready') |

    Then debug wait "cm-mt-vpc"

    When resource "vpc" is deleted

    Then resource "vpc" does not exist
