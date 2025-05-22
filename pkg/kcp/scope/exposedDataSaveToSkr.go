package scope

import (
	"context"

	"github.com/elliotchance/pie/v2"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	skrmanager "github.com/kyma-project/cloud-manager/pkg/skr/runtime/manager"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func exposedDataSaveToSkr(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	skrScheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(skrScheme))
	utilruntime.Must(cloudresourcesv1beta1.AddToScheme(skrScheme))
	utilruntime.Must(apiextensions.AddToScheme(skrScheme))

	skrManagerFactory := skrmanager.NewFactory(state.Cluster().ApiReader(), state.gardenerClusterSummary.Namespace, skrScheme)

	restConfig, err := skrManagerFactory.LoadRestConfig(ctx, state.gardenerClusterSummary.Name, state.gardenerClusterSummary.Key)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating rest config", composed.StopWithRequeue, ctx)
	}

	skrClient, err := ctrlclient.New(restConfig, ctrlclient.Options{Scheme: skrScheme})
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating k8s skr client", composed.StopWithRequeue, ctx)
	}

	cmName := types.NamespacedName{
		Namespace: "kyma-system",
		Name:      "kyma-info",
	}

	cm := &corev1.ConfigMap{}
	err = skrClient.Get(ctx, cmName, cm)
	if apierrors.IsNotFound(err) {
		cm = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: cmName.Namespace,
				Name:      cmName.Name,
			},
		}
	} else if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading SKR ConfigMap", composed.StopWithRequeue, ctx)
	}

	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}
	cm.Data["cloud.natGatewayIps"] = pie.Join(state.ObjAsScope().Status.ExposedData.NatGatewayIps, ", ")

	if cm.ResourceVersion == "" {
		err := skrClient.Create(ctx, cm)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error creating SKR ConfigMap", composed.StopWithRequeue, ctx)
		}
	} else {
		err := skrClient.Update(ctx, cm)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error updating SKR ConfigMap", composed.StopWithRequeue, ctx)
		}
	}

	return nil, ctx
}
