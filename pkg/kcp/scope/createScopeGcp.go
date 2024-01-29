package scope

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/composed"
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
				},
			},
		},
	}

	state.SetObj(scope)

	return nil, nil
}
