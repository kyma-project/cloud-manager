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
	"context"
	"flag"
	"os"

	awsexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/exposedData/client"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/exposedData/client"
	cceeconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/ccee/config"
	cceenfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/ccee/nfsinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/migrateFinalizers"

	"github.com/fsnotify/fsnotify"

	"github.com/elliotchance/pie/v2"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"

	"github.com/kyma-project/cloud-manager/pkg/config"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	featuretypes "github.com/kyma-project/cloud-manager/pkg/feature/types"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	"github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	vpcpeeringconfig "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/config"
	"github.com/kyma-project/cloud-manager/pkg/quota"

	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/iprange/client"
	awsnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nfsinstance/client"
	awsnukeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nuke/client"
	awsvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/vpcpeering/client"
	azureiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/iprange/client"
	azurenetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/network/client"
	azurenukeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/nuke/client"
	azureredisclusterclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/rediscluster/client"
	azureredisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/redisinstance/client"
	azurevpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vpcpeering/client"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	gcpnfsbackupclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client"
	gcpnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/client"
	gcpnfsrestoreclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client"
	gcpredisclusterclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/rediscluster/client"
	gcpredisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/redisinstance/client"
	gcpsubnetclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/subnet/client"
	gcpvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/vpcpeering/client"
	scopeclient "github.com/kyma-project/cloud-manager/pkg/kcp/scope/client"
	awsnfsvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumebackup/client"
	awsnfsvolumerestoreclient "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumerestore/client"
	azurerwxpvclient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxpv/client"

	azurerwxvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"

	"github.com/kyma-project/cloud-manager/pkg/util"

	"sigs.k8s.io/controller-runtime/pkg/client"

	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	skrruntimeconfig "github.com/kyma-project/cloud-manager/pkg/skr/runtime/config"

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
	utilruntime.Must(apiextensions.AddToScheme(kcpScheme))

	utilruntime.Must(clientgoscheme.AddToScheme(skrScheme))
	utilruntime.Must(cloudresourcesv1beta1.AddToScheme(skrScheme))
	utilruntime.Must(apiextensions.AddToScheme(skrScheme))
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

	cfg := loadConfig()
	cfg.Read()

	opts := zap.Options{}
	if gcpStructuredLogging {
		opts.EncoderConfigOptions = []zap.EncoderConfigOption{
			util.GcpZapEncoderConfigOption(),
		}
	} else {
		opts.Development = true
	}

	baseCtx := context.Background()
	baseCtx = feature.ContextBuilderFromCtx(baseCtx).
		Landscape(os.Getenv("LANDSCAPE")).
		Plane(featuretypes.PlaneKcp).
		Build(baseCtx)

	rootLogger := zap.New(zap.UseFlagOptions(&opts))
	rootLogger = rootLogger.WithSink(util.NewLogFilterSink(rootLogger.GetSink()))
	baseCtx = composed.LoggerIntoCtx(baseCtx, rootLogger)
	ctrl.SetLogger(rootLogger)

	setupLog.WithValues(
		"scheme", "KCP",
		"kinds", pie.Keys(kcpScheme.KnownTypes(cloudcontrolv1beta1.GroupVersion)),
	).Info("Schema dump")
	setupLog.WithValues(
		"scheme", "SKR",
		"kinds", pie.Keys(skrScheme.KnownTypes(cloudresourcesv1beta1.GroupVersion)),
	).Info("Schema dump")
	setupLog.WithValues("config", cfg.PrintJson()).
		Info("Config dump")

	skrRegistry := skrruntime.NewRegistry(skrScheme)
	activeSkrCollection := skrruntime.NewActiveSkrCollection()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		BaseContext: func() context.Context {
			return baseCtx
		},
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

	ctx := ctrl.SetupSignalHandler()

	ctx = feature.ContextBuilderFromCtx(ctx).
		Landscape(os.Getenv("LANDSCAPE")).
		Plane(featuretypes.PlaneKcp).
		Build(ctx)

	skrLoop := skrruntime.NewLooper(activeSkrCollection, mgr, skrScheme, skrRegistry, mgr.GetLogger())

	//Get env
	env := abstractions.NewOSEnvironment()

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

	if err = cloudresourcescontroller.SetupGcpNfsVolumeBackupReconciler(skrRegistry, gcpnfsbackupclient.NewFileBackupClientProvider(), env, setupLog); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GcpNfsVolumeBackup")
		os.Exit(1)
	}

	if err = cloudresourcescontroller.SetupGcpNfsVolumeRestoreReconciler(skrRegistry, gcpnfsrestoreclient.NewFileRestoreClientProvider(), env, setupLog); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GcpNfsVolumeRestore")
		os.Exit(1)
	}

	if err = cloudresourcescontroller.SetupAzureVpcPeeringReconciler(skrRegistry); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AzureVpcPeering")
		os.Exit(1)
	}

	if err = cloudresourcescontroller.SetupGcpRedisInstanceReconciler(skrRegistry); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GcpRedisInstance")
		os.Exit(1)
	}

	if err = cloudresourcescontroller.SetupGcpRedisClusterReconciler(skrRegistry); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GcpRedisCluster")
		os.Exit(1)
	}

	if err = cloudresourcescontroller.SetupAzureRedisInstanceReconciler(skrRegistry); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AzureRedisInstance")
		os.Exit(1)
	}

	if err = cloudresourcescontroller.SetupAwsRedisInstanceReconciler(skrRegistry); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AwsRedisInstance")
		os.Exit(1)
	}

	if err = cloudresourcescontroller.SetupAwsRedisClusterReconciler(skrRegistry); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AwsRedisCluster")
		os.Exit(1)
	}

	if err = cloudresourcescontroller.SetupAzureRedisClusterReconciler(skrRegistry); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AzureRedisCluster")
		os.Exit(1)
	}

	if err = cloudresourcescontroller.SetupAwsVpcPeeringReconciler(skrRegistry); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AwsVpcPeering")
		os.Exit(1)
	}

	if err = cloudresourcescontroller.SetupGcpVpcPeeringReconciler(skrRegistry); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GcpVpcPeering")
		os.Exit(1)
	}

	if err = cloudresourcescontroller.SetupGcpNfsBackupScheduleReconciler(skrRegistry, env); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GcpNfsBackupSchedule")
		os.Exit(1)
	}

	if err = cloudresourcescontroller.SetupCceeNfsVolumeReconciler(skrRegistry); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CceeNfsVolume")
		os.Exit(1)
	}

	if err = cloudresourcescontroller.SetupAwsNfsVolumeBackupReconciler(skrRegistry, awsnfsvolumebackupclient.NewClientProvider(), env); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AwsNfsVolumeBackup")
		os.Exit(1)
	}

	if err = cloudresourcescontroller.SetupAwsNfsBackupScheduleReconciler(skrRegistry, env); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AwsNfsBackupSchedule")
		os.Exit(1)
	}

	if err = cloudresourcescontroller.SetupAwsNfsVolumeRestoreReconciler(skrRegistry, awsnfsvolumerestoreclient.NewClientProvider(), env); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AwsNfsVolumeRestore")
		os.Exit(1)
	}

	//if err = cloudresourcescontroller.SetupAzureRwxBackupReconciler(skrRegistry, azurerwxvolumebackupclient.NewClientProvider()); err != nil {
	//	setupLog.Error(err, "unable to create controller", "controller", "AzureRwxVolumeBackup")
	//	os.Exit(1)
	//}

	if err = cloudresourcescontroller.SetupAzureRwxRestoreReconciler(skrRegistry, azurerwxvolumebackupclient.NewClientProvider()); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AzureRwxVolumeRestore")
		os.Exit(1)
	}

	if err = cloudresourcescontroller.SetupAzureRwxBackupScheduleReconciler(skrRegistry, env); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AzureRwxBackupSchedule")
		os.Exit(1)
	}

	if err = cloudresourcescontroller.SetupAzureRwxPvReconciler(skrRegistry, azurerwxpvclient.NewClientProvider()); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AzureRwxPV")
		os.Exit(1)
	}

	if err = cloudresourcescontroller.SetupGcpSubnetReconciler(skrRegistry); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GcpSubnet")
		os.Exit(1)
	}

	// KCP Controllers
	if err = cloudcontrolcontroller.SetupScopeReconciler(
		ctx,
		mgr,
		scopeclient.NewAwsStsGardenClientProvider(),
		activeSkrCollection,
		gcpclient.NewServiceUsageClientProvider(),
		awsexposeddataclient.NewClientProvider(),
		azureexposeddataclient.NewClientProvider(),
	); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Scope")
		os.Exit(1)
	}
	if err = cloudcontrolcontroller.SetupKymaReconciler(mgr, activeSkrCollection); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Kyma")
		os.Exit(1)
	}
	if err = cloudcontrolcontroller.SetupNfsInstanceReconciler(
		mgr,
		awsnfsinstanceclient.NewClientProvider(),
		gcpnfsinstanceclient.NewFilestoreClientProvider(),
		cceenfsinstanceclient.NewClientProvider(),
		env,
	); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NfsInstance")
		os.Exit(1)
	}
	if err = cloudcontrolcontroller.SetupVpcPeeringReconciler(
		mgr,
		awsvpcpeeringclient.NewClientProvider(),
		azurevpcpeeringclient.NewClientProvider(),
		gcpvpcpeeringclient.NewClientProvider(),
		env,
	); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "VpcPeering")
		os.Exit(1)
	}
	if err = cloudcontrolcontroller.SetupIpRangeReconciler(
		ctx,
		mgr,
		awsiprangeclient.NewClientProvider(),
		azureiprangeclient.NewClientProvider(),
		gcpiprangeclient.NewServiceNetworkingClient(),
		gcpiprangeclient.NewComputeClient(),
		env,
	); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "IpRange")
		os.Exit(1)
	}
	if err = cloudcontrolcontroller.SetupRedisInstanceReconciler(
		mgr,
		gcpredisinstanceclient.NewMemorystoreClientProvider(),
		azureredisinstanceclient.NewClientProvider(),
		awsclient.NewElastiCacheClientProvider(),
		env,
	); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RedisInstance")
		os.Exit(1)
	}
	if err = cloudcontrolcontroller.SetupNetworkReconciler(
		ctx,
		mgr,
		azurenetworkclient.NewClientProvider(),
	); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Network")
		os.Exit(1)
	}
	if err = cloudcontrolcontroller.SetupNukeReconciler(
		mgr,
		activeSkrCollection,
		gcpnfsbackupclient.NewFileBackupClientProvider(),
		awsnukeclient.NewClientProvider(),
		azurenukeclient.NewClientProvider(),
		env,
	); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Nuke")
		os.Exit(1)
	}
	if err = cloudcontrolcontroller.SetupRedisClusterReconciler(
		mgr,
		awsclient.NewElastiCacheClientProvider(),
		azureredisclusterclient.NewClientProvider(),
		env,
	); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RedisCluster")
		os.Exit(1)
	}
	if err = cloudcontrolcontroller.SetupGcpRedisClusterReconciler(
		mgr,
		gcpredisclusterclient.NewMemorystoreClientProvider(),
		env,
	); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GcpRedisCluster")
		os.Exit(1)
	}
	if err = cloudcontrolcontroller.SetupGcpSubnetReconciler(
		ctx,
		mgr,
		gcpsubnetclient.NewComputeClientProvider(),
		gcpsubnetclient.NewNetworkConnectivityClientProvider(),
		env,
	); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GcpSubnet")
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

	err = mgr.Add(skrLoop)
	if err != nil {
		setupLog.Error(err, "error adding SkrLooper to KCP manager")
		os.Exit(1)
	}

	setupLog.Info("starting manager")

	if err := feature.Initialize(ctx, rootLogger.WithName("ff")); err != nil {
		setupLog.Error(err, "problem initializing feature flags")
	}

	go func() {
		err := cfg.Watch(ctx.Done(), func(_ fsnotify.Event) {
			rootLogger.Info("Reloading config")
			cfg.Read()
			rootLogger.WithValues("config", cfg.PrintJson()).
				Info("Config reload dump")
		})
		if err != nil {
			rootLogger.Error(err, "Error from config watcher")
		}
	}()

	// TODO: Remove in next release - after 1.2.5 is released, aka in the 1.2.6
	// Finalizer migration
	func() {
		migLogger := setupLog.WithName("kcpFinalizerMigration")
		mig := migrateFinalizers.NewMigrationForKcp(mgr.GetAPIReader(), mgr.GetClient(), migLogger)
		_, err := mig.Run(ctx)
		if err != nil {
			migLogger.Error(err, "error running KCP finalizer migration")
		}
	}()

	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func loadConfig() config.Config {
	env := abstractions.NewOSEnvironment()
	configDir := env.Get("CONFIG_DIR")
	if len(configDir) < 1 {
		configDir = "./config/config"
	}
	cfg := config.NewConfig(env)
	cfg.BaseDir(configDir)

	awsconfig.InitConfig(cfg)
	azureconfig.InitConfig(cfg)
	cceeconfig.InitConfig(cfg)
	quota.InitConfig(cfg)
	skrruntimeconfig.InitConfig(cfg)
	scope.InitConfig(cfg)
	gcpclient.InitConfig(cfg)
	vpcpeeringconfig.InitConfig(cfg)

	cfg.Read()

	return cfg
}
