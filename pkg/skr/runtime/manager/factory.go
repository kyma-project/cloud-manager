package manager

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ Factory = &skrManagerFactory{}

func NewScopeNotFoundError(err error) *ScopeNotFoundError {
	return &ScopeNotFoundError{err: err}
}

var _ error = &ScopeNotFoundError{}

type ScopeNotFoundError struct {
	err error
}

func (e *ScopeNotFoundError) Unwrap() error {
	return e.err
}

func (e *ScopeNotFoundError) Error() string {
	return fmt.Sprintf("Scope not found: %s", e.err.Error())
}

type Factory interface {
	LoadRestConfig(ctx context.Context, secretName, secretKey string) (*rest.Config, error)
	CreateManager(ctx context.Context, kymaName string, logger logr.Logger) (SkrManager, *cloudcontrolv1beta1.Scope, error)
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

func (f *skrManagerFactory) LoadRestConfig(ctx context.Context, secretName, secretKey string) (*rest.Config, error) {
	secret := &corev1.Secret{}
	err := f.kcpClient.Get(ctx, types.NamespacedName{
		Namespace: f.namespace,
		Name:      secretName,
	}, secret)
	if err != nil {
		return nil, fmt.Errorf("error getting kubeconfig secret: %w", err)
	}
	b, ok := secret.Data[secretKey]
	if !ok {
		return nil, fmt.Errorf("the kubeconfig secret %s does not have the '%s' key", secretName, secretKey)
	}
	cc, err := clientcmd.NewClientConfigFromBytes(b)
	if err != nil {
		return nil, fmt.Errorf("error loading kubeconfig: %w", err)
	}
	restConfig, err := cc.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting rest config from kubeconfig: %w", err)
	}

	return restConfig, nil
}

func (f *skrManagerFactory) CreateManager(ctx context.Context, kymaName string, logger logr.Logger) (SkrManager, *cloudcontrolv1beta1.Scope, error) {
	// ideally secret name and key should come from the GardenerCluster using util.ExtractGardenerClusterSummary
	// but as legacy and since it's fixed so far we can keep it hard-coded like this
	secretName := fmt.Sprintf("kubeconfig-%s", kymaName)
	restConfig, err := f.LoadRestConfig(ctx, secretName, "config")
	if err != nil {
		return nil, nil, err
	}

	nn := types.NamespacedName{
		Namespace: f.namespace,
		Name:      kymaName,
	}
	scope := &cloudcontrolv1beta1.Scope{}
	err = f.kcpClient.Get(ctx, nn, scope)
	if apierrors.IsNotFound(err) {
		return nil, nil, NewScopeNotFoundError(err)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("error loading Scope for kyma %s: %w", nn, err)
	}

	mngr, err := New(restConfig, f.skrScheme, klog.ObjectRef{
		Name:      kymaName,
		Namespace: f.namespace,
	}, logger)

	return mngr, scope, err
}
