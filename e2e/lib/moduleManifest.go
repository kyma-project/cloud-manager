package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-logr/logr"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ModuleManifestComponent downloads a Kyma module's operator manifest and default CR from its
// GitHub release (as configured in config.Modules), caches them under CONFIG_DIR/moduleManifests,
// and installs/deletes them on a target cluster.
type ModuleManifestComponent interface {
	GetModuleName() string
	Load() error
	Manifests() []*unstructured.Unstructured
	DefaultCR() *unstructured.Unstructured
	InstallManifests(kubeClient client.Client, timeout time.Duration) error
	DeleteManifests(kubeClient client.Client, timeout time.Duration) error
	InstallDefaultCR(kubeClient client.Client, timeout time.Duration) error
	DeleteDefaultCR(kubeClient client.Client, timeout time.Duration) error
}

const (
	moduleNamespace    = "kyma-system"
	moduleManifestsDir = "moduleManifests"
	readyStateReady    = "Ready"
	readyStateError    = "Error"
	pollInterval       = 5 * time.Second
)

// ModuleManifest constructs a ModuleManifestComponent for the given module, loading the e2e config
// from the ambient CONFIG_DIR/PROJECTROOT/CWD lookup.
func ModuleManifest(logger logr.Logger, moduleName string) ModuleManifestComponent {
	return ModuleManifestWithConfig(logger, e2econfig.LoadConfig(), moduleName)
}

// ModuleManifestWithConfig constructs a ModuleManifestComponent with an explicit config, useful in tests.
func ModuleManifestWithConfig(logger logr.Logger, config *e2econfig.ConfigType, moduleName string) ModuleManifestComponent {
	return &moduleManifest{
		logger:     logger.WithName("ModuleManifest").WithValues("module", moduleName),
		moduleName: moduleName,
		config:     config,
	}
}

type moduleManifest struct {
	logger     logr.Logger
	moduleName string
	config     *e2econfig.ConfigType

	moduleCfg       e2econfig.ModuleConfig
	resolvedVersion string
	manifests       []*unstructured.Unstructured
	defaultCR       *unstructured.Unstructured
}

var _ ModuleManifestComponent = &moduleManifest{}

func (m *moduleManifest) GetModuleName() string {
	return m.moduleName
}

func (m *moduleManifest) Manifests() []*unstructured.Unstructured {
	return m.manifests
}

func (m *moduleManifest) DefaultCR() *unstructured.Unstructured {
	return m.defaultCR
}

// Load resolves the release version, downloads (or reads from cache) the module manifest and default
// CR, parses them into unstructured objects, and validates the default CR contains exactly one object.
func (m *moduleManifest) Load() error {
	moduleCfg, ok := m.config.Modules[m.moduleName]
	if !ok {
		return fmt.Errorf("module %q is not defined in config.modules", m.moduleName)
	}
	m.moduleCfg = moduleCfg

	owner, repo, err := parseGithubRepository(moduleCfg.Repository)
	if err != nil {
		return err
	}

	m.resolvedVersion, err = m.resolveVersion(owner, repo)
	if err != nil {
		return err
	}

	cacheDir := filepath.Join(m.config.ConfigDir, moduleManifestsDir, m.moduleName, m.resolvedVersion)
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return fmt.Errorf("error creating module cache dir %s: %w", cacheDir, err)
	}

	manifestBytes, err := m.loadOrDownload(cacheDir, owner, repo, moduleCfg.Manifest)
	if err != nil {
		return fmt.Errorf("error loading module manifest: %w", err)
	}
	defaultBytes, err := m.loadOrDownload(cacheDir, owner, repo, moduleCfg.Default)
	if err != nil {
		return fmt.Errorf("error loading module default CR: %w", err)
	}

	manifests, err := util.YamlMultiDecodeToUnstructured(manifestBytes)
	if err != nil {
		return fmt.Errorf("error parsing module manifest yaml: %w", err)
	}
	m.manifests = util.FlattenLists(manifests)

	defaultCRs, err := util.YamlMultiDecodeToUnstructured(defaultBytes)
	if err != nil {
		return fmt.Errorf("error parsing module default CR yaml: %w", err)
	}
	defaultCRs = util.FlattenLists(defaultCRs)
	if len(defaultCRs) != 1 {
		return fmt.Errorf("module %q default CR manifest must contain exactly one resource, found %d", m.moduleName, len(defaultCRs))
	}
	m.defaultCR = defaultCRs[0]

	return nil
}

