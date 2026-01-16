package dsl

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/testinfra"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WithSubscriptionSpecGarden(bindingName string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudcontrolv1beta1.Subscription); ok {
				x.Spec.Details = cloudcontrolv1beta1.SubscriptionDetails{
					Garden: &cloudcontrolv1beta1.SubscriptionGarden{
						BindingName: bindingName,
					},
				}
			} else {
				panic(fmt.Sprintf("expected Subscription but got type %T", obj))
			}
		},
	}
}

func CreateSubscription(
	ctx context.Context,
	infra testinfra.Infra,
	subscription *cloudcontrolv1beta1.Subscription,
	objActions ...ObjAction,
) error {
	if subscription == nil {
		subscription = &cloudcontrolv1beta1.Subscription{}
	}

	NewObjActions(objActions...).
		Append(WithNamespace(DefaultKcpNamespace)).
		ApplyOnObject(subscription)

	err := infra.KCP().Client().Create(ctx, subscription)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

func SubscriptionPatchStatusReadyAws(ctx context.Context, infra testinfra.Infra, subscription *cloudcontrolv1beta1.Subscription, awsAccountId string) error {
	return composed.NewStatusPatcher(subscription).
		MutateStatus(func(sub *cloudcontrolv1beta1.Subscription) {
			sub.Status.Provider = cloudcontrolv1beta1.ProviderAws
			sub.Status.SubscriptionInfo = &cloudcontrolv1beta1.SubscriptionInfo{
				Aws: &cloudcontrolv1beta1.SubscriptionInfoAws{
					Account: awsAccountId,
				},
			}
			sub.SetStatusReady()
		}).
		Patch(ctx, infra.KCP().Client())
}

func SubscriptionPatchStatusReadyAzure(ctx context.Context, infra testinfra.Infra, subscription *cloudcontrolv1beta1.Subscription, azureTenantId string, azureSubscriptionId string) error {
	return composed.NewStatusPatcher(subscription).
		MutateStatus(func(sub *cloudcontrolv1beta1.Subscription) {
			sub.Status.Provider = cloudcontrolv1beta1.ProviderAzure
			sub.Status.SubscriptionInfo = &cloudcontrolv1beta1.SubscriptionInfo{
				Azure: &cloudcontrolv1beta1.SubscriptionInfoAzure{
					TenantId:       azureTenantId,
					SubscriptionId: azureSubscriptionId,
				},
			}
			sub.SetStatusReady()
		}).
		Patch(ctx, infra.KCP().Client())
}
