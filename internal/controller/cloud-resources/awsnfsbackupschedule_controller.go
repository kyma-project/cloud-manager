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

	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/skr/backupschedule"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	reconcile2 "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
)

// AwsNfsBackupScheduleReconciler reconciles a AwsNfsBackupSchedule object
type AwsNfsBackupScheduleReconciler struct {
	Reconciler backupschedule.Reconciler
}

//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=awsnfsbackupschedules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=awsnfsbackupschedules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=awsnfsbackupschedules/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AwsNfsBackupSchedule object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *AwsNfsBackupScheduleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	return r.Reconciler.Run(ctx, req)
}

type AwsNfsBackupScheduleReconcilerFactory struct {
	env abstractions.Environment
}

func (f *AwsNfsBackupScheduleReconcilerFactory) New(args reconcile2.ReconcilerArguments) reconcile.Reconciler {
	return &AwsNfsBackupScheduleReconciler{
		Reconciler: backupschedule.NewReconciler(args.KymaRef, args.KcpCluster, args.SkrCluster, f.env, backupschedule.AwsNfsBackupSchedule),
	}
}

func SetupAwsNfsBackupScheduleReconciler(reg skrruntime.SkrRegistry, env abstractions.Environment) error {

	return reg.Register().
		WithFactory(&AwsNfsBackupScheduleReconcilerFactory{env: env}).
		For(&cloudresourcesv1beta1.AwsNfsBackupSchedule{}).
		Complete()
}
