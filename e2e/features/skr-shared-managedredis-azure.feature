Feature: AzureManagedRedis feature

  @skr @azure @managedredis
  Scenario: AzureManagedRedis S-tier scenario (non-HA, EnterpriseCluster, dev/test)

    Given there is shared SKR with "Azure" provider

    And resource declaration:
      | Alias  | Kind              | ApiVersion                              | Name                         | Namespace |
      | redis  | AzureManagedRedis | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                  |           |
      | secret | Secret            | v1                                      | ${redis.metadata.name ?? ''} |           |

    When resource "redis" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: AzureManagedRedis
      spec:
        tier: S1
      """

    Then eventually "redis.status.state == 'Ready'" is ok, unless:
      | redis.status.state == 'Error' |
      | #timeout=30m                  |

    And "findConditionTrue(redis, 'Ready')" is ok

    And eventually "secret.data.host" is ok
    And eventually "secret.data.port" is ok
    And eventually "secret.data.authString" is ok

    And Redis "PING" gives "PONG" with:
      | Host | Secret | ${redis.metadata.name} | host       |
      | Port | Secret | ${redis.metadata.name} | port       |
      | Auth | Secret | ${redis.metadata.name} | authString |
      | TLS  | True   |                        |            |

    When resource "redis" is deleted

    Then eventually resource "redis" does not exist

    And resource "secret" does not exist


  @skr @azure @managedredis
  Scenario: AzureManagedRedis P-tier scenario (HA, EnterpriseCluster)

    Given there is shared SKR with "Azure" provider

    And resource declaration:
      | Alias  | Kind              | ApiVersion                              | Name                         | Namespace |
      | redis  | AzureManagedRedis | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                  |           |
      | secret | Secret            | v1                                      | ${redis.metadata.name ?? ''} |           |

    When resource "redis" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: AzureManagedRedis
      spec:
        tier: P1
      """

    Then eventually "redis.status.state == 'Ready'" is ok, unless:
      | redis.status.state == 'Error' |
      | #timeout=30m                  |

    And "findConditionTrue(redis, 'Ready')" is ok

    And eventually "secret.data.host" is ok
    And eventually "secret.data.port" is ok
    And eventually "secret.data.authString" is ok

    And Redis "PING" gives "PONG" with:
      | Host | Secret | ${redis.metadata.name} | host       |
      | Port | Secret | ${redis.metadata.name} | port       |
      | Auth | Secret | ${redis.metadata.name} | authString |
      | TLS  | True   |                        |            |

    When resource "redis" is deleted

    Then eventually resource "redis" does not exist

    And resource "secret" does not exist


  @skr @azure @managedredis @rediscluster
  Scenario: AzureManagedRedis C-tier scenario (HA, OSSCluster sharded)

    Given there is shared SKR with "Azure" provider

    And resource declaration:
      | Alias  | Kind              | ApiVersion                              | Name                         | Namespace |
      | redis  | AzureManagedRedis | cloud-resources.kyma-project.io/v1beta1 | e2e-${id()}                  |           |
      | secret | Secret            | v1                                      | ${redis.metadata.name ?? ''} |           |

    When resource "redis" is created:
      """
      apiVersion: cloud-resources.kyma-project.io/v1beta1
      kind: AzureManagedRedis
      spec:
        tier: C5
      """

    Then eventually "redis.status.state == 'Ready'" is ok, unless:
      | redis.status.state == 'Error' |
      | #timeout=30m                  |

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
