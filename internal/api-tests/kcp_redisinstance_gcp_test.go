package api_tests

import (
	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

type redisInstanceBuilderGcp struct {
	tier         string
	replicaCount int32
	memorySizeGb int32
}

func (b *redisInstanceBuilderGcp) Build() *cloudcontrolv1beta1.RedisInstance {
	return &cloudcontrolv1beta1.RedisInstance{
		Spec: cloudcontrolv1beta1.RedisInstanceSpec{
			RemoteRef: cloudcontrolv1beta1.RemoteRef{
				Name:      uuid.NewString(),
				Namespace: "default",
			},
			IpRange: cloudcontrolv1beta1.IpRangeRef{
				Name: uuid.NewString(),
			},
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: uuid.NewString(),
			},
			Instance: cloudcontrolv1beta1.RedisInstanceInfo{
				Gcp: &cloudcontrolv1beta1.RedisInstanceGcp{
					MemorySizeGb: b.memorySizeGb,
					RedisVersion: "REDIS_7_0",
					AuthEnabled:  true,
					Tier:         b.tier,
					ReplicaCount: b.replicaCount,
				},
			},
		},
	}
}

var _ = Describe("Feature: KCP RedisInstance GCP", Ordered, func() {

	canCreate("RedisInstance GCP can be created with zero replicas if tier is BASIC", &redisInstanceBuilderGcp{tier: "BASIC", replicaCount: 0, memorySizeGb: 16})
	canCreate("RedisInstance GCP can be created with under memory size under 5GiB if tier is BASIC", &redisInstanceBuilderGcp{tier: "BASIC", replicaCount: 0, memorySizeGb: 1})
	canNotCreate("RedisInstance GCP can not be created with non-zero replicas if tier is BASIC", &redisInstanceBuilderGcp{tier: "BASIC", replicaCount: 1, memorySizeGb: 16}, "")

	canCreate("RedisInstance GCP can be created with non-zero replicas if tier is STANDARD_HA", &redisInstanceBuilderGcp{tier: "STANDARD_HA", replicaCount: 1, memorySizeGb: 16})
	canNotCreate("RedisInstance GCP can not be created with zero replicas if tier is STANDARD_HA", &redisInstanceBuilderGcp{tier: "STANDARD_HA", replicaCount: 0, memorySizeGb: 16}, "")
	canNotCreate("RedisInstance GCP can not be created with memory size under 5GiB if tier is STANDARD_HA", &redisInstanceBuilderGcp{tier: "STANDARD_HA", replicaCount: 0, memorySizeGb: 1}, "")
})
