Feature: AwsVpcPeering feature

  @skr @aws @peering
  Scenario: AwsVpcPeering scenario

    Given there is shared SKR with "AWS" provider

    Given resource declaration:
      | Alias   | Kind          | ApiVersion                              | Name        | Namespace |
      | peering | AwsVpcPeering | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()} |           |

    Given tf module "tf" is applied:
      | source             | ./aws-peering-target        |
      | provider           | hashicorp/aws@5.0           |
      | region             | "us-east-1"                 |
      | name               | "${_.peering.name}"         |
      | vpc_cidr           | "10.3.0.0/16"               |
      | public_subnet_cidr | "10.3.124.0/24"             |

    When resource "peering" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: AwsVpcPeering
      spec:
        remoteAccountId: "${tf.account_id}"
        remoteRegion: "us-east-1"
        remoteVpcId: "${tf.vpc_id}"
        deleteRemotePeering: true
      """

    Then eventually "peering.status.state == 'active'" is ok, unless:
      | peering.status.state == 'Error' |

    And HTTP operation succeeds:
      | Url            | http://${tf.private_ip_address}/base64/SFRUUEJJTiBpcyBhd2Vzb21l |
      | ExpectedOutput | HTTPBIN is awesome                                               |
      | MaxTime        | 30                                                               |

    When resource "peering" is deleted
    Then eventually resource "peering" does not exist

    Then tf module "tf" is destroyed
