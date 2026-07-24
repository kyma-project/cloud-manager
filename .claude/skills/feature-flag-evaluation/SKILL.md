---
name: feature-flag-evaluation
description: Use when answering questions about Cloud Manager feature flags — what value a flag resolves to for a given landscape/provider/globalAccount/subaccount/shoot/broker plan, what a flag's default is, where a feature is enabled, which clients (front-runners) have a feature turned on, or why an API/CRD is or isn't exposed on an SKR. Trigger on mentions of feature flags, apiDisabled, go-feature-flag, ff_ga.yaml/ff_edge.yaml, "is X enabled on dev/stage/prod", "who has feature Y", "front-runners for Z", or targeting rules — even when the user doesn't name a specific flag.
role: worker
user-invocable: true
---

# Cloud Manager Feature Flag Evaluation

> **Note — masked identifiers.** All examples in this skill and its references use **masked**
> identifying data: global/subaccount UUIDs are shown as the zero UUID
> `00000000-0000-0000-0000-000000000000`, shoot names as `c-000000`/`c-000001`, and customer / line-of-business
> names as generic placeholders like "Awesome Company". These are illustrative only — never treat them as
> real values. When you actually evaluate a question, fetch the **live** config (Step 1) and use the real
> identifiers from the asker's scenario; the masking applies only to the documentation examples. The
> "Phoenix" team name is kept because it refers to the cloud-manager developers, but its global account is
> masked.

