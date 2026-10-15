Feature: Certificate feature

  @skip @skr @aws @certificate
  Scenario: Certificate imports certificate to ACM then deletes it

    Given there is shared SKR with "AWS" provider

    And resource declaration:
      | Alias       | Kind        | ApiVersion                              | Name                 | Namespace |
      | certificate | Certificate | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}          |           |
      | secret      | Secret      | v1                                      | e2e-test-certificate | default   |


    # Create Certificate referencing the pre-existing Secret
    When resource "certificate" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: Certificate
      spec:
        secretRef:
          name: e2e-test-certificate
          namespace: default
      """

    Then eventually "certificate.status.state == 'Ready'" is ok, unless:
      | certificate.status.state == 'Error' |
      | #timeout=10m                        |

    And "findConditionTrue(certificate, 'Ready')" is ok
    And "certificate.status.providerId" is ok
    And "certificate.status.providerId.startsWith('arn:aws:acm:')" is ok

    # Clean up Certificate and Secret
    When resource "certificate" is deleted
    Then eventually resource "certificate" does not exist
