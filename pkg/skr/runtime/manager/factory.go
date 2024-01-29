package manager

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ Factory = &skrManagerFactory{}

type Factory interface {
	CreateManager(ctx context.Context, kymaName string, logger logr.Logger) (SkrManager, error)
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

func (f *skrManagerFactory) CreateManager(ctx context.Context, kymaName string, logger logr.Logger) (SkrManager, error) {
	secret := &corev1.Secret{}
	name := fmt.Sprintf("kubeconfig-%s", kymaName)
	err := f.kcpClient.Get(ctx, types.NamespacedName{
		Namespace: f.namespace,
		Name:      name,
	}, secret)
	if err != nil {
		return nil, fmt.Errorf("error getting kubeconfig secret: %w", err)
	}
	b, ok := secret.Data["config"]
	if !ok {
		return nil, fmt.Errorf("the kubeconfig secret for %s does not have the 'config' key", kymaName)
	}
	cc, err := clientcmd.NewClientConfigFromBytes(b)
	if err != nil {
		return nil, fmt.Errorf("error loading kubeconfig: %w", err)
	}
	restConfig, err := cc.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting rest config from kubeconfig: %w", err)
	}

	return New(restConfig, f.skrScheme, klog.ObjectRef{
		Name:      kymaName,
		Namespace: f.namespace,
	}, logger)
}
