package awswebacl

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	webAcl := state.ObjAsAwsWebAcl()

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, ctx
	}

	return composed.NewStatusPatcherComposed(webAcl).
		MutateStatus(func(acl *cloudresourcesv1beta1.AwsWebAcl) {
			acl.Status.Arn = ptr.Deref(state.awsWebAcl.ARN, "")
			acl.Status.Capacity = state.awsWebAcl.Capacity
			updateLoggingStatus(acl, state)
			acl.SetStatusReady()
		}).
		OnStatusChanged(
			composed.Log("AwsWebAcl is Ready"),
		).
		Run(ctx, state.Cluster().K8sClient())
}

func updateLoggingStatus(acl *cloudresourcesv1beta1.AwsWebAcl, state *State) {
	// Clear if disabled
	if acl.Spec.LoggingConfiguration == nil || !acl.Spec.LoggingConfiguration.Enabled {
		acl.Status.LoggingStatus = nil
		return
	}

	// Initialize if needed
	if acl.Status.LoggingStatus == nil {
		acl.Status.LoggingStatus = &cloudresourcesv1beta1.AwsWebAclLoggingStatus{}
	}

	// Update from AWS state
	if state.loggingConfig != nil {
		acl.Status.LoggingStatus.Enabled = true
		if len(state.loggingConfig.LogDestinationConfigs) > 0 {
			acl.Status.LoggingStatus.LogDestinationArn = state.loggingConfig.LogDestinationConfigs[0]
		}
		acl.Status.LoggingStatus.ManagedLogGroup = (state.managedLogGroupName != "")
		now := metav1.Now()
		acl.Status.LoggingStatus.LastConfigured = &now
	} else {
		acl.Status.LoggingStatus.Enabled = false
	}
}
