package awsnfsvolume

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	composed "github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

func testStateFactory(awsNfsVolume *cloudresourcesv1beta1.AwsNfsVolume) *State {
	skrScheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(skrScheme))
	utilruntime.Must(cloudresourcesv1beta1.AddToScheme(skrScheme))

	skrClient := fake.NewClientBuilder().
		WithScheme(skrScheme).
		Build()
	skrCluster := composed.NewStateCluster(skrClient, skrClient, nil, skrScheme)
	return &State{
		State:          composed.NewStateFactory(skrCluster).NewState(types.NamespacedName{}, awsNfsVolume),
		KcpCluster:     nil,
		SkrIpRange:     nil,
		KcpNfsInstance: nil,
		Volume:         nil,
		PVC:            nil,
	}
}
