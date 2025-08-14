package subscription

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func statusSaveOnCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	state.ObjAsSubscription().Status.Provider = state.provider

	var theErr error

	switch state.provider {
	case cloudcontrolv1beta1.ProviderGCP:
		project, ok := state.credentialData["project_id"]
		if !ok {
			theErr = multierror.Append(theErr, errors.New("gardener credential for gcp missing project_id key"))
		}
		if theErr != nil {
			break
		}
		state.ObjAsSubscription().Status.SubscriptionInfo = &cloudcontrolv1beta1.SubscriptionInfo{
			Gcp: &cloudcontrolv1beta1.SubscriptionInfoGcp{
				Project: project,
			},
		}

	case cloudcontrolv1beta1.ProviderAzure:
		subscriptionID, ok := state.credentialData["subscriptionID"]
		if !ok {
			theErr = multierror.Append(theErr, errors.New("gardener credentials for azure missing subscriptionID key"))
		}
		tenantID, ok := state.credentialData["tenantID"]
		if !ok {
			theErr = multierror.Append(theErr, errors.New("gardener credentials for azure missing tenantID key"))
		}
		if theErr != nil {
			break
		}
		state.ObjAsSubscription().Status.SubscriptionInfo = &cloudcontrolv1beta1.SubscriptionInfo{
			Azure: &cloudcontrolv1beta1.SubscriptionInfoAzure{
				TenantId:       tenantID,
				SubscriptionId: subscriptionID,
			},
		}

	case cloudcontrolv1beta1.ProviderAws:
		accessKeyID, ok := state.credentialData["accessKeyID"]
		if !ok {
			theErr = multierror.Append(theErr, errors.New("gardener credentials for aws missing accessKeyID key"))
		}
		secretAccessKey, ok := state.credentialData["secretAccessKey"]
		if !ok {
			theErr = multierror.Append(theErr, errors.New("gardener credentials for aws missing secretAccessKey key"))
		}

		if theErr != nil {
			break
		}

		stsClient, err := state.awsStsClientProvider(
			ctx,
			"us-east-1", // should not be important since IAM and STS are global
			accessKeyID,
			secretAccessKey,
		)
		if err != nil {
			theErr = multierror.Append(theErr, fmt.Errorf("error creating aws sts client: %w", err))
			break
		}

		callerIdentity, err := stsClient.GetCallerIdentity(ctx)
		if err != nil {
			theErr = multierror.Append(theErr, fmt.Errorf("error getting caller identity: %w", err))
			break
		}

		state.ObjAsSubscription().Status.SubscriptionInfo = &cloudcontrolv1beta1.SubscriptionInfo{
			Aws: &cloudcontrolv1beta1.SubscriptionInfoAws{
				Account: ptr.Deref(callerIdentity.Account, ""),
			},
		}

	case cloudcontrolv1beta1.ProviderOpenStack:
		domainName, ok := state.credentialData["domainName"]
		if !ok {
			theErr = multierror.Append(theErr, errors.New("gardener credentials for openstack missing domainName key"))
		}
		tenantName, ok := state.credentialData["tenantName"]
		if !ok {
			theErr = multierror.Append(theErr, errors.New("gardener credentials for openstack missing tenantName key"))
		}
		if theErr != nil {
			break
		}
		state.ObjAsSubscription().Status.SubscriptionInfo = &cloudcontrolv1beta1.SubscriptionInfo{
			OpenStack: &cloudcontrolv1beta1.SubscriptionInfoOpenStack{
				DomainName: domainName,
				TenantName: tenantName,
			},
		}
	} // case

	if theErr != nil {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(theErr, "error processing gardener cloud credentials")
		state.ObjAsSubscription().Status.State = cloudcontrolv1beta1.StateError

		return composed.PatchStatus(state.ObjAsSubscription()).
			SetExclusiveConditions(metav1.Condition{
				Type:               cloudcontrolv1beta1.ConditionTypeError,
				Status:             metav1.ConditionTrue,
				ObservedGeneration: state.ObjAsSubscription().Generation,
				Reason:             cloudcontrolv1beta1.ReasonGcpError,
				Message:            theErr.Error(),
			}).
			ErrorLogMessage("Error patching status for Subscription with error processing credentials").
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	state.ObjAsSubscription().Status.State = cloudcontrolv1beta1.StateReady
	return composed.PatchStatus(state.ObjAsSubscription()).
		SetExclusiveConditions(metav1.Condition{
			Type:               cloudcontrolv1beta1.ConditionTypeReady,
			Status:             metav1.ConditionTrue,
			ObservedGeneration: state.ObjAsSubscription().Generation,
			Reason:             cloudcontrolv1beta1.ConditionTypeReady,
			Message:            "Ready",
		}).
		ErrorLogMessage("Error patching status for Subscription with error processing credentials").
		SuccessErrorNil().
		Run(ctx, state)
}
