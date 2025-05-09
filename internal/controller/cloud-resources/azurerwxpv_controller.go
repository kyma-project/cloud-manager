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

	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"

	"github.com/kyma-project/cloud-manager/pkg/skr/azurerwxpv"
	"github.com/kyma-project/cloud-manager/pkg/skr/azurerwxpv/client"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	skrreconciler "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	ctrl "sigs.k8s.io/controller-runtime"
)

type AzureRwxPvReconcilerFactory struct {
	clientProvider azureclient.ClientProvider[client.Client]
}

func (f *AzureRwxPvReconcilerFactory) New(args skrreconciler.ReconcilerArguments) reconcile.Reconciler {
	return &AzureRwxPvReconciler{
		reconciler: azurerwxpv.NewReconciler(args, f.clientProvider),
	}
}

// AzureRwxPvReconciler reconciles a AzureRwxPv object
type AzureRwxPvReconciler struct {
	reconciler reconcile.Reconciler
}

// +kubebuilder:rbac:groups=v1,resources=persistentvolumes,verbs=get;list;watch
// +kubebuilder:rbac:groups=v1,resources=persistentvolumes/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PersistentVolumePv object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *AzureRwxPvReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.reconciler.Reconcile(ctx, req)
}

func SetupAzureRwxPvReconciler(reg skrruntime.SkrRegistry, clientProvider azureclient.ClientProvider[client.Client]) error {

	return reg.Register().
		WithFactory(&AzureRwxPvReconcilerFactory{
			clientProvider: clientProvider,
		}).
		For(&corev1.PersistentVolume{}).
		Complete()
}
