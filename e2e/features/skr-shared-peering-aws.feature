Feature: AwsVpcPeering feature

  @skr @aws @peering
  Scenario: AwsVpcPeering scenario

    Given there is shared SKR with "AWS" provider

    And resource declaration:
      | Alias   | Kind          | ApiVersion                              | Name                           | Namespace |
      | peering | AwsVpcPeering | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                    |           |
      | pod     | Pod           | v1                                      | ${peering.metadata.name ?? ''} |           |

    When resource "peering" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: AwsVpcPeering
      spec:
        remoteAccountId: "642531956841"
        remoteRegion: "us-east-1"
        remoteVpcId: "vpc-0709fb45c2be50920"
        deleteRemotePeering: true
      """

    Then eventually "peering.status.state == 'active'" is ok, unless:
      | peering.status.state == 'Error' |

    When resource "pod" is created:
      """
      apiVersion: v1
      kind: Pod
      spec:
        containers:
          - name: netcat
            resources:
              limits:
                memory: 512Mi
                cpu: "1"
              requests:
                memory: 256Mi
                cpu: "0.2"
            image: alpine
            command:
              - "nc"
            args:
              - "-zv"
              - "10.3.124.194"
              - "22"
        restartPolicy: Never
      """
    Then eventually "pod.status.phase == 'Succeeded'" is ok, unless:
      | pod.status.phase == 'Failed' |

    And logs of container "netcat" in pod "pod" contain "10.3.124.194 (10.3.124.194:22) open"

    When resource "pod" is deleted
    Then eventually resource "pod" does not exist

    When resource "peering" is deleted
    Then eventually resource "peering" does not exist
