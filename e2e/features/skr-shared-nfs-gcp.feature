Feature: GcpNfsVolume feature

  @skr @gcp @nfs
  Scenario: GcpNfsVolume/Backup/Restore scenario

    Given there is SKR with "GCP" provider and default IpRange

    And resource declaration:
      | Alias     | Kind                  | ApiVersion                              | Name                                            | Namespace |
      | vol       | GcpNfsVolume          | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                                     |           |
      | pv        | PersistentVolume      | v1                                      | ${vol.status.id ?? ''}                          |           |
      | pvc       | PersistentVolumeClaim | v1                                      | ${vol.metadata.name ?? ''}                      |           |
      | backup    | GcpNfsVolumeBackup    | cloud-resources.kyma-project.io/v1beta1 | ${vol.metadata.name ?? ''}                      |           |
      | restore   | GcpNfsVolumeRestore   | cloud-resources.kyma-project.io/v1beta1 | ${vol.metadata.name ?? ''}                      |           |
      | vol2      | GcpNfsVolume          | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                                     |           |
      | pv2       | PersistentVolume      | v1                                      | ${vol2.status.id ?? ''}                         |           |
      | pvc2      | PersistentVolumeClaim | v1                                      | ${vol2.metadata.name ?? ''}                     |           |
      | schedule  | GcpNfsBackupSchedule  | cloud-resources.kyma-project.io/v1beta1 | ${vol.metadata.name ?? ''}                      |           |
      | schBackup | GcpNfsVolumeBackup    | cloud-resources.kyma-project.io/v1beta1 | ${schedule.status.lastCreatedBackup.name ?? ''} |           |

    When resource "vol" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: GcpNfsVolume
      spec:
        capacityGb: 1024
      """

    Then eventually "vol.status.state == 'Ready'" is ok, unless:
      | vol.status.state == 'Error' |
    And eventually "pv.status.phase == 'Bound'" is ok
    And eventually "pvc.status.phase == 'Bound'" is ok

    # write initial content to the file, one that will be backed up
    And PVC "pvc" file operations succeed:
      | Operation | Path     | Content     |
      | Create    | test.txt | first value |

    When resource "backup" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: GcpNfsVolumeBackup
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

    When resource "restore" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: GcpNfsVolumeRestore
      spec:
        source:
          backup:
            name: ${backup.metadata.name}
            namespace: ${backup.metadata.namespace}
        destination:
          volume:
            name: ${vol.metadata.name}
      """

    Then eventually "restore.status.state == 'Done'" is ok

    # check file content matches the original restored value
    And PVC "pvc" file operations succeed:
      | Operation | Path     | Content     |
      | Contains  | test.txt | first value |

    When resource "restore" is deleted
    Then eventually resource "restore" does not exist

    When resource "vol" is deleted
    Then eventually resource "pvc" does not exist
    And eventually resource "pv" does not exist
    And eventually resource "vol" does not exist


    # create volume from backup

    When resource "vol2" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: GcpNfsVolume
      spec:
        capacityGb: 1024
        sourceBackup:
          name: ${backup.metadata.name}
          namespace: ${backup.metadata.namespace}
      """

    Then eventually "vol2.status.state == 'Ready'" is ok, unless:
      | vol2.status.state == 'Error' |
    And eventually "pv2.status.phase == 'Bound'" is ok
    And eventually "pvc2.status.phase == 'Bound'" is ok

    # check file content matches the original restored value
    And PVC "pvc2" file operations succeed:
      | Operation | Path     | Content     |
      | Contains  | test.txt | first value |

    When resource "schedule" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: GcpNfsBackupSchedule
      spec:
         nfsVolumeRef:
           name: ${vol2.metadata.name}
         schedule: "*/5 * * * *"
         prefix: ${vol2.metadata.name}-sch
         deleteCascade: true
      """

    Then eventually "schedule.status.state == 'Active'" is ok
    And eventually "schedule.status.lastCreatedBackup.name" is ok
    And eventually "schBackup.status.state == 'Ready'" is ok

    When resource "schedule" is deleted
    Then eventually resource "schBackup" does not exist
    And eventually resource "schedule" does not exist

    When resource "vol2" is deleted
    Then eventually resource "pvc2" does not exist
    And eventually resource "pv2" does not exist
    And eventually resource "vol2" does not exist

    When resource "backup" is deleted
    Then eventually resource "backup" does not exist
