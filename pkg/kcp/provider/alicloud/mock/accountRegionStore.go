package mock

import (
	alicloudiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/iprange/client"
	alicloudvpcnetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/vpcnetwork/client"
)

var _ AccountRegion = (*accountRegionStore)(nil)

type accountRegionStore struct {
	*vpcStore
	region string
}

func newAccountRegionStore(region string) *accountRegionStore {
	return &accountRegionStore{region: region, vpcStore: newVpcStore()}
}

func (s *accountRegionStore) Region() string { return s.region }

func (s *accountRegionStore) IpRangeClient() alicloudiprangeclient.Client {
	return &iprangeClientView{vpcStore: s.vpcStore}
}

func (s *accountRegionStore) VpcNetworkClient() alicloudvpcnetworkclient.Client {
	return &vpcnetworkClientView{vpcStore: s.vpcStore}
}
