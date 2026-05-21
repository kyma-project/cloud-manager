# Runtime Flow Congruence Check Prompt

Use this prompt to verify that the CloudManager implementation matches the flows specified in
`docs/kcp/runtime.md`. Run it whenever the docs or implementation change.

---

## Instructions

Read the following files fresh — do not rely on cached knowledge or prior analysis:

**Documentation:**
- `docs/kcp/runtime.md` — the authoritative spec for all flows

**Implementation entry points — read all of these:**

*Runtime reconciler:*
- `internal/controller/cloud-control/runtime_controller.go`
- `pkg/kcp/runtime/reconciler.go`
- `pkg/kcp/runtime/subscriptionLoad.go`
- `pkg/kcp/runtime/subscriptionCreate.go`
- `pkg/kcp/runtime/subscriptionWaitReady.go`
- `pkg/kcp/runtime/vpcNetworkLoad.go`
- `pkg/kcp/runtime/vpcNetworkCreate.go`
- `pkg/kcp/runtime/vpcNetworkDelete.go`
- `pkg/kcp/runtime/vpcNetworkWaitReady.go`

*Subscription reconciler:*
- `internal/controller/cloud-control/subscription_controller.go`
- `pkg/kcp/subscription/reconciler.go`
- `pkg/kcp/subscription/gardenerClientCreate.go`
- `pkg/kcp/subscription/gardenerCredentialsRead.go`
- `pkg/kcp/subscription/statusSaveOnCreate.go`
- `pkg/kcp/subscription/labelBindingName.go`
- `pkg/kcp/subscription/resourcesLoad.go`
- `pkg/kcp/subscription/statusSaveOnDelete.go`

*VpcNetwork reconciler (common):*
- `internal/controller/cloud-control/vpcnetwork_controller.go`
- `pkg/kcp/vpcnetwork/reconciler.go`
- `pkg/kcp/vpcnetwork/nameDetermine.go`
- `pkg/kcp/vpcnetwork/specCidrBlocksValidate.go`
- `pkg/kcp/vpcnetwork/statusReady.go`
- `pkg/common/commonVpcName.go`

*VpcNetwork reconciler (provider-specific — read all four):*
- `pkg/kcp/provider/azure/vpcnetwork/new.go` (and infraObserve.go, infraCreateUpdate.go, infraDelete.go)
- `pkg/kcp/provider/aws/vpcnetwork/new.go` (and infraCreateUpdate.go, infraDelete.go)
- `pkg/kcp/provider/gcp/vpcnetwork/new.go` (and infraCreateUpdate.go, infraDelete.go)
- `pkg/kcp/provider/sap/vpcnetwork/new.go` (and infraCreateUpdate.go, infraDelete.go)

*Scope reconciler:*
- `internal/controller/cloud-control/scope_controller.go` (follow through to `pkg/kcp/scope/...`)

*Controller tests with mocked providers — read all of these:*
Extract Gherkin scenarios and expected outcomes from the ginkgo nodes (Describe, It, By...)
- `internal/controller/cloud-control/runtime_*_test.go`
- `internal/controller/cloud-control/scope_*_test.go`
- `internal/controller/cloud-control/subscription_*_test.go`
- `internal/controller/cloud-control/vpcnetwork_*_test.go`

---

## Analysis Methodology

**Treat CloudManager as a whole** — the docs describe what CloudManager does across all its
reconcilers (Runtime, Subscription, VpcNetwork, Scope) without necessarily naming which reconciler
does each step. Verify the end-to-end effect, not whether a specific reconciler performs a specific
step. For example: if the doc says "CloudManager creates Subscription and labels it", verify that the
label ends up on the Subscription at some point, regardless of which reconciler sets it.

**Two distinct flows** — analyze each separately:
1. **Legacy Gardener network flow** — `Runtime` is created without `vpcNetwork` field; CloudManager
   creates the `VpcNetwork` of type Gardener.
