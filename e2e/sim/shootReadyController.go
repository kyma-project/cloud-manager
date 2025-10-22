package sim

import (
	"context"
	"fmt"

	gardenerapicore "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardenerhelper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewGardenManager(gardenRestConfig *rest.Config, scheme *runtime.Scheme, logger logr.Logger) (manager.Manager, error) {
	mgr, err := manager.New(gardenRestConfig, manager.Options{
		Scheme: scheme,
		Logger: logger.WithName("garden-manager"),
		Metrics: metricsserver.Options{
			BindAddress: "0", // disable
		},
		LeaderElection: false, // disable
		HealthProbeBindAddress: "0", // disable
		Client: client.Options{
			Cache: &client.CacheOptions{
				Unstructured: true,
			},
		},
	})
	return mgr, err
}

// SetupShootReadyController creates a controller that updates
// each Shoot with ready conditions. DO NOT USE on real Garden!!!
// Only to be used in tests with the kind cluster acting as the Garden.
func SetupShootReadyController(gardenManager manager.Manager) error {
	return ctrl.NewControllerManagedBy(gardenManager).
		For(&gardenerapicore.Shoot{}).
		Complete(&defaultShootReadyController{gardenClient: gardenManager.GetClient()})
}

func xxxNewShootReadyController(gardenClient client.Client) (controller.Controller, error) {
	c, err := controller.NewTypedUnmanaged("shootReady", controller.Options{})
	return c, err
}

type defaultShootReadyController struct {
	gardenClient client.Client
}

func (r *defaultShootReadyController) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	shoot := &gardenerapicore.Shoot{}
	err := r.gardenClient.Get(ctx, request.NamespacedName, shoot)
	if client.IgnoreNotFound(err) != nil {
		return reconcile.Result{}, fmt.Errorf("error loading shoot from ShootReadyController: %w", err)
	}
	if err != nil {
		return reconcile.Result{}, nil
	}

	changed := false
	for _, ct := range GardenerConditionTypes {
		cond := gardenerhelper.GetCondition(shoot.Status.Conditions, ct)
		if cond == nil {
			cond = &gardenerapicore.Condition{
				Type:               ct,
				Status:             gardenerapicore.ConditionTrue,
				LastTransitionTime: metav1.Now(),
				LastUpdateTime:     metav1.Now(),
				Reason:             string(ct),
				Message:            string(ct),
			}
			shoot.Status.Conditions = append(shoot.Status.Conditions, *cond)
			changed = true
		}
		if cond.Status != gardenerapicore.ConditionTrue {
			cond.Status = gardenerapicore.ConditionTrue
			cond.LastUpdateTime = metav1.Now()
			cond.Reason = string(cond.Reason)
			cond.Message = string(cond.Message)
			changed = true
		}
	}

	if changed {
		err = r.gardenClient.Update(ctx, shoot)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error updating shoot status with ready conditions: %w", err)
		}
	}

	return reconcile.Result{}, nil
}
