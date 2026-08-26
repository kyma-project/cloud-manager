Feature: AlicloudRedisCluster feature

  @skr @alicloud @rediscluster
  Scenario: AlicloudRedisCluster scenario

    Given there is shared SKR with "AliCloud" provider

    Given eventually timeout is "30m"

    And resource declaration:
      | Alias  | Kind                   | ApiVersion                              | Name                         | Namespace |
      | redis  | AlicloudRedisCluster   | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                  |           |
      | secret | Secret                 | v1                                      | ${redis.metadata.name ?? ''} |           |

    When resource "redis" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: AlicloudRedisCluster
      spec:
        redisTier: C3
        shardCount: 4
        replicasPerShard: 0
        engineVersion: "5.0"
        parameters:
          maxmemory-policy: volatile-lru
      """

    Then eventually "redis.status.state == 'Ready'" is ok, unless:
      | redis.status.state == 'Error' |
      | #timeout=30m                  |

    And "findConditionTrue(redis, 'Ready')" is ok

    And eventually "secret.data.host" is ok
    And eventually "secret.data.port" is ok
    And eventually "secret.data.authString" is ok
    And eventually "secret.data['CaCert.pem']" is ok

    And Redis "PING" gives "PONG" with:
      | Host        | Secret | ${redis.metadata.name} | host       |
      | Port        | Secret | ${redis.metadata.name} | port       |
      | Auth        | Secret | ${redis.metadata.name} | authString |
      | TLS         | True   |                        |            |
      | CA          | Secret | ${redis.metadata.name} | CaCert.pem |
      | ClusterMode | True   |                        |            |

    When resource "redis" is deleted

    Then eventually resource "secret" does not exist
    And eventually resource "redis" does not exist
