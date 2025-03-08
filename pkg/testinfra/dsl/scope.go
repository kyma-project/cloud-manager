package dsl

import (
	"context"
	"fmt"
	"strings"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/testinfra"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WithScope(scopeName string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(focal.CommonObject); ok {
				if x.ScopeRef().Name == "" {
					x.SetScopeRef(cloudcontrolv1beta1.ScopeRef{Name: scopeName})
				}
			} else {
				panic(fmt.Sprintf("type %T does not implement focal.CommonObject", obj))
			}
		},
	}
}

func CreateScopeAws(ctx context.Context, infra testinfra.Infra, scope *cloudcontrolv1beta1.Scope, opts ...ObjAction) error {
	if scope == nil {
		scope = &cloudcontrolv1beta1.Scope{}
	}

	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultKcpNamespace),
		).
		ApplyOnObject(scope)

	project := strings.TrimPrefix(scope.Namespace, "garden-")

	scope.Spec = cloudcontrolv1beta1.ScopeSpec{
		KymaName:  scope.Name,
		ShootName: scope.Name,
		Region:    "eu-west-1",
		Provider:  cloudcontrolv1beta1.ProviderAws,
		Scope: cloudcontrolv1beta1.ScopeInfo{
			Aws: &cloudcontrolv1beta1.AwsScope{
				AccountId:  infra.AwsMock().GetAccount(),
				VpcNetwork: fmt.Sprintf("shoot--%s--%s", project, scope.Name),
				Network: cloudcontrolv1beta1.AwsNetwork{
					Nodes:    "10.250.0.0/22",
					Pods:     "10.96.0.0/13",
					Services: "10.104.0.0/13",
					VPC: cloudcontrolv1beta1.AwsVPC{
						CIDR: "10.250.0.0/22",
					},
					Zones: []cloudcontrolv1beta1.AwsZone{
						{
							Name:     "eu-west-1a",
							Internal: "10.180.48.0/20",
							Public:   "10.180.32.0/20",
							Workers:  "10.180.0.0/19",
						},
						{
							Name:     "eu-west-1b",
							Internal: "10.180.112.0/20",
							Public:   "10.180.96.0/20",
							Workers:  "10.180.64.0/19",
						},
						{
							Name:     "eu-west-1c",
							Internal: "10.180.176.0/20",
							Public:   "10.180.160.0/20",
							Workers:  "10.180.128.0/19",
						},
					},
				},
			},
		},
	}

	err := infra.KCP().Client().Create(ctx, scope)
	if err != nil {
		return err
	}

	kyma := util.NewKymaUnstructured()
	if err := CreateKymaCR(ctx, infra, kyma,
		WithName(scope.Name),
		WithKymaModuleListedInSpec(),
		WithKymaStatusModuleState(util.KymaModuleStateReady),
		WithKymaSpecChannel("fast"),
	); err != nil {
		return fmt.Errorf("error creating kyma: %w", err)
	}

	return nil
}

func CreateScopeAzure(ctx context.Context, infra testinfra.Infra, scope *cloudcontrolv1beta1.Scope, opts ...ObjAction) error {
	if scope == nil {
		scope = &cloudcontrolv1beta1.Scope{}
	}

	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultKcpNamespace),
		).
		ApplyOnObject(scope)

	project := strings.TrimPrefix(scope.Namespace, "garden-")

	scope.Spec = cloudcontrolv1beta1.ScopeSpec{
		KymaName:  scope.Name,
		ShootName: scope.Name,
		Region:    "westeurope",
		Provider:  cloudcontrolv1beta1.ProviderAzure,
		Scope: cloudcontrolv1beta1.ScopeInfo{
			Azure: &cloudcontrolv1beta1.AzureScope{
				TenantId:       "fdd97055-c316-462f-8769-f99b670c2c4d",
				SubscriptionId: "2bfba5a4-c5d1-4b03-a7db-4ead64232fd6",
				VpcNetwork:     fmt.Sprintf("shoot--%s--%s", project, scope.Name),
				Network: cloudcontrolv1beta1.AzureNetwork{
					Cidr: "10.250.0.0/22",
					Zones: []cloudcontrolv1beta1.AzureNetworkZone{
						{
							Name: "1",
							Cidr: "10.250.0.0/25",
						},
						{
							Name: "2",
							Cidr: "10.250.0.128/25",
						},
						{
							Name: "3",
							Cidr: "10.250.1.0/25",
						},
					},
					Nodes:    "10.250.0.0/22",
					Pods:     "10.96.0.0/13",
					Services: "10.104.0.0/13",
				},
			},
		},
	}

	err := infra.KCP().Client().Create(ctx, scope)
	if err != nil {
		return err
	}

	kyma := util.NewKymaUnstructured()
	if err := CreateKymaCR(ctx, infra, kyma,
		WithName(scope.Name),
		WithKymaModuleListedInSpec(),
		WithKymaStatusModuleState(util.KymaModuleStateReady),
		WithKymaSpecChannel("fast"),
	); err != nil {
		return fmt.Errorf("error creating kyma: %w", err)
	}

	return nil
}

