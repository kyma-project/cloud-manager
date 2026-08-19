
# AliCloud Redis — Multi-PR Deployment Plan

**Status:** PR2 open as #2102 (feat/alicloud-redis-api → main). PR3 and PR4 rebased and pushed, awaiting PR2 merge.
**Rounds completed:** 15 deep review rounds across all three PRs. All branches are CI-green.

---

## PR sequence (current)

| # | Branch | GitHub PR | Tip commit | Status |
|---|--------|-----------|------------|--------|
| 2 | `feat/alicloud-redis-api` | #2102 | `897d48d0` | Open, awaiting review |
| 3 | `feat/alicloud-redis-reconcilers` | not yet opened | `ecb3dcd9` | Rebased onto PR2 tip, pushed |
| 4 | `feat/alicloud-redis-e2e` | not yet opened | `6e9901cf` | Rebased onto PR3 tip, pushed |

> PR 1 (shared runtime fixes) was folded into the main feature work. It is not a separate PR.

---

## Branch topology (current, post-rebase)

Branches are now properly stacked:

```
main
└── feat/alicloud-redis-api       (PR2, 9 commits above main)
      └── feat/alicloud-redis-reconcilers  (PR3, 7 commits above PR2)
            └── feat/alicloud-redis-e2e    (PR4, 5 commits above PR3)
```

When PR2 merges, rebase PR3 onto new `main` (no conflicts expected — PR3 already sits on PR2 tip).
When PR3 merges, rebase PR4 onto new `main` similarly.

---

## What each PR contains

### PR2 — `feat/alicloud-redis-api` (PR #2102)
Purely additive: new types, CRDs, feature flags, and the iprange provider.

- `api/cloud-control/v1beta1/redis_alicloud.go` — KCP type definitions for both resources
- `api/cloud-resources/v1beta1/alicloudredisinstance_*.go` — SKR type, builder, tiers
- `api/cloud-resources/v1beta1/alicloudrediscluster_*.go` — SKR type, builder, tiers
- `pkg/feature/ffAlicloud.go`, `types/types.go` — feature flag Go defs
- `pkg/feature/ff_ga.yaml`, `ff_edge.yaml` — GA disabled via `apiDisabled` provider-scoped rules; edge enabled
- `pkg/kcp/provider/alicloud/iprange/` — AliCloud iprange provider (vSwitch lifecycle)
- `config/crd/bases/`, `config/dist/`, `config/samples/` — generated CRDs and samples
- `config/sync.sh`, `config/patchAfterMakeManifests.sh`, `PROJECT` — tooling updates
- `internal/api-tests/skr_alicloudredisinstance_test.go` — API validation tests
- `internal/api-tests/skr_alicloudrediscluster_test.go` — API validation tests
- `pkg/skr/runtime/looper/installer_test.go` — alicloud CRD installer assertions
- `go.mod`, `go.sum` — adds `alibabacloud-go/r-kvstore` and related deps

**Key design decisions in PR2:**
- `AlicloudRedisInstance.SpecificToFeature()` returns `"redis"` (not a new key) — gated by `feature == "redis" and provider == "alicloud"` rule in `apiDisabled`
- `AlicloudRedisCluster.SpecificToFeature()` returns `"rediscluster"` — gated by existing `Disable RedisCluster` rule plus new `provider == "alicloud"` variant
- No `ffAlicloudRedis.go` — standalone flag blocks removed; AliCloud Redis is disabled via `apiDisabled` targeting rules only, consistent with all other disabled-by-default features

### PR3 — `feat/alicloud-redis-reconcilers`
All reconciler logic. Depends on PR2 being merged first.

