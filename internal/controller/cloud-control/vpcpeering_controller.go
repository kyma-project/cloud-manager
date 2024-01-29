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

package cloudresources

import (
	"context"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
)

func SetupVpcPeeringReconciler(
	kcpManager manager.Manager,
) error {
	return NewVpcPeeringReconciler(
		kcpManager,
	).SetupWithManager(kcpManager)
}

func NewVpcPeeringReconciler(mgr manager.Manager) *VpcPeeringReconciler {
	return &VpcPeeringReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
}

type VpcPeeringReconciler struct {
	client.Client
	record.EventRecorder
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=vpcpeerings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=vpcpeerings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=vpcpeerings/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VpcPeering object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *VpcPeeringReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	//// TODO: this should be moved into separate reconciler package
	//state := scope.NewState(
	//	focal.NewState(
	//		composed.NewState(r.K8sClient, r.EventRecorder, r.Scheme, req.NamespacedName, &cloudresourcesv1beta1.VpcPeering{}),
	//	),
	//	abstractions.NewFileReader(),
	//)
	//action := actions.New()
	//err, _ := action(ctx, state)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VpcPeeringReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudresourcesv1beta1.VpcPeering{}).
		Complete(r)
}
