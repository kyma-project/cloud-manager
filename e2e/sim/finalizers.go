package sim

// FinalizerE2E is the finalizer the e2e/sim framework places on the non-cloud-manager
// resources it manages on behalf of simulated KCP components (KLM's KCP/SKR Kyma). It is
// deliberately distinct from api.CommonFinalizerDeletionHook so that cloud-manager's own
// stale-finalizer removal (kymaRemoveStaleCloudManagerFinalizer) does not interfere with
// the sim's ordered-teardown logic during e2e tests.
const FinalizerE2E = "cloud-manager.kyma-project.io/deletion-hook-e2e"
