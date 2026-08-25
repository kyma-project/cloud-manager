Feature: AliCloud NfsVolume feature

  # SKR-level NFS (AliCloud NAS) flow against a real AliCloud shoot.
  #
  # Runs the full stack: SKR AlicloudNfsVolume -> KCP NfsInstance -> real AliCloud NAS
  # file system + mount target -> PersistentVolume/PersistentVolumeClaim -> a pod that
  # mounts the PVC and reads/writes a file.
  #
  # Prerequisites (see e2e/cmd/README.md):
  #   - a KCP running cloud-manager built from this branch with the "alicloud" feature flag enabled
  #   - GARDEN_KUBECONFIG with an "alicloud" CloudProfile and an AliCloud SecretBinding
  #   - a shared SKR runtime on an AliCloud shoot (alias "shared-alicloud"), or provisioning enabled
  #
  # Run with:
  #   go test ./e2e/tests -godog.tags="@skr && @alicloud && @nfs" -godog.format=pretty
  @skr @alicloud @nfs
  Scenario: AlicloudNfsVolume scenario

    Given there is shared SKR with "AliCloud" provider

    And resource declaration:
      | Alias | Kind                  | ApiVersion                              | Name                       | Namespace |
      | vol   | AlicloudNfsVolume     | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                |           |
      | pv    | PersistentVolume      | v1                                      | ${vol.status.id ?? ''}     |           |
      | pvc   | PersistentVolumeClaim | v1                                      | ${vol.metadata.name ?? ''} |           |

    When resource "vol" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: AlicloudNfsVolume
      spec:
        capacity: 20G
        storageType: Performance
      """

    Then eventually "vol.status.state == 'Ready'" is ok, unless:
      | vol.status.state == 'Error' |
      | #timeout=20m                |

    And eventually "pv.status.phase == 'Bound'" is ok
    And eventually "pvc.status.phase == 'Bound'" is ok

    # mount the PVC in a pod and verify read/write over the real NAS mount
    And PVC "pvc" file operations succeed:
      | Operation | Path     | Content   |
      | Create    | test.txt | hello nfs |
      | Contains  | test.txt | hello nfs |

    When resource "vol" is deleted

    Then eventually resource "pvc" does not exist
    And eventually resource "pv" does not exist
    And eventually resource "vol" does not exist
