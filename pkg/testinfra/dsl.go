package testinfra

import (
	"context"
	"fmt"
	"strings"

	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	scopeconfig "github.com/kyma-project/cloud-manager/pkg/kcp/scope/config"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ InfraDSL = &infraDSL{}

type infraDSL struct {
	i Infra
}

func (dsl *infraDSL) WhenKymaModuleStateUpdates(kymaName string, state util.KymaModuleState) error {
	kymaCR := util.NewKymaUnstructured()
	err := dsl.i.KCP().Client().Get(dsl.i.Ctx(), types.NamespacedName{
		Namespace: dsl.i.KCP().Namespace(),
		Name:      kymaName,
	}, kymaCR)
	if err != nil {
		return err
	}

	err = util.SetKymaModuleStateToStatus(kymaCR, "cloud-manager", state)
	if err != nil {
		return err
	}

	err = dsl.i.KCP().Client().Status().Update(dsl.i.Ctx(), kymaCR)
	if err != nil {
		return err
	}

	return nil
}

func (dsl *infraDSL) WhenSkrIpRangeIsCreated(ctx context.Context, ns, name, cidr string, id string, conditions ...metav1.Condition) error {
	skrIpRange, err := dsl.createSkrIpRangeExists(ctx, ns, name, cidr)
	if err != nil {
		return err
	}
	if id != "" {
		skrIpRange.Status.Id = id
		err = dsl.i.SKR().Client().Update(ctx, skrIpRange)
		if err != nil {
			return fmt.Errorf("error updating SKR IpRange status with id")
		}
	}
	return dsl.i.SKR().GivenConditionIsSet(ctx, skrIpRange, conditions...)
}

func (dsl *infraDSL) GivenSkrIpRangeExists(ctx context.Context, ns, name, cidr string, id string, conditions ...metav1.Condition) error {
	skrIpRange, err := dsl.createSkrIpRangeExists(ctx, ns, name, cidr)
	if client.IgnoreAlreadyExists(err) != nil {
		return client.IgnoreAlreadyExists(err)
	}
	if id != "" {
		skrIpRange.Status.Id = id
		err = dsl.i.SKR().Client().Update(ctx, skrIpRange)
		if err != nil {
			return fmt.Errorf("error updating SKR IpRange status with id")
		}
	}
	return dsl.i.SKR().GivenConditionIsSet(ctx, skrIpRange, conditions...)
}

func (dsl *infraDSL) createSkrIpRangeExists(ctx context.Context, ns, name, cidr string) (*cloudresourcesv1beta1.IpRange, error) {
	skrIpRange := &cloudresourcesv1beta1.IpRange{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: cloudresourcesv1beta1.IpRangeSpec{
			Cidr: cidr,
		},
	}
	if err := dsl.i.SKR().Client().Create(ctx, skrIpRange); err != nil {
		return nil, err
	}
	if err := dsl.i.SKR().Client().Get(ctx, client.ObjectKeyFromObject(skrIpRange), skrIpRange); err != nil {
		return nil, err
	}
	return skrIpRange, nil
}

func (dsl *infraDSL) GivenGardenShootGcpExists(name string) error {
	// Gardener-credentials secret
	{
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: dsl.i.KCP().Namespace(),
				Name:      "gardener-credentials",
			},
		}
		err := dsl.i.KCP().Client().Get(dsl.i.Ctx(), client.ObjectKeyFromObject(secret), secret)
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		if apierrors.IsNotFound(err) {
			b, err := kubeconfigToBytes(restConfigToKubeconfig(dsl.i.Garden().Cfg()))
			if err != nil {
				return fmt.Errorf("error getting garden kubeconfig bytes: %w", err)
			}
			secret.Data = map[string][]byte{
				"kubeconfig": b,
			}

			err = dsl.i.KCP().Client().Create(dsl.i.Ctx(), secret)
			if client.IgnoreAlreadyExists(err) != nil {
				return fmt.Errorf("error creating gardener-credentials secret: %w", err)
			}
		}
	}

	// Shoot
	{
		shoot := &gardenertypes.Shoot{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: dsl.i.Garden().Namespace(),
				Name:      name,
			},
			Spec: gardenertypes.ShootSpec{
				Region: "us-west1",
				Provider: gardenertypes.Provider{
					Type: "gcp",
				},
				// SA1019 keep using SecretBinding until migrated to CredentialsBinding
				// nolint:staticcheck
				SecretBindingName: ptr.To(name),
			},
		}
		err := dsl.i.Garden().Client().Create(dsl.i.Ctx(), shoot)
		if err != nil {
			return err
		}
	}

	// SecretBinding
	{
		// SA1019 keep using SecretBinding until migrated to CredentialsBinding
		// nolint:staticcheck
		secretBinding := &gardenertypes.SecretBinding{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: dsl.i.Garden().Namespace(),
				Name:      name,
			},
			// SA1019 keep using SecretBinding until migrated to CredentialsBinding
			// nolint:staticcheck
			Provider: &gardenertypes.SecretBindingProvider{
				Type: "aws",
			},
			SecretRef: corev1.SecretReference{
				Name:      name,
				Namespace: dsl.i.Garden().Namespace(),
			},
		}
		err := dsl.i.Garden().Client().Create(dsl.i.Ctx(), secretBinding)
		if err != nil {
			return err
		}
	}

	// Secret
	{
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: dsl.i.Garden().Namespace(),
				Name:      name,
			},
			StringData: map[string]string{
				"serviceaccount.json": "{}",
			},
		}
		err := dsl.i.Garden().Client().Create(dsl.i.Ctx(), secret)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dsl *infraDSL) GivenScopeGcpExists(name string) error {
	shootNamespace := scopeconfig.ScopeConfig.GardenerNamespace // os.Getenv("GARDENER_NAMESPACE")
	project := strings.TrimPrefix(shootNamespace, "garden-")
	scope := &cloudcontrolv1beta1.Scope{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: dsl.i.KCP().Namespace(),
			Name:      name,
		},
		Spec: cloudcontrolv1beta1.ScopeSpec{
			KymaName:  name,
			ShootName: name,
			Region:    "us-west1",
			Provider:  cloudcontrolv1beta1.ProviderGCP,
			Scope: cloudcontrolv1beta1.ScopeInfo{
				Gcp: &cloudcontrolv1beta1.GcpScope{
					Project:    project,
					VpcNetwork: fmt.Sprintf("shoot--%s--%s", project, name),
					Network: cloudcontrolv1beta1.GcpNetwork{
						Nodes:    "10.250.0.0/22",
						Pods:     "10.96.0.0/13",
						Services: "10.104.0.0/13",
					},
					Workers: []cloudcontrolv1beta1.GcpWorkers{
						{
							Zones: []string{"us-west1-a", "us-west1-b", "us-west1-c"},
						},
					},
				},
			},
		},
	}
	err := dsl.i.KCP().Client().Get(dsl.i.Ctx(), client.ObjectKeyFromObject(scope), scope)
	if client.IgnoreNotFound(err) != nil {
		return err
	}
	if apierrors.IsNotFound(err) {
		err := dsl.i.KCP().Client().Create(dsl.i.Ctx(), scope)
		if err != nil {
			return err
		}
	}
	return nil
}

