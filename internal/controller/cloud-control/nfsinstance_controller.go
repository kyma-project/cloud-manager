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
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsnfsinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nfsinstance"
	awsnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nfsinstance/client"
	azurenfsinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/nfsinstance"
	cceeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/ccee/client"
	cceenfsinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/ccee/nfsinstance"
	cceenfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/ccee/nfsinstance/client"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpnfsinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance"
	gcpnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/client"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func SetupNfsInstanceReconciler(
	kcpManager manager.Manager,
	awsSkrProvider awsclient.SkrClientProvider[awsnfsinstanceclient.Client],
	filestoreClientProvider gcpclient.ClientProvider[gcpnfsinstanceclient.FilestoreClient],
	cceeProvider cceeclient.CceeClientProvider[cceenfsinstanceclient.Client],
	env abstractions.Environment,
) error {
	if env == nil {
		env = abstractions.NewOSEnvironment()
	}
	return NewNfsInstanceReconciler(
		nfsinstance.NewNfsInstanceReconciler(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(kcpManager)),
			focal.NewStateFactory(),
			awsnfsinstance.NewStateFactory(awsSkrProvider),
			azurenfsinstance.NewStateFactory(),
			gcpnfsinstance.NewStateFactory(filestoreClientProvider, env),
			cceenfsinstance.NewStateFactory(cceeProvider),
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
func (r *NfsInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.Reconciler.Reconcile(ctx, req)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NfsInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudcontrolv1beta1.NfsInstance{}, builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Complete(r)
}
