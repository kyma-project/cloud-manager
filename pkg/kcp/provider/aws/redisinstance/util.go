package redisinstance

import (
	"fmt"
	"strings"

	elasticacheTypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"k8s.io/utils/ptr"
)

func GetAwsElastiCacheSubnetGroupName(name string) string {
	return fmt.Sprintf("cm-%s", name)
}

func GetAwsElastiCacheParameterGroupName(name string) string {
	return fmt.Sprintf("cm-%s", name)
}

func GetAwsElastiCacheUserGroupName(name string) string {
	return fmt.Sprintf("cm-%s", name)
}

func GetAwsElastiCacheParameterGroupFamily(engineVersion string) string {
	if strings.HasPrefix(engineVersion, "2.6") {
		return "redis2.6"
	}
	if strings.HasPrefix(engineVersion, "2.8") {
		return "redis2.8"
	}
	if strings.HasPrefix(engineVersion, "3.2") {
		return "redis3.2"
	}
	if strings.HasPrefix(engineVersion, "4.0") {
		return "redis4.0"
	}
	if strings.HasPrefix(engineVersion, "5.0") {
		return "redis5.0"
	}
	if strings.HasPrefix(engineVersion, "6.") {
		return "redis6.x"
	}
	if strings.HasPrefix(engineVersion, "7.") {
		return "redis7"
	}

	return ""
}

func GetAwsElastiCacheClusterName(name string) string {
	return fmt.Sprintf("cm-%s", name)
}

func GetAwsAuthTokenSecretName(name string) string {
	return fmt.Sprintf("cm-%s/authToken", name)
}

func MapParameters(parameters []elasticacheTypes.Parameter) map[string]string {
	result := map[string]string{}

	for _, parameters := range parameters {
		result[ptr.Deref(parameters.ParameterName, "")] = ptr.Deref(parameters.ParameterValue, "")
	}

	return result
}

func GetDesiredParameters(defaultParameters, userDefinedParameters map[string]string) map[string]string {
	result := map[string]string{}

	for key, value := range defaultParameters {
		result[key] = value
	}

	for key, value := range userDefinedParameters {
		result[key] = value
	}

	return result
}

func GetMissmatchedParameters(currentParameters, desiredParameters map[string]string) map[string]string {
	result := map[string]string{}

	for key := range desiredParameters {
		if desiredParameters[key] == currentParameters[key] {
			continue
		}

		result[key] = desiredParameters[key]
	}

	return result
}

func ToParametersSlice(parametersMap map[string]string) []elasticacheTypes.ParameterNameValue {
	result := []elasticacheTypes.ParameterNameValue{}

	for key, value := range parametersMap {
		result = append(result, elasticacheTypes.ParameterNameValue{
			ParameterName:  ptr.To(key),
			ParameterValue: ptr.To(value),
		})
	}

	return result
}