- `pkg/kcp/provider/alicloud/util.go` — shared utilities (CA cert, password gen, CIDR helpers)
- `pkg/kcp/provider/alicloud/redisinstance/` — KCP instance reconciler (14 action files + client)
- `pkg/kcp/provider/alicloud/rediscluster/` — KCP cluster reconciler (15 action files + client)
- `pkg/kcp/provider/alicloud/mock/` — mock server for controller tests (5 files)
- `pkg/kcp/redisinstance/reconciler.go` — wires alicloud into KCP RedisInstance dispatcher
- `pkg/kcp/rediscluster/reconciler.go` — wires alicloud into KCP RedisCluster dispatcher
- `pkg/skr/alicloudredisinstance/` — SKR instance reconciler (20 files)
- `pkg/skr/alicloudrediscluster/` — SKR cluster reconciler (20 files)
- `pkg/skr/iprange/preventDeleteOn*.go` — IpRange deletion guards for both resources
- `internal/controller/cloud-control/redisinstance_alicloud_test.go` — KCP controller tests
- `internal/controller/cloud-control/rediscluster_alicloud_test.go`
- `internal/controller/cloud-resources/alicloudredisinstance_controller.go` + test
- `internal/controller/cloud-resources/alicloudrediscluster_controller.go` + test
- `cmd/main.go` — wires up all four controllers
- `pkg/testinfra/dsl/alicloudRedisInstance.go` — DSL helpers for tests

### PR4 — `feat/alicloud-redis-e2e`
e2e tests, CI workflows, IAM tooling. Depends on PR3 being merged first.

- `e2e/features/skr-shared-redis-alicloud.feature`
- `e2e/features/skr-shared-rediscluster-alicloud.feature`
- `e2e/lib/runtimeBuilder.go`, `e2e/lib/consts.go`, `e2e/lib/initializeKcp.go`
- `e2e/scripts/alicloud_infra.sh`, `alicloud_key.sh`, `_common-alicloud.sh`
- `e2e/cloud/credentials.go`, `e2e/config/config.go`
- `e2e/sim/shootBuilder.go` — AliCloud shoot simulation
- `.github/actions/e2e-init/action.yaml`, `.github/workflows/e2e-tests.yaml`
- `docs/contributor/permissions/alicloud/policy-CloudManagerAccess.json` — IAM policy

**IAM policy:** `r-kvstore:*` (13 actions) + `vpc:*` (10 actions, exact subset) + `nas:*` (11 actions).
No `kvstore:*` (legacy API never called), no `ecs:*`, `slb:*`, `ram:*`, `ros:*`.

---

## Future carving plan: PR3 → 4 sub-PRs

When ready for GA, PR3 (`feat/alicloud-redis-reconcilers`) can be carved into 4 independent
reviewable PRs. Each must compile, pass `make build`, and have green tests on its own before
the next is opened.

### Merge order

```
PR2 (API/CRDs — merged)
  └── PR-A: KCP RedisInstance
        ├── PR-B: SKR AlicloudRedisInstance   (depends on PR-A)
        └── PR-C: KCP RedisCluster            (depends on PR-A — shares mock)
              └── PR-D: SKR AlicloudRedisCluster  (depends on PR-B + PR-C)
```

PR-B and PR-C can be reviewed in parallel once PR-A merges. PR-D needs both.

---

### PR-A: AliCloud Redis Instance — KCP

**What it adds:** The KCP reconciler that calls r-kvstore to provision a standard HA Redis instance.
Includes the shared mock (instance tests need it; cluster extends it in PR-C).

**Branch strategy:** Cut from `main` after PR2 merges. Contains only the files below.
`cmd/main.go` registers only the KCP RedisInstance controller — SKR and cluster registrations are not present yet.

