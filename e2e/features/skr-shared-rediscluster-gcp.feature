Feature: GcpRedisCluster feature

  @skr @gcp @rediscluster
  Scenario: GcpRedisCluster scenario

    Given there is shared SKR with "GCP" provider

    And resource declaration:
      | Alias  | Kind            | ApiVersion                              | Name                         | Namespace |
      | redis  | GcpRedisCluster | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                  |           |
      | secret | Secret          | v1                                      | ${redis.metadata.name ?? ''} |           |

    When resource "redis" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: GcpRedisCluster
      spec:
        redisTier: C1
        shardCount: 3
        replicasPerShard: 1
        redisConfigs:
          maxmemory-policy: volatile-lru
      """

    Then eventually "redis.status.state == 'Ready'" is ok, unless:
      | redis.status.state == 'Error' |

    And "findConditionTrue(redis, 'Ready')" is ok

    And eventually "secret.data.host" is ok
    And eventually "secret.data.port" is ok
    And eventually "secret.data.authString" is ok

    And Redis "PING" gives "PONG" with:
      | Host        | Secret | ${redis.metadata.name} | host       |
      | Port        | Secret | ${redis.metadata.name} | port       |
      | Auth        | Secret | ${redis.metadata.name} | authString |
      | TLS         | True   |                        |            |
      | ClusterMode | True   |                        |            |

    When resource "redis" is deleted

    Then eventually resource "redis" does not exist

    And resource "secret" does not exist