// resolveVersion returns moduleCfg.Version verbatim, or resolves "latest"/"" to the latest release tag.
func (m *moduleManifest) resolveVersion(owner, repo string) (string, error) {
	v := strings.TrimSpace(m.moduleCfg.Version)
	if v != "" && v != "latest" {
		return v, nil
	}
	return resolveLatestTag(owner, repo)
}

// loadOrDownload reads the asset from the cache dir, downloading it from the GitHub release if absent.
func (m *moduleManifest) loadOrDownload(cacheDir, owner, repo, assetFilename string) ([]byte, error) {
	if assetFilename == "" {
		return nil, fmt.Errorf("asset filename is empty for module %q", m.moduleName)
	}
	dest := filepath.Join(cacheDir, assetFilename)
	if _, err := os.Stat(dest); err == nil {
		return os.ReadFile(dest)
	}

	url := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", owner, repo, m.resolvedVersion, assetFilename)
	if err := downloadToFile(url, dest); err != nil {
		return nil, err
	}
	return os.ReadFile(dest)
}

// InstallManifests installs the module manifest objects in dependency order, waiting for readiness of
// CRDs (Established) and Deployments (Available) before proceeding. Other kinds (RBAC, Services,
// ConfigMaps, Secrets, webhooks, ...) have no meaningful readiness condition and are only created.
func (m *moduleManifest) InstallManifests(kubeClient client.Client, timeout time.Duration) error {
	ctx := context.Background()

	m.logger.Info("Installing module manifests", "version", m.resolvedVersion, "resourceCount", len(m.manifests))
	for _, obj := range m.manifests {
		m.logger.Info("Manifest resource", "apiVersion", obj.GetAPIVersion(), "kind", obj.GetKind(), "name", obj.GetName())
	}

	// The module resources are namespaced into kyma-system, but the Namespace object itself is
	// usually not part of the manifest - ensure it exists first.
	if err := ensureNamespace(ctx, kubeClient, moduleNamespace); err != nil {
		return err
	}

	groups := groupByKind(m.manifests)

	// 1. CRDs - apply and wait for Established.
	crds := groups["CustomResourceDefinition"]
	if err := applyObjects(ctx, kubeClient, crds); err != nil {
		return err
	}
	for _, crd := range crds {
		if err := waitCrdEstablished(ctx, m.logger, kubeClient, crd.GetName(), timeout); err != nil {
			return err
		}
	}

	// 2. RBAC - apply, no readiness.
	rbacKinds := []string{"ServiceAccount", "Role", "RoleBinding", "ClusterRole", "ClusterRoleBinding"}
	if err := applyObjects(ctx, kubeClient, collectKinds(groups, rbacKinds...)); err != nil {
		return err
	}

	// 3. Services - apply, no readiness.
	if err := applyObjects(ctx, kubeClient, groups["Service"]); err != nil {
		return err
	}

	// 4. Everything else except Deployments - apply, no readiness.
	handled := map[string]bool{
		"CustomResourceDefinition": true,
		"Service":                  true,
		"Deployment":               true,
	}
	for _, k := range rbacKinds {
		handled[k] = true
	}
	var rest []*unstructured.Unstructured
	for kind, objs := range groups {
		if handled[kind] {
			continue
		}
		rest = append(rest, objs...)
	}
	if err := applyObjects(ctx, kubeClient, rest); err != nil {
		return err
	}

	// 5. Deployments - apply and wait for Available.
	deployments := groups["Deployment"]
	if err := applyObjects(ctx, kubeClient, deployments); err != nil {
		return err
	}
	for _, dep := range deployments {
		if err := waitDeploymentAvailable(ctx, m.logger, kubeClient, dep.GetNamespace(), dep.GetName(), timeout); err != nil {
			return err
		}
	}

	return nil
}

