Feature: Certificate feature

  @skr @aws @certificate
  Scenario: Certificate imports certificate to ACM then deletes it

    Given there is shared SKR with "AWS" provider

    And resource declaration:
      | Alias        | Kind        | ApiVersion                              | Name               | Namespace    |
      | certificate  | Certificate | cloud-resources.kyma-project.io/v1beta1 | e2e-${scenarioId}  |              |
      | secret       | Secret      | v1                                      | kyma-gateway-certs | istio-system |
      | gardenerCert | Certificate | cert.gardener.cloud/v1alpha1            | kyma-tls-cert      | istio-system |


    # Wait for pre-existing gardener Certificate (kyma-tls-cert) to be Ready
    Then eventually "findConditionTrue(gardenerCert, 'Ready')" is ok, unless:
      | findConditionFalse(gardenerCert, 'Ready') |
      | #timeout=10m                              |


    # Create Certificate referencing the pre-existing Secret
    When resource "certificate" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: Certificate
      spec:
        secretRef:
          name: ${_.secret.name}
          namespace: ${_.secret.namespace}
      """

    Then eventually "findConditionTrue(certificate, 'Ready')" is ok, unless:
      | findConditionFalse(certificate, 'Ready') |
      | #timeout=10m                             |

    And "certificate.status.providerId" is ok
    And "certificate.status.providerId.startsWith('arn:aws:acm:')" is ok

    # Clean up
    When resource "certificate" is deleted
    Then eventually resource "certificate" does not exist
