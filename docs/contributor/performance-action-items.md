# Engineering Performance Action Items
**Goal:** Double monthly commit count from ~7/month to ~14/month consistently.
**Baseline:** 49 total commits, recent avg 7.2/month, peak 12 (Aug 2026).

---

## Immediate — Merge the Pipeline (Sep 2026)

These branches exist and are ready or near-ready. Each is a commit.

| # | Branch | Action |
|---|--------|--------|
| 1 | `pr-c/alicloud-redis-kcp-cluster` | Push and open PR — 9 commits waiting to land |
| 2 | `pr-d/alicloud-redis-skr-cluster` | Open PR immediately after pr-c merges — 10 commits ready |
| 3 | `feat/azure-managed-redis-config-fields` | Open PR for issue #1989 — 5 commits ready |
| 4 | `feat/alicloud-redis-e2e` | Open PR after pr-d merges — e2e coverage |
| 5 | `feat/gcp-deletion-protection` | Open PR — 5 commits ready, independent of alicloud |
| 6 | `feat/10392-gcp-warning-severity` | Open PR — 2 commits, small and mergeable now |

**Expected Sep yield: ~14 commits from existing work alone.**

---

## Structural — Change How You Work

### 1. One logical unit = one PR
Your current pattern bundles reconciler + tests + CRDs + docs. Split them:
- API types / CRDs → PR 1
- KCP reconciler → PR 2
- SKR reconciler → PR 3
- Controller tests → PR 4 (or with reconciler)
- Docs → PR 5

The AliCloud Redis series already does this (`pr-a` through `pr-d`). Apply this pattern to everything.

### 2. Fix bugs as standalone PRs
When you find a bug while working on a feature, fix it in a separate branch and merge it first.
Do not bundle: `fix(amr): HA + terminal state + DNS 409 + IpRange teardown` = 4 separate PRs.

### 3. Keep a "gap filler" backlog
July dropped to 3 commits — you finished AMR and hadn't started AliCloud yet.
Maintain a list of small standalone items to pick from during transition months:
- Test flake fixes (you already do this well — `#2138`)
- Dead link / doc fixes
- Small chores (Go bump, dependency updates, lint)
- Review-triggered fixes (when you review a PR and spot something, fix it yourself)

### 4. Start the next branch before the current one merges
Don't wait for `pr-c` to merge before touching `pr-d`. Stack the work.
Target: always have 2 branches in flight — one in review, one in development.

---

## Monthly Targets

| Commit type | Target/month |
|---|---|
| Feature PRs (main stream) | 4–6 |
| Fix PRs (bugs, flakes, review-triggered) | 3–4 |
| Chore / docs | 2–3 |
| **Total** | **9–13** |

---

## Tracker — Sep 2026

- [ ] Open PR: `pr-c/alicloud-redis-kcp-cluster`
- [ ] Open PR: `feat/gcp-deletion-protection`
- [ ] Open PR: `feat/10392-gcp-warning-severity`
- [ ] Open PR: `feat/azure-managed-redis-config-fields` (issue #1989)
- [ ] Open PR: `pr-d/alicloud-redis-skr-cluster` (after pr-c merges)
- [ ] Open PR: `feat/alicloud-redis-e2e` (after pr-d merges)
- [ ] Pick 1 item from gap-filler backlog
- [ ] Measure: count commits at end of month, compare to 7.2 baseline
