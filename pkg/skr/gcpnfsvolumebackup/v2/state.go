package v2

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	v2client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v2"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/klog/v2"
)

// State represents the v2 state using modern GCP protobuf types
type State struct {
	composed.State
	KymaRef    klog.ObjectRef
	KcpCluster composed.StateCluster
	SkrCluster composed.StateCluster

	Scope        *cloudcontrolv1beta1.Scope
	GcpNfsVolume *cloudresourcesv1beta1.GcpNfsVolume

	fileBackup       *filestorepb.Backup // Modern protobuf type
	fileBackupClient v2client.FileBackupClient
}

type StateFactory interface {
	NewState(ctx context.Context, baseState composed.State) (*State, error)
}

func NewStateFactory(
	kymaRef klog.ObjectRef,
	kcpCluster composed.StateCluster,
	skrCluster composed.StateCluster,
	fileBackupClientProvider gcpclient.GcpClientProvider[v2client.FileBackupClient],
) StateFactory {
	return &stateFactory{
		kymaRef:                  kymaRef,
		kcpCluster:               kcpCluster,
		skrCluster:               skrCluster,
		fileBackupClientProvider: fileBackupClientProvider,
	}
}

type stateFactory struct {
	kymaRef                  klog.ObjectRef
	kcpCluster               composed.StateCluster
	skrCluster               composed.StateCluster
	fileBackupClientProvider gcpclient.GcpClientProvider[v2client.FileBackupClient]
}

func (f *stateFactory) NewState(ctx context.Context, baseState composed.State) (*State, error) {
	return &State{
		State:            baseState,
		KymaRef:          f.kymaRef,
		KcpCluster:       f.kcpCluster,
		SkrCluster:       f.skrCluster,
		fileBackupClient: f.fileBackupClientProvider(),
	}, nil
}

func (s *State) ObjAsGcpNfsVolumeBackup() *cloudresourcesv1beta1.GcpNfsVolumeBackup {
	return s.Obj().(*cloudresourcesv1beta1.GcpNfsVolumeBackup)
}

func (s *State) isTimeForCapacityUpdate() bool {
	backup := s.ObjAsGcpNfsVolumeBackup()

	lastUpdate := backup.Status.LastCapacityUpdate
	configInterval := config.GcpConfig.GcpCapacityCheckInterval
	capacityUpdateDue := lastUpdate == nil || lastUpdate.Time.IsZero() || time.Since(lastUpdate.Time) > configInterval

	return capacityUpdateDue
}

func stopAndRequeueForCapacity() error {
	return composed.StopWithRequeueDelay(gcpclient.GcpCapacityCheckInterval)
}

// getAccessibleFromTargets returns the list of targets for accessibility labels.
// For Type=All, returns ["all"]. For Type=Specific, returns the Targets slice.
// Returns nil if AccessibleFrom is not set.
func (s *State) getAccessibleFromTargets() []string {
	backup := s.ObjAsGcpNfsVolumeBackup()
	accessibleFrom := backup.Spec.AccessibleFrom

	if accessibleFrom == nil {
		return nil
	}

	if accessibleFrom.Type == cloudresourcesv1beta1.AccessibleFromTypeAll {
		return []string{"all"}
	}

	return accessibleFrom.Targets
}

func (s *State) specCommaSeparatedAccessibleFrom() string {
	targets := s.getAccessibleFromTargets()
	if targets == nil {
		return ""
	}
	// Sort targets to have a consistent label value
	sortedTargets := make([]string, len(targets))
	copy(sortedTargets, targets)
	sort.Strings(sortedTargets)
	return strings.Join(sortedTargets, ",")
}

func (s *State) GetTargetGcpNfsVolumeNamespace() string {
	backup := s.ObjAsGcpNfsVolumeBackup()
	if backup.Spec.Source.Volume.Namespace != "" {
		return backup.Spec.Source.Volume.Namespace
	}

	return backup.Namespace
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
	if s.fileBackup.Labels[util.GcpLabelSkrVolumeNamespace] != s.GetTargetGcpNfsVolumeNamespace() {
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

	targets := s.getAccessibleFromTargets()
	for _, target := range targets {
		if s.fileBackup.Labels[ConvertToAccessibleFromKey(target)] != util.GcpLabelBackupAccessibleFrom {
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

		if slices.Contains(targets, StripAccessibleFromPrefix(key)) {
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
	if backup.Status.FileStoreBackupLabels[util.GcpLabelSkrVolumeNamespace] != s.GetTargetGcpNfsVolumeNamespace() {
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

	targets := s.getAccessibleFromTargets()
	for _, target := range targets {
		if backup.Status.FileStoreBackupLabels[ConvertToAccessibleFromKey(target)] != util.GcpLabelBackupAccessibleFrom {
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

		if slices.Contains(targets, StripAccessibleFromPrefix(key)) {
			continue
		}

		return false
	}

	return true
}

func (s *State) SetFilestoreLabels() {
	backup := s.ObjAsGcpNfsVolumeBackup()
	targets := s.getAccessibleFromTargets()

	if s.fileBackup.Labels == nil {
		s.fileBackup.Labels = make(map[string]string)
	}
	s.fileBackup.Labels[gcpclient.ManagedByKey] = gcpclient.ManagedByValue
	s.fileBackup.Labels[gcpclient.ScopeNameKey] = s.Scope.Name
	s.fileBackup.Labels[util.GcpLabelSkrVolumeName] = backup.Spec.Source.Volume.Name
	s.fileBackup.Labels[util.GcpLabelSkrVolumeNamespace] = s.GetTargetGcpNfsVolumeNamespace()
	s.fileBackup.Labels[util.GcpLabelSkrBackupName] = backup.Name
	s.fileBackup.Labels[util.GcpLabelSkrBackupNamespace] = backup.Namespace
	s.fileBackup.Labels[util.GcpLabelShootName] = s.Scope.Spec.ShootName
	for _, target := range targets {
		s.fileBackup.Labels[ConvertToAccessibleFromKey(target)] = util.GcpLabelBackupAccessibleFrom
	}

	// delete any stale accessibleFrom labels
	for key, fixedLabelValue := range s.fileBackup.Labels {
		if fixedLabelValue != util.GcpLabelBackupAccessibleFrom {
			continue
		}

		if !IsAccessibleFromKey(key) {
			continue
		}

		if slices.Contains(targets, StripAccessibleFromPrefix(key)) {
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

func StripAccessibleFromPrefix(key string) string {
	return strings.TrimPrefix(key, "cm-allow-")
}
