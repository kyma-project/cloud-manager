package alicloudrediscluster

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"strings"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"

	"github.com/kyma-project/cloud-manager/pkg/util"
)

func getAuthSecretName(alicloudRedis *cloudresourcesv1beta1.AlicloudRedisCluster) string {
	if alicloudRedis.Spec.AuthSecret != nil && len(alicloudRedis.Spec.AuthSecret.Name) > 0 {
		return alicloudRedis.Spec.AuthSecret.Name
	}

	return alicloudRedis.Name
}

func getAuthSecretLabels(alicloudRedis *cloudresourcesv1beta1.AlicloudRedisCluster) map[string]string {
	labelsBuilder := util.NewLabelBuilder()

	if alicloudRedis.Spec.AuthSecret != nil {
		for labelName, labelValue := range alicloudRedis.Spec.AuthSecret.Labels {
			labelsBuilder.WithCustomLabel(labelName, labelValue)
		}
	}

	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelRedisClusterStatusId, alicloudRedis.Status.Id)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelRedisClusterNamespace, alicloudRedis.Namespace)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelCloudManaged, "true")
	labelsBuilder.WithCloudManagerDefaults()
	pvLabels := labelsBuilder.Build()

	return pvLabels
}

func getAuthSecretAnnotations(alicloudRedis *cloudresourcesv1beta1.AlicloudRedisCluster) map[string]string {
	if alicloudRedis.Spec.AuthSecret == nil {
		return nil
	}
	result := map[string]string{}
	maps.Copy(result, alicloudRedis.Spec.AuthSecret.Annotations)
	return result
}

func getAuthSecretBaseData(kcpRedis *cloudcontrolv1beta1.RedisCluster) map[string][]byte {
	result := map[string][]byte{}

	if len(kcpRedis.Status.DiscoveryEndpoint) > 0 {
		result["discoveryEndpoint"] = []byte(kcpRedis.Status.DiscoveryEndpoint)

		splitEndpoint := strings.Split(kcpRedis.Status.DiscoveryEndpoint, ":")
		if len(splitEndpoint) >= 2 {
			host := splitEndpoint[0]
			port := splitEndpoint[1]
			result["host"] = []byte(host)
			result["port"] = []byte(port)
		}
	}

	if len(kcpRedis.Status.AuthString) > 0 {
		result["authString"] = []byte(kcpRedis.Status.AuthString)
	}

	if len(kcpRedis.Status.CaCert) > 0 {
		result["CaCert.pem"] = []byte(kcpRedis.Status.CaCert)
	}

	return result
}

func parseAuthSecretExtraData(extraDataTemplates map[string]string, authSecretBaseData map[string][]byte) map[string][]byte {
	baseDataStringMap := map[string]string{}
	for k, v := range authSecretBaseData {
		baseDataStringMap[k] = string(v)
	}

	return util.ParseTemplatesMapToBytesMap(extraDataTemplates, baseDataStringMap)
}

// alicloudRedisClusterTierMemoryGbMap maps each tier to its per-shard memory in GB.
// All AliCloud international regions use proxy-based sharding
// (redis.logic.sharding.*). The instance class is constructed dynamically in
// redisTierToInstanceClass because the proxy count is encoded in the class name
// alongside the shard count.
var alicloudRedisClusterTierMemoryGbMap = map[cloudresourcesv1beta1.AlicloudRedisClusterTier]int32{
	cloudresourcesv1beta1.AlicloudRedisClusterTierC3: 4,
	cloudresourcesv1beta1.AlicloudRedisClusterTierC4: 8,
	cloudresourcesv1beta1.AlicloudRedisClusterTierC5: 16,
	cloudresourcesv1beta1.AlicloudRedisClusterTierC6: 32,
	cloudresourcesv1beta1.AlicloudRedisClusterTierC7: 64,
}

func redisTierToInstanceClass(tier cloudresourcesv1beta1.AlicloudRedisClusterTier, shardCount int32) (string, error) {
	memGb, exists := alicloudRedisClusterTierMemoryGbMap[tier]
	if !exists {
		return "", errors.New("unknown redis cluster tier")
	}
	proxyCount := shardCount
	if proxyCount < 4 {
		proxyCount = 4
	}
	return fmt.Sprintf("redis.logic.sharding.%dg.%ddb.0rodb.%dproxy.default", memGb, shardCount, proxyCount), nil
}

// skrProxyShardTokensRe matches the shard-count and proxy-count tokens in
// proxy class names, e.g. ".4db.0rodb.8proxy." in
// "redis.logic.sharding.4g.4db.0rodb.8proxy.default".
var skrProxyShardTokensRe = regexp.MustCompile(`\.\d+db\.0rodb\.\d+proxy\.`)

// proxyClassTierKey strips the shard-count and proxy-count tokens from a
// proxy class name so that classes that differ only in shardCount compare as
// equal. Non-proxy class strings are returned unchanged.
func proxyClassTierKey(class string) string {
	if !strings.HasPrefix(class, "redis.logic.sharding.") &&
		!strings.HasPrefix(class, "redis.amber.logic.sharding.") {
		return class
	}
	return skrProxyShardTokensRe.ReplaceAllLiteralString(class, ".<N>db.0rodb.<N>proxy.")
}
