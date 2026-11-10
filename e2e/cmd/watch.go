package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/elliotchance/pie/v2"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	e2elib "github.com/kyma-project/cloud-manager/e2e/lib"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/common/watcher"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

const aliasKcp = "kcp"

type cmdWatchOptionsType struct {
	alias      string
	runtimeID  string
	kubeconfig string

	skr bool
	kcp bool

	output      string
	noDiffDedup bool

	timeout time.Duration
}

var cmdWatchOptions cmdWatchOptionsType

var cmdWatch = &cobra.Command{
	Use:   "watch",
	Short: "Watch object changes on a KCP or SKR cluster and stream them as YAML",
	RunE: func(cmd *cobra.Command, args []string) error {
		o := &cmdWatchOptions

		ctx := rootCtx
		cancel := func() {}
		if o.timeout > 0 {
			ctx, cancel = context.WithTimeout(rootCtx, o.timeout)
		}
		defer cancel()

		// --skr/--kcp are only valid with --kubeconfig
		if o.kubeconfig == "" && (o.skr || o.kcp) {
			return fmt.Errorf("--skr/--kcp only valid with --kubeconfig")
		}
		if o.kubeconfig != "" && !(o.skr || o.kcp) {
			return fmt.Errorf("--kubeconfig requires one of --skr or --kcp")
		}

		var mgr manager.Manager
		var watchFn func() error

		switch {
		case o.alias == aliasKcp:
			m, err := e2elib.NewKcpClientFactory(nil).CreateManager(ctx)
			if err != nil {
				return fmt.Errorf("failed to create KCP manager: %w", err)
			}
			mgr = m
			watchFn = nil // set after watcher is created

		case o.alias != "":
			keb, err := e2ekeb.Create(rootCtx, config)
			if err != nil {
				return fmt.Errorf("failed to create keb: %w", err)
			}
			arr, err := keb.List(rootCtx, e2ekeb.WithAlias(o.alias))
			if err != nil {
				return fmt.Errorf("failed to list instances: %w", err)
			}
			if len(arr) == 0 {
				return fmt.Errorf("instance %q not found", o.alias)
			}
			if len(arr) > 1 {
				return fmt.Errorf("more than one instance found for alias %q: %v", o.alias, pie.Map(arr, func(x e2ekeb.InstanceDetails) string {
					return x.RuntimeID
				}))
			}
			m, err := keb.CreateSkrManager(ctx, arr[0].RuntimeID, e2ekeb.WithLogger(rootLogger))
			if err != nil {
				return fmt.Errorf("failed to create SKR manager: %w", err)
			}
			mgr = m
			watchFn = nil // set after watcher is created

		case o.runtimeID != "":
			keb, err := e2ekeb.Create(rootCtx, config)
			if err != nil {
				return fmt.Errorf("failed to create keb: %w", err)
			}
			m, err := keb.CreateSkrManager(ctx, o.runtimeID, e2ekeb.WithLogger(rootLogger))
			if err != nil {
				return fmt.Errorf("failed to create SKR manager: %w", err)
			}
			mgr = m
			watchFn = nil // set after watcher is created

		case o.kubeconfig != "":
			b, err := os.ReadFile(o.kubeconfig)
			if err != nil {
				return fmt.Errorf("failed to read kubeconfig: %w", err)
			}
			restConfig, err := clientcmd.RESTConfigFromKubeConfig(b)
			if err != nil {
				return fmt.Errorf("failed to parse kubeconfig: %w", err)
			}
			scheme := commonscheme.KcpScheme
			if o.skr {
				scheme = commonscheme.SkrScheme
			}
			m, err := manager.New(restConfig, manager.Options{
				Scheme: scheme,
				Metrics: metricsserver.Options{
					BindAddress: "0",
				},
				LeaderElection:         false,
				HealthProbeBindAddress: "0",
				Client: client.Options{
					Cache: &client.CacheOptions{
						Unstructured: true,
					},
				},
				Logger: rootLogger,
			})
			if err != nil {
				return fmt.Errorf("failed to create manager: %w", err)
			}
			mgr = m
			watchFn = nil // set after watcher is created
		}

		// Resolve output target
		toStdout := cmd.Flags().Changed("output") && o.output == ""
		var outputPath string
		if !toStdout {
			if cmd.Flags().Changed("output") {
				outputPath = o.output
			} else {
				outputPath = filepath.Join(config.ConfigDir, "tmp", "ctrl-watcher-log-e2e.yaml")
			}
			if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}
		}

		// Build handler
		dedup := !o.noDiffDedup
		var handler watcher.ObjectHandler
		switch {
		case toStdout && dedup:
			handler = watcher.NewDiffObjectHandler(watcher.NewYamlLoggingObjectHandler(watcher.NewStdoutLogWriter()))
		case toStdout && !dedup:
			handler = watcher.NewYamlLoggingObjectHandler(watcher.NewStdoutLogWriter())
		case !toStdout && dedup:
			handler = watcher.NewDefaultObjectHandler(outputPath)
		default:
			handler = watcher.NewYamlLoggingObjectHandler(watcher.NewFileLogWriter(outputPath))
		}

		// Create watcher and bind watchFn
		w := watcher.New(mgr, handler)
		switch {
		case o.alias == aliasKcp:
			watchFn = w.ForKCP
		case o.alias != "" || o.runtimeID != "":
			watchFn = w.ForSKR
		case o.kcp:
			watchFn = w.ForKCP
		default:
			watchFn = w.ForSKR
		}

		if err := watchFn(); err != nil {
			return fmt.Errorf("failed to set up watcher: %w", err)
		}

		go func() {
			if err := mgr.Start(ctx); err != nil {
				rootLogger.Error(err, "manager stopped")
			}
		}()

		cacheCtx, cacheCancel := context.WithTimeout(ctx, 30*time.Second)
		defer cacheCancel()
		if ok := mgr.GetCache().WaitForCacheSync(cacheCtx); !ok {
			return fmt.Errorf("error waiting for cache to sync: %w", cacheCtx.Err())
		}

		if toStdout {
			fmt.Fprintf(os.Stderr, "Watching... (press Ctrl+C to stop)\n")
		} else {
			fmt.Fprintf(os.Stderr, "Watching... writing to %s (press Ctrl+C to stop)\n", outputPath)
		}

		<-ctx.Done()
		return nil
	},
}

