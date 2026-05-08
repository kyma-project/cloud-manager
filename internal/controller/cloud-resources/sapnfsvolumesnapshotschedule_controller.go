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

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	reconcile2 "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	"github.com/kyma-project/cloud-manager/pkg/skr/sapnfssnapshotschedule"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// SapNfsVolumeSnapshotScheduleReconciler reconciles a SapNfsVolumeSnapshotSchedule object
type SapNfsVolumeSnapshotScheduleReconciler struct {
	reconciler *sapnfssnapshotschedule.Reconciler
}

//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=sapnfsvolumesnapshotschedules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=sapnfsvolumesnapshotschedules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=sapnfsvolumesnapshotschedules/finalizers,verbs=update

func (r *SapNfsVolumeSnapshotScheduleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.reconciler.Run(ctx, req)
}

type SapNfsVolumeSnapshotScheduleReconcilerFactory struct {
	env abstractions.Environment
	clk clock.Clock
}

func (f *SapNfsVolumeSnapshotScheduleReconcilerFactory) New(args reconcile2.ReconcilerArguments) reconcile.Reconciler {
	reconciler := sapnfssnapshotschedule.NewReconciler(
		args.ScopeProvider, args.KcpCluster, args.SkrCluster, f.env, f.clk,
	)
	return &SapNfsVolumeSnapshotScheduleReconciler{reconciler: &reconciler}
}

func SetupSapNfsVolumeSnapshotScheduleReconciler(
	reg skrruntime.SkrRegistry,
	env abstractions.Environment,
	clk clock.Clock,
) error {
	return reg.Register().
		WithFactory(&SapNfsVolumeSnapshotScheduleReconcilerFactory{env: env, clk: clk}).
		For(&cloudresourcesv1beta1.SapNfsVolumeSnapshotSchedule{}).
		Complete()
}
