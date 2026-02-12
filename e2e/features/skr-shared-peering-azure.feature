Feature: AzureVpcPeering feature

  @skr @azure @peering
  Scenario: AzureVpcPeering with dynamically allocated target scenario

    Given there is shared SKR with "Azure" provider

    Given resource declaration:
      | Alias   | Kind            | ApiVersion                              | Name                           | Namespace |
      | peering | AzureVpcPeering | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                    |           |
      | pod     | Pod             | v1                                      | ${peering.metadata.name ?? ''} |           |

    Given tf module "tf" is applied:
      | source                        | ./azure-peering-target       |
      | provider                      | hashicorp/azurerm 4.55.0     |
      | provider                      | hashicorp/random 3.7.2       |
      | location                      | "westeurope"                 |
      | name                          | "${_.peering.name}"          |
      | virtual_network_address_space | "192.168.255.0/25"           |

    When resource "peering" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: AzureVpcPeering
      spec:
        remotePeeringName: e2e
        remoteVnet: ${tf.vnet_id}
        deleteRemotePeering: true
      """

    Then eventually "peering.status.state == 'Connected'" is ok, unless:
      | peering.status.state == 'Error' |

    And HTTP operation succeeds:
      | Url            | http://${tf.private_ip_address}/base64/SFRUUEJJTiBpcyBhd2Vzb21l |
      | ExpectedOutput | HTTPBIN is awesome                                               |
      | MaxTime        | 30                                                               |

    When resource "peering" is deleted
    Then eventually resource "peering" does not exist

    Then tf module "tf" is destroyed
