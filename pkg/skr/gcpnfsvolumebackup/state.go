package gcpnfsvolumebackup

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
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

func (s *State) specCommaSeparatedAccessibleFrom() string {
	backup := s.ObjAsGcpNfsVolumeBackup()
	//sort backup.Spec.AccessibleFrom to have a consistent label value
	sort.Strings(backup.Spec.AccessibleFrom)
	return strings.Join(backup.Spec.AccessibleFrom, ",")
}

func (s *State) ShouldShortCircuit() bool {
	backup := s.ObjAsGcpNfsVolumeBackup()
	backupState := backup.Status.State
	return backupState == cloudresourcesv1beta1.GcpNfsBackupReady &&
		backup.Status.AccessibleFrom == s.specCommaSeparatedAccessibleFrom() &&
		!s.isTimeForCapacityUpdate() &&
		s.HasAllStatusLabels()
}

func (s *State) HasProperLabels() bool {
	backup := s.ObjAsGcpNfsVolumeBackup()

	if s.fileBackup.Labels == nil {
		return false
	}

	if s.fileBackup.Labels[gcpclient.ManagedByKey] != gcpclient.ManagedByValue {
		return false
	}

	if s.fileBackup.Labels[gcpclient.ScopeNameKey] != s.Scope.Name {
		return false
	}

	if s.fileBackup.Labels[util.GcpLabelSkrVolumeName] != backup.Spec.Source.Volume.Name {
		return false
	}
	if s.fileBackup.Labels[util.GcpLabelSkrVolumeNamespace] != backup.Spec.Source.Volume.Namespace {
		return false
	}

	if s.fileBackup.Labels[util.GcpLabelSkrBackupName] != backup.Name {
		return false
	}
	if s.fileBackup.Labels[util.GcpLabelSkrBackupNamespace] != backup.Namespace {
		return false
	}

	if s.fileBackup.Labels[util.GcpLabelShootName] != s.Scope.Spec.ShootName {
		return false
	}

	for _, shoot := range backup.Spec.AccessibleFrom {
		if s.fileBackup.Labels[ConvertToAccessibleFromKey(shoot)] != util.GcpLabelBackupAccessibleFrom {
			return false
		}
	}

	for key, fixedLabelValue := range s.fileBackup.Labels {
		if fixedLabelValue != util.GcpLabelBackupAccessibleFrom {
			continue
		}

		if !IsAccessibleFromKey(key) {
			continue
		}

		if slices.Contains(backup.Spec.AccessibleFrom, StripAccessibleFrompPrefix(key)) {
			continue
		}

		return false
	}

	return true
}

func (s *State) HasAllStatusLabels() bool {
	backup := s.ObjAsGcpNfsVolumeBackup()

	if backup.Status.FileStoreBackupLabels == nil {
		return false
	}

	if backup.Status.FileStoreBackupLabels[gcpclient.ManagedByKey] != gcpclient.ManagedByValue {
		return false
	}

	if backup.Status.FileStoreBackupLabels[gcpclient.ScopeNameKey] != s.Scope.Name {
		return false
	}

	if backup.Status.FileStoreBackupLabels[util.GcpLabelSkrVolumeName] != backup.Spec.Source.Volume.Name {
		return false
	}
	if backup.Status.FileStoreBackupLabels[util.GcpLabelSkrVolumeNamespace] != backup.Spec.Source.Volume.Namespace {
		return false
	}

	if backup.Status.FileStoreBackupLabels[util.GcpLabelSkrBackupName] != backup.Name {
		return false
	}
	if backup.Status.FileStoreBackupLabels[util.GcpLabelSkrBackupNamespace] != backup.Namespace {
		return false
	}

	if backup.Status.FileStoreBackupLabels[util.GcpLabelShootName] != s.Scope.Spec.ShootName {
		return false
	}

	for _, shoot := range backup.Spec.AccessibleFrom {
		if backup.Status.FileStoreBackupLabels[ConvertToAccessibleFromKey(shoot)] != util.GcpLabelBackupAccessibleFrom {
			return false
		}
	}

	for key, fixedLabelValue := range backup.Status.FileStoreBackupLabels {
		if fixedLabelValue != util.GcpLabelBackupAccessibleFrom {
			continue
		}

		if !IsAccessibleFromKey(key) {
			continue
		}

		if slices.Contains(backup.Spec.AccessibleFrom, StripAccessibleFrompPrefix(key)) {
			continue
		}

		return false
	}

	return true
}

func (s *State) SetFilestoreLabels() {
	backup := s.ObjAsGcpNfsVolumeBackup()

	if s.fileBackup.Labels == nil {
		s.fileBackup.Labels = make(map[string]string)
	}
	s.fileBackup.Labels[gcpclient.ManagedByKey] = gcpclient.ManagedByValue
	s.fileBackup.Labels[gcpclient.ScopeNameKey] = s.Scope.Name
	s.fileBackup.Labels[util.GcpLabelSkrVolumeName] = backup.Spec.Source.Volume.Name
	s.fileBackup.Labels[util.GcpLabelSkrVolumeNamespace] = backup.Spec.Source.Volume.Namespace
	s.fileBackup.Labels[util.GcpLabelSkrBackupName] = backup.Name
	s.fileBackup.Labels[util.GcpLabelSkrBackupNamespace] = backup.Namespace
	s.fileBackup.Labels[util.GcpLabelShootName] = s.Scope.Spec.ShootName
	for _, shoot := range backup.Spec.AccessibleFrom {
		s.fileBackup.Labels[ConvertToAccessibleFromKey(shoot)] = util.GcpLabelBackupAccessibleFrom
	}

	// delete any stale accessibleFrom labels
	for key, fixedLabelValue := range s.fileBackup.Labels {
		if fixedLabelValue != util.GcpLabelBackupAccessibleFrom {
			continue
		}

		if !IsAccessibleFromKey(key) {
			continue
		}

		if slices.Contains(backup.Spec.AccessibleFrom, StripAccessibleFrompPrefix(key)) {
			continue
		}

		delete(s.fileBackup.Labels, key)
	}
}

func ConvertToAccessibleFromKey(name string) string {
	return fmt.Sprintf("cm-allow-%s", name)
}

func IsAccessibleFromKey(key string) bool {
	return strings.HasPrefix(key, "cm-allow-")
}

func StripAccessibleFrompPrefix(key string) string {
	return strings.TrimPrefix(key, "cm-allow-")
}
