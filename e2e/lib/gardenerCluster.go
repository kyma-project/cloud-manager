package lib

import (
	"fmt"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"k8s.io/utils/clock"
)

var ErrGardenerClusterCredentialsExpired = fmt.Errorf("gardenercluster credentials expired")

// IsGardenerClusterSyncNeeded checks if the given GardenerCluster credentials are expired or about to expire.
// It returns true if the credentials are expired or will expire soon, along with the expiration time.
func IsGardenerClusterSyncNeeded(gc *infrastructuremanagerv1.GardenerCluster, clck clock.Clock) (bool, time.Duration) {
	if gc.Status.State != infrastructuremanagerv1.ReadyState {
		return true, time.Minute
	}
	if gc.Annotations == nil {
		return true, time.Minute
	}
	if _, ok := gc.Annotations[ForceKubeconfigRotationAnnotation]; ok {
		return true, time.Minute
	}
	expiresAt := clck.Now()
	val, ok := gc.Annotations[ExpiresAtAnnotation]
	if ok {
		ea, err := time.Parse(time.RFC3339, val)
		if err == nil {
			expiresAt = ea
		}
	}

	expiresIn := expiresAt.Sub(clck.Now())
	if expiresIn < time.Minute {
		return true, time.Minute
	}

	return false, expiresIn - time.Second
}