**Files:**
```
pkg/kcp/provider/alicloud/util.go                           shared CA cert / password / CIDR utils
pkg/kcp/provider/alicloud/redisinstance/client/client.go    r-kvstore SDK wrapper
pkg/kcp/provider/alicloud/redisinstance/state.go
pkg/kcp/provider/alicloud/redisinstance/new.go
pkg/kcp/provider/alicloud/redisinstance/createRedis.go
pkg/kcp/provider/alicloud/redisinstance/loadRedis.go
pkg/kcp/provider/alicloud/redisinstance/deleteRedis.go
pkg/kcp/provider/alicloud/redisinstance/waitRedisAvailable.go
pkg/kcp/provider/alicloud/redisinstance/waitRedisDeleted.go
pkg/kcp/provider/alicloud/redisinstance/updateStatus.go
pkg/kcp/provider/alicloud/redisinstance/addUpdatingCondition.go
pkg/kcp/provider/alicloud/redisinstance/removeReadyCondition.go
pkg/kcp/provider/alicloud/redisinstance/enableSsl.go
pkg/kcp/provider/alicloud/redisinstance/setSecurityIps.go
pkg/kcp/provider/alicloud/redisinstance/modifyInstanceClass.go
pkg/kcp/provider/alicloud/redisinstance/modifyParameters.go
pkg/kcp/redisinstance/reconciler.go                         wire alicloud into KCP dispatcher
pkg/kcp/provider/alicloud/mock/type.go                      mock types (shared, created here)
pkg/kcp/provider/alicloud/mock/accountRegionStore.go
pkg/kcp/provider/alicloud/mock/redisStore.go                instance methods only
pkg/kcp/provider/alicloud/mock/redisClientViews.go          instance client view
pkg/kcp/provider/alicloud/mock/server.go                    mock server wiring
internal/controller/cloud-control/redisinstance_controller.go
internal/controller/cloud-control/redisinstance_alicloud_test.go
internal/controller/cloud-control/suite_test.go             instance mock registration only
cmd/main.go                                                 KCP RedisInstance provider registration
go.mod / go.sum                                             r-kvstore SDK dependency
```

**Compile gate:** `make build` must pass with only KCP RedisInstance wired. The cluster
reconciler package does not exist yet — `pkg/kcp/rediscluster/reconciler.go` must not reference
alicloud until PR-C.

**Test gate:**
```bash
make build && make test-ff
go test ./pkg/kcp/provider/alicloud/redisinstance/...
go test ./internal/controller/cloud-control/ -run TestControllers --ginkgo.focus="AliCloud.*RedisInstance"
```

~2,800 lines production · ~600 lines tests

---

### PR-B: AliCloud Redis Instance — SKR

**What it adds:** The SKR reconciler managing `AlicloudRedisInstance` objects on the user cluster.
Creates KCP `RedisInstance`, waits for it to become Ready, syncs connection details back as an auth Secret.

**Branch strategy:** Cut from `main` after PR-A merges.

**Files:**
```
pkg/skr/alicloudredisinstance/state.go
pkg/skr/alicloudredisinstance/reconciler.go
pkg/skr/alicloudredisinstance/ignorant.go
pkg/skr/alicloudredisinstance/createKcpRedisInstance.go
pkg/skr/alicloudredisinstance/modifyKcpRedisInstance.go
pkg/skr/alicloudredisinstance/loadKcpRedisInstance.go
pkg/skr/alicloudredisinstance/deleteKcpRedisInstance.go
pkg/skr/alicloudredisinstance/waitKcpRedisInstanceDeleted.go
pkg/skr/alicloudredisinstance/waitKcpStatusUpdate.go
pkg/skr/alicloudredisinstance/createAuthSecret.go
pkg/skr/alicloudredisinstance/modifyAuthSecret.go
pkg/skr/alicloudredisinstance/loadAuthSecret.go
pkg/skr/alicloudredisinstance/deleteAuthSecret.go
pkg/skr/alicloudredisinstance/removeAuthSecretFinalizer.go
pkg/skr/alicloudredisinstance/waitAuthSecretDeleted.go
pkg/skr/alicloudredisinstance/updateId.go
pkg/skr/alicloudredisinstance/updateStatus.go
pkg/skr/alicloudredisinstance/waitSkrStatusReady.go
pkg/skr/alicloudredisinstance/util.go                       tier→instanceClass mapping
pkg/skr/alicloudredisinstance/util_test.go
pkg/skr/iprange/preventDeleteOnAlicloudRedisInstanceUsage.go
pkg/skr/iprange/reconciler.go                               add instance deletion guard
internal/controller/cloud-resources/alicloudredisinstance_controller.go
internal/controller/cloud-resources/alicloudredisinstance_test.go
internal/controller/cloud-resources/suite_test.go           instance controller registration
internal/api-tests/skr_alicloudredisinstance_test.go
pkg/skr/runtime/looper/installer_test.go                    alicloudredisinstance CRD entry
pkg/testinfra/dsl/alicloudRedisInstance.go
pkg/testinfra/dsl/commonIpRange.go
pkg/testinfra/dsl/kcpIpRange.go
pkg/testinfra/run.go                                        instance DSL registration
cmd/main.go                                                 SKR AlicloudRedisInstance controller
```

