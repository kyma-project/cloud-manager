package keb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	gardenerapicore "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	e2elib "github.com/kyma-project/cloud-manager/e2e/lib"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/clock"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// KEB =============================================

type InstanceLister interface {
	List(ctx context.Context, opts ...ListOption) ([]InstanceDetails, error)
}

type Keb interface {
	SkrManagerFactory

	Config() *e2econfig.ConfigType

	KcpClient() client.Client
	GardenClient() client.Client

	CreateInstance(ctx context.Context, opts ...CreateOption) (InstanceDetails, error)
	GetInstance(ctx context.Context, runtimeID string) (*InstanceDetails, error)
	List(ctx context.Context, opts ...ListOption) ([]InstanceDetails, error)
	DeleteInstance(ctx context.Context, opts ...DeleteOption) error

	GetInstanceKubeconfig(ctx context.Context, runtimeID string) ([]byte, time.Time, error)
	CreateInstanceClient(ctx context.Context, runtimeID string) (client.Client, error)
	RenewInstanceKubeconfig(ctx context.Context, runtimeID string) error
}

var _ InstanceLister = (Keb)(nil)

func NewKeb(kcpClient client.Client, gardenClient client.Client, managerFactory SkrManagerFactory, cpl e2elib.CloudProfileLoader, skrKubeconfigProvider e2elib.SkrKubeconfigProvider, config *e2econfig.ConfigType) Keb {
	return &defaultKeb{
		SkrManagerFactory:     managerFactory,
		kcpClient:             kcpClient,
		gardenClient:          gardenClient,
		cpl:                   cpl,
		skrKubeconfigProvider: skrKubeconfigProvider,
		config:                config,
		clock:                 &clock.RealClock{},
	}
}

// CreateOption =========================================================

type CreateOption interface {
	ApplyOnCreate(*createOptions)
}

type createOptions struct {
	alias         string
	globalAccount string
	subAccount    string
	provider      cloudcontrolv1beta1.ProviderType
	region        string
	nodesRange    string
	podsRange     string
	servicesRange string

	shootCreatedTimeout  time.Duration
	shootCreatedInterval time.Duration
}

func (in *createOptions) validate() error {
	var result error
	if in.alias == "" {
		result = multierror.Append(result, errors.New("alias required"))
	}
	if in.provider == "" {
		result = multierror.Append(result, errors.New("provider required"))
	}

	if in.globalAccount == "" {
		result = multierror.Append(result, errors.New("global account required"))
	}
	if in.subAccount == "" {
		result = multierror.Append(result, errors.New("subAccount required"))
	}
	if in.region == "" {
		result = multierror.Append(result, errors.New("region required"))
	}

	if in.shootCreatedTimeout > 0 && in.shootCreatedInterval == 0 {
		result = multierror.Append(result, errors.New("if timeout is specified then interval must be also specified"))
	}

	return result
}

// defaultCreateOptions is not a var as others, but func so it always evaluates to different uuids that are used
// as defaults for globalAccount and subAccount
func defaultCreateOptions() []CreateOption {
	return []CreateOption{
		WithGlobalAccount(uuid.NewString()),
		WithSubAccount(uuid.NewString()),
		WithNodesRange("10.250.0.0/16"),
		WithPodsRange("10.96.0.0/13"),
		WithServicesRange("10.104.0.0/13"),
		WithInterval(time.Second),
		WithTimeout(5 * time.Minute), // Wait up to 5 minutes for shoot object to be created in garden cluster
	}
}

// ListOption ============================================================

type ListOption interface {
	ApplyOnList(*listOptions)
}

type listOptions struct {
	alias         string
	runtimeId     string
	globalAccount string
	subAccount    string
	provider      cloudcontrolv1beta1.ProviderType
}

func (o listOptions) MatchingLabels() client.MatchingLabels {
	result := client.MatchingLabels{}
	if o.alias != "" {
		result[e2elib.AliasLabel] = o.alias
	}
	if o.globalAccount != "" {
		result[cloudcontrolv1beta1.LabelScopeGlobalAccountId] = o.globalAccount
	}
	if o.subAccount != "" {
		result[cloudcontrolv1beta1.LabelScopeSubaccountId] = o.subAccount
	}
	if o.provider != "" {
		result[cloudcontrolv1beta1.LabelScopeBrokerPlanName] = strings.ToLower(string(o.provider))
	}
	return result
}

