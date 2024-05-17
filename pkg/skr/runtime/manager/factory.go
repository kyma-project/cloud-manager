package manager

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ Factory = &skrManagerFactory{}

type Factory interface {
	CreateManager(ctx context.Context, kymaName string, logger logr.Logger) (SkrManager, *cloudcontrolv1beta1.Scope, *unstructured.Unstructured, error)
}

func NewFactory(kcpClient client.Reader, namespace string, skrScheme *runtime.Scheme) Factory {
	return &skrManagerFactory{
		kcpClient: kcpClient,
		namespace: namespace,
		skrScheme: skrScheme,
	}
}

type skrManagerFactory struct {
	kcpClient client.Reader
	namespace string
	skrScheme *runtime.Scheme
}

func (f *skrManagerFactory) CreateManager(ctx context.Context, kymaName string, logger logr.Logger) (SkrManager, *cloudcontrolv1beta1.Scope, *unstructured.Unstructured, error) {
	secret := &corev1.Secret{}
	name := fmt.Sprintf("kubeconfig-%s", kymaName)
	err := f.kcpClient.Get(ctx, types.NamespacedName{
		Namespace: f.namespace,
		Name:      name,
	}, secret)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error getting kubeconfig secret: %w", err)
	}
	b, ok := secret.Data["config"]
	if !ok {
		return nil, nil, nil, fmt.Errorf("the kubeconfig secret for %s does not have the 'config' key", kymaName)
	}
	cc, err := clientcmd.NewClientConfigFromBytes(b)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error loading kubeconfig: %w", err)
	}
	restConfig, err := cc.ClientConfig()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error getting rest config from kubeconfig: %w", err)
	}

	nn := types.NamespacedName{
		Namespace: f.namespace,
		Name:      kymaName,
	}
	scope := &cloudcontrolv1beta1.Scope{}
	err = f.kcpClient.Get(ctx, nn, scope)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error loading Scope for kyma %s: %w", kymaName, err)
	}

	kyma := util.NewKymaUnstructured()
	err = f.kcpClient.Get(ctx, nn, kyma)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error loading Kyma for kyma %s: %w", kymaName, err)
	}

	mngr, err := New(restConfig, f.skrScheme, klog.ObjectRef{
		Name:      kymaName,
		Namespace: f.namespace,
	}, logger)

	return mngr, scope, kyma, err
}