**Test gate:**
```bash
make build && make test-ff
go test ./pkg/skr/alicloudredisinstance/...
go test ./internal/controller/cloud-resources/ -run TestControllers --ginkgo.focus="AlicloudRedisInstance"
go test ./internal/api-tests/ -run TestControllers --ginkgo.focus="AlicloudRedisInstance"
go test ./pkg/skr/runtime/looper/ -run TestInstaller
```

~1,100 lines production · ~725 lines tests

---

### PR-C: AliCloud Redis Cluster — KCP

**What it adds:** The KCP reconciler for sharded cluster mode. The cluster client embeds the
instance client (reuses all single-instance operations) and adds `AddShardingNode` /
`DeleteShardingNode`. The mock `redisStore.go` is extended with cluster methods.

**Branch strategy:** Cut from `main` after PR-A merges (parallel with PR-B).
`cmd/main.go` adds only KCP RedisCluster registration.

**Files:**
```
pkg/kcp/provider/alicloud/rediscluster/client/client.go     embeds instance client + adds sharding ops
pkg/kcp/provider/alicloud/rediscluster/state.go
pkg/kcp/provider/alicloud/rediscluster/new.go
pkg/kcp/provider/alicloud/rediscluster/createRedis.go
pkg/kcp/provider/alicloud/rediscluster/loadRedis.go
pkg/kcp/provider/alicloud/rediscluster/deleteRedis.go
pkg/kcp/provider/alicloud/rediscluster/waitRedisAvailable.go
pkg/kcp/provider/alicloud/rediscluster/waitRedisDeleted.go
pkg/kcp/provider/alicloud/rediscluster/updateStatus.go
pkg/kcp/provider/alicloud/rediscluster/addUpdatingCondition.go
pkg/kcp/provider/alicloud/rediscluster/removeReadyCondition.go
pkg/kcp/provider/alicloud/rediscluster/enableSsl.go
pkg/kcp/provider/alicloud/rediscluster/setSecurityIps.go
pkg/kcp/provider/alicloud/rediscluster/modifyInstanceClass.go
pkg/kcp/provider/alicloud/rediscluster/modifyParameters.go
pkg/kcp/provider/alicloud/rediscluster/modifyShardCount.go  cluster-only: shard scale up/down
pkg/kcp/rediscluster/reconciler.go                          wire alicloud into KCP dispatcher
pkg/kcp/provider/alicloud/mock/redisStore.go                amend: add cluster store methods
pkg/kcp/provider/alicloud/mock/redisClientViews.go          amend: add cluster client view
internal/controller/cloud-control/rediscluster_controller.go
internal/controller/cloud-control/rediscluster_alicloud_test.go
internal/controller/cloud-control/suite_test.go             cluster mock registration
cmd/main.go                                                 KCP RedisCluster provider registration
```

**Note on mock amendment:** `redisStore.go` already exists from PR-A. PR-C adds cluster-specific
methods (`createCluster`, `describeCluster`, `addShardingNode`, etc.) to the same file. The PR
diff will show only the additions.