2. **New Kyma network flow** — `Runtime` is created with `vpcNetwork` already set; KEB pre-creates
   `VpcNetwork` (type Kyma) and `Subscription` before creating `Runtime`.

**For each flow, verify these behaviors:**

### Legacy Gardener flow — provisioning

- Does CloudManager discover/create `Subscription` using both lookup strategies (by name = secretBindingName, then by label `cloud-manager.kyma-project.io/binding-name` = secretBindingName)?
- Does Subscription creation use name = `Runtime.spec.shoot.secretBindingName`?
- Does the `cloud-manager.kyma-project.io/binding-name` label eventually get set on the Subscription?
- Does CloudManager create a `VpcNetwork` of type Gardener when `Runtime.spec.shoot.networking.vpcNetwork` is empty?
- Is the created VpcNetwork named equal to the Runtime name?
- Is the cloud VPC network name set to the Gardener format derived from shoot namespace and shoot name?
- Does CloudManager patch the `Runtime.spec.shoot.networking.vpcNetwork` field with the created VpcNetwork name?
- Does the Gardener-type VpcNetwork **observe** (read-only) the existing cloud VPC created by Gardener, writing identifiers to status? Check this for each provider (Azure, AWS, GCP, OpenStack) — note which providers have a proper observe path and which still use create/update.
- Does CloudManager read `CredentialsBinding` (not Shoot) to determine cloud subscription scope?

### Legacy Gardener flow — deprovisioning

- When Runtime is deleted, does CloudManager delete the `VpcNetwork` only if it is type Gardener?
- Does CloudManager NOT delete the `Subscription`?
- When Kyma is deleted and a Scope exists: does CloudManager create a `Nuke` resource and then delete the `Scope`? (This is handled by the Scope reconciler, not the Runtime reconciler — verify the Scope reconciler's trigger condition and action chain.)

### New Kyma network flow — provisioning

- When `Runtime.spec.shoot.networking.vpcNetwork` is set to a name different from the Runtime name, does CloudManager skip VpcNetwork creation?
- For Kyma-type VpcNetworks: does CloudManager **create** the cloud VPC network (not merely observe it)?
- Does VpcNetwork provisioning wait for a ready `Subscription` before proceeding?
- Are cloud VPC identifiers written to `VpcNetwork.status.identifiers`?
- Is the VPC network name written to `status.identifiers.name` (from `spec.vpcNetworkName` if set, otherwise from the KymaVpcName function)?

### New Kyma network flow — deprovisioning

- Does CloudManager NOT delete the `VpcNetwork` when the Runtime is deleted?
- Does CloudManager NOT delete the `Subscription` when the Runtime is deleted?
- When the `VpcNetwork` itself is explicitly deleted: does CloudManager verify no dependent resources exist, delete the cloud VPC, and then remove the finalizer?
- When the `Subscription` itself is explicitly deleted: does CloudManager verify no dependent resources exist and then remove the finalizer?

---

## Report Format

Produce a **congruence report** with the following structure:

**For each mismatch found:**
- What the doc says (with line number)
- What the implementation actually does
- Which providers are affected (if provider-specific)
- A suggested doc fix (one sentence)

**Classification:**
- Mark as `MISMATCH` when the implementation does something materially different from or missing vs the doc
- Mark as `PARTIAL` when behavior is correct for some providers but not others (e.g., observe path only on Azure)
- Do NOT report things that are implementation details not covered by the doc (e.g., cooldown mechanisms, CIDR validation) — those belong in the "extras" section

**Extras section (brief):**
List features present in the implementation that are not described in the docs. One line each.
Do not evaluate whether these extras are correct — just note their existence.

**Do NOT report as mismatches:**
- The doc describes CloudManager collectively; a behavior happening in a different reconciler than
  you might expect is NOT a mismatch as long as the end effect is correct
- Minor timing windows (e.g., a label being set one reconciliation cycle after object creation)
  unless they cause a functional problem
- Naming convention details that are stable and work correctly in practice
