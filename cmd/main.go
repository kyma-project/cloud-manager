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
	"github.com/elliotchance/pie/v2"
	awsiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/iprange/client"
	awsnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nfsinstance/client"
	gcpiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	gcpFilestoreClient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/client"
	scopeclient "github.com/kyma-project/cloud-manager/pkg/kcp/scope/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"os"

	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
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

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	cloudcontrolcontroller "github.com/kyma-project/cloud-manager/internal/controller/cloud-control"
	cloudresourcescontroller "github.com/kyma-project/cloud-manager/internal/controller/cloud-resources"
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

	utilruntime.Must(clientgoscheme.AddToScheme(skrScheme))
	utilruntime.Must(cloudresourcesv1beta1.AddToScheme(skrScheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var gcpStructuredLogging bool
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&gcpStructuredLogging, "gcp-structured-logging", false, "Enable GCP structured logging")
	flag.Parse()

	opts := zap.Options{}
	if gcpStructuredLogging {
		opts.EncoderConfigOptions = []zap.EncoderConfigOption{
			util.GcpZapEncoderConfigOption(),
		}
	} else {
		opts.Development = true
	}

	rootLogger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(rootLogger)

	setupLog.WithValues(
		"scheme", "KCP",
		"kinds", pie.Keys(kcpScheme.KnownTypes(cloudcontrolv1beta1.GroupVersion)),
	).Info("Schema dump")
	setupLog.WithValues(
		"scheme", "SKR",
		"kinds", pie.Keys(skrScheme.KnownTypes(cloudresourcesv1beta1.GroupVersion)),
	).Info("Schema dump")

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 kcpScheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "445827a5.kyma-project.io",
		Logger:                 rootLogger,
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
	skrLoop := skrruntime.NewLooper(mgr, skrScheme, skrRegistry, mgr.GetLogger())

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
	if err = cloudresourcescontroller.SetupGcpNfsVolumeReconciler(skrRegistry); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GcpNfsVolume")
		os.Exit(1)
	}

	// KCP Controllers
	if err = cloudcontrolcontroller.SetupScopeReconciler(mgr, scopeclient.NewAwsStsGardenClientProvider(), skrLoop); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Scope")
		os.Exit(1)
	}
	if err = cloudcontrolcontroller.SetupNfsInstanceReconciler(
		mgr,
		awsnfsinstanceclient.NewClientProvider(),
		gcpFilestoreClient.NewFilestoreClientProvider(),
	); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NfsInstance")
		os.Exit(1)
	}
	if err = cloudcontrolcontroller.SetupVpcPeeringReconciler(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "VpcPeering")
		os.Exit(1)
	}
	if err = cloudcontrolcontroller.SetupIpRangeReconciler(
		mgr,
		awsiprangeclient.NewClientProvider(),
		gcpiprangeclient.NewServiceNetworkingClient(),
		gcpiprangeclient.NewComputeClient(),
	); err != nil {
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

	//skrLoop.AddKymaName("dffb0722-a18c-11ee-8c90-0242ac120002")
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
	// 2024-03-04T14:18
}
