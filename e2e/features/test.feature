Feature: AWS NfsVolume feature

  @test @skr
  Scenario: Test scenario

    Given there is shared SKR with "AWS" provider

    Given resource declaration:
      | Alias  | Kind            | ApiVersion                              | Name                         | Namespace |
      | cm     | ConfigMap       | v1                                      | e2e-${id()}                  |           |

    Given tf module "tf" is applied:
      | source             | ./noop  |

    When resource "cm" is created:
        """
        apiVersion: v1
        kind: ConfigMap
        data:
          noop: ${tf.noop}
        """
    
    Then debug wait "mt-test"
    
    Then tf module "tf" is destroyed
