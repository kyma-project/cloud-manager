package sim

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/elliotchance/pie/v2"
	gardenerapicore "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// KEB =============================================

type Keb interface {
	CreateInstance(ctx context.Context, opts ...CreateOption) (InstanceDetails, error)
	WaitProvisioningCompleted(ctx context.Context, opts ...WaitOption) error
	WaitDeleted(ctx context.Context, opts ...WaitOption) error
	GetInstance(ctx context.Context, runtimeID string) (*InstanceDetails, error)
	List(ctx context.Context, opts ...ListOption) ([]InstanceDetails, error)
	DeleteInstance(ctx context.Context, opts ...DeleteOption) error
}

func NewKeb(kcpClient client.Client, gardenClient client.Client, cpl CloudProfileLoader, config *e2econfig.ConfigType) Keb {
	return &defaultKeb{
		kcpClient:    kcpClient,
		gardenClient: gardenClient,
		cpl:          cpl,
		config:       config,
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
		WithTimeout(10 * time.Second),
	}
}

// WaitOption ============================================================

type WaitOption interface {
	ApplyOnWait(*waitOptions)
}

type waitOptions struct {
	runtimeIds       []string
	timeout          time.Duration
	interval         time.Duration
	progressCallback func(WaitProgress)
}

func (o waitOptions) validate() error {
	var result error
	if len(o.runtimeIds) == 0 {
		result = multierror.Append(result, errors.New("no runtime IDs specified to wait for"))
	}
	return result
}

type WaitProgress struct {
	Done    []InstanceDetails
	Pending []InstanceDetails
	WithErr []InstanceDetails
}

func (in WaitProgress) DoneAliases() []string {
	return pie.Map(in.Done, func(x InstanceDetails) string {
		return x.Alias
	})
}

func (in WaitProgress) PendingAliases() []string {
	return pie.Map(in.Pending, func(x InstanceDetails) string {
		return x.Alias
	})
}

func (in WaitProgress) ErrAliases() []string {
	return pie.Map(in.WithErr, func(x InstanceDetails) string {
		return x.Alias
	})
}

func (in WaitProgress) Hash() string {
	arr := make([]string, 0, len(in.Done)+1+len(in.Pending))
	for _, i := range in.Done {
		arr = append(arr, i.RuntimeID)
	}
	arr = append(arr, "|")
	for _, i := range in.Pending {
		arr = append(arr, i.RuntimeID)
	}
	txt := strings.Join(arr, ",")
	hasher := sha256.New()
	hasher.Write([]byte(txt))
	sum := hasher.Sum(nil)
	return hex.EncodeToString(sum)
}

var defaultWaitOptions = []WaitOption{
	WithTimeout(15 * time.Minute),
	WithInterval(2 * time.Second),
	WithProgressCallback(func(WaitProgress) {}),
}

// ListOption ============================================================

type ListOption interface {
	ApplyOnList(*listOptions)
}

type listOptions struct {
	alias         string
	globalAccount string
	subAccount    string
	provider      cloudcontrolv1beta1.ProviderType
}

