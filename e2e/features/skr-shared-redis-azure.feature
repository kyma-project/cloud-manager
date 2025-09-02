Feature: AzureRedisInstance feature

  @skr @azure @redis
  Scenario: AzureRedisInstance scenario

    Given there is SKR with "Azure" provider and default IpRange

    And resource declaration:
      | Alias  | Kind               | ApiVersion                              | Name                         | Namespace |
      | redis  | AzureRedisInstance | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                  |           |
      | secret | Secret             | v1                                      | ${redis.metadata.name ?? ''} |           |

    When resource "redis" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: AzureRedisInstance
      spec:
        redisTier: P1
        redisConfiguration:
          maxclients: "8"
        redisVersion: "6.0"
      """

    Then eventually "redis.status.state == 'Ready'" is ok, unless:
      | redis.status.state == 'Error' |

    And Redis "PING" gives "PONG" with:
      | Host | Secret | ${redis.metadata.name} | host       |
      | Port | Secret | ${redis.metadata.name} | port       |
      | Auth | Secret | ${redis.metadata.name} | authString |
      | TLS  | True   |                        |            |

    When resource "redis" is deleted
    Then eventually resource "secret" does not exist
    And eventually resource "redis" does not exist