Cloud Manager gates provider features and exposed APIs behind feature flags evaluated with
[go-feature-flag](https://gofeatureflag.org/). This skill answers questions about **what a flag
resolves to** in a given context, **why**, and **where a feature is enabled**.

Answer three things, always: **the resolved value**, **the rule that decided it**, and **the trace of
rules it skipped to get there**. A bare "enabled/disabled" is rarely what the asker needs — they want to
know *why*, so they can change it or trust it.

## Step 1 — Pick the config source (landscape decides)

The config a running cloud-manager uses depends on the SKR's landscape. Match that when answering:

| Landscape | Config source | How to get it |
|---|---|---|
| `dev` | Live SAP deployment config | `gh` fetch (see below) |
| `stage` | Live SAP deployment config | `gh` fetch |
| `prod` | Live SAP deployment config | `gh` fetch |
| anything else (restricted / air-gapped markets, local, unknown) | **embedded** `pkg/feature/ff_ga.yaml` | read from repo |

The dev/stage/prod configs live in the **SAP-internal** repo `kyma/management-plane-config` on
`github.tools.sap`, under `manager.configFiles.featureFlags` in each `values.yaml`. Fetch the current
version — never reason from a stale copy:

```bash
gh api --hostname github.tools.sap \
  repos/kyma/management-plane-config/contents/argoenv/cloud-manager/<ENV>/values.yaml \
  --jq '.content' | base64 -d
```

`<ENV>` is `dev`, `stage`, or `prod`. The feature-flag block is nested under `manager.configFiles.featureFlags:`.

If `gh` isn't authenticated for `github.tools.sap`, the fetch fails. Ask the user to run
`gh auth login --hostname github.tools.sap` (suggest they type `! gh auth login --hostname github.tools.sap`),
then retry. Do not silently fall back to the embedded config for a dev/stage/prod question — the answer
would be wrong. Say what happened instead.

Two embedded configs live in the repo:
- `pkg/feature/ff_ga.yaml` — **the production fallback**, compiled into the binary via `//go:embed`. This
  is what runs on any landscape without a live SAP config (restricted markets, air-gapped, local without
  `FEATURE_FLAG_CONFIG_URL`/`FILE`). GA features on, non-GA off.
- `pkg/feature/ff_edge.yaml` — the e2e-test config (everything enabled). Used by CI e2e workflows and
  `tools/e2e/*` local runs. Not a real deployment; mention it only if the user asks about e2e/testing.

## Step 2 — Evaluate the rules in order

Each flag is a go-feature-flag toggle:

```yaml
apiDisabled:
  variations:
    enabled: false      # variation name -> value. Note: names are arbitrary labels.
    disabled: true      # here "disabled" resolves to true (API IS disabled)
  targeting:            # ordered list — FIRST matching query wins
    - name: Enable all for Phoenix
      query: globalAccount == "00000000-0000-0000-0000-000000000000"   # Phoenix team GA (masked)
      variation: enabled
    - name: Disable NfsBackup
      query: feature == "nfsBackup"
      variation: disabled
  defaultRule:          # used only if NO targeting rule matched
    variation: enabled
```

Evaluation algorithm — walk `targeting` top to bottom, **first query that matches wins**; if none match,
`defaultRule` applies. This ordering is the whole game: a broad "Disable X" rule lower down is overridden
by a specific "Enable X for client Y" rule placed above it. When you trace, list the rules you skipped and
*why each didn't match*, then the one that did.

**Watch the variation labels.** For `apiDisabled`, `enabled` → `false` and `disabled` → `true`, because the
flag means "is the API disabled." So "variation: enabled" means the API is **available**. Always resolve the
label to its actual boolean via the `variations` block; don't assume `enabled == true`.

## Step 3 — Build the evaluation context

Queries reference context keys. The asker's scenario supplies their values. Full list is in
`references/context-and-query.md`; the ones that matter most for targeting:

| Key | Meaning |
|---|---|
| `landscape` | `dev` / `stage` / `prod` — also picks the config (Step 1) |
| `feature` | derived from the CRD kind, **not** free text (see below) |
| `provider` | `aws`, `gcp`, `azure`, `openstack`, `alicloud` |
| `globalAccount`, `subAccount` | BTP account IDs — how per-client targeting is done |
| `shoot` | Gardener shoot name — the finest-grained client targeting |
| `brokerPlan` | KEB plan, e.g. `trial` |
| `region` | SKR region |

**The `feature` key is derived from the resource kind, not typed by the user.** A CRD implements
`SpecificToFeature()` returning a constant. Map the resource the user asks about to its feature name before
evaluating — otherwise a rule like `feature == "nfsBackup"` won't match. The mapping and how `apiDisabled`
special-cases the `CloudResources` CRD are in `references/features-and-apidisabled.md`. Read it whenever the
question involves `apiDisabled`, a specific CRD, or "which API is exposed."

**Source of truth for available features is the code, not the README.** The authoritative list of feature
names is the `FeatureName` constants in `pkg/feature/types/types.go` — read that file when you need the
current set. The table in `pkg/feature/README.md` is a human-maintained summary that lags the code (e.g. it
omits `rediscluster`, `azureManagedRedis`, `vpcdnslink`); use it for prose descriptions only, never as the
definitive list.

## Step 4 — Answer with verdict + winning rule + trace

Structure the answer so the reasoning is auditable:

```
**Flag:** apiDisabled
**Context:** landscape=prod, feature=nfsBackup, provider=aws, shoot=c-000001
**Resolved value:** false (API enabled) — from variation "enabled"

**Winning rule:** "Enable NFS Backup and Restore for Awesome Company except on Azure"
  query: feature == "nfsBackup" and provider != "azure" and shoot in ["c-000001", ...]  ✅ matches

**Trace (rules skipped above the winner):**
  1. "Enable all for Phoenix" — globalAccount != Phoenix GA ❌
  (winner is rule 2)
Rules below the winner ("Disable NfsBackup", "Disable all on trial") never evaluated.
```

For "where is feature X enabled?" or "who are the front-runners for X?": scan all three live configs (and
note the embedded fallback governs everything else), collect the `Enable …` targeting rules that mention
that feature, and report each rule's name + the accounts/shoots it targets. Rule names follow the convention
`(Disable|Enable) FEATURE (for CLIENT) (on PROVIDER|PLAN)`, so the `for CLIENT` part usually names the
front-runner directly. See `references/common-questions.md` for worked patterns.

## References

- `references/context-and-query.md` — all context keys, query operators (`==`, `!=`, `in`, `and`, `co`/contains), gotchas.
- `references/features-and-apidisabled.md` — CRD-kind→feature mapping, how `apiDisabled` controls exposed APIs, the `CloudResources` code override.
- `references/common-questions.md` — worked examples for each question type (value in context, default, where enabled, front-runners).