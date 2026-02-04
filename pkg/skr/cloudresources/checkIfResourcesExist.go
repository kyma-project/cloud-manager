package cloudresources

import (
	"context"
	"fmt"
	"strings"
	"sync"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func checkIfResourcesExist(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// Collect GVKs to check first (synchronous, fast in-memory operations)
	type gvkToCheck struct {
		gvk     schema.GroupVersionKind
		listGvk schema.GroupVersionKind
	}
	var gvksToCheck []gvkToCheck

	scheme := state.Cluster().Scheme()

	for gvk := range scheme.AllKnownTypes() {
		if gvk.Group != cloudresourcesv1beta1.GroupVersion.Group {
			continue
		}
		if gvk.Kind == "CloudResources" {
			continue
		}
		if strings.HasSuffix(gvk.Kind, "List") {
			continue
		}

		listGvk := schema.GroupVersionKind{
			Group:   gvk.Group,
			Version: gvk.Version,
			Kind:    gvk.Kind + "List",
		}
		if !scheme.Recognizes(listGvk) {
			continue
		}

		// Validate that we can create a list object and it implements client.ObjectList
		listObj, err := scheme.New(listGvk)
		if runtime.IsNotRegisteredError(err) {
			continue
		}
		if err != nil {
			logger.
				WithValues(
					"errorType", fmt.Errorf("%T", err),
					"gvk", listGvk.String(),
				).
				Error(err, "Error instantiating GVK list object")
			continue
		}
		if _, ok := listObj.(client.ObjectList); !ok {
			logger.
				WithValues("gvk", listGvk.String()).
				Info("List object does not implement client.ObjectList")
			continue
		}

		gvksToCheck = append(gvksToCheck, gvkToCheck{gvk: gvk, listGvk: listGvk})
	}

	// Limit concurrent API calls to avoid overwhelming the API server
	const maxConcurrency = 10
	semaphore := make(chan struct{}, maxConcurrency)

	var foundKinds []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Get the client once before spawning goroutines
	k8sClient := state.Cluster().K8sClient()

	for _, item := range gvksToCheck {
		wg.Add(1)
		go func(gvk, listGvk schema.GroupVersionKind) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				return
			}

			// Create a fresh list object for this goroutine
			listObj, err := scheme.New(listGvk)
			if err != nil {
				logger.
					WithValues(
						"errorType", fmt.Sprintf("%T", err),
						"gvk", listGvk.String(),
					).
					Error(err, "Error creating list object in goroutine")
				return
			}
			list := listObj.(client.ObjectList) // Safe: validated during gvksToCheck build

			err = k8sClient.List(ctx, list)
			if meta.IsNoMatchError(err) {
				// this CRD is not installed
				return
			}
			if err != nil {
				logger.
					WithValues(
						"errorType", fmt.Sprintf("%T", err),
						"gvk", gvk.String(),
						"listGvk", listGvk.String(),
					).
					Error(err, "Error listing GVK")
				return
			}

			if meta.LenList(list) == 0 {
				return
			}

			mu.Lock()
			foundKinds = append(foundKinds, gvk.Kind)
			mu.Unlock()
		}(item.gvk, item.listGvk)
	}

	wg.Wait()

	// If context was cancelled, we may have incomplete results - requeue to try again
	if ctx.Err() != nil {
		return composed.StopWithRequeue, ctx
	}

	if len(foundKinds) == 0 {
		return nil, nil
	}

	logger.
		WithValues("existingResourceKinds", foundKinds).
		Info("Can not deactivate module due to found resources")

	state.ObjAsCloudResources().Status.State = cloudresourcesv1beta1.ModuleState(util.KymaModuleStateWarning)

	return composed.UpdateStatus(state.ObjAsCloudResources()).
		RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeWarning,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonResourcesExist,
			Message: fmt.Sprintf("Can not deactivate module while cloud resources exist: %v", foundKinds),
		}).
		SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
		Run(ctx, state)
}
