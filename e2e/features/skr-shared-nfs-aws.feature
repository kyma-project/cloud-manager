Feature: AWS NfsVolume feature

  @skr @aws @nfs
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
      | #timeout=20m                    |

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


  @skr @aws @nfs @wip
  Scenario: AwsNfsVolume in custom IpRange scenario

    Given there is shared SKR with "AWS" provider

    Given resource declaration:
      | Alias     | Kind                  | ApiVersion                              | Name                                            | Namespace |
      | vol1      | AwsNfsVolume          | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                                     |           |
      | range1    | IpRange               | cloud-resources.kyma-project.io/v1beta1 | default                                         |           |
      | pv1       | PersistentVolume      | v1                                      | ${vol1.status.id ?? ''}                         |           |
      | pvc1      | PersistentVolumeClaim | v1                                      | ${vol1.metadata.name ?? ''}                     |           |
      | vol2      | AwsNfsVolume          | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                                     |           |
      | range2    | IpRange               | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                                     |           |
      | pv2       | PersistentVolume      | v1                                      | ${vol2.status.id ?? ''}                         |           |
      | pvc2      | PersistentVolumeClaim | v1                                      | ${vol2.metadata.name ?? ''}                     |           |

    # create first nfs to trigger default IpRange creation
    When resource "vol1" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: AwsNfsVolume
      spec:
        capacity: 10G
      """

    # wait for default IpRange to be created and ready
    Then eventually "findConditionTrue(range1, 'Ready')" is ok, unless:
      | findConditionTrue(range1, 'Error') |
      | #timeout=20m                       |

    # create custom IpRange
    When resource "range2" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: IpRange
      spec:
        cidr: 10.252.0.0/16
      """

    # wait custom IpRange to be created and ready
    Then eventually "findConditionTrue(range2, 'Ready')" is ok, unless:
      | findConditionTrue(range1, 'Error') |
      | #timeout=20m                       |

    # create second NFS in custom IpRange
    When resource "vol2" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: AwsNfsVolume
      spec:
        capacity: 10G
        ipRange:
          name: ${_.range2.name}
      """

    # verify first NFS works
    Then eventually "findConditionTrue(vol1, 'Ready')" is ok, unless:
      | findConditionTrue(vol1, 'Error') |
      | #timeout=20m                    |
    And "vol1.status.state == 'Ready'" is ok
    And eventually "pv1.status.phase == 'Bound'" is ok
    And eventually "pvc1.status.phase == 'Bound'" is ok
    And PVC "pvc1" file operations succeed:
      | Operation | Path     | Content     |
      | Create    | test.txt | first value |

    # verify second NFS works
    Then eventually "findConditionTrue(vol2, 'Ready')" is ok, unless:
      | findConditionTrue(vol2, 'Error') |
      | #timeout=20m                    |
    And "vol2.status.state == 'Ready'" is ok
    And eventually "pv2.status.phase == 'Bound'" is ok
    And eventually "pvc2.status.phase == 'Bound'" is ok
    And PVC "pvc2" file operations succeed:
      | Operation | Path     | Content     |
      | Create    | test.txt | first value |

    # clean-up and delete --------------------

    Then debug wait "mtwait"

    When resource "vol1" is deleted
    When resource "vol2" is deleted
    When resource "range2" is deleted

    Then eventually resource "vol1" does not exist
    Then eventually resource "vol2" does not exist
    Then eventually resource "range2" does not exist
