package efsrecovery

import (
	"context"
	"fmt"
	"time"

	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	"github.com/kyma-project/cloud-manager/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	// observerInterval is how often phase 3 re-checks for convergence.
	observerInterval = 30 * time.Second
	// maxSweepPasses bounds phase 2 retries so a genuinely stuck orphan cannot loop forever.
	maxSweepPasses = 50
)

// Recovery is the disposable, leader-gated garbage collector that reclaims orphaned EFS created
// by the NfsInstance create loop. It runs once per process start, then exits. Remove after close-out.
type Recovery struct {
	kcpClient   client.Client
	cfg         *awsconfig.EfsOrphanRecoveryConfig
	efsProvider EfsClientProvider
}

var _ manager.Runnable = (*Recovery)(nil)
var _ manager.LeaderElectionRunnable = (*Recovery)(nil)

// New constructs the recovery Runnable. Callers should only register it when cfg.Mode != off.
func New(kcpClient client.Client, cfg *awsconfig.EfsOrphanRecoveryConfig) *Recovery {
	return &Recovery{
		kcpClient:   kcpClient,
		cfg:         cfg,
		efsProvider: newEfsClientProvider(),
	}
}

// NeedLeaderElection ensures only the leader replica runs the GC, so the outgoing pod never
// deletes concurrently with the incoming one during a redeploy overlap.
func (r *Recovery) NeedLeaderElection() bool { return true }

func (r *Recovery) labels() prometheus.Labels {
	return prometheus.Labels{"shoot": r.cfg.ShootId, "region": r.cfg.Region, "mode": r.cfg.Mode}
}

// Start runs the recovery sweep once in a goroutine and returns when ctx is done. Any error is
// logged (not returned) so a recovery failure never crashes the manager.
func (r *Recovery) Start(ctx context.Context) error {
	logger := log.FromContext(ctx).WithName("efs-orphan-recovery").WithValues(
		"shoot", r.cfg.ShootId, "region", r.cfg.Region, "mode", r.cfg.Mode,
	)
	go func() {
		if err := r.run(ctx, logger); err != nil {
			logger.Error(err, "EFS orphan recovery finished with error")
		}
	}()
	<-ctx.Done()
	return nil
}

func (r *Recovery) run(ctx context.Context, logger logr.Logger) error {
	if r.cfg.Mode == awsconfig.EfsOrphanRecoveryModeOff {
		return nil // defensive: should not be registered when off
	}
	if r.cfg.ShootId == "" || r.cfg.AwsAccount == "" || r.cfg.Region == "" {
		return fmt.Errorf("incomplete EFS orphan recovery config: shootId/awsAccount/region required")
	}

	efsClient, err := r.efsProvider(ctx, r.cfg.AwsAccount, r.cfg.Region)
	if err != nil {
		return fmt.Errorf("building EFS client: %w", err)
	}

	lbl := r.labels()

	scoped, statusIds, mtCounts, err := r.loadAndScope(ctx, efsClient)
	if err != nil {
		return err
	}
	c := classify(scoped, statusIds, mtCounts)

	metrics.EfsRecoveryCandidates.With(lbl).Set(float64(len(scoped)))
	metrics.EfsRecoveryKeepers.With(lbl).Set(float64(len(c.keepers)))
	metrics.EfsRecoveryOrphans.With(lbl).Set(float64(len(c.orphans)))
	metrics.EfsRecoveryAmbiguous.With(lbl).Set(float64(len(c.ambiguous)))
	metrics.EfsRecoveryRemainingTagged.With(lbl).Set(float64(len(scoped)))

	logClassification(logger, scoped, c)

	if r.cfg.Mode == awsconfig.EfsOrphanRecoveryModeDryRun {
		logger.Info("EFS recovery dry-run complete; no deletes performed",
			"candidates", len(scoped), "keepers", len(c.keepers), "orphans", len(c.orphans), "ambiguous", len(c.ambiguous))
		return nil
	}

	// mode == on
	deleteMountTargets(ctx, logger, efsClient, c.orphans, r.cfg.DeleteSleep, lbl)

	remaining := c.orphans
	for pass := 0; pass < maxSweepPasses && len(remaining) > 0 && ctx.Err() == nil; pass++ {
		remaining = deleteFileSystems(ctx, logger, efsClient, remaining, r.refreshStatusIds, r.cfg.DeleteSleep, lbl)
		if len(remaining) > 0 {
			sleep(ctx, r.cfg.DeleteSleep)
		}
	}
	if len(remaining) > 0 {
		logger.Info("EFS recovery left stuck orphans after max passes", "stuck", len(remaining))
	}

	return r.observe(ctx, logger, efsClient, len(c.keepers), lbl)
}

