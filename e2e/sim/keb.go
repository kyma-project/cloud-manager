package sim

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CreateInstanceInput struct {
	Alias         string
	GlobalAccount string
	SubAccount    string
	Provider      cloudcontrolv1beta1.ProviderType
	Region        string
}

func (in *CreateInstanceInput) Validate() error {
	var result error
	if in.Alias == "" {
		result = multierror.Append(result, errors.New("alias required"))
	}
	if in.Provider == "" {
		result = multierror.Append(result, errors.New("provider required"))
	}

	if result != nil {
		return result
	}

	if in.GlobalAccount == "" {
		in.GlobalAccount = uuid.NewString()
	}
	if in.SubAccount == "" {
		in.SubAccount = uuid.NewString()
	}
	if in.Region == "" {
		in.Region = defaultRegions[in.Provider]
	}
	return nil
}

type InstanceDetails struct {
	Alias         string
	GlobalAccount string
	SubAccount    string
	Provider      cloudcontrolv1beta1.ProviderType
	Region        string

	ProvisioningCompleted bool

	RuntimeID string
	ShootName string
}

// list

type listOptions struct {
	alias         string
	globalAccount string
	subAccount    string
	provider      cloudcontrolv1beta1.ProviderType
}

func (lo listOptions) MatchingLabels() client.MatchingLabels {
	result := client.MatchingLabels{}
	if lo.alias != "" {
		result[aliasLabel] = lo.alias
	}
	if lo.globalAccount != "" {
		result[cloudcontrolv1beta1.LabelScopeGlobalAccountId] = lo.globalAccount
	}
	if lo.subAccount != "" {
		result[cloudcontrolv1beta1.LabelScopeSubaccountId] = lo.subAccount
	}
	if lo.provider != "" {
		result[cloudcontrolv1beta1.LabelScopeBrokerPlanName] = strings.ToLower(string(lo.provider))
	}
	return result
}

type ListOption interface {
	ApplyOnList(*listOptions)
}

type WithAlias string

func (o WithAlias) ApplyOnList(opt *listOptions) {
	opt.alias = string(o)
}

type WithGlobalAccount string

func (o WithGlobalAccount) ApplyOnList(opt *listOptions) {
	opt.globalAccount = string(o)
}

type WithSubAccount string

func (o WithSubAccount) ApplyOnList(opt *listOptions) {
	opt.globalAccount = string(o)
}

type WithProvider cloudcontrolv1beta1.ProviderType

func (o WithProvider) ApplyOnList(opt *listOptions) {
	opt.provider = cloudcontrolv1beta1.ProviderType(o)
}

// wait

type waitOptions struct {
	runtimeIds []string
	timeout    time.Duration
	interval   time.Duration
}

type WaitOption interface {
	ApplyOnWait(*waitOptions)
}

var defaultWaitOptions = []WaitOption{
	WithTimeout(time.Minute),
	WithInterval(2 * time.Second),
}

type WithRuntimes []string

func (o WithRuntimes) ApplyOnWait(opt *waitOptions) {
	opt.runtimeIds = append(opt.runtimeIds, []string(o)...)
}

type WithTimeout time.Duration

func (o WithTimeout) ApplyOnWait(opt *waitOptions) {
	opt.timeout = time.Duration(o)
}

type WithInterval time.Duration

func (o WithInterval) ApplyOnWait(opt *waitOptions) {
	opt.interval = time.Duration(o)
}

// keb

type Keb interface {
	CreateInstance(ctx context.Context, in CreateInstanceInput) (InstanceDetails, error)
	WaitProvisioningCompleted(ctx context.Context, opts ...WaitOption) error
	GetInstance(ctx context.Context, runtimeID string) (*InstanceDetails, error)
	List(ctx context.Context, opts ...ListOption) ([]InstanceDetails, error)
	DeleteInstance(ctx context.Context, runtimeID string) error
}

func NewKeb(kcp client.Client, cpl CloudProfileLoader) Keb {
	return &defaultKeb{
		kcp: kcp,
		cpl: cpl,
	}
}

// keb implementation =================

var _ Keb = &defaultKeb{}

type defaultKeb struct {
	kcp client.Client
	cpl CloudProfileLoader
}

func RuntimeToInstanceDetails(rt *infrastructuremanagerv1.Runtime) InstanceDetails {
	return InstanceDetails{
		Alias:                 rt.Labels[aliasLabel],
		GlobalAccount:         rt.Labels[cloudcontrolv1beta1.LabelScopeGlobalAccountId],
		SubAccount:            rt.Labels[cloudcontrolv1beta1.LabelScopeSubaccountId],
		Provider:              cloudcontrolv1beta1.ProviderType(rt.Spec.Shoot.Provider.Type),
		Region:                rt.Spec.Shoot.Region,
		ProvisioningCompleted: rt.Status.ProvisioningCompleted,
		RuntimeID:             rt.Name,
		ShootName:             rt.Spec.Shoot.Name,
	}
}

