/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloudcontrol

import (
	"context"

	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	kcpsubscription "github.com/kyma-project/cloud-manager/pkg/kcp/subscription"
	subscriptionclient "github.com/kyma-project/cloud-manager/pkg/kcp/subscription/client"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
)

func SetupSubscriptionReconciler(
	kcpManager manager.Manager,
	awsStsClientProvider awsclient.GardenClientProvider[subscriptionclient.AwsStsClient],
) error {
	return NewSubscriptionReconciler(
		kcpsubscription.New(kcpManager, awsStsClientProvider),
	).SetupWithManager(kcpManager)
}

func NewSubscriptionReconciler(r kcpsubscription.SubscriptionReconciler) *SubscriptionReconciler {
	return &SubscriptionReconciler{
		Reconciler: r,
	}
}

// SubscriptionReconciler reconciles a Subscription object
type SubscriptionReconciler struct {
	Reconciler kcpsubscription.SubscriptionReconciler
}

// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=subscriptions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=subscriptions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=subscriptions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SubscriptionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.Reconciler.Reconcile(ctx, req)
}

// SetupWithManager sets up the controller with the Manager.
func (r *SubscriptionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudcontrolv1beta1.Subscription{}).
		Named("cloud-control-subscription").
		Complete(r)
}
