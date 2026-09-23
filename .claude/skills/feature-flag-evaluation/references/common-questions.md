# Common questions — worked patterns

Each pattern: what to fetch, how to evaluate, how to present. Always end with **verdict + winning rule +
trace**.

---

## 1. "What is the value of flag X for <context> on <landscape>?"

1. Landscape picks the config (dev/stage/prod → `gh` fetch; else embedded `ff_ga.yaml`).
2. Find flag X's block. Map any resource in the question to its `feature` name.
3. Walk `targeting` top→bottom; first matching `query` wins. Resolve the winning `variation` through
   `variations`. If none match, use `defaultRule`.
4. Present verdict, the matched rule, and the skipped rules with the reason each missed.

**Example** — "Is `nfsBackup` enabled on prod for an AWS shoot `c-000002`?"
- Fetch prod config. feature=nfsBackup, provider=aws, shoot=c-000002.
- Rule "Enable NFS Backup and Restore for Awesome Company" — `feature=="nfsBackup" and provider!="azure" and shoot in ["c-000002"]` ✅.
- Verdict: `apiDisabled=false` → **NFS Backup API is exposed**. Winner: Awesome Company rule. Skipped above:
  Phoenix (GA mismatch), another backup front-runner (shoot not in list). Below (Disable NfsBackup) never reached.

---

## 2. "What is the default value of flag X?"

The default is `defaultRule.variation` resolved through `variations` — the value when **no** targeting rule
matches. Report it per-landscape if they differ (the embedded `ff_ga.yaml` often disables what dev enables).
Note that "default" ≠ "the value for my cluster" — a targeting rule may override it. Say which you're
answering.

**Example** — "What's the default for `apiDisabled`?" → `defaultRule: enabled` → `false` → APIs available by
default; non-GA features are turned off by explicit `Disable …` rules, not by the default.

---

## 3. "Where is feature X enabled?" / "Which landscapes have X on?"

Scan all three live configs (dev, stage, prod) and note the embedded `ff_ga.yaml` governs every other
landscape. For each, determine the resolved value for a generic context (no client-specific keys) — i.e.
what a non-front-runner cluster gets — and separately list any `Enable …` targeting rules that switch it on
for specific clients. Distinguish "on by default here" from "on only for these clients here."

**Example** — "Where is RedisCluster enabled?" `feature=="rediscluster"`:
- dev: `Disable RedisCluster` rule may be absent → check; typically enabled on dev.
- stage/prod: `Disable RedisCluster` rule present → off by default, **except** front-runner rules like
  "Enable Redis for Awesome Company" (`shoot in ["c-000003"]`).
- Report: default-off on stage/prod, enabled for Awesome Company (c-000003); on for all of dev.

---

## 4. "Who are the front-runners for feature X?"

Front-runners = clients with an `Enable X for CLIENT …` targeting rule above the disable rules. Grep the
live configs for `Enable` rules mentioning feature X, and read off the `name` (names the client) and the
`globalAccount` / `subAccount` / `shoot` list they target.

**Example** — "Front-runners for `vpcdnslink`?" → prod rule "Enable VpcDnsLink for Awesome Company on Azure",
targeting `provider=="azure" and shoot in [c-000004, c-000005, …]`. Report client = Awesome Company, the shoots, and
that it's Azure-only.

Present as a table: client (from rule name) | landscape | targeting key + values | provider constraint.

---

## 5. "Why is API/CRD Y (not) exposed on this SKR?"

This is an `apiDisabled` question. First check the `CloudResources` override: if Y is the `CloudResources`
CRD, it's always exposed by code (features-and-apidisabled.md). Otherwise map Y's kind → `feature`, then run
pattern 1 against `apiDisabled` for that landscape/context. Remember the inverted variations: `enabled` →
`apiDisabled=false` → exposed.

---

## Fetching a config quickly

```bash
gh api --hostname github.tools.sap \
  repos/kyma/management-plane-config/contents/argoenv/cloud-manager/prod/values.yaml \
  --jq '.content' | base64 -d
```

Swap `prod` for `dev`/`stage`. The flags are under `manager.configFiles.featureFlags:`. If auth fails, ask
the user to `gh auth login --hostname github.tools.sap`. For non-dev/stage/prod landscapes, read
`pkg/feature/ff_ga.yaml` from the repo instead — that's the embedded fallback those clusters actually run.