// there's no
// * validate() since the filter may be empty which means return all instances
// * default values since default is list all instances

// DeleteOption ============================================================

type DeleteOption interface {
	ApplyOnDelete(*deleteOptions)
}

type deleteOptions struct {
	runtimeId                      string
	alias                          string
	shootMarkedForDeletionTimeout  time.Duration
	shootMarkedForDeletionInterval time.Duration
}

func (in deleteOptions) validate() error {
	var result error
	if in.runtimeId == "" && in.alias == "" {
		result = multierror.Append(result, errors.New("runtimeId or alias is required"))
	}
	if in.shootMarkedForDeletionTimeout > 0 && in.shootMarkedForDeletionInterval == 0 {
		result = multierror.Append(result, errors.New("if timeout is specified then internal must also be specified"))
	}
	return result
}

var defaultDeleteOptions = []DeleteOption{
	WithTimeout(30 * time.Second),
	WithInterval(2 * time.Second),
}

// WithAlias =====================================================

// InstanceDetails =============================================

type InstanceDetails struct {
	Alias         string                           `json:"alias" yaml:"alias"`
	GlobalAccount string                           `json:"globalAccount" yaml:"globalAccount"`
	SubAccount    string                           `json:"subAccount" yaml:"subAccount"`
	Provider      cloudcontrolv1beta1.ProviderType `json:"provider" yaml:"provider"`
	Region        string                           `json:"region" yaml:"region"`

	ProvisioningCompleted bool `json:"provisioningCompleted" yaml:"provisioningCompleted"`

	RuntimeID string `json:"runtimeID" yaml:"runtimeID"`
	ShootName string `json:"shootName" yaml:"shootName"`

	// State has value of the runtime.status.state
	State string `json:"state" yaml:"state"`
	// Message has value of message of the Condition type Error
	Message string `json:"message" yaml:"message"`

	BeingDeleted bool `json:"beingDeleted" yaml:"beingDeleted"`

	Ignored bool `json:"ignored" yaml:"ignored"`
}

func (id InstanceDetails) AddLoggerValues(log logr.Logger) logr.Logger {
	return log.WithValues(
		"alias", id.Alias,
		"runtimeId", id.RuntimeID,
		"shootName", id.ShootName,
		"provider", id.Provider,
		"region", id.Region,
		"globalAccount", id.GlobalAccount,
		"subAccount", id.SubAccount,
	)
}

// keb implementation =================

var _ Keb = &defaultKeb{}

type defaultKeb struct {
	SkrManagerFactory

	kcpClient             client.Client
	gardenClient          client.Client
	cpl                   e2elib.CloudProfileLoader
	skrKubeconfigProvider e2elib.SkrKubeconfigProvider
	config                *e2econfig.ConfigType
	clock                 clock.Clock
}

