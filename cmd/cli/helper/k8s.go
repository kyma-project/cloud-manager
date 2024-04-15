package helper

import (
	"context"
	"errors"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func mustGetRestConfig(kubeconfigVars []string, contextVars []string) *rest.Config {
	kubeconfigVars = append(kubeconfigVars, "KUBECONFIG")
	var kubeconfig string
	for _, envVarName := range kubeconfigVars {
		kubeconfig = os.Getenv(envVarName)
		if len(kubeconfig) > 0 {
			break
		}
	}

	if len(kubeconfig) == 0 {
		home := homedir.HomeDir()
		if home == "" {
			panic(errors.New("unable to locate KCP kubeconfig, use CM_CONFIF_[KCP|SKR|GARDEN] or KUBECONFIG with CM_CONTEXT_[KCP|SKR|GARDEN] env vars"))
		}
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	var contextToUse string
	if len(contextVars) > 0 {
		cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
			&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}})
		raw, _ := cfg.RawConfig()

		var contexts []string
		for _, envVarName := range contextVars {
			context := os.Getenv(envVarName)
			if len(context) != 0 {
				contexts = append(contexts, context)
			}
		}
	loop:
		for contextName := range raw.Contexts {
			for _, cn := range contexts {
				if strings.Contains(contextName, cn) {
					contextToUse = contextName
					break loop
				}
			}
		}
	}

	if len(contextToUse) > 0 {
		fmt.Printf("Using context %s from kubeconfig %s\n", contextToUse, kubeconfig)
		rc, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
			&clientcmd.ConfigOverrides{
				ClusterInfo:    clientcmdapi.Cluster{Server: ""},
				CurrentContext: contextToUse,
			}).ClientConfig()
		if err != nil {
			panic(fmt.Errorf("error loading rest config with context %s from kubeconfig file %s", contextToUse, kubeconfig))
		}
		return rc
	}

	fmt.Printf("Using kubeconfig %s\n", kubeconfig)
	rc, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(fmt.Errorf("error loading rest config from kubeconfig file %s", kubeconfig))
	}
	return rc
}

func NewKcpConfig() *rest.Config {
	return mustGetRestConfig([]string{"CM_CONFIG_KCP"}, []string{"CM_CONTEXT_KCP"})
}

func NewGardenConfig() *rest.Config {
	return mustGetRestConfig([]string{"CM_CONFIG_GARDEN"}, []string{"CM_CONTEXT_GARDEN"})
}

func NewSkrConfig() *rest.Config {
	return mustGetRestConfig([]string{"CM_CONFIG_SKR"}, []string{"CM_CONTEXT_SKR"})
}

func NewKcpClient() client.Client {
	cfg := NewKcpConfig()
	c, err := client.New(cfg, client.Options{Scheme: KcpScheme})
	if err != nil {
		panic(fmt.Errorf("error creating KCP client: %w", err))
	}

	return c
}

func NewGardenClient() client.Client {
	cfg := NewGardenConfig()
	c, err := client.New(cfg, client.Options{Scheme: GardenScheme})
	if err != nil {
		panic(fmt.Errorf("error creating GARDEN client: %w", err))
	}

	return c
}

func NewSkrClient() client.Client {
	cfg := NewSkrConfig()
	c, err := client.New(cfg, client.Options{Scheme: SkrScheme})
	if err != nil {
		panic(fmt.Errorf("error creating SKR client: %w", err))
	}

	return c
}

func LoadKymaCRWithClient(c client.Client, ns, name string) (*unstructured.Unstructured, error) {
	kymaCR := util.NewKymaUnstructured()
	err := c.Get(context.Background(), types.NamespacedName{
		Namespace: ns,
		Name:      name,
	}, kymaCR)
	if err != nil {
		return nil, fmt.Errorf("error loading Kyma CR %s/%s: %w", ns, name, err)
	}
	return kymaCR, nil
}

func LoadKymaCR(ns, name string) (*unstructured.Unstructured, error) {
	c := NewKcpClient()
	return LoadKymaCRWithClient(c, ns, name)
}
