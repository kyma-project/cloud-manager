package mock

import (
	alicloudiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/iprange/client"
	alicloudnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/nfsinstance/client"
	alicloudvpcnetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/vpcnetwork/client"
)

var _ AccountRegion = (*accountRegionStore)(nil)

type accountRegionStore struct {
	*vpcStore
	*nasStore
	region string
}

func newAccountRegionStore(region string) *accountRegionStore {
	return &accountRegionStore{region: region, vpcStore: newVpcStore(), nasStore: newNasStore()}
}

func (s *accountRegionStore) Region() string { return s.region }

func (s *accountRegionStore) IpRangeClient() alicloudiprangeclient.Client {
	return &iprangeClientView{vpcStore: s.vpcStore}
}

func (s *accountRegionStore) VpcNetworkClient() alicloudvpcnetworkclient.Client {
	return &vpcnetworkClientView{vpcStore: s.vpcStore}
}

func (s *accountRegionStore) NfsInstanceClient() alicloudnfsinstanceclient.Client {
	return &nfsInstanceClientView{nasStore: s.nasStore}
}
