Feature: AWS NfsVolume feature

  @skr @aws @nfs
  Scenario: AwsNfsVolume scenario

    Given there is shared SKR with "AWS" provider

    Given resource declaration:
      | Alias     | Kind                  | ApiVersion                              | Name                                            | Namespace |
      | vol       | AwsNfsVolume          | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                                     |           |
      | pv        | PersistentVolume      | v1                                      | ${vol.status.id ?? ''}                          |           |
      | pvc       | PersistentVolumeClaim | v1                                      | ${vol.metadata.name ?? ''}                      |           |
      | backup    | AwsNfsVolumeBackup    | cloud-resources.kyma-project.io/v1beta1 | ${vol.metadata.name ?? ''}                      |           |
      | restore   | AwsNfsVolumeRestore   | cloud-resources.kyma-project.io/v1beta1 | ${vol.metadata.name ?? ''}                      |           |
      | schedule  | AwsNfsBackupSchedule  | cloud-resources.kyma-project.io/v1beta1 | ${vol.metadata.name ?? ''}                      |           |
      | schBackup | AwsNfsVolumeBackup    | cloud-resources.kyma-project.io/v1beta1 | ${schedule.status.lastCreatedBackup.name ?? ''} |           |

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

    When resource "backup" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: AwsNfsVolumeBackup
      spec:
        source:
          volume:
            name: ${vol.metadata.name}
      """

    Then eventually "backup.status.state == 'Ready'" is ok

    # write some other content to the file, one that will be overwritten by restore
    When PVC "pvc" file operations succeed:
      | Operation | Path     | Content      |
      | Create    | test.txt | second value |

    And resource "restore" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: AwsNfsVolumeRestore
      spec:
        source:
          backup:
            name: ${backup.metadata.name}
            namespace: ${backup.metadata.namespace}
      """

    Then eventually "restore.status.state == 'Done'" is ok

    # check file content matches the original restored value
    And PVC "pvc" file operations succeed:
      | Operation | Path     | Content     |
      | Contains  | test.txt | first value |

    When resource "schedule" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: AwsNfsBackupSchedule
      spec:
         nfsVolumeRef:
           name: ${vol.metadata.name}
         schedule: "*/5 * * * *"
         prefix: ${vol.metadata.name}-sch
         deleteCascade: true
      """

    # there will be a race condition between schedule and a pod that would write a new fourth value
    # to be overwritten by the scheduled backup, so we're not doing that nor testing
    # if that sceduled backup would do the restore, since we already tested above
    # that restore overwrittes with backed up content

    Then eventually "schedule.status.state == 'Active'" is ok
    And eventually "schedule.status.lastCreatedBackup.name" is ok
    And eventually "schBackup.status.state == 'Ready'" is ok


    When resource "schedule" is deleted
    Then eventually resource "schedule" does not exist
    And eventually resource "schBackup" does not exist

    When resource "restore" is deleted
    Then eventually resource "restore" does not exist

    When resource "backup" is deleted
    Then eventually resource "backup" does not exist

    When resource "vol" is deleted
    Then eventually resource "pvc" does not exist
    And eventually resource "pv" does not exist
    And eventually resource "vol" does not exist
