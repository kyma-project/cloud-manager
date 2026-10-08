# SKR Looper — dev-time fairness simulation

`fairness_sim_test.go` is a **manual, dev-time** simulation of the two-sleeve SKR looper. It is
guarded by the `looper_sim` build tag, so `make test`, `go build`, and CI never compile or run it.

## What it does

It runs the **real shipped fairness code** — `Queue`, `SkrGate`, `processOne`, `cyclicReAdd`,
`Notify`, `recordNotifConnect`, and the real `add()` path — with a dummy sleeping handler instead of a
real per-SKR manager (no IO, no real 10s connect). It seeds a large fleet, designates a few "hot" SKRs
that fire notifications constantly, drives a skewed notification stream, periodically re-activates the
whole fleet via `AddKyma` (mirroring the KCP Kyma reconciler's periodic resync), and measures per-SKR
reconnect gaps.

Time is **scaled**: fairness depends on ratios (hot-vs-cold notification rate, connect time vs
`notifMinInterval`, fleet/workers), not absolute magnitudes. A ~30s wall-clock run at the scaled
values reproduces many hours of production rotation.

Use it to validate any change to the looper's queueing/fairness logic. It reproduces **two** distinct
production symptoms and asserts both are gone:

- **Notification starvation** — a fixed set of SKRs starved with >20-min gaps while hot SKRs stay
  fresh (fixed in #2083).
- **AddKyma strand** — the periodic `AddKyma` re-activation racing the cyclic connect lifecycle,
  stranding a persistent core of live members that connect once and never cycle again. This ONLY
  reproduces with the `-sim.addKymaEvery` driver active: against the unfixed `add()` the sim reports
  FAIRNESS FAIL with a max gap in the tens of seconds (100×+ mean); against the fixed `add()` it stays
  tight (~1× mean).

## Run

```
go test -tags looper_sim -run TestLooperFairnessSim -v -timeout 30m \
    ./pkg/skr/runtime/looper/ \
    -args -sim.fleet=730 -sim.cyclic=24 -sim.notif=8 -sim.hot=3 \
          -sim.connect=10ms -sim.duration=30s -sim.seed=1 -sim.addKymaEvery=30ms
```

`-v` is required to see the report (Go hides `t.Log` output on success). No `FEATURE_FLAG_CONFIG_FILE`
is needed — that is only for `internal/controller` tests.

## Flags (after `-args`)

| Flag | Default | Meaning |
|---|---|---|
| `-sim.fleet` | 730 | active SKR count |
| `-sim.cyclic` | 24 | cyclic worker count |
| `-sim.notif` | 8 | notification worker count |
| `-sim.hot` | 3 | number of hyper-active SKRs |
| `-sim.connect` | 10ms | scaled per-connect duration (stands in for ~10s) |
| `-sim.notifMin` | 10ms | per-SKR notification rate-limit interval (scaled) |
| `-sim.hotEvery` | 2ms | how often each hot SKR fires a notification |
| `-sim.addKymaEvery` | 30ms | period between full-fleet `AddKyma` re-activation sweeps (0 disables); models the KCP Kyma reconciler's periodic resync — the driver that reproduces the strand |
| `-sim.duration` | 30s | total wall-clock run |
| `-sim.seed` | 1 | RNG seed (fixed for reproducibility; vary to explore) |

## Reading the report

- **theoretical fair mean gap** = `fleet × connect / cyclic` — the gap every SKR should see under a
  perfectly fair rotation.
- **observed gap mean/p50/p95/p99/max** — should cluster near the fair mean with a tight spread.
- **worst 10 SKRs by max gap (× mean)** — the long tail. Healthy: all near 1×. Broken (starvation):
  a fixed set at 5–20× mean.
- **hot SKR connect counts** — hot SKRs should get many bonus connects yet NOT appear in the worst
  list (bonus without starving others).
- **FAIRNESS OK/FAIL** — the test fails if any SKR's max gap exceeds `maxGapFactor` (3×) the mean. The
  factor is generous on purpose: starvation is 5–20× mean, scheduling jitter is well under 2×.
