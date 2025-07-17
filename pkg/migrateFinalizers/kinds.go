package migrateFinalizers

import (
	"reflect"
	"strings"

	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
)

const kcpNamespace = "kcp-system"

func newKindInfo(list client.ObjectList, title string) KindInfo {
	if title == "" {
		x := reflect.ValueOf(list).Elem().Type().Name()
		title = strings.TrimSuffix(x, "List")
	}
	return KindInfo{
		Title: title,
		List:  list,
	}
}

type kindInfoProvider func() []KindInfo

func newKindsForKcp() []KindInfo {
	return pie.Map([]KindInfo{
		newKindInfo(&cloudcontrolv1beta1.IpRangeList{}, "KcpIpRange"),
		newKindInfo(&cloudcontrolv1beta1.NetworkList{}, ""),
		newKindInfo(&cloudcontrolv1beta1.NfsInstanceList{}, ""),
		newKindInfo(&cloudcontrolv1beta1.NukeList{}, ""),
		newKindInfo(&cloudcontrolv1beta1.RedisInstanceList{}, ""),
		newKindInfo(&cloudcontrolv1beta1.ScopeList{}, ""),
		newKindInfo(&cloudcontrolv1beta1.SkrStatusList{}, ""),
		newKindInfo(&cloudcontrolv1beta1.VpcPeeringList{}, ""),
		{
			Title: "Kyma",
			List:  util.NewKymaListUnstructured(),
		},
	}, func(x KindInfo) KindInfo {
		x.Namespace = kcpNamespace
		return x
	})
}

func newKindsForSkr() []KindInfo {
	return []KindInfo{
		newKindInfo(&cloudresourcesv1beta1.AwsNfsBackupScheduleList{}, ""),
		newKindInfo(&cloudresourcesv1beta1.AwsNfsVolumeList{}, ""),
		newKindInfo(&cloudresourcesv1beta1.AwsNfsVolumeBackupList{}, ""),
		newKindInfo(&cloudresourcesv1beta1.AwsNfsVolumeRestoreList{}, ""),
		newKindInfo(&cloudresourcesv1beta1.AwsRedisInstanceList{}, ""),
		newKindInfo(&cloudresourcesv1beta1.AwsVpcPeeringList{}, ""),
		newKindInfo(&cloudresourcesv1beta1.AzureRedisInstanceList{}, ""),
		newKindInfo(&cloudresourcesv1beta1.AzureVpcPeeringList{}, ""),
		newKindInfo(&cloudresourcesv1beta1.SapNfsVolumeList{}, ""),
		newKindInfo(&cloudresourcesv1beta1.CloudResourcesList{}, ""),
		newKindInfo(&cloudresourcesv1beta1.GcpNfsBackupScheduleList{}, ""),
		newKindInfo(&cloudresourcesv1beta1.GcpNfsVolumeList{}, ""),
		newKindInfo(&cloudresourcesv1beta1.GcpNfsVolumeBackupList{}, ""),
		newKindInfo(&cloudresourcesv1beta1.GcpNfsVolumeRestoreList{}, ""),
		newKindInfo(&cloudresourcesv1beta1.GcpRedisInstanceList{}, ""),
		newKindInfo(&cloudresourcesv1beta1.GcpVpcPeeringList{}, ""),
		newKindInfo(&cloudresourcesv1beta1.IpRangeList{}, "SkrIpRange"),
		newKindInfo(&corev1.ConfigMapList{}, ""),
		newKindInfo(&corev1.SecretList{}, ""),
		newKindInfo(&corev1.PersistentVolumeList{}, ""),
		newKindInfo(&corev1.PersistentVolumeClaimList{}, ""),
	}
}
