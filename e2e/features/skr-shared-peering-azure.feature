Feature: AzureVpcPeering feature

  @skr @azure @peering
  Scenario: AzureVpcPeering with dynamically allocated target scenario

    Given there is shared SKR with "Azure" provider

    And resource declaration:
      | Alias   | Kind            | ApiVersion                              | Name                           | Namespace |
      | peering | AzureVpcPeering | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                    |           |
      | pod     | Pod             | v1                                      | ${peering.metadata.name ?? ''} |           |

    When resource "peering" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: AzureVpcPeering
      spec:
        remotePeeringName: e2e
        remoteVnet: ${params.peeringVnetId}
        deleteRemotePeering: true
      """

    Then eventually "peering.status.state == 'Connected'" is ok, unless:
      | peering.status.state == 'Error' |

    And HTTP operation succeeds:
      | Url            | http://${params.peeringTargetIp}/base64/SFRUUEJJTiBpcyBhd2Vzb21l |
      | ExpectedOutput | HTTPBIN is awesome                                               |
      | MaxTime        | 30                                                               |

    When resource "peering" is deleted
    Then eventually resource "peering" does not exist