func CreateScopeGcp(ctx context.Context, infra testinfra.Infra, scope *cloudcontrolv1beta1.Scope, opts ...ObjAction) error {
	if scope == nil {
		scope = &cloudcontrolv1beta1.Scope{}
	}

	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultKcpNamespace),
		).
		ApplyOnObject(scope)

	project := strings.TrimPrefix(scope.Namespace, "garden-")

	scope.Spec = cloudcontrolv1beta1.ScopeSpec{
		KymaName:  scope.Name,
		Region:    "us-central1",
		ShootName: scope.Name,
		Provider:  cloudcontrolv1beta1.ProviderGCP,
		Scope: cloudcontrolv1beta1.ScopeInfo{
			Gcp: &cloudcontrolv1beta1.GcpScope{
				VpcNetwork: fmt.Sprintf("shoot--%s--%s", project, scope.Name),
				Project:    "sap-gcp-skr-dev-cust-00002",
				Network: cloudcontrolv1beta1.GcpNetwork{
					Nodes:    "10.250.0.0/22",
					Pods:     "10.96.0.0/13",
					Services: "10.104.0.0/13",
				},
				Workers: []cloudcontrolv1beta1.GcpWorkers{
					{
						Zones: []string{"us-central1-a", "us-central1-b", "us-central1-c"},
					},
				},
			},
		},
	}

	err := infra.KCP().Client().Create(ctx, scope)
	if err != nil {
		return err
	}

	kyma := util.NewKymaUnstructured()
	if err := CreateKymaCR(ctx, infra, kyma,
		WithName(scope.Name),
		WithKymaModuleListedInSpec(),
		WithKymaStatusModuleState(util.KymaModuleStateReady),
		WithKymaSpecChannel("fast"),
	); err != nil {
		return fmt.Errorf("error creating kyma: %w", err)
	}

	return nil
}

func CreateScopeCcee(ctx context.Context, infra testinfra.Infra, scope *cloudcontrolv1beta1.Scope, opts ...ObjAction) error {
	if scope == nil {
		scope = &cloudcontrolv1beta1.Scope{}
	}

	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultKcpNamespace),
		).
		ApplyOnObject(scope)

	project := strings.TrimPrefix(scope.Namespace, "garden-")

	scope.Spec = cloudcontrolv1beta1.ScopeSpec{
		KymaName:  scope.Name,
		ShootName: scope.Name,
		Region:    "eu-de-1",
		Provider:  cloudcontrolv1beta1.ProviderOpenStack,
		Scope: cloudcontrolv1beta1.ScopeInfo{
			OpenStack: &cloudcontrolv1beta1.OpenStackScope{
				VpcNetwork: fmt.Sprintf("shoot--%s--%s", project, scope.Name),
				DomainName: "kyma",
				TenantName: "kyma-dev-02",
				Network: cloudcontrolv1beta1.OpenStackNetwork{
					Nodes:    "10.250.0.0/22",
					Pods:     "10.96.0.0/13",
					Services: "10.104.0.0/13",
					Zones:    []string{"eu-de-1d", "eu-de-1a", "eu-de-1b"},
				},
			},
		},
	}

	err := infra.KCP().Client().Create(ctx, scope)
	if err != nil {
		return err
	}

	kyma := util.NewKymaUnstructured()
	if err := CreateKymaCR(ctx, infra, kyma,
		WithName(scope.Name),
		WithKymaModuleListedInSpec(),
		WithKymaStatusModuleState(util.KymaModuleStateReady),
		WithKymaSpecChannel("fast"),
	); err != nil {
		return fmt.Errorf("error creating kyma: %w", err)
	}

	return nil
}

func GivenScopeAwsExists(ctx context.Context, infra testinfra.Infra, scope *cloudcontrolv1beta1.Scope, opts ...ObjAction) error {

	if scope == nil {
		scope = &cloudcontrolv1beta1.Scope{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultKcpNamespace),
		).
		ApplyOnObject(scope)

	err := infra.KCP().Client().Get(ctx, client.ObjectKeyFromObject(scope), scope)
	if client.IgnoreNotFound(err) != nil {
		return err
	}
	if apierrors.IsNotFound(err) {
		err := CreateScopeAws(ctx, infra, scope)
		if err != nil {
			return err
		}
	}
	return nil
}

func GivenScopeAzureExists(ctx context.Context, infra testinfra.Infra, scope *cloudcontrolv1beta1.Scope, opts ...ObjAction) error {

	if scope == nil {
		scope = &cloudcontrolv1beta1.Scope{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultKcpNamespace),
		).
		ApplyOnObject(scope)

	err := infra.KCP().Client().Get(ctx, client.ObjectKeyFromObject(scope), scope)
	if client.IgnoreNotFound(err) != nil {
		return err
	}
	if apierrors.IsNotFound(err) {
		err := CreateScopeAzure(ctx, infra, scope)
		if err != nil {
			return err
		}
	}
	return nil
}
