# Context keys and query language

## Full context key list

Built by `ContextBuilder` (`pkg/feature/context.go`) and read at evaluation time. Not every key is set
for every evaluation — SKR-scoped keys are empty for KCP-only features.

| Key | Values / format | Notes |
|---|---|---|
| `landscape` | `dev`, `stage`, `prod` | Always set. Also selects the config file (see SKILL.md Step 1). |
| `feature` | e.g. `nfs`, `nfsBackup`, `peering`, `redis`, `rediscluster`, `azureManagedRedis`, `vpcdnslink` | **Derived from the resource kind**, not free text. See features-and-apidisabled.md. Empty when the object isn't specific to one feature. |
| `plane` | `kcp`, `skr` | Which reconciliation loop. |
| `provider` | `aws`, `gcp`, `azure`, `openstack`, `alicloud` | From the Shoot. `openstack` is CCEE. |
| `brokerPlan` | KEB plan name, e.g. `trial` | Used to disable everything on trial. |
| `globalAccount` | BTP Global Account UUID | Per-client targeting. Phoenix team GA (masked) = `00000000-0000-0000-0000-000000000000`. |
| `subAccount` | BTP Subaccount UUID | Per-client targeting. |
| `kyma` | KCP Kyma name | SKR only. |
| `shoot` | Gardener shoot name, e.g. `c-000000` | Finest-grained per-client targeting. |
| `region` | SKR region | |
| `objKindGroup` | `lower(kind).lower(group)` | The object being reconciled. |
| `crdKindGroup` | `lower(kind).lower(group)` | Set when a CRD is handled. |
| `busolaKindGroup` | `lower(kind).lower(group)` | Busola extension ConfigMap. |
| `allKindGroups` | `objKindGroup,crdKindGroup,busolaKindGroup` | Comma-joined; use with `co`/contains. |

SKR-scoped keys (`globalAccount`, `subAccount`, `kyma`, `shoot`) are **not defined** for non-SKR (KCP)
features. A query on `globalAccount` simply won't match when the key is absent.

## General feature names

**Authoritative source: the `FeatureName` constants in `pkg/feature/types/types.go`** — read that file for
the current, complete set. As of writing:

| Constant | Value |
|---|---|
| `FeatureNfs` | `nfs` |
| `FeatureNfsBackup` | `nfsBackup` |
| `FeaturePeering` | `peering` |
| `FeatureRedis` | `redis` |
| `FeatureRedisCluster` | `rediscluster` |
| `FeatureAzureManagedRedis` | `azureManagedRedis` |
| `FeatureVpcDnsLink` | `vpcdnslink` |

The feature table in `pkg/feature/README.md` is a prose summary and is **incomplete** — it lists only
`nfs`, `nfsBackup`, `peering`, `redis` and omits the rest. Treat it as documentation, not the source of
truth; when the exact set matters, grep `pkg/feature/types/types.go` for `FeatureName`.

## go-feature-flag query operators

The `query` field is a go-feature-flag rule expression. Operators seen and supported:

| Operator | Meaning | Example |
|---|---|---|
| `==` | equals | `provider == "aws"` |
| `!=` | not equals | `provider != "azure"` |
| `in` | membership in list | `shoot in ["c-000000", "c-000001"]` |
| `and` | logical and | `feature == "nfsBackup" and provider != "azure"` |
| `or` | logical or | `provider == "aws" or provider == "gcp"` |
| `co` / `contains` | substring/element contains | `allKindGroups co "cloudresources.cloud-resources.kyma-project.io"` |
| `sw` / `ew` | starts-with / ends-with | |

Strings are double-quoted. Lists are `["a", "b"]`. Evaluation is against the string context values.

### Gotchas
- **A missing context key does not match** an `==` comparison. If the scenario doesn't set `shoot`, a rule
  keyed on `shoot in [...]` is skipped.
- Values are strings — `region == "eu"` matches only exact `eu`, not a prefix, unless `sw` is used.
- `and`/`or` precedence: parenthesize when mixing; the SAP configs mostly chain `and`.