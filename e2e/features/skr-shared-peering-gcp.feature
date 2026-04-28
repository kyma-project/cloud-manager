Feature: GcpVpcPeering feature

  @skr @gcp @peering
  Scenario: GcpVpcPeering scenario

    Given there is shared SKR with "GCP" provider

    And resource declaration:
      | Alias   | Kind          | ApiVersion                              | Name                           | Namespace |
      | peering | GcpVpcPeering | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                    |           |
      | pod     | Pod           | v1                                      | ${peering.metadata.name ?? ''} |           |

    Given tf module "tf" is applied:
      | source                        | ./gcp-peering-target         |
      | provider                      | hashicorp/google@7.29.0      |
      | location                      | "us-east1"                   |
      | name                          | "${_.peering.name}"          |
      | subnet_cidr                   | "192.168.255.0/25"           |

    When resource "peering" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: GcpVpcPeering
      spec:
        remotePeeringName: "${_.peering.name}"
        remoteProject: "${tf.project_id}"
        remoteVpc: "${tf.vpc_id}"
        deleteRemotePeering: true
        importCustomRoutes: false
      """

    Then eventually "peering.status.state == 'Connected'" is ok, unless:
      | peering.status.state == 'Error' |
      | #timeout=10m                    |

    And HTTP operation succeeds:
      | Url            | http://${tf.instance_ip_address}|
      | ExpectedOutput | GCP VPC Peering is working!     |
      | MaxTime        | 30                              |

    When resource "peering" is deleted
    Then eventually resource "peering" does not exist

    Then tf module "tf" is destroyed
