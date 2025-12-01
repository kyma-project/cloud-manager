package gcpnfsvolume

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpnfsbackupclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client"
	"github.com/kyma-project/cloud-manager/pkg/skr/common/defaultiprange"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"google.golang.org/api/file/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

type State struct {
	composed.State
	KymaRef           klog.ObjectRef
	KcpCluster        composed.StateCluster
	KcpNfsInstance    *cloudcontrolv1beta1.NfsInstance
	KcpIpRange        *cloudcontrolv1beta1.IpRange
	SkrIpRange        *cloudresourcesv1beta1.IpRange
	SkrCluster        composed.StateCluster
	PV                *corev1.PersistentVolume
	PVC               *corev1.PersistentVolumeClaim
	Scope             *cloudcontrolv1beta1.Scope
	SrcBackupFullPath string

	fileBackup *file.Backup

	fileBackupClient gcpnfsbackupclient.FileBackupClient
}

type StateFactory interface {
	NewState(ctx context.Context, baseState composed.State) (*State, error)
}

func NewStateFactory(
	kymaRef klog.ObjectRef,
	kcpCluster composed.StateCluster,
	skrCluster composed.StateCluster,
	fileBackupClientProvider gcpclient.ClientProvider[gcpnfsbackupclient.FileBackupClient],
	env abstractions.Environment,
) StateFactory {
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
	fileBackupClientProvider gcpclient.ClientProvider[gcpnfsbackupclient.FileBackupClient]
	env                      abstractions.Environment
}

func (f *stateFactory) NewState(ctx context.Context, baseState composed.State) (*State, error) {
	fbc, err := f.fileBackupClientProvider(ctx, gcpclient.GcpConfig.CredentialsFile)
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

func (s *State) ObjAsGcpNfsVolume() *cloudresourcesv1beta1.GcpNfsVolume {
	return s.Obj().(*cloudresourcesv1beta1.GcpNfsVolume)
}

func (s *State) IsChanged() bool {
	return s.KcpNfsInstance != nil && s.KcpNfsInstance.Spec.Instance.Gcp.CapacityGb != s.ObjAsGcpNfsVolume().Spec.CapacityGb
}

func (s *State) GetSkrIpRange() *cloudresourcesv1beta1.IpRange {
	return s.SkrIpRange
}

func (s *State) SetSkrIpRange(skrIpRange *cloudresourcesv1beta1.IpRange) {
	s.SkrIpRange = skrIpRange
}

func (s *State) ObjAsObjWithIpRangeRef() defaultiprange.ObjWithIpRangeRef {
	return s.ObjAsGcpNfsVolume()
}

func (s *State) IsAllowedToRestoreBackup() bool {
	if s.fileBackup == nil {
		return false
	}

	labels := s.fileBackup.Labels
	if labels == nil {
		return false
	}

	shootName := s.Scope.Spec.ShootName

	managed, exists := s.fileBackup.Labels[gcpclient.ManagedByKey]
	if !exists || managed != gcpclient.ManagedByValue {
		return false
	}

	allowed, exists := labels[ConvertToAccessibleFromKey(shootName)]
	if exists && allowed == util.GcpLabelBackupAccessibleFrom {
		return true
	}

	owner, exists := s.fileBackup.Labels[util.GcpLabelShootName]
	if exists && owner == shootName {
		return true
	}

	return false
}

func ConvertToAccessibleFromKey(name string) string {
	return fmt.Sprintf("cm-allow-%s", name)
}
