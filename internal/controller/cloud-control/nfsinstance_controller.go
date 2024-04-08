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
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsnfsinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nfsinstance"
	awsnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nfsinstance/client"
	azurenfsinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/nfsinstance"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpnfsinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance"
	gcpnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/client"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func SetupNfsInstanceReconciler(
	kcpManager manager.Manager,
	awsSkrProvider awsclient.SkrClientProvider[awsnfsinstanceclient.Client],
	filestoreClientProvider gcpclient.ClientProvider[gcpnfsinstanceclient.FilestoreClient],
	env abstractions.Environment,
) error {
	if env == nil {
		env = abstractions.NewOSEnvironment()
	}
	return NewNfsInstanceReconciler(
		nfsinstance.NewNfsInstanceReconciler(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(kcpManager)),
			focal.NewStateFactory(),
			awsnfsinstance.NewStateFactory(awsSkrProvider, env),
			azurenfsinstance.NewStateFactory(),
			gcpnfsinstance.NewStateFactory(filestoreClientProvider, env),
		),
	).SetupWithManager(kcpManager)
}

func NewNfsInstanceReconciler(
	reconciler nfsinstance.NfsInstanceReconciler,
) *NfsInstanceReconciler {
	return &NfsInstanceReconciler{
		Reconciler: reconciler,
	}
}

type NfsInstanceReconciler struct {
	Reconciler nfsinstance.NfsInstanceReconciler
}

//+kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=nfsinstances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=nfsinstances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=nfsinstances/finalizers,verbs=update

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
	return r.Reconciler.Reconcile(ctx, req)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NfsInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudresourcesv1beta1.NfsInstance{}).
		Complete(r)
}
