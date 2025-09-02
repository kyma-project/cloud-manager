Feature: GcpVpcPeering feature

  @skr @gcp @peering
  Scenario: GcpVpcPeering scenario

    Given there is SKR with "GCP" provider and default IpRange

    And resource declaration:
      | Alias   | Kind          | ApiVersion                              | Name                           | Namespace |
      | peering | GcpVpcPeering | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                    |           |
      | pod     | Pod           | v1                                      | ${peering.metadata.name ?? ''} |           |

    When resource "peering" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: GcpVpcPeering
      spec:
        remotePeeringName: "${_.peering.name}"
        remoteProject: "sap-sc-learn"
        remoteVpc: "vpc-peering-e2e-tests"
        deleteRemotePeering: true
        importCustomRoutes: false
      """

    Then eventually "peering.status.state == 'Connected'" is ok, unless:
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
            - "10.240.254.2"
            - "22"
        restartPolicy: Never
      """

    Then eventually "pod.status.phase == 'Succeeded'" is ok, unless:
      | pod.status.phase == 'Failed' |

    And logs of container "netcat" in pod "pod" contain "10.240.254.2 (10.240.254.2:22) open"

    When resource "pod" is deleted
    Then eventually resource "pod" does not exist

    When resource "peering" is deleted
    Then eventually resource "peering" does not exist