// InstallDefaultCR applies the default module CR and waits for its status.state to reach Ready, or
// returns an error if it reaches Error.
func (m *moduleManifest) InstallDefaultCR(kubeClient client.Client, timeout time.Duration) error {
	ctx := context.Background()

	m.logger.Info("Installing default CR", "apiVersion", m.defaultCR.GetAPIVersion(), "kind", m.defaultCR.GetKind(), "name", m.defaultCR.GetName())
	if err := applyObjects(ctx, kubeClient, []*unstructured.Unstructured{m.defaultCR}); err != nil {
		return err
	}

	key := client.ObjectKeyFromObject(m.defaultCR)
	gvk := m.defaultCR.GroupVersionKind()

	m.logger.Info("Waiting for default CR to become ready", "name", m.defaultCR.GetName())
	err := wait.PollUntilContextTimeout(ctx, pollInterval, timeout, true, func(ctx context.Context) (bool, error) {
		current := &unstructured.Unstructured{}
		current.SetGroupVersionKind(gvk)
		if err := kubeClient.Get(ctx, key, current); err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			return false, fmt.Errorf("error getting default CR %s while waiting ready: %w", key, err)
		}
		state, _, err := unstructured.NestedString(current.Object, "status", "state")
		if err != nil {
			return false, fmt.Errorf("error reading default CR %s status.state: %w", key, err)
		}
		switch state {
		case readyStateReady:
			return true, nil
		case readyStateError:
			return false, fmt.Errorf("default CR %s is in Error state", key)
		default:
			return false, nil
		}
	})
	if err != nil {
		return err
	}
	m.logger.Info("Default CR is ready", "name", m.defaultCR.GetName())
	return nil
}

// DeleteDefaultCR deletes the default module CR and waits for it to disappear from the cluster.
func (m *moduleManifest) DeleteDefaultCR(kubeClient client.Client, timeout time.Duration) error {
	ctx := context.Background()
	m.logger.Info("Deleting default CR", "apiVersion", m.defaultCR.GetAPIVersion(), "kind", m.defaultCR.GetKind(), "name", m.defaultCR.GetName())
	return deleteAndWaitGone(ctx, m.logger, kubeClient, []*unstructured.Unstructured{m.defaultCR}, timeout)
}

// DeleteManifests deletes all module manifest objects and waits for them to disappear. It does not
// rely on the module having cleaned up after its default CR was deleted.
func (m *moduleManifest) DeleteManifests(kubeClient client.Client, timeout time.Duration) error {
	ctx := context.Background()
	m.logger.Info("Deleting module manifests", "resourceCount", len(m.manifests))
	for _, obj := range m.manifests {
		m.logger.Info("Deleting manifest resource", "apiVersion", obj.GetAPIVersion(), "kind", obj.GetKind(), "name", obj.GetName())
	}
	return deleteAndWaitGone(ctx, m.logger, kubeClient, m.manifests, timeout)
}

// -- helpers ---------------------------------------------------------------------------------------

func parseGithubRepository(repository string) (owner, repo string, err error) {
	trimmed := strings.TrimPrefix(strings.TrimSpace(repository), "github.com/")
	parts := strings.Split(trimmed, "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid github repository %q, expected github.com/<owner>/<repo>", repository)
	}
	return parts[0], parts[1], nil
}

func resolveLatestTag(owner, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	resp, err := http.Get(url) //nolint:gosec // url built from trusted config values
	if err != nil {
		return "", fmt.Errorf("error resolving latest release for %s/%s: %w", owner, repo, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error resolving latest release for %s/%s: unexpected status %s", owner, repo, resp.Status)
	}
	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("error decoding latest release for %s/%s: %w", owner, repo, err)
	}
	if payload.TagName == "" {
		return "", fmt.Errorf("latest release for %s/%s has empty tag_name", owner, repo)
	}
	return payload.TagName, nil
}

func downloadToFile(url, dest string) error {
	resp, err := http.Get(url) //nolint:gosec // url built from trusted config values
	if err != nil {
		return fmt.Errorf("error downloading %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error downloading %s: unexpected status %s", url, resp.Status)
	}

	tmp := dest + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("error creating %s: %w", tmp, err)
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("error writing %s: %w", tmp, err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("error closing %s: %w", tmp, err)
	}
	// Rename only after a fully successful download so a partial file is never cached.
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("error moving %s to %s: %w", tmp, dest, err)
	}
	return nil
}

func groupByKind(objs []*unstructured.Unstructured) map[string][]*unstructured.Unstructured {
	result := map[string][]*unstructured.Unstructured{}
	for _, obj := range objs {
		kind := obj.GetKind()
		result[kind] = append(result[kind], obj)
	}
	return result
}

func collectKinds(groups map[string][]*unstructured.Unstructured, kinds ...string) []*unstructured.Unstructured {
	var result []*unstructured.Unstructured
	for _, k := range kinds {
		result = append(result, groups[k]...)
	}
	return result
}

func ensureNamespace(ctx context.Context, kubeClient client.Client, name string) error {
	ns := &corev1.Namespace{}
	err := kubeClient.Get(ctx, types.NamespacedName{Name: name}, ns)
	if err == nil {
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return fmt.Errorf("error getting namespace %s: %w", name, err)
	}
	ns = &corev1.Namespace{Name: name}
	if err := kubeClient.Create(ctx, ns); err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("error creating namespace %s: %w", name, err)
	}
	return nil
}