func init() {
	cmdRoot.AddCommand(cmdWatch)

	o := &cmdWatchOptions
	cmdWatch.Flags().StringVarP(&o.alias, "alias", "a", "", "e2e instance alias; 'kcp' selects the KCP cluster")
	cmdWatch.Flags().StringVarP(&o.runtimeID, "runtime-id", "r", "", "e2e SKR instance runtime id")
	cmdWatch.Flags().StringVar(&o.kubeconfig, "kubeconfig", "", "path to a kubeconfig for an arbitrary KCP or SKR cluster")
	cmdWatch.Flags().BoolVar(&o.skr, "skr", false, "with --kubeconfig: use SKR scheme and watch SKR resources")
	cmdWatch.Flags().BoolVar(&o.kcp, "kcp", false, "with --kubeconfig: use KCP scheme and watch KCP resources")
	cmdWatch.Flags().StringVarP(&o.output, "output", "o", "", "log YAML filename; empty string => stdout")
	cmdWatch.Flags().BoolVar(&o.noDiffDedup, "no-diff-dedup", false, "disable diff deduplication (enabled by default)")
	cmdWatch.Flags().DurationVarP(&o.timeout, "timeout", "t", 0, "run for this duration then quit (0 = until signal)")

	cmdWatch.MarkFlagsMutuallyExclusive("alias", "runtime-id", "kubeconfig")
	cmdWatch.MarkFlagsOneRequired("alias", "runtime-id", "kubeconfig")
	cmdWatch.MarkFlagsMutuallyExclusive("skr", "kcp")
}
