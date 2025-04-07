package looper

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	"io"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"time"
)

type Cluster interface {
	GetScheme() *runtime.Scheme
	GetAPIReader() client.Reader
	GetClient() client.Client
}

var _ Cluster = cluster.Cluster(nil)

func ToCluster(clstr cluster.Cluster) Cluster {
	return NewCluster(clstr.GetScheme(), clstr.GetAPIReader(), clstr.GetClient())
}

func NewCluster(scheme *runtime.Scheme, apiReader client.Reader, clnt client.Client) Cluster {
	return &clusterImpl{
		scheme:    scheme,
		apiReader: apiReader,
		client:    clnt,
	}
}

type clusterImpl struct {
	scheme    *runtime.Scheme
	apiReader client.Reader
	client    client.Client
}

func (c *clusterImpl) GetScheme() *runtime.Scheme {
	return c.scheme
}

func (c *clusterImpl) GetAPIReader() client.Reader {
	return c.apiReader
}

func (c *clusterImpl) GetClient() client.Client {
	return c.client
}

type Installer interface {
	Handle(ctx context.Context, provider string, skrCluster Cluster) error
}

var _ Installer = &installer{}

type installer struct {
	skrStatus        *SkrStatus
	skrProvidersPath string
	logger           logr.Logger
}

func (i *installer) Handle(ctx context.Context, provider string, skrCluster Cluster) error {
	ctx = feature.ContextBuilderFromCtx(ctx).
		Provider(provider).
		Build(ctx)
	dir := path.Join(i.skrProvidersPath, provider)
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return fmt.Errorf("installer directory %s does not exist: %w", dir, err)
	}
	if err != nil {
		return fmt.Errorf("error checking provider dir %s: %w", dir, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("error listing SKR provider directory %s: %w", dir, err)
	}
	var files []string
	rx := regexp.MustCompile(`.+\.yaml`)
	for _, en := range entries {
		if rx.Match([]byte(en.Name())) {
			files = append(files, path.Join(dir, en.Name()))
		}
	}

	docCount := 0
	for _, f := range files {
		cnt, err := i.applyFile(ctx, skrCluster, f, provider)
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

func (i *installer) applyFile(ctx context.Context, skrCluster Cluster, fn string, provider string) (int, error) {
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
		desired, ok := obj.(*unstructured.Unstructured)
		if !ok {
			unstructuredData, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
			if err != nil {
				return docCount - 1, fmt.Errorf("error converting obj to unstructured in document #%d in %s: %w", docCount, fn, err)
			}

			desired = &unstructured.Unstructured{Object: unstructuredData}
		}

		objCtx := feature.ContextBuilderFromCtx(ctx).
			KindsFromObject(desired, skrCluster.GetScheme()).
			FeatureFromObject(desired, skrCluster.GetScheme()).
			Build(ctx)

		handle := i.skrStatus.Handle(objCtx, "InstallerManifest")
		handle.WithObj(desired)
		handle.WithFilename(filepath.Base(fn))

		logger := feature.DecorateLogger(objCtx, i.logger).
			WithValues(
				"objName", desired.GetName(),
				"objNamespace", desired.GetNamespace(),
				"manifestFile", filepath.Base(fn),
			)

		if !common.ObjSupportsProvider(desired, skrCluster.GetScheme(), provider) {
			handle.NotSupportedByProvider()
			//logger.Info("Object Kind does not support this provider")
			continue
		}

		existing := desired.DeepCopy()
		err = skrCluster.GetAPIReader().Get(ctx, client.ObjectKeyFromObject(existing), existing)
		if client.IgnoreNotFound(err) != nil {
			return docCount - 1, fmt.Errorf("error getting obj %s of kind %s/%s to check if it exist: %w", existing.GetName(), existing.GetAPIVersion(), existing.GetKind(), err)
		}

		if err == nil {
			// It already exists
			// Even if desired belongs to disabled API, since it's already applied we must update it
			// so feature flag will be checked in create branch only

			desiredVersion := i.getVersion(desired)
			existingVersion := i.getVersion(existing)
			if desiredVersion == existingVersion {
				handle.AlreadyExistsWithSameVersion(desiredVersion)
				continue
			}

			err = i.copyForUpdate(desired, existing)
			if err != nil {
				handle.SpecCopyError(err)
				// leave this log since it indicates a developer logical error that happens rarely and only if copy spec code is invalid
				logger.Error(err, fmt.Sprintf("Error copying spec for %s/%s/%s before update", desired.GetAPIVersion(), desired.GetKind(), desired.GetName()))
				continue
			}
			handle.Updating(existingVersion, desiredVersion)
			//logger.Info(fmt.Sprintf("Updating %s/%s/%s from version %s to %s", desired.GetAPIVersion(), desired.GetKind(), desired.GetName(), existingVersion, desiredVersion))
			err = skrCluster.GetClient().Update(ctx, existing)
		} else {
			err = nil // clear the not found error, so we only return Create error if any, and not this not found
			if feature.ApiDisabled.Value(objCtx) {
				handle.ApiDisabled()
				//logger.Info(fmt.Sprintf("Skipping installation of disabled API of %s/%s/%s", desired.GetAPIVersion(), desired.GetKind(), desired.GetName()))
			} else {
				handle.Creating()
				//logger.Info(fmt.Sprintf("Creating %s/%s/%s", desired.GetAPIVersion(), desired.GetKind(), desired.GetName()))
				err = skrCluster.GetClient().Create(ctx, desired)
			}
		}

		if err != nil {
			handle.Error(err)
			return docCount - 1, fmt.Errorf("error applying %s/%s/%s: %w", desired.GetAPIVersion(), desired.GetKind(), desired.GetName(), err)
		}

		handle.Success()
	}

	return docCount, nil
}

func (i *installer) getVersion(u *unstructured.Unstructured) string {
	if u.GetAnnotations() == nil {
		return ""
	}
	result, ok := u.GetAnnotations()["cloud-resources.kyma-project.io/version"]
	if !ok {
		return ""
	}
	return result
}

func (i *installer) copyForUpdate(from, to *unstructured.Unstructured) (err error) {
	if err = i.copyField(from, to, "spec"); err != nil {
		return fmt.Errorf("error copying spec field: %w", err)
	}
	if err = i.copyField(from, to, "data"); err != nil {
		return fmt.Errorf("error copying data field: %w", err)
	}
	if err = i.copyField(from, to, "metadata", "labels"); err != nil {
		return fmt.Errorf("error copying labels field: %w", err)
	}
	if err = i.copyField(from, to, "metadata", "annotations"); err != nil {
		return fmt.Errorf("error copying labels field: %w", err)
	}
	return nil
}

func (i *installer) copyField(from, to *unstructured.Unstructured, fields ...string) error {
	fromSpec, exists, err := unstructured.NestedMap(from.Object, fields...)
	if !exists {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error getting fields from source: %w", err)
	}
	err = unstructured.SetNestedMap(to.Object, fromSpec, fields...)
	if err != nil {
		return fmt.Errorf("error setting fields to destination: %w", err)
	}
	return nil
}
