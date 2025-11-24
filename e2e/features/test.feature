Feature: AWS NfsVolume feature

  @test @skr
  Scenario: AwsNfsVolume scenario

    Given there is shared SKR with "AWS" provider

    Given resource declaration:
      | Alias     | Kind                  | ApiVersion                              | Name                                            | Namespace |
      | vol       | AwsNfsVolume          | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                                     |           |
      | pv        | PersistentVolume      | v1                                      | ${vol.status.id ?? ''}                          |           |
      | pvc       | PersistentVolumeClaim | v1                                      | ${vol.metadata.name ?? ''}                      |           |

    When resource "vol" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: AwsNfsVolume
      spec:
        capacity: 10G
      """

    Then eventually "findConditionTrue(vol, 'Ready')" is ok, unless:
      | findConditionTrue(vol, 'Error') |

    And "vol.status.state == 'Ready'" is ok
    And eventually "pv.status.phase == 'Bound'" is ok
    And eventually "pvc.status.phase == 'Bound'" is ok

    # write initial content to the file, one that will be backed up
    And PVC "pvc" file operations succeed:
      | Operation | Path     | Content     |
      | Create    | test.txt | first value |

    When resource "vol" is deleted
    Then eventually resource "pvc" does not exist
    And eventually resource "pv" does not exist
    And eventually resource "vol" does not exist
