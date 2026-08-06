Feature: SapNfsVolume feature

  @skr @openstack @nfs
  Scenario: SapNfsVolume Snapshot and Restore scenario

    Given there is shared SKR with "OpenStack" provider

    And resource declaration:
      | Alias       | Kind                             | ApiVersion                              | Name                                              | Namespace |
      | vol         | SapNfsVolume                     | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                                       |           |
      | pv          | PersistentVolume                 | v1                                      | ${vol.status.id ?? ''}                            |           |
      | pvc         | PersistentVolumeClaim            | v1                                      | ${vol.metadata.name ?? ''}                        |           |
      | snapshot    | SapNfsVolumeSnapshot             | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                                       |           |
      | restore     | SapNfsVolumeSnapshotRestore      | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                                       |           |
      | newrestore  | SapNfsVolumeSnapshotRestore      | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                                       |           |
      | vol2        | SapNfsVolume                     | cloud-resources.kyma-project.io/v1beta1 | ${vol.metadata.name ?? ''}-restored               |           |
      | pv2         | PersistentVolume                 | v1                                      | ${vol2.status.id ?? ''}                           |           |
      | pvc2        | PersistentVolumeClaim            | v1                                      | ${vol2.metadata.name ?? ''}                       |           |
      | schedule    | SapNfsVolumeSnapshotSchedule     | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                                       |           |
      | schSnapshot | SapNfsVolumeSnapshot             | cloud-resources.kyma-project.io/v1beta1 | ${schedule.status.lastCreatedBackup.name ?? ''}   |           |

    # ── Phase A: Volume Setup ──

    When resource "vol" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: SapNfsVolume
      spec:
        capacityGb: 1024
      """

    Then eventually "vol.status.state == 'Ready'" is ok, unless:
      | vol.status.state == 'Error' |
      | #timeout=20m                |

    And eventually "pv.status.phase == 'Bound'" is ok
    And eventually "pvc.status.phase == 'Bound'" is ok

    # write initial content to the file, one that will be snapshotted
    And PVC "pvc" file operations succeed:
      | Operation | Path     | Content     |
      | Create    | test.txt | first value |
      | Contains  | test.txt | first value |

    # ── Phase B: Snapshot Creation ──

    When resource "snapshot" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: SapNfsVolumeSnapshot
      spec:
        sourceVolume:
          name: ${vol.metadata.name}
          namespace: ${vol.metadata.namespace}
      """

    Then eventually "snapshot.status.state == 'Ready'" is ok

    # ── Phase C: In-Place Restore (ExistingVolume) ──

    # write some other content to the file, one that will be overwritten by restore
    When PVC "pvc" file operations succeed:
      | Operation | Path     | Content      |
      | Create    | test.txt | second value |
      | Contains  | test.txt | second value |

    When resource "restore" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: SapNfsVolumeSnapshotRestore
      spec:
        sourceSnapshot:
          name: ${snapshot.metadata.name}
          namespace: ${snapshot.metadata.namespace}
        destination:
          existingVolume:
            name: ${vol.metadata.name}
            namespace: ${vol.metadata.namespace}
      """

    Then eventually "restore.status.state == 'Done'" is ok

    # check file content matches the original restored value
    And PVC "pvc" file operations succeed:
      | Operation | Path     | Content     |
      | Contains  | test.txt | first value |

    When resource "restore" is deleted
    Then eventually resource "restore" does not exist

    # ── Phase D: New-Volume Restore (NewVolume) ──

    When resource "newrestore" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: SapNfsVolumeSnapshotRestore
      spec:
        sourceSnapshot:
          name: ${snapshot.metadata.name}
          namespace: ${snapshot.metadata.namespace}
        destination:
          newVolume:
            metadata:
              name: ${vol.metadata.name}-restored
            spec:
              capacityGb: 1024
      """

    Then eventually "newrestore.status.state == 'Done'" is ok

    Then eventually "vol2.status.state == 'Ready'" is ok, unless:
      | vol2.status.state == 'Error' |
      | #timeout=20m                 |

    And eventually "pv2.status.phase == 'Bound'" is ok
    And eventually "pvc2.status.phase == 'Bound'" is ok

    # check file content matches the original restored value on the new volume
    And PVC "pvc2" file operations succeed:
      | Operation | Path     | Content     |
      | Contains  | test.txt | first value |

    When resource "newrestore" is deleted
    When resource "vol2" is deleted

    Then eventually resource "newrestore" does not exist
    Then eventually resource "pvc2" does not exist
    And eventually resource "pv2" does not exist
    And eventually resource "vol2" does not exist

    # ── Phase E: Snapshot Schedule ──

    When resource "schedule" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: SapNfsVolumeSnapshotSchedule
      spec:
        schedule: "*/5 * * * *"
        prefix: ${vol.metadata.name}-sch
        deleteCascade: true
        template:
          spec:
            sourceVolume:
              name: ${vol.metadata.name}
              namespace: ${vol.metadata.namespace}
      """

    Then eventually "schedule.status.state == 'Active'" is ok
    And eventually "schedule.status.lastCreatedBackup.name" is ok
    And eventually "schSnapshot.status.state == 'Ready'" is ok

    # ── Phase F: Cleanup ──

    When resource "schedule" is deleted
    When resource "snapshot" is deleted
    When resource "vol" is deleted

    Then eventually resource "schSnapshot" does not exist
    And eventually resource "schedule" does not exist

    Then eventually resource "snapshot" does not exist

    Then eventually resource "pvc" does not exist
    And eventually resource "pv" does not exist
    And eventually resource "vol" does not exist
