package mock

import (
	alicloudiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/iprange/client"
	alicloudnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/nfsinstance/client"
	alicloudredisclusterclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/rediscluster/client"
	alicloudredisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
	alicloudvpcnetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/vpcnetwork/client"
)

var _ AccountRegion = (*accountRegionStore)(nil)

type accountRegionStore struct {
	*vpcStore
	*nasStore
	*redisStore
	region string
}

func newAccountRegionStore(region string) *accountRegionStore {
	return &accountRegionStore{
		region:     region,
		vpcStore:   newVpcStore(),
		nasStore:   newNasStore(),
		redisStore: newRedisStore(),
	}
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

func (s *accountRegionStore) RedisInstanceClient() alicloudredisinstanceclient.Client {
	return &redisInstanceClientView{redisStore: s.redisStore}
}

func (s *accountRegionStore) RedisClusterClient() alicloudredisclusterclient.Client {
	return &redisClusterClientView{redisStore: s.redisStore}
}