**Test gate:**
```bash
make build && make test-ff
go test ./pkg/kcp/provider/alicloud/rediscluster/...
go test ./internal/controller/cloud-control/ -run TestControllers --ginkgo.focus="AliCloud.*RedisCluster"
```

~1,100 lines production · ~828 lines tests

---

### PR-D: AliCloud Redis Cluster — SKR

**What it adds:** The SKR reconciler managing `AlicloudRedisCluster` objects. Mirrors PR-B
for cluster mode — creates KCP `RedisCluster`, syncs auth Secret back.

**Branch strategy:** Cut from `main` after both PR-B and PR-C merge.

**Files:**
```
pkg/skr/alicloudrediscluster/state.go
pkg/skr/alicloudrediscluster/reconciler.go
pkg/skr/alicloudrediscluster/ignorant.go
pkg/skr/alicloudrediscluster/createKcpRedisCluster.go
pkg/skr/alicloudrediscluster/modifyKcpRedisCluster.go
pkg/skr/alicloudrediscluster/loadKcpRedisCluster.go
pkg/skr/alicloudrediscluster/deleteKcpRedisCluster.go
pkg/skr/alicloudrediscluster/waitKcpRedisClusterDeleted.go
pkg/skr/alicloudrediscluster/waitKcpStatusUpdate.go
pkg/skr/alicloudrediscluster/createAuthSecret.go
pkg/skr/alicloudrediscluster/modifyAuthSecret.go
pkg/skr/alicloudrediscluster/loadAuthSecret.go
pkg/skr/alicloudrediscluster/deleteAuthSecret.go
pkg/skr/alicloudrediscluster/removeAuthSecretFinalizer.go
pkg/skr/alicloudrediscluster/waitAuthSecretDeleted.go
pkg/skr/alicloudrediscluster/updateId.go
pkg/skr/alicloudrediscluster/updateStatus.go
pkg/skr/alicloudrediscluster/waitSkrStatusReady.go
pkg/skr/alicloudrediscluster/util.go                        tier→instanceClass mapping
pkg/skr/alicloudrediscluster/util_test.go
pkg/skr/iprange/preventDeleteOnAlicloudRedisClusterUsage.go
pkg/skr/iprange/reconciler.go                               add cluster deletion guard
internal/controller/cloud-resources/alicloudrediscluster_controller.go
internal/controller/cloud-resources/alicloudrediscluster_test.go
internal/controller/cloud-resources/suite_test.go           cluster controller registration
internal/api-tests/skr_alicloudrediscluster_test.go
pkg/skr/runtime/looper/installer_test.go                    alicloudrediscluster CRD entry
pkg/testinfra/run.go                                        cluster DSL registration
cmd/main.go                                                 SKR AlicloudRedisCluster controller
```

**Test gate:**
```bash
make build && make test-ff
go test ./pkg/skr/alicloudrediscluster/...
go test ./internal/controller/cloud-resources/ -run TestControllers --ginkgo.focus="AlicloudRedisCluster"
go test ./internal/api-tests/ -run TestControllers --ginkgo.focus="AlicloudRedisCluster"
go test ./pkg/skr/runtime/looper/ -run TestInstaller
```

~1,117 lines production · ~894 lines tests

---

### Shared file split strategy

| File | PR-A | PR-B | PR-C | PR-D |
|------|------|------|------|------|
| `pkg/kcp/provider/alicloud/util.go` | Create | — | — | — |
| `mock/` (5 files) | Create (instance only) | — | Amend redisStore + clientViews | — |
| `suite_test.go` KCP | Instance mock wiring | — | Cluster mock wiring | — |
| `suite_test.go` SKR | — | Instance registration | — | Cluster registration |
| `pkg/skr/iprange/reconciler.go` | — | Instance guard | — | Cluster guard |
| `installer_test.go` | — | Instance CRD entry | — | Cluster CRD entry |
| `cmd/main.go` | KCP instance | SKR instance | KCP cluster | SKR cluster |
| `pkg/testinfra/run.go` | — | Instance DSL | — | Cluster DSL |

