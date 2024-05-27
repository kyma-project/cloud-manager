package dsl

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/testinfra"
	"strings"
)

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
					VPC: cloudcontrolv1beta1.AwsVPC{
						CIDR: "10.180.0.0/16",
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
		Region:    "eu-west-1",
		Provider:  cloudcontrolv1beta1.ProviderAzure,
		Scope: cloudcontrolv1beta1.ScopeInfo{
			Azure: &cloudcontrolv1beta1.AzureScope{
				TenantId:       "fdd97055-c316-462f-8769-f99b670c2c4d",
				SubscriptionId: "2bfba5a4-c5d1-4b03-a7db-4ead64232fd6",
				VpcNetwork:     fmt.Sprintf("shoot--%s--%s", project, scope.Name),
			},
		},
	}

	err := infra.KCP().Client().Create(ctx, scope)
	if err != nil {
		return err
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
			},
		},
	}

	err := infra.KCP().Client().Create(ctx, scope)
	if err != nil {
		return err
	}

	return nil
}
