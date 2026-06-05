package azuremanagedredis

import (
	"fmt"
	"maps"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redisenterprise/armredisenterprise/v3"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// AzureManagedRedisSpecTuple is the resolved KCP-side configuration that the
// SKR-facing tier letter expands to.
type AzureManagedRedisSpecTuple struct {
	SKU              armredisenterprise.SKUName
	HighAvailability bool
	ClusteringPolicy armredisenterprise.ClusteringPolicy
}

// tierToSpec is the canonical S/P/C → AMR cluster spec mapping.
//
// Notes:
//   - All P and C tiers share the same SKUs and pricing — they differ only in
//     ClusteringPolicy at the database level.
//   - C tiers start at C3 because OSSCluster requires multiple shards which
//     are only available on ComputeOptimized SKUs (X5+).
//   - The integer suffixes are not contiguous between families on purpose;
//     they reflect Azure-side capacity steps, not consecutive sizing.
var tierToSpec = map[cloudresourcesv1beta1.AzureManagedRedisTier]AzureManagedRedisSpecTuple{
	// S — Balanced, non-HA, EnterpriseCluster.
	cloudresourcesv1beta1.AzureManagedRedisTierS1: {armredisenterprise.SKUNameBalancedB0, false, armredisenterprise.ClusteringPolicyEnterpriseCluster},
	cloudresourcesv1beta1.AzureManagedRedisTierS2: {armredisenterprise.SKUNameBalancedB3, false, armredisenterprise.ClusteringPolicyEnterpriseCluster},
	cloudresourcesv1beta1.AzureManagedRedisTierS3: {armredisenterprise.SKUNameBalancedB5, false, armredisenterprise.ClusteringPolicyEnterpriseCluster},
	cloudresourcesv1beta1.AzureManagedRedisTierS4: {armredisenterprise.SKUNameBalancedB10, false, armredisenterprise.ClusteringPolicyEnterpriseCluster},
	cloudresourcesv1beta1.AzureManagedRedisTierS5: {armredisenterprise.SKUNameBalancedB20, false, armredisenterprise.ClusteringPolicyEnterpriseCluster},

	// P — ComputeOptimized, HA, EnterpriseCluster.
	cloudresourcesv1beta1.AzureManagedRedisTierP1: {armredisenterprise.SKUNameComputeOptimizedX5, true, armredisenterprise.ClusteringPolicyEnterpriseCluster},
	cloudresourcesv1beta1.AzureManagedRedisTierP2: {armredisenterprise.SKUNameComputeOptimizedX10, true, armredisenterprise.ClusteringPolicyEnterpriseCluster},
	cloudresourcesv1beta1.AzureManagedRedisTierP3: {armredisenterprise.SKUNameComputeOptimizedX20, true, armredisenterprise.ClusteringPolicyEnterpriseCluster},
	cloudresourcesv1beta1.AzureManagedRedisTierP4: {armredisenterprise.SKUNameComputeOptimizedX50, true, armredisenterprise.ClusteringPolicyEnterpriseCluster},
	cloudresourcesv1beta1.AzureManagedRedisTierP5: {armredisenterprise.SKUNameComputeOptimizedX100, true, armredisenterprise.ClusteringPolicyEnterpriseCluster},

	// C — ComputeOptimized, HA, OSSCluster.
	cloudresourcesv1beta1.AzureManagedRedisTierC3: {armredisenterprise.SKUNameComputeOptimizedX5, true, armredisenterprise.ClusteringPolicyOSSCluster},
	cloudresourcesv1beta1.AzureManagedRedisTierC4: {armredisenterprise.SKUNameComputeOptimizedX10, true, armredisenterprise.ClusteringPolicyOSSCluster},
	cloudresourcesv1beta1.AzureManagedRedisTierC5: {armredisenterprise.SKUNameComputeOptimizedX20, true, armredisenterprise.ClusteringPolicyOSSCluster},
	cloudresourcesv1beta1.AzureManagedRedisTierC6: {armredisenterprise.SKUNameComputeOptimizedX50, true, armredisenterprise.ClusteringPolicyOSSCluster},
	cloudresourcesv1beta1.AzureManagedRedisTierC7: {armredisenterprise.SKUNameComputeOptimizedX100, true, armredisenterprise.ClusteringPolicyOSSCluster},
}

// TierToSpec resolves a Kyma tier letter to the underlying AMR cluster spec.
func TierToSpec(tier cloudresourcesv1beta1.AzureManagedRedisTier) (AzureManagedRedisSpecTuple, error) {
	spec, ok := tierToSpec[tier]
	if !ok {
		return AzureManagedRedisSpecTuple{}, fmt.Errorf("unknown AzureManagedRedis tier %q", tier)
	}
	return spec, nil
}

// --- auth secret helpers ---

func getAuthSecretName(amr *cloudresourcesv1beta1.AzureManagedRedis) string {
	if amr.Spec.AuthSecret != nil && len(amr.Spec.AuthSecret.Name) > 0 {
		return amr.Spec.AuthSecret.Name
	}
	return amr.Name
}

func getAuthSecretLabels(amr *cloudresourcesv1beta1.AzureManagedRedis) map[string]string {
	labelsBuilder := util.NewLabelBuilder()

	if amr.Spec.AuthSecret != nil {
		for labelName, labelValue := range amr.Spec.AuthSecret.Labels {
			labelsBuilder.WithCustomLabel(labelName, labelValue)
		}
	}

	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelRedisInstanceStatusId, amr.Status.Id)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelRedisInstanceNamespace, amr.Namespace)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelCloudManaged, "true")
	labelsBuilder.WithCloudManagerDefaults()

	return labelsBuilder.Build()
}

func getAuthSecretAnnotations(amr *cloudresourcesv1beta1.AzureManagedRedis) map[string]string {
	if amr.Spec.AuthSecret == nil {
		return nil
	}
	result := map[string]string{}
	maps.Copy(result, amr.Spec.AuthSecret.Annotations)
	return result
}

// getAuthSecretBaseData mirrors the data layout used by AzureRedisInstance/AzureRedisCluster
// auth secrets. AMR exposes PrimaryEndpoint as a hostname and Port as a separate
// int32 field, so host == PrimaryEndpoint by contract.
func getAuthSecretBaseData(kcpAMR *cloudcontrolv1beta1.AzureManagedRedis) map[string][]byte {
	result := map[string][]byte{}

	if len(kcpAMR.Status.PrimaryEndpoint) > 0 {
		result["primaryEndpoint"] = []byte(kcpAMR.Status.PrimaryEndpoint)
		result["host"] = []byte(kcpAMR.Status.PrimaryEndpoint)
	}

	if kcpAMR.Status.Port > 0 {
		result["port"] = fmt.Appendf(nil, "%d", kcpAMR.Status.Port)
	}

	if len(kcpAMR.Status.AuthString) > 0 {
		result["authString"] = []byte(kcpAMR.Status.AuthString)
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
