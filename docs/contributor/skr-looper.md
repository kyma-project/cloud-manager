# SKR Looper — Contributor & Agent Guide

Audience: an engineer or AI agent picking up the SKR looper (`pkg/skr/runtime/looper/`) cold. This
doc captures the **non-obvious insights** — the things that took real investigation to understand. It
deliberately does **not** re-explain what you can read directly in the code or in client-go's
`workqueue` (Add/Get/Done, dirty/processing sets, slice FIFO). Read those; this fills the gaps.

## What the looper is

One active looper (leader-elected; a single replica reconciles) repeatedly "connects" to each active
SKR: builds a per-SKR manager, verifies CRDs, and runs the SKR reconcilers for a bounded window. A
"connect" is one full pass of `handleOneSkr`.

## Two sleeves, and why

There are **two independent worker pools ("sleeves")** draining **two separate queues**:

| Sleeve | Queue | Trigger | Role |
|---|---|---|---|
| **Cyclic** | `cyclicQueue` | background round-robin | the fairness-critical base cadence — every active SKR is revisited on rotation |
| **Notification** | `notifQueue` | runtime-watcher events (`Notify`) | fast, event-driven **bonus** connects for SKRs that just changed |

Key mental model: **the cyclic sleeve owns fairness; the notification sleeve is an optional accelerator.**
`cyclicQueue` membership is the source of truth for "is this SKR active" (`Contains`). The notification
sleeve must never be *required* for correctness and must never degrade cyclic fairness.

### The historical coupling that caused the bug

The two sleeves were **not** as independent as the table suggests. On a successful notification
connect, the notification worker used to call `CyclicQueue().Delay(kymaName)` — a **cross-sleeve write
into the cyclic queue**. That single line coupled them and was the root cause of the starvation
described below. It has been removed; keep them decoupled.

## The single-manager invariant — three states, don't conflate them