func RuntimeToInstanceDetails(rt *infrastructuremanagerv1.Runtime) InstanceDetails {
	id := InstanceDetails{
		Alias:                 rt.Labels[e2elib.AliasLabel],
		GlobalAccount:         rt.Labels[cloudcontrolv1beta1.LabelScopeGlobalAccountId],
		SubAccount:            rt.Labels[cloudcontrolv1beta1.LabelScopeSubaccountId],
		Provider:              cloudcontrolv1beta1.ProviderType(rt.Spec.Shoot.Provider.Type),
		Region:                rt.Spec.Shoot.Region,
		ProvisioningCompleted: rt.Status.ProvisioningCompleted,
		RuntimeID:             rt.Name,
		ShootName:             rt.Spec.Shoot.Name,
		State:                 string(rt.Status.State),
		BeingDeleted:          rt.DeletionTimestamp != nil,
		Ignored:               rt.Labels[e2elib.DoNotReconcile] != "",
	}
	errCond := meta.FindStatusCondition(rt.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	if errCond != nil && errCond.Status == metav1.ConditionTrue {
		id.Message = errCond.Message
	}
	return id
}

func (k *defaultKeb) Config() *e2econfig.ConfigType {
	return k.config
}

func (k *defaultKeb) KcpClient() client.Client {
	return k.kcpClient
}

func (k *defaultKeb) GardenClient() client.Client {
	return k.gardenClient
}

func (k *defaultKeb) GetInstance(ctx context.Context, runtimeID string) (*InstanceDetails, error) {
	rt := &infrastructuremanagerv1.Runtime{}
	err := k.kcpClient.Get(ctx, client.ObjectKey{Namespace: k.config.KcpNamespace, Name: runtimeID}, rt)
	if client.IgnoreNotFound(err) != nil {
		return nil, fmt.Errorf("error getting runtime %q: %w", runtimeID, err)
	}
	if err != nil {
		return nil, nil
	}
	return ptr.To(RuntimeToInstanceDetails(rt)), nil
}

func (k *defaultKeb) CreateInstance(ctx context.Context, opts ...CreateOption) (InstanceDetails, error) {
	if k.config.ShootPrefix == "" {
		return InstanceDetails{}, fmt.Errorf("required config shootPrefix not set")
	}
	options := &createOptions{}
	for _, o := range append(append([]CreateOption{}, defaultCreateOptions()...), opts...) {
		o.ApplyOnCreate(options)
	}
	if err := options.validate(); err != nil {
		return InstanceDetails{}, err
	}

	arr, err := k.List(ctx, WithAlias(options.alias))
	if err != nil {
		return InstanceDetails{}, fmt.Errorf("error listing instances to check if duplucate alias %q exists: %w", options.alias, err)
	}
	if len(arr) != 0 {
		return InstanceDetails{}, fmt.Errorf("instance with alias %q already exists", options.alias)
	}

	if k.config.GardenNamespace == "" {
		return InstanceDetails{}, fmt.Errorf("config garden namespace not set")
	}
	if k.config.KcpNamespace == "" {
		return InstanceDetails{}, fmt.Errorf("config kcp namespace not set")
	}
	subscription := k.config.Subscriptions.GetDefaultForProvider(options.provider)
	if subscription == nil {
		return InstanceDetails{}, fmt.Errorf("subscription not found for provider %q", options.provider)
	}
	cpr, err := k.cpl.Load(ctx)
	if err != nil {
		return InstanceDetails{}, fmt.Errorf("error loading cloud profiles: %w", err)
	}
	if options.region == "" {
		options.region = e2elib.DefaultRegions[options.provider]
	}
	rtBuilder := e2elib.NewRuntimeBuilder(cpr, k.config).
		WithAlias(options.alias).
		WithProvider(options.provider, options.region).
		WithSecretBindingName(subscription.Name).
		WithGlobalAccount(options.globalAccount).
		WithSubAccount(options.subAccount)
	if err := rtBuilder.Validate(); err != nil {
		return InstanceDetails{}, fmt.Errorf("invalid create instance request: %w", err)
	}
	rt := rtBuilder.Build()

	err = k.kcpClient.Create(ctx, rt)
	if err != nil {
		return InstanceDetails{}, fmt.Errorf("error creating runtime: %w", err)
	}

	id := InstanceDetails{
		Alias:                 options.alias,
		GlobalAccount:         options.globalAccount,
		SubAccount:            options.subAccount,
		Provider:              options.provider,
		Region:                rt.Spec.Shoot.Region,
		ProvisioningCompleted: false,
		RuntimeID:             rt.Name,
		ShootName:             rt.Spec.Shoot.Name,
	}

	time.Sleep(time.Second)

	if options.shootCreatedTimeout > 0 && k.gardenClient != nil {
		// wait for shoot to get created, so afterward this func returns, even if sim is stopped the gardener will
		// keep creating the cluster instance
		logger := ctrl.Log.WithName("keb")
		logger.
			WithValues(
				"shoot", rt.Spec.Shoot.Name,
				"runtimeID", rt.Name,
				"gardenNamespace", k.config.GardenNamespace,
			).
			Info("Waiting for shoot to get created")
		err = wait.PollUntilContextTimeout(ctx, options.shootCreatedInterval, options.shootCreatedTimeout, false, func(ctx context.Context) (bool, error) {
			shoot := &gardenerapicore.Shoot{}
			err := k.gardenClient.Get(ctx, types.NamespacedName{
				Namespace: k.config.GardenNamespace,
				Name:      rt.Spec.Shoot.Name,
			}, shoot)
			if err == nil {
				return true, nil
			}
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		})
		if err != nil {
			return id, fmt.Errorf("error checking if shoot is created: %w", err)
		}
	}

	return id, err
}

func (k *defaultKeb) List(ctx context.Context, options ...ListOption) ([]InstanceDetails, error) {
	opts := &listOptions{}
	for _, o := range options {
		o.ApplyOnList(opts)
	}

	var loArr []client.ListOption
	if ml := opts.MatchingLabels(); len(ml) > 0 {
		loArr = append(loArr, ml)
	}

	rtList := &infrastructuremanagerv1.RuntimeList{}
	if opts.runtimeId != "" {
		rt := &infrastructuremanagerv1.Runtime{}
		if err := k.kcpClient.Get(ctx, client.ObjectKey{Namespace: k.config.KcpNamespace, Name: opts.runtimeId}, rt); err != nil {
			return nil, fmt.Errorf("runtime %s not found: %w", opts.runtimeId, err)
		}
		rtList.Items = append(rtList.Items, *rt)
	} else {
		if err := k.kcpClient.List(ctx, rtList, loArr...); err != nil {
			return nil, fmt.Errorf("error listing runtimes: %w", err)
		}
	}

	var results []InstanceDetails
	for _, rt := range rtList.Items {
		results = append(results, RuntimeToInstanceDetails(&rt))
	}
	return results, nil
}

func (k *defaultKeb) DeleteInstance(ctx context.Context, opts ...DeleteOption) error {
	options := &deleteOptions{}
	for _, o := range append(append([]DeleteOption{}, defaultDeleteOptions...), opts...) {
		o.ApplyOnDelete(options)
	}
	if err := options.validate(); err != nil {
		return err
	}

	rt := &infrastructuremanagerv1.Runtime{}
	err := k.kcpClient.Get(ctx, client.ObjectKey{Namespace: k.config.KcpNamespace, Name: options.runtimeId}, rt)
	if client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("error getting runtime %q: %w", options.runtimeId, err)
	}
	if apierrors.IsNotFound(err) {
		return nil
	}
	err = k.kcpClient.Delete(ctx, rt)
	if err != nil {
		return fmt.Errorf("error deleting runtime %q: %w", options.runtimeId, err)
	}

	if rt.Spec.Shoot.Name != "" && options.shootMarkedForDeletionTimeout > 0 && k.gardenClient != nil {
		// wait until shoot is marked for deletion
		err := wait.PollUntilContextTimeout(ctx, options.shootMarkedForDeletionInterval, options.shootMarkedForDeletionTimeout, false, func(ctx context.Context) (bool, error) {
			shoot := &gardenerapicore.Shoot{}
			err := k.gardenClient.Get(ctx, types.NamespacedName{
				Namespace: k.config.GardenNamespace,
				Name:      rt.Spec.Shoot.Name,
			}, shoot)
			if client.IgnoreNotFound(err) != nil {
				return false, err
			}
			if err != nil {
				// not found
				return true, nil
			}
			if shoot.DeletionTimestamp != nil {
				return true, nil
			}
			return false, nil
		})
		if err != nil {
			return fmt.Errorf("error waiting for shoot to get deletion timestamp: %w", err)
		}
	}

	return nil
}

