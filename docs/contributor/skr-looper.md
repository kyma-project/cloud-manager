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

## Invariants that MUST hold

These are the load-bearing constraints the looper's fairness and correctness rest on. Each was learned
the hard way (a production incident + a fix). If any breaks, fairness or correctness fails. Stated as
rule → why → what breaks it:

1. **Nothing outside the cyclic worker may write to the cyclic queue for an SKR that is already an
   active member.** Not the notification sleeve, not `AddKyma` re-activation. *Why:* any external
   `Add`/`Delay`/`AddAfter` on an already-queued-or-processing member either reshuffles the tail (→
   starvation, the #2083 bug) or races the connect lifecycle (→ stranding, the AddKyma-strand bug).
   *Breaks it:* the notification success path's old `CyclicQueue().Delay` (fixed); the unconditional
   `cyclicQueue.Add` in `add()` on re-activation (fixed — guarded on `!alreadyActive`). Re-adds belong
   to the cyclic worker's own success path (`cyclicReAdd`) **only**.

2. **workqueue dedup protects only *queued* items, never *processing* ones.** *Why:* client-go's
   `Add` short-circuits (`dirty.Has` → `Touch`-or-return) only while the item sits in the queue; once a
   worker has `Get`'d it, it is in the `processing` set, not `dirty`/`queue`. *Breaks it:* assuming "the
   dirty-set dedups, so a re-add is a harmless no-op" — false during the connect window. An external
   `Add` interleaved with the connect's own `Get`/re-add/`Done` transitions can leave the SKR a member
   present in neither the dirty set nor the ready queue → stranded. This false assumption caused the
   AddKyma strand.

3. **The single-manager gate is the only cross-sleeve exclusion; membership ≠ queued ≠ processing ≠
   gate-claimed.** *Why:* these four states are independent (see the three-states section above; the
   gate adds the fourth). *Breaks it:* conflating any two — e.g. treating "is a member" as "is queued",
   or "gate free" as "not processing". This is the recurring trap behind both bugs.

4. **`AddKyma` must be idempotent for already-active SKRs.** *Why:* the KCP Kyma reconciler calls it
   periodically (resync) for live SKRs, not just once at activation. *Breaks it:* doing queue work
   (enqueue/Touch/count) on re-activation — activation adds to rotation and counts module-active
   exactly once; re-activation must be a pure no-op.

5. **The ~10s `Start` window is required work; fairness is achieved by ordering, not throughput.**
   *Why:* the worker keeps the per-SKR manager and its reconcilers alive to do real work; the window is
   not idle slack. *Breaks it:* trying to shrink gaps by cutting the connect time or tuning
   `workerTimeout` — the wrong lever (see "Why throughput is NOT the lever").

6. **A connect must always end by either re-queuing the SKR (success) or being a deliberate,
   documented drop.** *Why:* a live member silently left out of the queue never cycles again — that is
   exactly the strand. *Breaks it:* any code path through `processOne` that neither re-adds nor is an
   explicit, commented drop (membership-drop, gate-conflict-drop, shutdown).

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

## The AddKyma strand — the second reasoning chain

This is a **distinct, latent bug** from the starvation above. It predates the fairness work (the
unconditional `Add` was always there) but only became visible once per-SKR gaps were measured closely,
and it survived the fairness fix. Follow the chain — it took several days, multiple PRs, and prod
deploys to diagnose.

Symptom: after the fairness fix landed on stage, `count(increase(...reconcile_total[20m]) == 0)`
climbed and oscillated (26→43→59→…). It was **two populations**: a churning tail (~30, normal for a
6.6-min-mean saturated queue probed at the 3×-mean `[20m]` window) **plus a persistent stranded core
(~16)**. A single stranded SKR told the whole story: `reconcile_total` **flat at 1 for its entire
70-min container lifetime** — it connected exactly once. It was an active member
(`module_active_count=1`), had **no** notification activity, **no** gate conflicts, **no** panic, **no**
error log. The *only* activity after its single connect: repeated "Adding Kyma to SkrLooper" from the
KCP Kyma reconciler (`pkg/kcp/kyma/skrActivate.go` → `AddKyma`), firing periodically.

The race: `activeSkrCollection.add()` called `cyclicQueue.Add(kymaName)` **unconditionally**, even for
an SKR already in rotation. The periodic `AddKyma` therefore issued `q.Add(X)` for SKRs already
cycling, **racing the cyclic connect lifecycle** (the worker's `Get` clears dirty; `cyclicReAdd`/`Done`
re-push based on the dirty/processing state). Under the wrong interleaving between a mid-flight
connect's `Get`/re-add/`Done` and a concurrent external `Add`, X ends up a member absent from **both**
the workqueue dirty set and the ready queue — and nothing re-adds it (cyclic re-add only happens on the
*next* successful connect, which never comes). Result: stranded — member, one connect, never cycles.

Why a **persistent core + churning tail** (not a uniform tail): the strand is a one-way trap. Once an
SKR falls out of rotation it stays out (its counter flatlines), so the same names recur in every
`increase[20m]==0` snapshot — that is the persistent core. The churning tail is the ordinary,
healthy rotation of a saturated queue probed at a window wider than the mean gap.

The fix: guard the queue write on membership — an already-active SKR is already in rotation and must
not be re-added (invariants 1 and 4 above). In `add()`, `cyclicQueue.Add` and the module-active count
move **inside** the `!alreadyActive` block. A genuinely new activation still enqueues + counts; a
periodic re-activation of an already-cycling SKR becomes a pure no-op and cannot race its in-flight
connect. Note the old comment ("the workqueue dirty-set dedups, so a re-add is a no-op there") was the
flawed assumption — invariant 2: dedup holds only while the item is *queued*, never while it is being
*processed* (mid-connect), which is exactly the window the race exploits.

## Diagnostic playbook (re-derive it fast)

- **Is any SKR starved?** `count by (kyma)(increase(cloud_manager_skr_runtime_reconcile_total[20m]) == 0)`
  — returns a series per SKR idle > 20 min. Should trend to zero after the fix. Probe other windows
  (`[9m]`, `[11m]`, `[15m]`) to sketch the gap CDF.
- **Fixed set vs rotating tail?** Export the above as a time series and check whether the same `kyma`
  labels recur (starvation) or differ each sample (normal tail). This distinction is what proved it
  was a bug, not saturation.
- **Persistent-core vs churning-tail (the strand tell)?** Take **two `increase[20m]==0` snapshots ~20
  min apart** and diff the `kyma` sets. Names in **both** are the persistent stranded core (the AddKyma
  strand); names rotating in/out are the normal saturated-queue tail. A pure saturation tail has an
  empty "in both" set.
- **Single-SKR strand tell:** a **flat `reconcile_total` (stuck at 1) for a live member** — active
  (`module_active_count=1`), connected once, then no growth — with no notification activity, no gate
  conflict, no panic, no error, and only periodic "Adding Kyma to SkrLooper" logs. That is a stranded
  SKR, not a slow one.
- **Are connects actually fast?** `sum by (le)(rate(cloud_manager_skr_looper_connect_total_seconds_bucket[30m]))`
  — if all mass is ≤ the 20s bucket, no connect is held; gaps are queueing, not hangs.
- **Is the worker timeout firing?** `cloud_manager_skr_looper_connect_total_seconds_count{timeout="true"}`
  and the "SKR worker timeout exceeded" log. Empty/silent = timeout not involved (expected).
- **Sanity on saturation:** `gate_in_flight` ≈ cyclic + active-notification workers; beat rate
  `sum(rate(cloud_manager_skr_runtime_reconcile_total[15m]))` ≈ fleet / mean-gap.
- Mind the `timeout` label semantics from step 2 above when reading the phase histogram.

## Validating fairness changes

Use the dev-time simulation (`fairness_sim_test.go`, build tag `looper_sim`) — see
`pkg/skr/runtime/looper/README.md`. It runs the real queues/gate/`processOne`/`Notify` — and, via the
`-sim.addKymaEvery` driver, the real `add()` path — with a dummy sleeping handler, a hot/cold
notification skew, and periodic full-fleet `AddKyma` re-activation (mirroring the KCP Kyma reconciler's
resync). It reports mean vs max gaps and fails if any SKR's max gap exceeds 3× the mean.

It now reproduces **both** bugs and confirms both are gone — without a cluster:

- **Starvation** — driven by the hot/cold notification skew.
- **AddKyma strand** — driven by the periodic `AddKyma` sweep. This is why the `AddKyma` driver
  matters: the notification-only sim reported "FAIRNESS OK" while prod was stranding SKRs, because it
  never exercised `add()`. Against the unfixed `add()`, the sim with the driver reports FAIRNESS FAIL
  with a max gap in the tens of seconds (100×+ mean); against the fixed `add()` it stays ~1× mean. Set
  `-sim.addKymaEvery=0` to disable the driver.
