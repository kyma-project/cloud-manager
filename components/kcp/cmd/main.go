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

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/actions/scope"
	scopeclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/common/actions/scope/client"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/iprange"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/nfsinstance"
	awsiprange "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/aws/iprange"
	awsiprangeclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/aws/iprange/client"
	awsnfsinstance "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/aws/nfsinstance"
	awsnfsinstanceclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/aws/nfsinstance/client"
	azureiprange "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/azure/iprange"
	azurenfsinstance "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/azure/nfsinstance"
	gcpiprange "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/iprange"
	gcpiprangeclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/iprange/client"
	gcpnfsinstance "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/nfsinstance"
	gcpFilestoreClient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/nfsinstance/client"
	skrruntime "github.com/kyma-project/cloud-manager/components/kcp/pkg/skr/runtime"
	"github.com/kyma-project/cloud-manager/components/lib/composed"

	"sigs.k8s.io/controller-runtime/pkg/client"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-resources/v1beta1"
	cloudcontrolcontroller "github.com/kyma-project/cloud-manager/components/kcp/internal/controller/cloud-control"
	cloudresourcescontroller "github.com/kyma-project/cloud-manager/components/kcp/internal/controller/cloud-resources"
	//+kubebuilder:scaffold:imports
)

var (
	kcpScheme = runtime.NewScheme()
	skrScheme = runtime.NewScheme()
	setupLog  = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(kcpScheme))
	utilruntime.Must(cloudcontrolv1beta1.AddToScheme(kcpScheme))

	utilruntime.Must(cloudresourcesv1beta1.AddToScheme(skrScheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	for k := range skrScheme.KnownTypes(cloudresourcesv1beta1.GroupVersion) {
		fmt.Println(k)
	}
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 kcpScheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "445827a5.kyma-project.io",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
		Client: client.Options{
			Cache: &client.CacheOptions{
				Unstructured: true,
			},
		},
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	skrRegistry := skrruntime.NewRegistry(skrScheme)

	// SKR Controllers
	if err = cloudresourcescontroller.SetupCloudResourcesReconciler(skrRegistry); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CloudResources")
		os.Exit(1)
	}
	if err = cloudresourcescontroller.SetupIpRangeReconciler(skrRegistry); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "IpRange")
		os.Exit(1)
	}
	if err = cloudresourcescontroller.SetupAwsNfsVolumeReconciler(skrRegistry); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AwsNfsVolume")
		os.Exit(1)
	}

	// KCP Controllers
	if err = (&cloudcontrolcontroller.NfsInstanceReconciler{
		Reconciler: nfsinstance.NewNfsInstanceReconciler(
			composed.NewStateFactory(composed.NewStateClusterFromManager(mgr)),
			focal.NewStateFactory(),
			scope.NewStateFactory(abstractions.NewFileReader(), scopeclient.NewAwsStsGardenClientProvider()),
			awsnfsinstance.NewStateFactory(awsnfsinstanceclient.NewClientProvider(), abstractions.NewOSEnvironment()),
			azurenfsinstance.NewStateFactory(),
			gcpnfsinstance.NewStateFactory(gcpFilestoreClient.NewFilestoreClient(), abstractions.NewOSEnvironment()),
		),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NfsInstance")
		os.Exit(1)
	}
	if err = (&cloudcontrolcontroller.VpcPeeringReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "VpcPeering")
		os.Exit(1)
	}
	if err = (&cloudcontrolcontroller.IpRangeReconciler{
		Reconciler: iprange.NewIPRangeReconciler(
			composed.NewStateFactory(composed.NewStateClusterFromManager(mgr)),
			focal.NewStateFactory(),
			scope.NewStateFactory(abstractions.NewFileReader(), scopeclient.NewAwsStsGardenClientProvider()),
			awsiprange.NewStateFactory(awsiprangeclient.NewClientProvider(), abstractions.NewOSEnvironment()),
			azureiprange.NewStateFactory(nil),
			gcpiprange.NewStateFactory(gcpiprangeclient.NewServiceNetworkingClient(), gcpiprangeclient.NewComputeClient(), abstractions.NewOSEnvironment()),
		),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "IpRange")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	skrLoop := skrruntime.NewLooper(mgr, skrScheme, skrRegistry, mgr.GetLogger())
	skrLoop.AddKymaName("dffb0722-a18c-11ee-8c90-0242ac120002")
	//skrLoop.AddKymaName("134c0a3c-873d-436a-81c3-9b830a27b73a")
	//skrLoop.AddKymaName("264bb633-80f7-455b-83b2-f86630a57635")
	//skrLoop.AddKymaName("3f6f5a93-1c75-425a-b07e-4c82a0db1526")
	//skrLoop.AddKymaName("46059d75-d7b0-4d6a-955d-ce828ec4bb37")
	//skrLoop.AddKymaName("511d7132-448d-4672-a90d-420ad61b2365")
	//skrLoop.AddKymaName("eb693381-3e9c-4818-9f5d-378ba0c47314")

	err = mgr.Add(skrLoop)
	if err != nil {
		setupLog.Error(err, "error adding SkrLooper to KCP manager")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
