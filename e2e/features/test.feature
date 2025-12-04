Feature: AWS NfsVolume feature

  @test @skr
  Scenario: Cloud API

    Given AWS VPC network is created:
      | Name         | some-name     |
      | CIDR         | 10.250.0.0/16 |

    Given AWS subnet is created:
      | Name                 | some-subnet        |
      | VPC                  | some-name          |
      | CIDR                 | 10.250.0.0/16      |
