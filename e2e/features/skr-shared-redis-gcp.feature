Feature: GcpRedisInstance feature

  @skr @gcp @redis
  Scenario: GcpRedisInstance scenario

    Given there is shared SKR with "GCP" provider

    And resource declaration:
      | Alias  | Kind             | ApiVersion                              | Name                         | Namespace |
      | redis  | GcpRedisInstance | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                  |           |
      | secret | Secret           | v1                                      | ${redis.metadata.name ?? ''} |           |

    When resource "redis" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: GcpRedisInstance
      spec:
        redisTier: P1
        redisVersion: REDIS_7_0
        authEnabled: true
        redisConfigs:
          maxmemory-policy: volatile-lru
          activedefrag: "yes"
        maintenancePolicy:
          dayOfWeek:
            day: "SATURDAY"
            startTime:
                hours: 15
                minutes: 45
      """

    Then eventually "redis.status.state == 'Ready'" is ok, unless:
      | redis.status.state == 'Error' |

    And "findConditionTrue(redis, 'Ready')" is ok

    And eventually "secret.data.host" is ok
    And eventually "secret.data.port" is ok
    And eventually "secret.data.authString" is ok
    And eventually "secret.data['CaCert.pem']" is ok

    And Redis "PING" gives "PONG" with:
      | Host | Secret | ${redis.metadata.name} | host       |
      | Port | Secret | ${redis.metadata.name} | port       |
      | Auth | Secret | ${redis.metadata.name} | authString |
      | TLS  | True   |                        |            |
      | CA   | Secret | ${redis.metadata.name} | CaCert.pem |

    When resource "redis" is deleted
    Then eventually resource "secret" does not exist
    And eventually resource "redis" does not exist
