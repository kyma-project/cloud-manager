package testinfra

import (
	"context"
	"fmt"
	gardenerTypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	awsgardener "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/aws/gardener"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/util"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ InfraDSL = &infraDSL{}

type infraDSL struct {
	i Infra
}

func (dsl *infraDSL) GivenKymaCRExists(name string) error {
	kymaCR := util.NewKymaUnstructured()
	kymaCR.SetLabels(map[string]string{
		"kyma-project.io/shoot-name": name,
	})
	kymaCR.SetName(name)
	kymaCR.SetNamespace(dsl.i.KCP().Namespace())

	err := dsl.i.KCP().Client().Get(dsl.i.Ctx(), client.ObjectKeyFromObject(kymaCR), kymaCR)
	if err == nil {
		// already exist
		return nil
	}
	if client.IgnoreNotFound(err) != nil {
		// some error
		return err
	}
	err = dsl.i.KCP().Client().Create(dsl.i.Ctx(), kymaCR)
	if err != nil {
		return err
	}

	// Kubeconfig secret
	{
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: dsl.i.KCP().Namespace(),
				Name:      fmt.Sprintf("kubeconfig-%s", name),
			},
		}
		err := dsl.i.KCP().Client().Get(dsl.i.Ctx(), client.ObjectKeyFromObject(secret), secret)
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		if apierrors.IsNotFound(err) {
			b, err := kubeconfigToBytes(restConfigToKubeconfig(dsl.i.SKR().Cfg()))
			if err != nil {
				return fmt.Errorf("error getting SKR kubeconfig bytes: %w", err)
			}
			secret.Data = map[string][]byte{
				"config": b,
			}

			err = dsl.i.KCP().Client().Create(dsl.i.Ctx(), secret)
			if client.IgnoreAlreadyExists(err) != nil {
				return fmt.Errorf("error creating SKR secret: %w", err)
			}
		}
	}

	return nil
}

func (dsl *infraDSL) GivenGardenShootAwsExists(name string) error {
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
		shoot := &gardenerTypes.Shoot{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: dsl.i.Garden().Namespace(),
				Name:      name,
			},
			Spec: gardenerTypes.ShootSpec{
				Region: "us-central1",
				Networking: &gardenerTypes.Networking{
					IPFamilies: []gardenerTypes.IPFamily{gardenerTypes.IPFamilyIPv4},
					Nodes:      pointer.String("10.180.0.0/16"),
					Pods:       pointer.String("100.64.0.0/12"),
					Services:   pointer.String("100.104.0.0/13"),
				},
				Provider: gardenerTypes.Provider{
					Type: "aws",
					InfrastructureConfig: &runtime.RawExtension{
						Object: &awsgardener.InfrastructureConfig{
							TypeMeta: metav1.TypeMeta{
								Kind:       "InfrastructureConfig",
								APIVersion: "aws.provider.extensions.gardener.cloud/v1alpha1",
							},
							Networks: awsgardener.Networks{
								VPC: awsgardener.VPC{
									CIDR: pointer.String("10.180.0.0/16"),
								},
								Zones: []awsgardener.Zone{
									{
										Name:     "eu-west-1a",
										Internal: "10.180.48.0/20",
										Public:   "10.180.32.0/20",
										Workers:  "10.180.0.0/19",
									},
									{
										Name:     "eu-west-1b",
										Internal: "10.180.112.0/20",
										Public:   "10.180.96.0/20",
										Workers:  "10.180.64.0/19",
									},
									{
										Name:     "eu-west-1c",
										Internal: "10.180.176.0/20",
										Public:   "10.180.160.0/20",
										Workers:  "10.180.128.0/19",
									},
								},
							},
						},
					},
				},
				SecretBindingName: pointer.String(name),
			},
		}
		err := dsl.i.Garden().Client().Create(dsl.i.Ctx(), shoot)
		if err != nil {
			return err
		}
	}

	// SecretBinding
	{
		secretBinding := &gardenerTypes.SecretBinding{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: dsl.i.Garden().Namespace(),
				Name:      name,
			},
			Provider: &gardenerTypes.SecretBindingProvider{
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
				"accessKeyID":     "accessKeyID",
				"secretAccessKey": "secretAccessKey",
			},
		}
		err := dsl.i.Garden().Client().Create(dsl.i.Ctx(), secret)
		if err != nil {
			return err
		}
	}

	return nil
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

	err = util.SetKymaModuleState(kymaCR, "cloud-manager", state)
	if err != nil {
		return err
	}

	err = dsl.i.KCP().Client().Status().Update(dsl.i.Ctx(), kymaCR)
	if err != nil {
		return err
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
	return err

}
