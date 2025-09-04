package gcpnfsvolumebackup

import (
	"context"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client"
	"google.golang.org/api/file/v1"
	"k8s.io/klog/v2"
)

type State struct {
	composed.State
	KymaRef    klog.ObjectRef
	KcpCluster composed.StateCluster
	SkrCluster composed.StateCluster

	Scope        *cloudcontrolv1beta1.Scope
	GcpNfsVolume *cloudresourcesv1beta1.GcpNfsVolume

	fileBackup *file.Backup

	fileBackupClient client.FileBackupClient
}

type StateFactory interface {
	NewState(ctx context.Context, baseState composed.State) (*State, error)
}

func NewStateFactory(kymaRef klog.ObjectRef, kcpCluster composed.StateCluster, skrCluster composed.StateCluster,
	fileBackupClientProvider gcpclient.ClientProvider[client.FileBackupClient], env abstractions.Environment) StateFactory {

	return &stateFactory{
		kymaRef:                  kymaRef,
		kcpCluster:               kcpCluster,
		skrCluster:               skrCluster,
		fileBackupClientProvider: fileBackupClientProvider,
		env:                      env,
	}
}

type stateFactory struct {
	kymaRef                  klog.ObjectRef
	kcpCluster               composed.StateCluster
	skrCluster               composed.StateCluster
	fileBackupClientProvider gcpclient.ClientProvider[client.FileBackupClient]
	env                      abstractions.Environment
}

func (f *stateFactory) NewState(ctx context.Context, baseState composed.State) (*State, error) {
	fbc, err := f.fileBackupClientProvider(
		ctx,
		f.env.Get("GCP_SA_JSON_KEY_PATH"),
	)
	if err != nil {
		return nil, err
	}
	return &State{
		State:            baseState,
		KymaRef:          f.kymaRef,
		KcpCluster:       f.kcpCluster,
		SkrCluster:       f.skrCluster,
		fileBackupClient: fbc,
	}, nil
}

func (s *State) ObjAsGcpNfsVolumeBackup() *cloudresourcesv1beta1.GcpNfsVolumeBackup {
	return s.Obj().(*cloudresourcesv1beta1.GcpNfsVolumeBackup)
}

func (s *State) isTimeForCapacityUpdate() bool {
	backup := s.ObjAsGcpNfsVolumeBackup()

	lastUpdate := backup.Status.LastCapacityUpdate
	configInterval := gcpclient.GcpConfig.GcpCapacityCheckInterval
	capacityUpdateDue := lastUpdate == nil || lastUpdate.Time.IsZero() || time.Since(lastUpdate.Time) > configInterval

	return capacityUpdateDue

}

func stopAndRequeueForCapacity() error {
	return composed.StopWithRequeueDelay(gcpclient.GcpCapacityCheckInterval)
}

func StopAndRequeueForCapacityAction() composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		return stopAndRequeueForCapacity(), nil
	}
}
