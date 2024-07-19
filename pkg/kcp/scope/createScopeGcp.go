package scope

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elliotchance/pie/v2"
	gardenerv1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func createScopeGcp(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	js, ok := state.credentialData["serviceaccount.json"]
	if !ok {
		err := errors.New("gardener credential for gcp missing serviceaccount.json key")
		logger.Error(err, "error defining GCP scope")
		return composed.StopAndForget, nil // no requeue
	}

	var data map[string]string
	err := json.Unmarshal([]byte(js), &data)
	if err != nil {
		err := fmt.Errorf("error decoding serviceaccount.json: %w", err)
		logger.Error(err, "error defining GCP scope")
		return composed.StopAndForget, nil // no requeue
	}

	project, ok := data["project_id"]
	if !ok {
		err := errors.New("gardener gcp credentials missing project_id")
		logger.Error(err, "error defining GCP scope")
		return composed.StopAndForget, nil // no requeue
	}

	// just create the scope with GCP specifics, the ensureScopeCommonFields will set common values
	scope := &cloudcontrolv1beta1.Scope{
		Spec: cloudcontrolv1beta1.ScopeSpec{
			Scope: cloudcontrolv1beta1.ScopeInfo{
				Gcp: &cloudcontrolv1beta1.GcpScope{
					Project:    project,
					VpcNetwork: commonVpcName(state.shootNamespace, state.shootName),
					Network: cloudcontrolv1beta1.GcpNetwork{
						Nodes:    ptr.Deref(state.shoot.Spec.Networking.Nodes, ""),
						Pods:     ptr.Deref(state.shoot.Spec.Networking.Pods, ""),
						Services: ptr.Deref(state.shoot.Spec.Networking.Services, ""),
					},
					Workers: pie.Map(state.shoot.Spec.Provider.Workers, func(w gardenerv1beta1.Worker) cloudcontrolv1beta1.GcpWorkers {
						return cloudcontrolv1beta1.GcpWorkers{
							Zones: w.Zones,
						}
					}),
				},
			},
		},
	}

	// Preserve loaded obj resource version before getting overwritten by newly created scope
	if st.Obj() != nil && st.Obj().GetName() != "" {
		scope.ResourceVersion = st.Obj().GetResourceVersion()
	}
	state.SetObj(scope)

	return nil, nil
}
