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
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/skr/iprange"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	reconcile2 "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	ctrl "sigs.k8s.io/controller-runtime"
)

type IpRangeReconcilerFactory struct{}

func (f *IpRangeReconcilerFactory) New(args reconcile2.ReconcilerArguments) reconcile.Reconciler {
	ff := iprange.NewReconcilerFactory()
	return &IpRangeReconciler{
		reconciler: ff.New(args),
	}
}

// IpRangeReconciler reconciles a IpRange object
type IpRangeReconciler struct {
	reconciler reconcile.Reconciler

	kymaRef    klog.ObjectRef
	kcpCluster cluster.Cluster
	skrCluster cluster.Cluster
}

//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=ipranges,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=ipranges/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud-resources.kyma-project.io,resources=ipranges/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the IpRange object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *IpRangeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.reconciler.Reconcile(ctx, req)
}

func SetupIpRangeReconciler(reg skrruntime.SkrRegistry) error {
	reg.IndexField(&cloudresourcesv1beta1.AwsNfsVolume{}, cloudresourcesv1beta1.IpRangeField, func(object client.Object) []string {
		nfsVol := object.(*cloudresourcesv1beta1.AwsNfsVolume)
		if nfsVol.Spec.IpRange.Name == "" {
			return nil
		}
		ns := nfsVol.Spec.IpRange.Namespace
		if len(ns) == 0 {
			ns = nfsVol.Namespace
		}
		return []string{fmt.Sprintf("%s/%s", ns, nfsVol.Spec.IpRange.Name)}
	})

	return reg.Register().
		WithFactory(&IpRangeReconcilerFactory{}).
		For(&cloudresourcesv1beta1.IpRange{}).
		Complete()
}
