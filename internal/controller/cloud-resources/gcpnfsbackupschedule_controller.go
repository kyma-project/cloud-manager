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
	"github.com/kyma-project/cloud-manager/pkg/feature"
	backupschedulev1 "github.com/kyma-project/cloud-manager/pkg/skr/backupschedule/v1"
	"github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsbackupschedule"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	reconcile2 "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// gcpNfsBackupScheduleRunner is a common interface for v1 and v2 reconcilers
type gcpNfsBackupScheduleRunner interface {
	Run(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
}

// GcpNfsBackupScheduleReconciler reconciles a GcpNfsBackupSchedule object
type GcpNfsBackupScheduleReconciler struct {
	reconciler gcpNfsBackupScheduleRunner
}

//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=gcpnfsbackupschedules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=gcpnfsbackupschedules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=gcpnfsbackupschedules/finalizers,verbs=update

func (r *GcpNfsBackupScheduleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.reconciler.Run(ctx, req)
}

type GcpNfsBackupScheduleReconcilerFactory struct {
	env abstractions.Environment
	clk clock.Clock
}

func (f *GcpNfsBackupScheduleReconcilerFactory) New(args reconcile2.ReconcilerArguments) reconcile.Reconciler {
	if feature.BackupScheduleV2.Value(context.Background()) {
		reconciler := gcpnfsbackupschedule.NewReconciler(
			args.KymaRef, args.KcpCluster, args.SkrCluster, f.env, f.clk,
		)
		return &GcpNfsBackupScheduleReconciler{reconciler: &reconciler}
	}

	reconciler := backupschedulev1.NewReconciler(
		args.KymaRef, args.KcpCluster, args.SkrCluster, f.env, backupschedulev1.GcpNfsBackupSchedule,
	)
	return &GcpNfsBackupScheduleReconciler{reconciler: &reconciler}
}

func SetupGcpNfsBackupScheduleReconciler(
	reg skrruntime.SkrRegistry,
	env abstractions.Environment,
	clk clock.Clock,
) error {
	return reg.Register().
		WithFactory(&GcpNfsBackupScheduleReconcilerFactory{env: env, clk: clk}).
		For(&cloudresourcesv1beta1.GcpNfsBackupSchedule{}).
		Complete()
}
