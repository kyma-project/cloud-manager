package dsl

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/testinfra"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func CreateGardenerClusterCR(ctx context.Context, infra testinfra.Infra, gardenerCluster *unstructured.Unstructured, kymaName, shootName string, provider cloudcontrolv1beta1.ProviderType, opts ...ObjAction) error {
	if gardenerCluster == nil {
		gardenerCluster = util.NewGardenerClusterUnstructured()
	}

	// first default name and namespace, so they can be used in second pass
	NewObjActions(opts...).
		Append(
			WithName(kymaName),
			WithNamespace(DefaultKcpNamespace),
		).
		ApplyOnObject(gardenerCluster)

	if gardenerCluster.GetName() == "" {
		return fmt.Errorf("gardenerCluster name is empty: %w", common.ErrLogical)
	}

	givenSummary, _ := util.ExtractGardenerClusterSummary(gardenerCluster)
	if givenSummary == nil {
		givenSummary = &util.GardenerClusterSummary{}
	}

	// if shootName is given then it takes precedence over the value in the resource
	if shootName == "" {
		shootName = givenSummary.Shoot
	}
	givenSummary.Shoot = shootName

	if shootName == "" {
		return fmt.Errorf("gardenerCluster shoot is empty: %w", common.ErrLogical)
	}

	if provider == "" {
		return fmt.Errorf("gardenerCluster provider is empty: %w", common.ErrLogical)
	}

	// now gardenerCluster has defaulted name and namespace if not given
	NewObjActions().
		Append(
			WithLabels(map[string]string{
				cloudcontrolv1beta1.LabelScopeGlobalAccountId: "ga-account-id",
				cloudcontrolv1beta1.LabelScopeSubaccountId:    "subaccount-id",
				cloudcontrolv1beta1.LabelScopeShootName:       shootName,
				cloudcontrolv1beta1.LabelScopeRegion:          "region-some",
				cloudcontrolv1beta1.LabelScopeBrokerPlanName:  string(provider),
			}),
		).
		ApplyOnObject(gardenerCluster)

	if givenSummary.Name == "" {
		givenSummary.Name = fmt.Sprintf("kubeconfig-%s", gardenerCluster.GetName())
	}
	if givenSummary.Namespace == "" {
		givenSummary.Namespace = DefaultKcpNamespace
	}
	if givenSummary.Key == "" {
		givenSummary.Key = "config"
	}
	err := util.SetGardenerClusterSummary(gardenerCluster, *givenSummary)
	if err != nil {
		return fmt.Errorf("gardenerCluster failed to set summary: %w", err)
	}

	err = CreateObj(ctx, infra.KCP().Client(), gardenerCluster)
	if err != nil {
		return fmt.Errorf("failed creating GardenerCluster CR: %w", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      givenSummary.Name,
			Namespace: givenSummary.Namespace,
		},
	}
	b, err := kubeconfigToBytes(restConfigToKubeconfig(infra.SKR().Cfg()))
	if err != nil {
		return fmt.Errorf("error getting SKR kubeconfig bytes: %w", err)
	}
	secret.Data = map[string][]byte{
		givenSummary.Key: b,
	}

	err = infra.KCP().Client().Create(ctx, secret)
	if err != nil {
		return fmt.Errorf("failed creating GardenerCluster kubeconfig secret: %w", err)
	}

	return nil
}