func (o listOptions) MatchingLabels() client.MatchingLabels {
	result := client.MatchingLabels{}
	if o.alias != "" {
		result[aliasLabel] = o.alias
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
	shootMarkedForDeletionTimeout  time.Duration
	shootMarkedForDeletionInterval time.Duration
}

func (in deleteOptions) validate() error {
	var result error
	if in.runtimeId == "" {
		result = multierror.Append(result, errors.New("runtimeId required"))
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

// ===================================================================
// OPTIONS IMPLEMENTATION
// ===================================================================

// WithNodesRange =====================================================

// WithNodesRange specifies nodes rage for create
type WithNodesRange string

func (o WithNodesRange) ApplyOnCreate(opts *createOptions) {
	opts.nodesRange = string(o)
}

// WithPodsRange =====================================================

// WithPodsRange specifies pods range for create
type WithPodsRange string

func (o WithPodsRange) ApplyOnCreate(opts *createOptions) {
	opts.podsRange = string(o)
}

// WithServicesRange =====================================================

// WithServicesRange specifies services range for create
type WithServicesRange string

func (o WithServicesRange) ApplyOnCreate(opts *createOptions) {
	opts.servicesRange = string(o)
}

// WithAlias =====================================================

// WithAlias specifies alias for list and create
type WithAlias string

func (o WithAlias) ApplyOnList(opt *listOptions) {
	opt.alias = string(o)
}

func (o WithAlias) ApplyOnCreate(opt *createOptions) {
	opt.alias = string(o)
}

// WithGlobalAccount =====================================================

// WithGlobalAccount specifies alias for list and create
type WithGlobalAccount string

func (o WithGlobalAccount) ApplyOnList(opt *listOptions) {
	opt.globalAccount = string(o)
}

func (o WithGlobalAccount) ApplyOnCreate(opt *createOptions) {
	opt.globalAccount = string(o)
}

// WithSubAccount =====================================================

// WithSubAccount specifies alias for list and create
type WithSubAccount string

func (o WithSubAccount) ApplyOnList(opt *listOptions) {
	opt.globalAccount = string(o)
}

func (o WithSubAccount) ApplyOnCreate(opt *createOptions) {
	opt.subAccount = string(o)
}

// WithRegion =======================================================

type WithRegion string

func (o WithRegion) ApplyOnCreate(opt *createOptions) {
	opt.region = string(o)
}

// WithProvider =====================================================

// WithProvider specifies alias for list and create
type WithProvider cloudcontrolv1beta1.ProviderType

func (o WithProvider) ApplyOnList(opt *listOptions) {
	opt.provider = cloudcontrolv1beta1.ProviderType(o)
}

func (o WithProvider) ApplyOnCreate(opt *createOptions) {
	opt.provider = cloudcontrolv1beta1.ProviderType(o)
	if opt.region == "" {
		opt.region = defaultRegions[opt.provider]
	}
}

// WithProgressCallback =====================================================

// WithProgressCallback specifies progress callback for wait operations in WaitProvisioningCompleted
// and WaitDeleted
type WithProgressCallback func(WaitProgress)

func (o WithProgressCallback) ApplyOnWait(opt *waitOptions) {
	opt.progressCallback = o
}

// WithRuntimes =====================================================

// WithRuntimes specifies multiple runtime ids for wait operations in WaitProvisioningCompleted and WaitDeleted
type WithRuntimes []string

func (o WithRuntimes) ApplyOnWait(opt *waitOptions) {
	opt.runtimeIds = append(opt.runtimeIds, []string(o)...)
}

// WithRuntime =====================================================

// WithRuntime specifies single runtime id for WaitProvisioningCompleted, WaitDeleted, and DeleteInstance
// If applied multiple times or in the combination with WithRuntimes then
// * for the WaitProvisioningCompleted and WaitDeleted it has cumulative effect
// * for DeleteInstance it overwrites previously set runtime id
type WithRuntime string

func (o WithRuntime) ApplyOnWait(opt *waitOptions) {
	opt.runtimeIds = append(opt.runtimeIds, string(o))
}

func (o WithRuntime) ApplyOnDelete(opt *deleteOptions) {
	opt.runtimeId = string(o)
}

// WithTimeout =====================================================

// WithTimeout specifies timeout for
// * wait ops WaitProvisioningCompleted and WaitDeleted for how much it wil be waited for instances to be provisioned or deleted
// * DeleteInstance for how much it will be waited for shoot to be marked as deleted, if zero no wait is done
// * CreateInstance for how much it will be waited for shoot to be created, if zero no wait is done
type WithTimeout time.Duration

func (o WithTimeout) ApplyOnWait(opt *waitOptions) {
	opt.timeout = time.Duration(o)
}

func (o WithTimeout) ApplyOnDelete(opt *deleteOptions) {
	opt.shootMarkedForDeletionTimeout = time.Duration(o)
}

func (o WithTimeout) ApplyOnCreate(opt *createOptions) {
	opt.shootCreatedTimeout = time.Duration(o)
}

// WithInterval =====================================================

// WithInterval specifies interval duration for timeout if any set, in wait ops, DeleteInstance and CreateInstance
type WithInterval time.Duration

func (o WithInterval) ApplyOnWait(opt *waitOptions) {
	opt.interval = time.Duration(o)
}

func (o WithInterval) ApplyOnDelete(opt *deleteOptions) {
	opt.shootMarkedForDeletionInterval = time.Duration(o)
}

func (o WithInterval) ApplyOnCreate(opt *createOptions) {
	opt.shootCreatedInterval = time.Duration(o)
}

// InstanceDetails =============================================

type InstanceDetails struct {
	Alias         string
	GlobalAccount string
	SubAccount    string
	Provider      cloudcontrolv1beta1.ProviderType
	Region        string

	ProvisioningCompleted bool

	RuntimeID string
	ShootName string

	// State has value of the runtime.status.state
	State string
	// Message has value of message of the Condition type Error
	Message string

	BeingDeleted bool
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
	kcpClient    client.Client
	gardenClient client.Client
	cpl          CloudProfileLoader
	config       *e2econfig.ConfigType
}

func RuntimeToInstanceDetails(rt *infrastructuremanagerv1.Runtime) InstanceDetails {
	id := InstanceDetails{
		Alias:                 rt.Labels[aliasLabel],
		GlobalAccount:         rt.Labels[cloudcontrolv1beta1.LabelScopeGlobalAccountId],
		SubAccount:            rt.Labels[cloudcontrolv1beta1.LabelScopeSubaccountId],
		Provider:              cloudcontrolv1beta1.ProviderType(rt.Spec.Shoot.Provider.Type),
		Region:                rt.Spec.Shoot.Region,
		ProvisioningCompleted: rt.Status.ProvisioningCompleted,
		RuntimeID:             rt.Name,
		ShootName:             rt.Spec.Shoot.Name,
		State:                 string(rt.Status.State),
		BeingDeleted:          rt.DeletionTimestamp != nil,
	}
	errCond := meta.FindStatusCondition(rt.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	if errCond != nil && errCond.Status == metav1.ConditionTrue {
		id.Message = errCond.Message
	}
	return id
}

func (k *defaultKeb) WaitProvisioningCompleted(ctx context.Context, opts ...WaitOption) error {
	options := &waitOptions{}
	for _, o := range append(append([]WaitOption{}, defaultWaitOptions...), opts...) {
		o.ApplyOnWait(options)
	}
	if err := options.validate(); err != nil {
		return fmt.Errorf("waitProvisioningCompleted invalid input: %w", err)
	}

	lastNotifiedWith := "-"

	err := wait.PollUntilContextTimeout(ctx, options.interval, options.timeout, false, func(ctx context.Context) (bool, error) {
		arr, err := k.List(ctx)
		if err != nil {
			return false, fmt.Errorf("error listing instances in WaitProvisioningCompleted: %w", err)
		}
		var done []InstanceDetails
		var pending []InstanceDetails
		var withErr []InstanceDetails
		for _, id := range arr {
			if !slices.Contains(options.runtimeIds, id.RuntimeID) {
				continue
			}
			if id.BeingDeleted {
				id.Message = "being deleted"
				withErr = append(withErr, id)
			} else if id.State == infrastructuremanagerv1.RuntimeStateFailed {
				withErr = append(withErr, id)
			} else {
				if id.ProvisioningCompleted {
					done = append(done, id)
				} else {
					pending = append(pending, id)
				}
			}
		}

		wp := WaitProgress{
			Done:    done,
			Pending: pending,
			WithErr: withErr,
		}
		currentNotifyWith := wp.Hash()

		if currentNotifyWith != lastNotifiedWith {
			lastNotifiedWith = currentNotifyWith
			options.progressCallback(wp)
		}

		err = nil
		for _, id := range withErr {
			err = multierror.Append(err, fmt.Errorf("instance %s %s has error %s", id.Alias, id.RuntimeID, id.Message))
		}

		return len(pending) == 0, err
	})

	if err != nil {
		return fmt.Errorf("error waiting for instance(s) to become provisioned: %w", err)
	}

	return nil
}

func (k *defaultKeb) WaitDeleted(ctx context.Context, opts ...WaitOption) error {
	options := &waitOptions{}
	for _, o := range append(append([]WaitOption{}, defaultWaitOptions...), opts...) {
		o.ApplyOnWait(options)
	}
	if err := options.validate(); err != nil {
		return fmt.Errorf("waitDeleted invalid input: %w", err)
	}

	lastNotifiedWith := "-"

	err := wait.PollUntilContextTimeout(ctx, options.interval, options.timeout, false, func(ctx context.Context) (bool, error) {
		arr, err := k.List(ctx)
		if err != nil {
			return false, fmt.Errorf("error listing instances in WaitDeleted: %w", err)
		}
		var done []InstanceDetails
		var pending []InstanceDetails
		var withErr []InstanceDetails
		for _, runtimeID := range options.runtimeIds {
			var id *InstanceDetails
			for _, x := range arr {
				if x.Alias == runtimeID {
					id = &x
					break
				}
			}
			if id == nil {
				done = append(done, InstanceDetails{RuntimeID: runtimeID})
			} else {
				if !id.BeingDeleted {
					id.Message = "not being deleted"
					withErr = append(withErr, *id)
				} else if id.State == infrastructuremanagerv1.RuntimeStateFailed {
					withErr = append(withErr, *id)
				} else {
					pending = append(pending, *id)
				}
			}
		}

		wp := WaitProgress{
			Done:    done,
			Pending: pending,
			WithErr: withErr,
		}
		currentNotifyWith := wp.Hash()

		if currentNotifyWith != lastNotifiedWith {
			lastNotifiedWith = currentNotifyWith
			options.progressCallback(wp)
		}

		err = nil
		for _, id := range withErr {
			err = multierror.Append(err, fmt.Errorf("instance %s %s has error %s", id.Alias, id.RuntimeID, id.Message))
		}

		return len(pending) == 0, err
	})

	if err != nil {
		return fmt.Errorf("error waiting for instance(s) to become provisioned: %w", err)
	}

	return nil
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
		options.region = defaultRegions[options.provider]
	}
	rtBuilder := NewRuntimeBuilder(cpr, k.config).
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

	if options.shootCreatedTimeout > 0 {
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
	if err := k.kcpClient.List(ctx, rtList, loArr...); err != nil {
		return nil, fmt.Errorf("error listing runtimes: %w", err)
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

	if rt.Spec.Shoot.Name != "" && options.shootMarkedForDeletionTimeout > 0 {
		// wait until shoot is marked for deletion
		err := wait.PollUntilContextTimeout(ctx, time.Second, 30*time.Second, false, func(ctx context.Context) (bool, error) {
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
