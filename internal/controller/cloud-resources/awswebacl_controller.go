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
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	"github.com/kyma-project/cloud-manager/pkg/skr/awswebacl"
	"github.com/kyma-project/cloud-manager/pkg/skr/awswebacl/client"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	reconcile2 "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type AwsWebAclReconcilerFactory struct {
	webAclProvider awsclient.SkrClientProvider[client.Client]
	env            abstractions.Environment
}

func (f *AwsWebAclReconcilerFactory) New(args reconcile2.ReconcilerArguments) reconcile.Reconciler {
	return &AwsWebAclReconciler{
		reconciler: awswebacl.NewReconcilerFactory(f.webAclProvider, f.env).New(args),
	}
}

// AwsWebAclReconciler reconciles a AwsWebAcl object
type AwsWebAclReconciler struct {
	reconciler reconcile.Reconciler
}

//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=awswebacls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=awswebacls/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=awswebacls/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AwsWebAcl object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *AwsWebAclReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.reconciler.Reconcile(ctx, req)
}

func SetupAwsWebAclReconciler(reg skrruntime.SkrRegistry, provider awsclient.SkrClientProvider[client.Client], env abstractions.Environment) error {
	return reg.Register().
		WithFactory(&AwsWebAclReconcilerFactory{
			webAclProvider: provider,
			env:            env,
		}).
		For(&cloudresourcesv1beta1.AwsWebAcl{}).
		Complete()
}
