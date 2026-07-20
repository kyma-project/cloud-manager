Feature: NfsInstance AliCloud

  # KCP-level NFS (AliCloud NAS) flow against real AliCloud.
  #
  # NOTE: @skip — this scenario needs a KCP Scope provisioned for an AliCloud shoot
  # (and the "alicloud" feature flag enabled). The e2e godog harness does not yet
  # expose a "Scope exists" step (only "Subscription exists"), and KCP IpRange /
  # NfsInstance require spec.scope. Enabling this scenario is a follow-up once the
  # harness can provision an AliCloud Scope.
  #
  # The no-dev-cluster coverage for this flow already exists as an envtest controller
  # test driven against the AliCloud mock:
  #   internal/controller/cloud-control/nfsinstance_alicloud_test.go
  @skip @kcp @alicloud @nfs
  Scenario: NfsInstance AliCloud is created and deleted

    Given current cluster is "kcp"

    Given resource declaration:
      | Alias   | Kind        | ApiVersion                            | Name        | Namespace |
      | iprange | IpRange     | cloud-control.kyma-project.io/v1beta1 | e2e-${id()} |           |
      | nfs     | NfsInstance | cloud-control.kyma-project.io/v1beta1 | e2e-${id()} |           |

    Given Subscription "subscription" exists for "alicloud" provider

    When resource "iprange" is created:
      """
      apiVersion: cloud-control.kyma-project.io/v1beta1
      kind: IpRange
      spec:
        remoteRef:
          namespace: kcp-system
          name: e2e-${id()}
        scope:
          name: ${subscription.status.scopeRef.name ?? ''}
        cidr: "10.250.4.0/22"
      """

    Then eventually "findConditionTrue(iprange, 'Ready')" is ok, unless:
      | findConditionFalse(iprange, 'Ready') |

    When resource "nfs" is created:
      """
      apiVersion: cloud-control.kyma-project.io/v1beta1
      kind: NfsInstance
      spec:
        remoteRef:
          namespace: kcp-system
          name: e2e-${id()}
        ipRange:
          name: ${iprange.metadata.name}
        scope:
          name: ${subscription.status.scopeRef.name ?? ''}
        instance:
          alicloud:
            storageType: Performance
            protocolType: NFS
      """

    Then eventually "findConditionTrue(nfs, 'Ready')" is ok, unless:
      | findConditionFalse(nfs, 'Ready') |

    And "nfs.status.host != ''" is ok
    And "nfs.status.path == '/'" is ok

    When resource "nfs" is deleted
    Then eventually resource "nfs" does not exist

    When resource "iprange" is deleted
    Then eventually resource "iprange" does not exist
