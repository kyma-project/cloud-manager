package looper

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"io"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"path"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"time"
)

type Installer interface {
	Handle(ctx context.Context, provider string, skrCluster cluster.Cluster) error
}

var _ Installer = &installer{}

type installer struct {
	skrProvidersPath string
	logger           logr.Logger
}

func (i *installer) Handle(ctx context.Context, provider string, skrCluster cluster.Cluster) error {
	dir := path.Join(i.skrProvidersPath, provider)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("error listing SKR provider directory %s: %w", dir, err)
	}
	var files []string
	rx := regexp.MustCompile(".+\\.yaml")
	for _, en := range entries {
		if rx.Match([]byte(en.Name())) {
			files = append(files, path.Join(dir, en.Name()))
		}
	}

	docCount := 0
	for _, f := range files {
		cnt, err := i.applyFile(ctx, skrCluster, f)
		if err != nil {
			return fmt.Errorf("error installing SKR provider dependencies: %w", err)
		}
		docCount += cnt
	}

	if docCount > 0 {
		time.Sleep(2 * time.Second)
	}

	return nil
}

func (i *installer) applyFile(ctx context.Context, skrCluster cluster.Cluster, fn string) (int, error) {
	b, err := os.ReadFile(fn)
	if err != nil {
		return 0, fmt.Errorf("error reading SKR install manifest %s: %w", fn, err)
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
			return docCount - 1, fmt.Errorf("error deconding document #%d in %s: %w", docCount, fn, err)
		}
		if rawObj.Raw == nil {
			// empty yaml doc
			continue
		}

		obj, _, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		if err != nil {
			return docCount - 1, fmt.Errorf("error deconding rawObj into UnstructuredJSONScheme in document #%d in %s: %w", docCount, fn, err)
		}
		u, ok := obj.(*unstructured.Unstructured)
		if !ok {
			unstructuredData, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
			if err != nil {
				return docCount - 1, fmt.Errorf("error converting obj to unstructured in document #%d in %s: %w", docCount, fn, err)
			}

			u = &unstructured.Unstructured{Object: unstructuredData}
		}

		uu := u.DeepCopy()
		err = skrCluster.GetAPIReader().Get(ctx, client.ObjectKeyFromObject(uu), uu)
		if client.IgnoreNotFound(err) != nil {
			return docCount - 1, fmt.Errorf("error getting obj %s of kind %s/%s to check if it exist: %w", uu.GetName(), uu.GetAPIVersion(), uu.GetKind(), err)
		}

		if err == nil {
			i.logger.Info(fmt.Sprintf("Updating %s/%s/%s", u.GetAPIVersion(), u.GetKind(), u.GetName()))
			err = skrCluster.GetClient().Update(ctx, u)
		} else {
			i.logger.Info(fmt.Sprintf("Creating %s/%s/%s", u.GetAPIVersion(), u.GetKind(), u.GetName()))
			err = skrCluster.GetClient().Create(ctx, u)
		}

		if err != nil {
			return docCount - 1, fmt.Errorf("error applying %s/%s/%s: %w", u.GetAPIVersion(), u.GetKind(), u.GetName(), err)
		}
	}

	return docCount, nil
}
