Feature: AWS NfsVolume feature

  @test @skr
  Scenario: Cloud API

    Given tf module "peeringTarget" is applied:
      | source             | terraform-aws-modules/vpc/aws ~> 6.5.1  |
      | provider           | hashicorp/azurerm ~> 5.0 |
      | name               | "{id()}"             |
      | cidr               | "10.240.0.0/16"      |

    # git::https://github.com/kyma/cloud-manager.git//e2e/tf/aws-peering-taget?ref=my-test-branch

    When resource "peering" is created:
        """
        apiVersion: cloud-resources.kyma-project.io/v1beta1
        kind: AwsVPCPeering
        metadata:
          name: some-peering
        spec:
          vpc: ${peeringTarget.vpc_id}
        """

    When resource "pod" is created:
      """
        apiVersion: v1
        kind: Pod
        spec:
          containers:
          - name: aws-cli
            image: amazonlinux
            command: ["/bin/sh", "-c"]
            args:
                - |
                nc -zv ${peeringTarget.private_ip_address}
      """