// loadAndScope paginates all EFS in the target account/region, filters to the configured shoot
// tag, and builds the in-scope NfsInstance status.id set plus per-EFS mount-target counts.
func (r *Recovery) loadAndScope(ctx context.Context, efsClient awsclient.EfsClient) ([]efstypes.FileSystemDescription, map[string]struct{}, map[string]int, error) {
	all, err := efsClient.DescribeFileSystems(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("describing EFS: %w", err)
	}
	scoped := inScope(all, r.cfg.ShootId)

	statusIds, err := r.refreshStatusIds(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	return scoped, statusIds, countMountTargets(scoped), nil
}

// refreshStatusIds returns the set of non-empty status.id for NfsInstances whose scope shoot
// matches the configured shoot id.
func (r *Recovery) refreshStatusIds(ctx context.Context) (map[string]struct{}, error) {
	var scopes cloudcontrolv1beta1.ScopeList
	if err := r.kcpClient.List(ctx, &scopes); err != nil {
		return nil, fmt.Errorf("listing Scopes: %w", err)
	}
	shootByScopeName := map[string]string{}
	for i := range scopes.Items {
		shootByScopeName[scopes.Items[i].Name] = scopes.Items[i].Spec.ShootName
	}

	var nfsList cloudcontrolv1beta1.NfsInstanceList
	if err := r.kcpClient.List(ctx, &nfsList); err != nil {
		return nil, fmt.Errorf("listing NfsInstances: %w", err)
	}
	ids := map[string]struct{}{}
	for i := range nfsList.Items {
		n := &nfsList.Items[i]
		if shootByScopeName[n.Spec.Scope.Name] != r.cfg.ShootId {
			continue
		}
		if n.Status.Id != "" {
			ids[n.Status.Id] = struct{}{}
		}
	}
	return ids, nil
}

// observe waits until only keepers remain (orphan count reaches zero), emitting the terminal summary.
func (r *Recovery) observe(ctx context.Context, logger logr.Logger, efsClient awsclient.EfsClient, expectedKeepers int, lbl prometheus.Labels) error {
	for ctx.Err() == nil {
		all, err := efsClient.DescribeFileSystems(ctx)
		if err != nil {
			logger.Error(err, "observer: error describing EFS, retrying")
			sleep(ctx, observerInterval)
			continue
		}
		scoped := inScope(all, r.cfg.ShootId)
		statusIds, err := r.refreshStatusIds(ctx)
		if err != nil {
			logger.Error(err, "observer: error refreshing status ids, retrying")
			sleep(ctx, observerInterval)
			continue
		}
		c := classify(scoped, statusIds, countMountTargets(scoped))
		metrics.EfsRecoveryRemainingTagged.With(lbl).Set(float64(len(scoped)))
		metrics.EfsRecoveryKeepers.With(lbl).Set(float64(len(c.keepers)))
		metrics.EfsRecoveryOrphans.With(lbl).Set(float64(len(c.orphans)))

		if len(c.orphans) == 0 {
			logger.Info(fmt.Sprintf("EFS recovery complete for shoot %s: remaining_tagged=%d expected_keepers=%d ambiguous=%d",
				r.cfg.ShootId, len(scoped), expectedKeepers, len(c.ambiguous)))
			return nil
		}
		logger.Info("observer: orphans remain, waiting", "orphans", len(c.orphans), "remaining_tagged", len(scoped))
		sleep(ctx, observerInterval)
	}
	return ctx.Err()
}

func countMountTargets(scoped []efstypes.FileSystemDescription) map[string]int {
	m := map[string]int{}
	for i := range scoped {
		m[ptrDeref(scoped[i].FileSystemId)] = int(scoped[i].NumberOfMountTargets)
	}
	return m
}

func logClassification(logger logr.Logger, scoped []efstypes.FileSystemDescription, c classification) {
	for i := range c.keepers {
		logger.Info("classified EFS", "decision", "keeper", "efsId", ptrDeref(c.keepers[i].FileSystemId), "creationTime", c.keepers[i].CreationTime)
	}
	for i := range c.orphans {
		logger.Info("classified EFS", "decision", "orphan", "efsId", ptrDeref(c.orphans[i].FileSystemId), "creationTime", c.orphans[i].CreationTime)
	}
	for i := range c.ambiguous {
		logger.Info("classified EFS", "decision", "ambiguous", "efsId", ptrDeref(c.ambiguous[i].FileSystemId), "creationTime", c.ambiguous[i].CreationTime)
	}
}

func ptrDeref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
