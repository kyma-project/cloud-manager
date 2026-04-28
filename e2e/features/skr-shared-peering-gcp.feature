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
            - "${tf.instance_ip_address}"
            - "22"
        restartPolicy: Never
      """

    Then eventually "pod.status.phase == 'Succeeded'" is ok, unless:
      | pod.status.phase == 'Failed' |
      | #timeout=5m                  |

    And logs of container "netcat" in pod "pod" contain "${tf.instance_ip_address} (${tf.instance_ip_address}:22) open":
      | #timeout=2m                  |

    When resource "pod" is deleted
    Then eventually resource "pod" does not exist

    When resource "peering" is deleted
    Then eventually resource "peering" does not exist

    Then tf module "tf" is destroyed