func (k *defaultKeb) GetInstanceKubeconfig(ctx context.Context, runtimeID string) ([]byte, time.Time, error) {
	t := time.Unix(0, 0)
	rt := &infrastructuremanagerv1.Runtime{}
	err := k.kcpClient.Get(ctx, client.ObjectKey{Namespace: k.config.KcpNamespace, Name: runtimeID}, rt)
	if err != nil {
		return nil, t, fmt.Errorf("error getting Runtime %q: %w", runtimeID, err)
	}

	gc := &infrastructuremanagerv1.GardenerCluster{}
	err = k.kcpClient.Get(ctx, client.ObjectKeyFromObject(rt), gc)
	if err != nil {
		return nil, t, fmt.Errorf("error getting GardenerCluster %q: %w", runtimeID, err)
	}

	hasExpired, _ := e2elib.IsGardenerClusterSyncNeeded(gc, k.clock)
	if hasExpired {
		return nil, t, e2elib.ErrGardenerClusterCredentialsExpired
	}

	ns := gc.Spec.Kubeconfig.Secret.Namespace
	if ns == "" {
		ns = rt.Namespace
	}
	secret := &corev1.Secret{}
	err = k.kcpClient.Get(ctx, types.NamespacedName{
		Namespace: ns,
		Name:      gc.Spec.Kubeconfig.Secret.Name,
	}, secret)
	if err != nil {
		return nil, t, fmt.Errorf("error getting SKR credentials secret: %w", err)
	}

	data, ok := secret.Data[gc.Spec.Kubeconfig.Secret.Key]
	if !ok {
		return nil, t, fmt.Errorf("skr credential secret does not have key %q as GardenerCluster defines", gc.Spec.Kubeconfig.Secret.Key)
	}

	if gc.Annotations != nil {
		tt, err := time.Parse(time.RFC3339, gc.Annotations[e2elib.ExpiresAtAnnotation])
		if err == nil {
			t = tt
		}
	}

	return data, t, nil
}

