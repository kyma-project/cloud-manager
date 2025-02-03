package testinfra

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/testinfra/infraTypes"
	"github.com/onsi/ginkgo/v2"
	"io"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/rest"
	"os"
	"path"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var _ Infra = &infra{}

type infra struct {
	InfraEnv
	InfraDSL

	clusters map[infraTypes.ClusterType]*clusterInfo
}

func (i *infra) KCP() ClusterInfo {
	return i.clusters[infraTypes.ClusterTypeKcp]
}

func (i *infra) SKR() ClusterInfo {
	return i.clusters[infraTypes.ClusterTypeSkr]
}

func (i *infra) Garden() ClusterInfo {
	return i.clusters[infraTypes.ClusterTypeGarden]
}

func (i *infra) Stop() error {
	i.stopControllers()

	var lastErr error
	for name, cluster := range i.clusters {
		ginkgo.By(fmt.Sprintf("Stopping cluster %s", name))
		if err := cluster.env.Stop(); err != nil {
			err = fmt.Errorf("error stopping env %s: %w", name, err)
			if lastErr == nil {
				lastErr = err
			} else {
				lastErr = fmt.Errorf("; %w", err)
			}
		}
	}

	return lastErr
}

// =======================

var _ ClusterInfo = &clusterInfo{}

type clusterInfo struct {
	ClusterEnv
	ClusterDSL

	crdDirs []string
	env     *envtest.Environment
	cfg     *rest.Config
	scheme  *runtime.Scheme
	client  client.Client
}

func (c *clusterInfo) Scheme() *runtime.Scheme {
	return c.scheme
}

func (c *clusterInfo) Client() client.Client {
	return c.client
}

func (c *clusterInfo) Cfg() *rest.Config {
	return c.cfg
}

func (c *clusterInfo) EnsureCrds(ctx context.Context) error {
	var files []string
	rx := regexp.MustCompile(`.+\.yaml`)
	for _, dir := range c.crdDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("error reading dir %s: %w", dir, err)
		}
		for _, en := range entries {
			if rx.Match([]byte(en.Name())) {
				files = append(files, path.Join(dir, en.Name()))
			}
		}
	}

	for _, fn := range files {
		b, err := os.ReadFile(fn)
		if err != nil {
			return fmt.Errorf("error reading CRD manifest %s: %w", fn, err)
		}

		decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(b), 1000)
		docCount := 0
		for {
			docCount++
			var rawObj runtime.RawExtension
			if err := decoder.Decode(&rawObj); err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("error deconding document #%d in %s: %w", docCount, fn, err)
			}
			if rawObj.Raw == nil {
				// empty yaml doc
				continue
			}

			obj, _, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
			if err != nil {
				return fmt.Errorf("error deconding rawObj into UnstructuredJSONScheme in document #%d in %s: %w", docCount, fn, err)
			}
			u, ok := obj.(*unstructured.Unstructured)
			if !ok {
				unstructuredData, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
				if err != nil {
					return fmt.Errorf("error converting obj to unstructured in document #%d in %s: %w", docCount, fn, err)
				}

				u = &unstructured.Unstructured{Object: unstructuredData}
			}

			uu := u.DeepCopy()
			err = c.client.Get(ctx, client.ObjectKeyFromObject(uu), uu)
			if client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("error getting obj %s of kind %s/%s to check if it exist: %w", uu.GetName(), uu.GetAPIVersion(), uu.GetKind(), err)
			}

			if err == nil {
				// this object exist
				continue
			}

			err = c.client.Create(ctx, u)
			if err != nil {
				return fmt.Errorf("error creating %s/%s/%s: %w", u.GetAPIVersion(), u.GetKind(), u.GetName(), err)
			}
		}
	}

	return nil
}
