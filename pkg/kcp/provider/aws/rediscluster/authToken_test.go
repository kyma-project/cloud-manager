package rediscluster

import (
	"context"
	"errors"
	"testing"

	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/meta"
)

func TestLoadAuthTokenSecret_NilValueContinues(t *testing.T) {
	stub := newElastiCacheStub() // getAuthTokenValue and getAuthTokenErr both nil
	state := newTestState(t, "abc", true, &elasticachetypes.ReplicationGroup{}, nil)
	state.awsClient = stub

	err, _ := loadAuthTokenSecret(context.Background(), state)

	assert.Nil(t, err, "a nil token value is a normal eventual-consistency window; must continue the pipeline")
	assert.Nil(t, state.authTokenValue, "authTokenValue must stay nil when AWS returns no value")
}

func TestLoadAuthTokenSecret_ErrorRequeues(t *testing.T) {
	stub := newElastiCacheStub()
	stub.getAuthTokenErr = errors.New("get boom")
	state := newTestState(t, "abc", true, &elasticachetypes.ReplicationGroup{}, nil)
	state.awsClient = stub

	err, _ := loadAuthTokenSecret(context.Background(), state)

	assert.NotNil(t, err, "a GetAuthTokenSecretValue error must requeue")
	assert.NotEqual(t, composed.StopAndForget, err, "must requeue, not terminate")
}

func TestCreateAuthTokenSecret_NonRetryableErrorSetsStateError(t *testing.T) {
	stub := newElastiCacheStub()
	stub.createAuthTokenErr = errors.New("create boom") // generic error is non-retryable
	state := newTestState(t, "abc", true, &elasticachetypes.ReplicationGroup{}, nil)
	state.awsClient = stub

	err, _ := createAuthTokenSecret(context.Background(), state)

	assert.NotNil(t, err)
	assert.NotEqual(t, composed.StopAndForget, err, "a non-retryable create error must requeue, not terminate")
	assert.Len(t, stub.createAuthTokenCalls, 1, "AWS CreateAuthTokenSecret must be attempted")
	assert.Equal(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisCluster().Status.State)
	cond := meta.FindStatusCondition(state.ObjAsRedisCluster().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	assert.NotNil(t, cond, "Error condition must be set on non-retryable create failure")
	assert.Equal(t, cloudcontrolv1beta1.ReasonCloudProviderError, cond.Reason)
}

func TestCreateAuthTokenSecret_AuthDisabledSkips(t *testing.T) {
	stub := newElastiCacheStub()
	state := newTestState(t, "abc", false, &elasticachetypes.ReplicationGroup{}, nil)
	state.awsClient = stub

	err, _ := createAuthTokenSecret(context.Background(), state)

	assert.Nil(t, err, "must skip (nil) when AuthEnabled is false")
	assert.Empty(t, stub.createAuthTokenCalls, "must not call AWS when auth is disabled")
}

// retryableError implements the aws-sdk retry package's RetryableError() bool
// interface, which retry.NewStandard().IsErrorRetryable (used by
// awsmeta.IsErrorRetryable) recognizes directly — a robust way to exercise the
// retryable branch without depending on specific AWS error codes.
type retryableError struct{}

func (retryableError) Error() string        { return "transient AWS error" }
func (retryableError) RetryableError() bool { return true }

func TestCreateAuthTokenSecret_RetryableErrorRequeuesWithoutStateError(t *testing.T) {
	stub := newElastiCacheStub()
	stub.createAuthTokenErr = retryableError{}
	state := newTestState(t, "abc", true, &elasticachetypes.ReplicationGroup{}, nil)
	state.awsClient = stub

	err, _ := createAuthTokenSecret(context.Background(), state)

	assert.Len(t, stub.createAuthTokenCalls, 1, "AWS CreateAuthTokenSecret must be attempted")
	assert.NotNil(t, err, "a retryable error must requeue")
	assert.NotEqual(t, composed.StopAndForget, err, "a retryable error must requeue, not terminate")
	assert.NotEqual(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisCluster().Status.State, "a retryable error must NOT set StateError")
	assert.Nil(t, meta.FindStatusCondition(state.ObjAsRedisCluster().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError), "a retryable error must not set the Error condition")
}
