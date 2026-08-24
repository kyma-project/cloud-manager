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

	"github.com/kyma-project/cloud-manager/pkg/skr/alicloudnfsvolume"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	reconcile2 "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type AlicloudNfsVolumeReconcilerFactory struct{}

func (f *AlicloudNfsVolumeReconcilerFactory) New(args reconcile2.ReconcilerArguments) reconcile.Reconciler {
	return &AlicloudNfsVolumeReconciler{
		reconciler: alicloudnfsvolume.NewReconcilerFactory().New(args),
	}
}

// AlicloudNfsVolumeReconciler reconciles a AlicloudNfsVolume object
type AlicloudNfsVolumeReconciler struct {
	reconciler reconcile.Reconciler
}

//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=alicloudnfsvolumes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=alicloudnfsvolumes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=alicloudnfsvolumes/finalizers,verbs=update

func (r *AlicloudNfsVolumeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.reconciler.Reconcile(ctx, req)
}

func SetupAlicloudNfsVolumeReconciler(reg skrruntime.SkrRegistry) error {
	return reg.Register().
		WithFactory(&AlicloudNfsVolumeReconcilerFactory{}).
		For(&cloudresourcesv1beta1.AlicloudNfsVolume{}).
		Complete()
}