func (k *defaultKeb) CreateInstanceClient(ctx context.Context, runtimeID string) (client.Client, error) {
	b, _, err := k.GetInstanceKubeconfig(ctx, runtimeID)
	if err != nil {
		return nil, fmt.Errorf("error getting instance kubeconfig: %w", err)
	}

	cc, err := clientcmd.NewClientConfigFromBytes(b)
	if err != nil {
		return nil, fmt.Errorf("error creating client config from kubeconfig: %w", err)
	}
	restConfig, err := cc.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting rest config from client config: %w", err)
	}
	clnt, err := client.New(restConfig, client.Options{Scheme: commonscheme.SkrScheme})
	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	return clnt, nil
}

func (k *defaultKeb) RenewInstanceKubeconfig(ctx context.Context, runtimeID string) error {
	rt := &infrastructuremanagerv1.Runtime{}
	err := k.kcpClient.Get(ctx, client.ObjectKey{Namespace: k.config.KcpNamespace, Name: runtimeID}, rt)
	if err != nil {
		return fmt.Errorf("error getting Runtime %q: %w", runtimeID, err)
	}

	gc := &infrastructuremanagerv1.GardenerCluster{}
	err = k.kcpClient.Get(ctx, client.ObjectKeyFromObject(rt), gc)
	if err != nil {
		return fmt.Errorf("error getting GardenerCluster %q: %w", runtimeID, err)
	}

	ns := gc.Spec.Kubeconfig.Secret.Namespace
	if ns == "" {
		ns = rt.Namespace
	}
	secret := &corev1.Secret{}
	err = k.kcpClient.Get(ctx, types.NamespacedName{
		Namespace: ns,
		Name:      gc.Spec.Kubeconfig.Secret.Name,
	}, secret)
	if apierrors.IsNotFound(err) {
		secret = nil
		err = nil
	}
	if err != nil {
		return fmt.Errorf("error getting SKR credential secret: %w", err)
	}

	data, err := k.skrKubeconfigProvider.CreateNewKubeconfig(ctx, rt.Spec.Shoot.Name)
	if err != nil {
		return fmt.Errorf("error creating new kubeconfig: %w", err)
	}

	if secret != nil {
		secret.Data[gc.Spec.Kubeconfig.Secret.Key] = data
		if err := k.kcpClient.Update(ctx, secret); err != nil {
			return fmt.Errorf("error updating SKR credential secret: %w", err)
		}
	} else {
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      gc.Spec.Kubeconfig.Secret.Name,
			},
			Data: map[string][]byte{
				gc.Spec.Kubeconfig.Secret.Key: data,
			},
		}
		if err := k.kcpClient.Create(ctx, secret); err != nil {
			return fmt.Errorf("error creating SKR credential secret: %w", err)
		}
	}

	_, err = composed.PatchObjMergeAnnotation(
		ctx,
		e2elib.ExpiresAtAnnotation,
		k.clock.Now().Add(k.skrKubeconfigProvider.ExpiresIn()).Format(time.RFC3339),
		gc, k.kcpClient,
	)
	if err != nil {
		return fmt.Errorf("error patching GardenerCluster expires-in annotation: %w", err)
	}

	return nil
}
