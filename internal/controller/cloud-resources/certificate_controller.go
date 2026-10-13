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
	"github.com/kyma-project/cloud-manager/pkg/skr/certificate"
	certificateclient "github.com/kyma-project/cloud-manager/pkg/skr/provider/aws/certificate/client"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	reconcile2 "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type CertificateReconcilerFactory struct {
	awsClientProvider awsclient.SkrClientProvider[certificateclient.Client]
	env               abstractions.Environment
}

func (f *CertificateReconcilerFactory) New(args reconcile2.ReconcilerArguments) reconcile.Reconciler {
	return &CertificateReconciler{
		reconciler: certificate.NewReconcilerFactory(f.awsClientProvider, f.env).New(args),
	}
}

// CertificateReconciler reconciles a Certificate object
type CertificateReconciler struct {
	reconciler reconcile.Reconciler
}

//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=certificates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=certificates/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *CertificateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.reconciler.Reconcile(ctx, req)
}

func SetupCertificateReconciler(reg skrruntime.SkrRegistry, provider awsclient.SkrClientProvider[certificateclient.Client], env abstractions.Environment) error {
	factory := &CertificateReconcilerFactory{
		awsClientProvider: provider,
		env:               env,
	}

	return reg.Register().
		WithFactory(factory).
		For(&cloudresourcesv1beta1.Certificate{}).
		Complete()
}
