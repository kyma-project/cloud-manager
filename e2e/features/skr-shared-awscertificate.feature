Feature: AwsCertificate feature

  @skr @aws @certificate
  Scenario: AwsCertificate imports certificate to ACM then deletes it

    Given there is shared SKR with "AWS" provider

    And resource declaration:
      | Alias       | Kind            | ApiVersion                              | Name                    | Namespace |
      | certificate | AwsCertificate  | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}             |           |
      | secret      | Secret          | v1                                      | e2e-test-certificate    | default   |

    # Note: Secret "e2e-test-certificate" is created by the workflow before running tests
    # For local development: Run ./e2e/scripts/create-test-certificate.sh and ./e2e/scripts/create-certificate-secret.sh

    # Create AwsCertificate referencing the pre-existing Secret
    When resource "certificate" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: AwsCertificate
      spec:
        secretRef:
          name: e2e-test-certificate
          namespace: default
      """

    Then eventually "certificate.status.state == 'Ready'" is ok, unless:
      | certificate.status.state == 'Error' |
      | #timeout=10m                        |

    And "findConditionTrue(certificate, 'Ready')" is ok
    And "certificate.status.arn" is ok
    And "certificate.status.arn.startsWith('arn:aws:acm:')" is ok
    And "certificate.status.expirationDate" is ok

    # Clean up AwsCertificate (Secret is cleaned up by workflow)
    When resource "certificate" is deleted
    Then eventually resource "certificate" does not exist