// applyObjects performs a create-or-update for each object, preserving the existing object's metadata
// on update (mirrors pkg/util.Apply, but used per ordered batch).
func applyObjects(ctx context.Context, kubeClient client.Client, objects []*unstructured.Unstructured) error {
	for _, desired := range objects {
		key := client.ObjectKeyFromObject(desired)
		existing := desired.DeepCopy()
		err := kubeClient.Get(ctx, key, existing)
		if client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error getting %s %s: %w", desired.GetKind(), key, err)
		}
		if err != nil {
			if err := kubeClient.Create(ctx, desired); err != nil {
				return fmt.Errorf("error creating %s %s: %w", desired.GetKind(), key, err)
			}
			continue
		}
		metadata, _, err := unstructured.NestedMap(existing.Object, "metadata")
		if err != nil {
			return fmt.Errorf("error reading existing %s %s metadata: %w", desired.GetKind(), key, err)
		}
		if err := unstructured.SetNestedField(desired.Object, metadata, "metadata"); err != nil {
			return fmt.Errorf("error setting %s %s metadata: %w", desired.GetKind(), key, err)
		}
		if err := kubeClient.Update(ctx, desired); err != nil {
			return fmt.Errorf("error updating %s %s: %w", desired.GetKind(), key, err)
		}
	}
	return nil
}

func deleteAndWaitGone(ctx context.Context, logger logr.Logger, kubeClient client.Client, objects []*unstructured.Unstructured, timeout time.Duration) error {
	for _, obj := range objects {
		if err := kubeClient.Delete(ctx, obj); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting %s %s: %w", obj.GetKind(), client.ObjectKeyFromObject(obj), err)
		}
	}
	logger.Info("Waiting for deleted resources to disappear", "resourceCount", len(objects))
	err := wait.PollUntilContextTimeout(ctx, pollInterval, timeout, true, func(ctx context.Context) (bool, error) {
		for _, obj := range objects {
			current := &unstructured.Unstructured{}
			current.SetGroupVersionKind(obj.GroupVersionKind())
			err := kubeClient.Get(ctx, client.ObjectKeyFromObject(obj), current)
			if err == nil {
				return false, nil
			}
			if !apierrors.IsNotFound(err) {
				return false, fmt.Errorf("error getting %s %s while waiting for deletion: %w", obj.GetKind(), client.ObjectKeyFromObject(obj), err)
			}
		}
		return true, nil
	})
	if err != nil {
		return err
	}
	logger.Info("Deleted resources are gone", "resourceCount", len(objects))
	return nil
}

func waitCrdEstablished(ctx context.Context, logger logr.Logger, kubeClient client.Client, name string, timeout time.Duration) error {
	logger.Info("Waiting for CRD to be established", "crd", name)
	err := wait.PollUntilContextTimeout(ctx, pollInterval, timeout, true, func(ctx context.Context) (bool, error) {
		crd := &apiextensionsv1.CustomResourceDefinition{}
		if err := kubeClient.Get(ctx, types.NamespacedName{Name: name}, crd); err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			return false, fmt.Errorf("error getting CRD %s while waiting established: %w", name, err)
		}
		for _, cond := range crd.Status.Conditions {
			if cond.Type == apiextensionsv1.Established && cond.Status == apiextensionsv1.ConditionTrue {
				return true, nil
			}
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	logger.Info("CRD is established", "crd", name)
	return nil
}

func waitDeploymentAvailable(ctx context.Context, logger logr.Logger, kubeClient client.Client, namespace, name string, timeout time.Duration) error {
	logger.Info("Waiting for Deployment to be ready", "namespace", namespace, "deployment", name)
	err := wait.PollUntilContextTimeout(ctx, pollInterval, timeout, true, func(ctx context.Context) (bool, error) {
		dep := &appsv1.Deployment{}
		if err := kubeClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, dep); err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			return false, fmt.Errorf("error getting Deployment %s/%s while waiting available: %w", namespace, name, err)
		}
		desiredReplicas := int32(1)
		if dep.Spec.Replicas != nil {
			desiredReplicas = *dep.Spec.Replicas
		}
		if dep.Status.AvailableReplicas != desiredReplicas {
			return false, nil
		}
		for _, cond := range dep.Status.Conditions {
			if cond.Type == appsv1.DeploymentAvailable && cond.Status == corev1.ConditionTrue {
				return true, nil
			}
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	logger.Info("Deployment is ready", "namespace", namespace, "deployment", name)
	return nil
}
