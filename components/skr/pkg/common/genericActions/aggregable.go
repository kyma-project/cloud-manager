package genericActions

import (
	"github.com/elliotchance/pie/v2"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/skr/api/cloud-resources/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AggregateInfoList interface {
	All() []AggregateInfo
	Get(index int) AggregateInfo
	Append(sr cloudresourcesv1beta1.SourceRef, spec any)
	Remove(index int)
}

type AggregateInfo interface {
	SetSpec(spec any)
	GetSourceRef() cloudresourcesv1beta1.SourceRef
}

type Aggregable interface {
	GetDeletionTimestamp() *metav1.Time
	GetSpec() any
	GetSourceRef() cloudresourcesv1beta1.SourceRef
}

// ======================================================================

type GcpVpcPeeringInfoListWrap struct {
	*cloudresourcesv1beta1.GcpVpcPeeringInfoList
}

func (l *GcpVpcPeeringInfoListWrap) All() []AggregateInfo {
	return pie.Map(l.Items, func(item *cloudresourcesv1beta1.GcpVpcPeeringInfo) AggregateInfo {
		return item
	})
}

func (l *GcpVpcPeeringInfoListWrap) Get(index int) AggregateInfo {
	return l.Items[index]
}

func (l *GcpVpcPeeringInfoListWrap) Append(sr cloudresourcesv1beta1.SourceRef, spec any) {
	l.Items = append(l.Items, &cloudresourcesv1beta1.GcpVpcPeeringInfo{
		Spec:      spec.(cloudresourcesv1beta1.GcpVpcPeeringSpec),
		SourceRef: sr,
	})
}

func (l *GcpVpcPeeringInfoListWrap) Remove(index int) {
	l.Items = pie.Delete(l.Items, index)
}

// ======================================================================

type AzureVpcPeeringInfoListWrap struct {
	*cloudresourcesv1beta1.AzureVpcPeeringInfoList
}

func (l *AzureVpcPeeringInfoListWrap) All() []AggregateInfo {
	return pie.Map(l.Items, func(item *cloudresourcesv1beta1.AzureVpcPeeringInfo) AggregateInfo {
		return item
	})
}

func (l *AzureVpcPeeringInfoListWrap) Get(index int) AggregateInfo {
	return l.Items[index]
}

func (l *AzureVpcPeeringInfoListWrap) Append(sr cloudresourcesv1beta1.SourceRef, spec any) {
	l.Items = append(l.Items, &cloudresourcesv1beta1.AzureVpcPeeringInfo{
		Spec:      spec.(cloudresourcesv1beta1.AzureVpcPeeringSpec),
		SourceRef: sr,
	})
}

func (l *AzureVpcPeeringInfoListWrap) Remove(index int) {
	pie.Delete(l.Items, index)
}

// ======================================================================

type AwsVpcPeeringInfoListWrap struct {
	*cloudresourcesv1beta1.AwsVpcPeeringInfoList
}

func (l *AwsVpcPeeringInfoListWrap) All() []AggregateInfo {
	return pie.Map(l.Items, func(item *cloudresourcesv1beta1.AwsVpcPeeringInfo) AggregateInfo {
		return item
	})
}

func (l *AwsVpcPeeringInfoListWrap) Get(index int) AggregateInfo {
	return l.Items[index]
}

func (l *AwsVpcPeeringInfoListWrap) Append(sr cloudresourcesv1beta1.SourceRef, spec any) {
	l.Items = append(l.Items, &cloudresourcesv1beta1.AwsVpcPeeringInfo{
		Spec:      spec.(cloudresourcesv1beta1.AwsVpcPeeringSpec),
		SourceRef: sr,
	})
}

func (l *AwsVpcPeeringInfoListWrap) Remove(index int) {
	pie.Delete(l.Items, index)
}

// ======================================================================

type NfsVolumeInfoListWrap struct {
	*cloudresourcesv1beta1.NfsVolumeInfoList
}

func (l *NfsVolumeInfoListWrap) All() []AggregateInfo {
	return pie.Map(l.Items, func(item *cloudresourcesv1beta1.NfsVolumeInfo) AggregateInfo {
		return item
	})
}

func (l *NfsVolumeInfoListWrap) Get(index int) AggregateInfo {
	return l.Items[index]
}

func (l *NfsVolumeInfoListWrap) Append(sr cloudresourcesv1beta1.SourceRef, spec any) {
	l.Items = append(l.Items, &cloudresourcesv1beta1.NfsVolumeInfo{
		Spec:      spec.(cloudresourcesv1beta1.NfsVolumeSpec),
		SourceRef: sr,
	})
}

func (l *NfsVolumeInfoListWrap) Remove(index int) {
	pie.Delete(l.Items, index)
}