---

## Rebase procedure (run after each PR merges to upstream main)

```bash
# After PR2 merges:
git fetch upstream
git checkout feat/alicloud-redis-reconcilers
git rebase upstream/main
# No conflicts expected — PR3 already sits on PR2 tip
make manifests && ./config/patchAfterMakeManifests.sh && ./config/sync.sh
make build && make test-ff
git push origin feat/alicloud-redis-reconcilers --force-with-lease

# After PR3 merges:
git checkout feat/alicloud-redis-e2e
git rebase upstream/main
make build && make test-ff
git push origin feat/alicloud-redis-e2e --force-with-lease
```

---

## Verification gate per PR

```bash
make build
make manifests && ./config/patchAfterMakeManifests.sh && ./config/sync.sh  # PR2, PR3
git diff --exit-code config/                                                 # generated files match
make test-ff
go test ./pkg/skr/runtime/looper/ -run TestInstaller
go test ./pkg/feature/... -v                                                 # feature flag unit tests
PROJECTROOT=$(pwd) KUBEBUILDER_ASSETS=$(go run ./hack/getEnvtestBinaryAssets) \
  go test ./internal/controller/cloud-control/ -run TestControllers --ginkgo.focus="AliCloud Redis"
PROJECTROOT=$(pwd) KUBEBUILDER_ASSETS=$(go run ./hack/getEnvtestBinaryAssets) \
  go test ./internal/controller/cloud-resources/ -run TestControllers --ginkgo.focus="AlicloudRedis"
bash tools/e2e/e2e-run-locally-alicloud.sh                                   # PR4 — full e2e
```

CI must be green on all PRs before merging in sequence.

---

## e2e test results (2026-08-18)

All 3 scenarios passed on `feat/alicloud-redis-e2e`:

| Scenario | Duration | Result |
|----------|----------|--------|
| VpcNetwork AliCloud is created and deleted | 8s | ✅ PASS |
| AlicloudRedisInstance scenario | 795s (~13min) | ✅ PASS |
| AlicloudRedisCluster scenario | 793s (~13min) | ✅ PASS |

35/35 steps green. Region: `ap-northeast-1` (Tokyo), shoot `sk-c69253`.

---

## Production-readiness: **B− (edge-ready, not GA)**

### Strengths
- Shipped disabled in GA via `apiDisabled` provider-scoped rules; zero production exposure until opt-in
- 15 rounds of deep review across all three PRs
- Full test coverage: KCP controller, SKR controller, API validation, e2e features
- Immutability via CEL: engineVersion, ipRange, authSecret.name, replicasPerShard==0
- e2e tests passing end-to-end on real AliCloud infrastructure

### GA gates (none block edge)
| # | Gate | Action |
|---|------|--------|
| 1 | No production traffic — only mock-validated + single e2e run | Real-cloud e2e soak on edge runtime |
| 2 | Busola UI forms not designed | Design and add for both resources |
| 3 | CA cert fetched at runtime over HTTP | Caching / fallback strategy |
| 4 | Observability not wired | Provisioning-latency + error-rate dashboards |

---

## Notes

- Default e2e region: `ap-northeast-1` (Tokyo, zones a/b/c). `eu-central-1` (Frankfurt) also configured.
- `ReplicasPerShard` is documented as "always 0 for current tiers" — field reserved for future non-proxy tier support. CEL rule `self == 0` enforces this invariant.
- Feature flag approach: `AlicloudRedisInstance` returns `SpecificToFeature() = "redis"`, `AlicloudRedisCluster` returns `"rediscluster"`. Gated by `feature == "redis" and provider == "alicloud"` and existing `Disable RedisCluster` + provider variant in `apiDisabled`. No standalone flag blocks needed.
- IAM policy (`policy-CloudManagerAccess.json`): minimal, verified against actual SDK calls. `kvstore:*` (legacy API) excluded — only `r-kvstore:*` is used.