// =======================

var _ ClusterDSL = &clusterDSL{}

type clusterDSL struct {
	ci  ClusterInfo
	ctx func() context.Context
}

func (dsl *clusterDSL) GivenSecretExists(name string, data map[string][]byte) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: dsl.ci.Namespace(),
			Name:      name,
		},
	}
	err := dsl.ci.Client().Get(dsl.ctx(), client.ObjectKeyFromObject(secret), secret)
	if client.IgnoreNotFound(err) != nil {
		return err
	}
	secret.Data = data
	if apierrors.IsNotFound(err) {
		err = dsl.ci.Client().Create(dsl.ctx(), secret)
	} else {
		err = dsl.ci.Client().Update(dsl.ctx(), secret)
	}
	return err

}

func (dsl *clusterDSL) GivenNamespaceExists(name string) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	err := dsl.ci.Client().Get(dsl.ctx(), client.ObjectKeyFromObject(ns), ns)
	if err == nil {
		// already exists
		return nil
	}
	if client.IgnoreNotFound(err) != nil {
		// some error
		return err
	}
	err = dsl.ci.Client().Create(dsl.ctx(), ns)
	return client.IgnoreAlreadyExists(err)

}

func (dsl *clusterDSL) GivenConditionIsSet(ctx context.Context, obj composed.ObjWithConditions, conditions ...metav1.Condition) error {
	return dsl.setCondition(ctx, obj, conditions)
}

func (dsl *clusterDSL) WhenConditionIsSet(ctx context.Context, obj composed.ObjWithConditions, conditions ...metav1.Condition) error {
	return dsl.setCondition(ctx, obj, conditions)
}

func (dsl *clusterDSL) setCondition(ctx context.Context, obj composed.ObjWithConditions, conditions []metav1.Condition) error {
	err := dsl.ci.Client().Get(ctx, client.ObjectKeyFromObject(obj), obj)
	if err != nil {
		return err
	}
	for _, cond := range conditions {
		_ = meta.SetStatusCondition(obj.Conditions(), cond)
	}
	err = dsl.ci.Client().Status().Update(ctx, obj)
	if err != nil {
		return err
	}
	err = dsl.ci.Client().Get(ctx, client.ObjectKeyFromObject(obj), obj)
	return err
}