func (k *defaultKeb) WaitProvisioningCompleted(ctx context.Context, opts ...WaitOption) error {
	options := &waitOptions{}
	for _, o := range append(append([]WaitOption{}, defaultWaitOptions...), opts...) {
		o.ApplyOnWait(options)
	}

	pollCtx, pollCancel := context.WithTimeout(ctx, options.timeout)
	defer pollCancel()

	err := wait.PollUntilContextTimeout(pollCtx, options.interval, options.timeout, false, func(ctx context.Context) (done bool, err error) {
		arr, err := k.List(ctx)
		if err != nil {
			return false, fmt.Errorf("error listing instances: %w", err)
		}
		allDone := true
		for _, i := range arr {
			if !slices.Contains(options.runtimeIds, i.RuntimeID) {
				continue
			}
			if !i.ProvisioningCompleted {
				allDone = false
				break
			}
		}

		return allDone, nil
	})

	if err != nil {
		return fmt.Errorf("error waiting for instance to become provisioned: %w", err)
	}

	return nil
}

func (k *defaultKeb) GetInstance(ctx context.Context, runtimeID string) (*InstanceDetails, error) {
	rt := &infrastructuremanagerv1.Runtime{}
	err := k.kcp.Get(ctx, client.ObjectKey{Namespace: e2econfig.Config.KcpNamespace, Name: runtimeID}, rt)
	if client.IgnoreNotFound(err) != nil {
		return nil, fmt.Errorf("error getting runtime %q: %w", runtimeID, err)
	}
	if err != nil {
		return nil, nil
	}
	return ptr.To(RuntimeToInstanceDetails(rt)), nil
}

func (k *defaultKeb) CreateInstance(ctx context.Context, in CreateInstanceInput) (InstanceDetails, error) {
	if err := in.Validate(); err != nil {
		return InstanceDetails{}, err
	}
	subscription := e2econfig.Config.Subscriptions.GetDefaultForProvider(in.Provider)
	if subscription == nil {
		return InstanceDetails{}, fmt.Errorf("subscription not found for provider %q", in.Provider)
	}
	cpr, err := k.cpl.Load(ctx)
	if err != nil {
		return InstanceDetails{}, fmt.Errorf("error loading cloud profiles: %w", err)
	}
	rtBuilder := NewRuntimeBuilder(cpr).
		WithAlias(in.Alias).
		WithProvider(in.Provider, in.Region).
		WithSecretBindingName(subscription.Name).
		WithGlobalAccount(in.GlobalAccount).
		WithSubAccount(in.SubAccount)
	if err := rtBuilder.Validate(); err != nil {
		return InstanceDetails{}, fmt.Errorf("invalid create instance request: %w", err)
	}
	rt := rtBuilder.Build()

	err = k.kcp.Create(ctx, rt)
	if err != nil {
		return InstanceDetails{}, fmt.Errorf("error creating runtime: %w", err)
	}

	return InstanceDetails{
		Alias:                 in.Alias,
		GlobalAccount:         in.GlobalAccount,
		SubAccount:            in.SubAccount,
		Provider:              in.Provider,
		Region:                in.Region,
		ProvisioningCompleted: false,
		RuntimeID:             rt.Name,
		ShootName:             rt.Spec.Shoot.Name,
	}, nil
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
	if err := k.kcp.List(ctx, rtList, loArr...); err != nil {
		return nil, fmt.Errorf("error listing runtimes: %w", err)
	}

	var results []InstanceDetails
	for _, rt := range rtList.Items {
		results = append(results, RuntimeToInstanceDetails(&rt))
	}
	return results, nil
}

func (k *defaultKeb) DeleteInstance(ctx context.Context, runtimeID string) error {
	rt := &infrastructuremanagerv1.Runtime{}
	err := k.kcp.Get(ctx, client.ObjectKey{Namespace: e2econfig.Config.KcpNamespace, Name: runtimeID}, rt)
	if client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("error getting runtime %q: %w", runtimeID, err)
	}
	if apierrors.IsNotFound(err) {
		return nil
	}
	err = k.kcp.Delete(ctx, rt)
	if err != nil {
		return fmt.Errorf("error deleting runtime %q: %w", runtimeID, err)
	}
	return nil
}