At most one live manager per SKR may exist **across both sleeves** at once. This is enforced by
`SkrGate` (a claim taken for the connect's lifetime), NOT by the queues. An SKR is described by three
**distinct** states that are easy to confuse:

- **membership** (`Queue.Contains`) — "is this SKR active / should it be in rotation." Set by the KCP
  reconciler via `AddKyma`; cleared by `Remove`.
- **gate claim** (`SkrGate`) — "a worker is connecting to it right now." Cross-sleeve, held for the
  whole connect.
- **workqueue processing** (client-go `processing` set) — "a worker has Get'd this item from *this*
  queue." Per-queue.

Removing membership does **not** abort a running manager or free the gate (graceful teardown — the
owning worker's `Release` does that). A gate conflict (`TryClaim` fails) means the *other* sleeve is
mid-connect; it is not a membership or queue problem.

## The starvation bug — the reasoning chain (this is the core value of this doc)

Symptom: individual SKRs showed large reconnect gaps (mode ~10 min, tail well past 20 min) in
`cloud_manager_skr_runtime_reconcile_total`, even after earlier cadence fixes.

The investigation ruled out several plausible-but-wrong hypotheses. Follow the chain so you can
re-derive it or avoid re-treading it:

1. **"A worker hangs/lags in the pre-amble" — WRONG.** A `workerTimeout` (10 min) was added to release
   a stuck worker. The connect-time histograms show it **never fires**:
   `cloud_manager_skr_looper_connect_total_seconds{timeout="true"}` is empty and the "SKR worker
   timeout exceeded" log is silent. Every connect completes in ~10–20s (100% of the `total_seconds`
   histogram lands in the 10–20s bucket; `+Inf == le=20`, i.e. none exceed 20s). No connect is slow.

2. **Beware the `timeout` label — two different contexts.** `connect_phase_seconds{phase="start"}`
   shows `timeout="true"` on *every* connect. That is the **intended** inner `reconcileTimeout` (~10s)
   bounding `skrManager.Start`, which blocks until its context is done — it is recorded from the inner
   `timeoutCtx`, a *different* context than the outer `workerCtx` used by `total_seconds`. So
   "start timed out" is normal and healthy; "total timed out" would be the real alarm and never
   happens.

3. **"The 10-min mode is the worker timeout cutting connects" — WRONG (red herring).** The 10-min mode
   coincided with `workerTimeout=10m`, which is misleading. Gaps **larger than 10 min are common**
   (`increase[20m]==0` returns many SKRs), which a hard 10-min slicer could never produce. The timeout
   is not involved.

4. **"It's just a saturated-queue tail" — WRONG (it's worse: a *fixed* starved set).** The pool is
   saturated (`gate_in_flight` ≈ cyclic + a few notification workers; connect beat ≈ fleet/mean).
   Fleet is ~730 SKRs → fair mean revisit ≈ 6.6 min, which is acceptable. But a CSV of the
   `increase[20m]==0` series over hours showed the **same** SKRs starved repeatedly (a fixed ~50+ set
   idle >20 min for ~30% of the window), while the hyper-active "hot" SKRs were **never** starved. A
   *fair* saturated FIFO cannot do that — the tail would rotate, not stick.

5. **Root cause: outside writes reshuffle the cyclic tail.** The cyclic sleeve alone is provably fair
   (`TestCyclicFairDistribution`, spread = 0). The notification success path wrote into the cyclic
   queue (`CyclicQueue().Delay`). In client-go's workqueue, `Add` on an item that is **waiting in the
   queue** (`dirty && !processing`) calls the queue's `Touch`, which our `slidingQueue` implements as
   **move-to-tail**. Hot SKRs fire notifications constantly → their cyclic entry is repeatedly yanked
   to the tail, **leapfrogging** whatever SKRs had drifted toward the tail, which then never reach the
   head. Result: a fixed starved set. (The gate-conflict re-add via `AddAfter` was a second, smaller
   source of the same reshuffling.)

Crucially, `Touch` fires **only** for an item that is *queued and not being processed*. An item a
worker has already `Get`'d is in `processing`; re-adding it there just marks it dirty and `Done`
re-pushes it to the **tail with no Touch**. This asymmetry is why the fix (below) works.

## Why throughput is NOT the lever

It is tempting to "speed up" connects to shrink gaps. Don't:

- The ~10s `skrManager.Start` window is **required reconcile work**, not idle slack — the worker is a
  placeholder keeping the manager and its reconcilers alive to do their job. Cutting it cuts the work.
- More workers is bounded by KCP/SKR apiserver load (each worker = one live manager hitting both
  apiservers). It cannot be raised much.

So the fair mean (~6.6 min at current scale) is a capacity floor, and it is acceptable. **Fairness —
not throughput — is the lever.** A fair rotation gives every SKR ~the mean with a tight spread; the
bug was the tail, not the mean.

## The fix and its invariants

Hold these invariants when touching the looper:

1. **The notification sleeve never writes to the cyclic queue.** On success it only stamps the
   notification-connect time for the rate limiter; it does not re-add/`Delay`/`Touch` cyclic. The SKR
   keeps its cyclic position and is served on its normal cyclic turn; the notification was a bonus.
2. **Gate-conflict handling is sleeve-specific:**
   - Cyclic worker hits a conflict → re-add via **plain FIFO tail** (`q.Add`, not `AddAfter`). At that
     point the item is `processing`, so `Done` re-pushes to the tail with no `Touch` — order preserved.
   - Notification worker hits a conflict → **drop**. The in-flight connect already satisfies the intent.
3. **Per-SKR notification rate limit** (`notifMinInterval`, default 10s, env
   `SKR_RUNTIME_NOTIF_MIN_INTERVAL`): a notification arriving within the interval of that SKR's last
   notification connect is coalesced (dropped, counted in
   `cloud_manager_skr_looper_notification_rate_limited_total`). Bounds how much a hot SKR occupies the
   notification sleeve. It never touches the cyclic sleeve.
4. **`workerTimeout` stays** as a safety net (it correctly never fires today). Don't tune it to chase
   gaps — that was the wrong lever.

## Diagnostic playbook (re-derive it fast)

- **Is any SKR starved?** `count by (kyma)(increase(cloud_manager_skr_runtime_reconcile_total[20m]) == 0)`
  — returns a series per SKR idle > 20 min. Should trend to zero after the fix. Probe other windows
  (`[9m]`, `[11m]`, `[15m]`) to sketch the gap CDF.
- **Fixed set vs rotating tail?** Export the above as a time series and check whether the same `kyma`
  labels recur (starvation) or differ each sample (normal tail). This distinction is what proved it
  was a bug, not saturation.
- **Are connects actually fast?** `sum by (le)(rate(cloud_manager_skr_looper_connect_total_seconds_bucket[30m]))`
  — if all mass is ≤ the 20s bucket, no connect is held; gaps are queueing, not hangs.
- **Is the worker timeout firing?** `cloud_manager_skr_looper_connect_total_seconds_count{timeout="true"}`
  and the "SKR worker timeout exceeded" log. Empty/silent = timeout not involved (expected).
- **Sanity on saturation:** `gate_in_flight` ≈ cyclic + active-notification workers; beat rate
  `sum(rate(cloud_manager_skr_runtime_reconcile_total[15m]))` ≈ fleet / mean-gap.
- Mind the `timeout` label semantics from step 2 above when reading the phase histogram.

## Validating fairness changes

Use the dev-time simulation (`fairness_sim_test.go`, build tag `looper_sim`) — see
`pkg/skr/runtime/looper/README.md`. It runs the real queues/gate/`processOne`/`Notify` with a dummy
sleeping handler and a hot/cold notification skew, then reports mean vs max gaps and fails if any SKR's
max gap exceeds 3× the mean. It reproduces the starvation and confirms it is gone — without a cluster.
