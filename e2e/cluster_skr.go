package e2e

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/config/crd"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SkrCluster interface {
	Cluster

	SubscriptionName() string
	Provider() cloudcontrolv1beta1.ProviderType
	RuntimeID() string
	ShootName() string

	EnsureKymaCR(ctx context.Context) error
	ModuleAdd(ctx context.Context, moduleName string)
}

type defaultSkrCluster struct {
	Cluster
	subscriptionName string
	provider         cloudcontrolv1beta1.ProviderType
	runtimeID        string
	shootName        string
	kubeConfigBytes  []byte
}

func NewSkrCluster(clstr Cluster, rt *infrastructuremanagerv1.Runtime) SkrCluster {
	return &defaultSkrCluster{
		Cluster:          clstr,
		subscriptionName: rt.Spec.Shoot.SecretBindingName,
		provider:         cloudcontrolv1beta1.ProviderType(rt.Spec.Shoot.Provider.Type),
		runtimeID:        rt.Name,
		shootName:        rt.Spec.Shoot.Name,
	}
}

func (s *defaultSkrCluster) SubscriptionName() string {
	return s.subscriptionName
}

func (s *defaultSkrCluster) Provider() cloudcontrolv1beta1.ProviderType {
	return s.provider
}

func (s *defaultSkrCluster) RuntimeID() string {
	return s.runtimeID
}

func (s *defaultSkrCluster) ShootName() string {
	return s.shootName
}

func (s *defaultSkrCluster) EnsureKymaCR(ctx context.Context) error {
	// install kyma crd if not present
	_, err := s.Cluster.GetClient().RESTMapper().RESTMapping(schema.GroupKind{
		Group: operatorv1beta2.GroupVersion.Group,
		Kind:  "Kyma",
	}, operatorv1beta2.GroupVersion.Version)
	if meta.IsNoMatchError(err) {
		arr, err := crd.KLM()
		if err != nil {
			return fmt.Errorf("error reading Kyma CRD: %w", err)
		}
		err = util.Apply(ctx, s.GetClient(), arr)
		if err != nil {
			return fmt.Errorf("error installing Kyma CRD: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("error checking for installed kyma crd: %w", err)
	}

	// create namespace kyma-system
	ns := &corev1.Namespace{}
	err = s.GetClient().Get(ctx, client.ObjectKey{Name: "kyma-system"}, ns)
	if client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("error getting kyma-system namespace: %w", err)
	}
	if apierrors.IsNotFound(err) {
		ns.Name = "kyma-system"
		err = s.GetClient().Create(ctx, ns)
		if err != nil {
			return fmt.Errorf("error creating kyma-system namespace: %w", err)
		}
	}

	// load kyma resource
	kyma := &operatorv1beta2.Kyma{}
	err = s.GetClient().Get(ctx, types.NamespacedName{
		Namespace: "kyma-system",
		Name:      "kyma",
	}, kyma)
	if client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("error loading Kyma CR: %w", err)
	}
	if apierrors.IsNotFound(err) {
		// kyma cr does not exist - create it
		kyma.Name = "kyma"
		kyma.Namespace = "kyma-system"
		kyma.Labels = map[string]string{
			cloudcontrolv1beta1.LabelRuntimeId:      s.RuntimeID(),
			cloudcontrolv1beta1.LabelScopeShootName: s.ShootName(),
			cloudcontrolv1beta1.LabelScopeProvider:  string(s.Provider()),
		}
		kyma.Spec.Channel = "regular"
		err = s.GetClient().Create(ctx, kyma)
		if err != nil {
			return fmt.Errorf("error creating kyma: %w", err)
		}
		kyma.Status.State = "Ready"
		err = composed.PatchObjStatus(ctx, kyma, s.GetClient())
		if err != nil {
			return fmt.Errorf("error patching kyma status: %w", err)
		}
	}

	if !s.Has("kyma") {
		err = s.AddResources(ctx, &ResourceDeclaration{
			Alias:      "kyma",
			Kind:       "Kyma",
			ApiVersion: operatorv1beta2.GroupVersion.String(),
			Name:       "kyma",
			Namespace:  "kyma-system",
		})
		if err != nil {
			return fmt.Errorf("error adding Kyma resource: %w", err)
		}
	}

	// create namespace Config.SkrNamespace
	ns = &corev1.Namespace{}
	err = s.GetClient().Get(ctx, client.ObjectKey{Name: Config.SkrNamespace}, ns)
	if client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("error getting %q namespace: %w", Config.SkrNamespace, err)
	}
	if apierrors.IsNotFound(err) {
		ns.Name = Config.SkrNamespace
		err = s.GetClient().Create(ctx, ns)
		if err != nil {
			return fmt.Errorf("error creating %q namespace: %w", Config.SkrNamespace, err)
		}
	}
	return nil
}
