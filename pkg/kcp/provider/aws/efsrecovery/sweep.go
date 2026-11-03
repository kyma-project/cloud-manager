package efsrecovery

import (
	"context"
	"errors"
	"time"

	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/go-logr/logr"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	"github.com/kyma-project/cloud-manager/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/utils/ptr"
)

// statusIdRefresher returns the current set of non-empty in-scope NfsInstance status.id values.
// It is called again immediately before each file-system delete to close the classify-then-delete gap.
type statusIdRefresher func(ctx context.Context) (map[string]struct{}, error)

// deleteMountTargets fires DeleteMountTarget for every mount target of every orphan (phase 1).
// It does not block waiting for mount targets to vanish; deleteSleep rate-limits between calls.
func deleteMountTargets(ctx context.Context, log logr.Logger, efsClient awsclient.EfsClient, orphans []efstypes.FileSystemDescription, deleteSleep time.Duration, lbl prometheus.Labels) {
	for i := range orphans {
		fsId := ptr.Deref(orphans[i].FileSystemId, "")
		mts, err := efsClient.DescribeMountTargets(ctx, fsId)
		if err != nil {
			if isThrottling(err) {
				metrics.EfsRecoveryThrottled.With(lbl).Inc()
			}
			log.Error(err, "error describing mount targets", "efsId", fsId)
			continue
		}
		for _, mt := range mts {
			mtId := ptr.Deref(mt.MountTargetId, "")
			if err := efsClient.DeleteMountTarget(ctx, mtId); err != nil {
				if isThrottling(err) {
					metrics.EfsRecoveryThrottled.With(lbl).Inc()
				}
				log.Error(err, "error deleting mount target", "efsId", fsId, "mountTargetId", mtId)
				continue
			}
			metrics.EfsRecoveryMountTargetsDeleted.With(lbl).Inc()
			log.Info("issued mount target delete", "efsId", fsId, "mountTargetId", mtId)
			sleep(ctx, deleteSleep)
		}
	}
}

// deleteFileSystems attempts DeleteFileSystem per orphan (phase 2), re-checking the refreshed
// status.id set immediately before each delete. FileSystemInUse means mount targets are not gone
// yet — skip and let a later pass retry. Returns the orphans still not deletable (stuck/in-use).
func deleteFileSystems(ctx context.Context, log logr.Logger, efsClient awsclient.EfsClient, orphans []efstypes.FileSystemDescription, refresh statusIdRefresher, deleteSleep time.Duration, lbl prometheus.Labels) []efstypes.FileSystemDescription {
	var remaining []efstypes.FileSystemDescription
	for i := range orphans {
		fsId := ptr.Deref(orphans[i].FileSystemId, "")

		// re-check keeper status against a fresh KCP view right before deleting
		ids, err := refresh(ctx)
		if err != nil {
			log.Error(err, "error refreshing NfsInstance status.id view; skipping delete this pass", "efsId", fsId)
			remaining = append(remaining, orphans[i])
			continue
		}
		if _, becameKeeper := ids[fsId]; becameKeeper {
			metrics.EfsRecoverySkipped.With(lbl).Inc()
			log.Info("skipping delete: EFS became a keeper", "efsId", fsId)
			continue
		}

		err = efsClient.DeleteFileSystem(ctx, fsId)
		if err != nil {
			var inUse *efstypes.FileSystemInUse
			if errors.As(err, &inUse) {
				metrics.EfsRecoverySkipped.With(lbl).Inc()
				log.Info("skipping delete: EFS still has mount targets, will retry", "efsId", fsId)
				remaining = append(remaining, orphans[i])
				continue
			}
			if isThrottling(err) {
				metrics.EfsRecoveryThrottled.With(lbl).Inc()
				log.Info("throttled deleting EFS, will retry", "efsId", fsId)
				remaining = append(remaining, orphans[i])
				sleep(ctx, deleteSleep)
				continue
			}
			log.Error(err, "error deleting EFS", "efsId", fsId)
			remaining = append(remaining, orphans[i])
			continue
		}
		metrics.EfsRecoveryDeleted.With(lbl).Inc()
		log.Info("issued EFS delete", "efsId", fsId)
		sleep(ctx, deleteSleep)
	}
	return remaining
}

func isThrottling(err error) bool {
	var t *efstypes.ThrottlingException
	return errors.As(err, &t)
}

// sleep pauses for d, returning early if ctx is cancelled.
func sleep(ctx context.Context, d time.Duration) {
	if d <= 0 {
		return
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}
