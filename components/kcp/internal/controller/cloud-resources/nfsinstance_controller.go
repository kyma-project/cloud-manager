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
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/util"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/kcp/api/cloud-resources/v1beta1"
)

// NfsInstanceReconciler reconciles a NfsInstance object
type NfsInstanceReconciler struct {
	client.Client
	record.EventRecorder
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=nfsinstances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=nfsinstances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=nfsinstances/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NfsInstance object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *NfsInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	//// TODO: this should be moved into separate reconciler package
	//state := scope.NewState(
	//	focal.NewState(
	//		composed.NewState(r.Client, r.EventRecorder, r.Scheme, req.NamespacedName, &cloudresourcesv1beta1.VpcPeering{}),
	//	),
	//	abstractions.NewFileReader(),
	//)
	//action := actions.New()
	//err, _ := action(ctx, state)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NfsInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	kymaUnstructured := util.NewKymaUnstructured()
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudresourcesv1beta1.NfsInstance{}).
		// Kyma CR should be watched on one place only so it gets into the cache
		// we're using empty handler since we're not interested into starting
		// reconciliation when Kyma CR changes, we just want them cached
		Watches(
			kymaUnstructured,
			handler.Funcs{},
		).
		Complete(r)
}
