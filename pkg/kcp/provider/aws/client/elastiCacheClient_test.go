package client

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	secretsmanagertypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

type secretsManagerApiFake struct {
	getErr    error
	createErr error
	deleteErr error
}

func (f *secretsManagerApiFake) GetSecretValue(_ context.Context, _ *secretsmanager.GetSecretValueInput, _ ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	return &secretsmanager.GetSecretValueOutput{}, f.getErr
}

func (f *secretsManagerApiFake) CreateSecret(_ context.Context, _ *secretsmanager.CreateSecretInput, _ ...func(*secretsmanager.Options)) (*secretsmanager.CreateSecretOutput, error) {
	return &secretsmanager.CreateSecretOutput{}, f.createErr
}

func (f *secretsManagerApiFake) DeleteSecret(_ context.Context, _ *secretsmanager.DeleteSecretInput, _ ...func(*secretsmanager.Options)) (*secretsmanager.DeleteSecretOutput, error) {
	return &secretsmanager.DeleteSecretOutput{}, f.deleteErr
}

func scheduledForDeletionErr() error {
	return &secretsmanagertypes.InvalidRequestException{
		Message: ptr.To("You can't perform this operation on the secret because it was marked for deletion."),
	}
}

func TestGetAuthTokenSecretValue(t *testing.T) {
	t.Run("returns value on success", func(t *testing.T) {
		c := &elastiCacheClient{secretsManagerSvc: &secretsManagerApiFake{}}
		out, err := c.GetAuthTokenSecretValue(context.Background(), "name")
		assert.NoError(t, err)
		assert.NotNil(t, out)
	})

	t.Run("treats not found as absent", func(t *testing.T) {
		c := &elastiCacheClient{secretsManagerSvc: &secretsManagerApiFake{
			getErr: &smithy.GenericAPIError{Code: (&secretsmanagertypes.ResourceNotFoundException{}).ErrorCode()},
		}}
		out, err := c.GetAuthTokenSecretValue(context.Background(), "name")
		assert.NoError(t, err)
		assert.Nil(t, out)
	})

	t.Run("treats scheduled for deletion as absent", func(t *testing.T) {
		c := &elastiCacheClient{secretsManagerSvc: &secretsManagerApiFake{getErr: scheduledForDeletionErr()}}
		out, err := c.GetAuthTokenSecretValue(context.Background(), "name")
		assert.NoError(t, err)
		assert.Nil(t, out)
	})

	t.Run("propagates unrelated InvalidRequestException", func(t *testing.T) {
		getErr := &secretsmanagertypes.InvalidRequestException{Message: ptr.To("some other invalid request")}
		c := &elastiCacheClient{secretsManagerSvc: &secretsManagerApiFake{getErr: getErr}}
		_, err := c.GetAuthTokenSecretValue(context.Background(), "name")
		assert.ErrorIs(t, err, getErr)
	})

	t.Run("propagates other errors", func(t *testing.T) {
		getErr := errors.New("some other error")
		c := &elastiCacheClient{secretsManagerSvc: &secretsManagerApiFake{getErr: getErr}}
		_, err := c.GetAuthTokenSecretValue(context.Background(), "name")
		assert.ErrorIs(t, err, getErr)
	})
}

func TestCreateAuthTokenSecret(t *testing.T) {
	t.Run("returns nil on success", func(t *testing.T) {
		c := &elastiCacheClient{secretsManagerSvc: &secretsManagerApiFake{}}
		assert.NoError(t, c.CreateAuthTokenSecret(context.Background(), "name", nil))
	})

	t.Run("swallows ResourceExistsException", func(t *testing.T) {
		c := &elastiCacheClient{secretsManagerSvc: &secretsManagerApiFake{
			createErr: &secretsmanagertypes.ResourceExistsException{},
		}}
		assert.NoError(t, c.CreateAuthTokenSecret(context.Background(), "name", nil))
	})

	t.Run("propagates other errors", func(t *testing.T) {
		createErr := errors.New("some other error")
		c := &elastiCacheClient{secretsManagerSvc: &secretsManagerApiFake{createErr: createErr}}
		assert.ErrorIs(t, c.CreateAuthTokenSecret(context.Background(), "name", nil), createErr)
	})
}

func TestDeleteAuthTokenSecret(t *testing.T) {
	t.Run("returns nil on success", func(t *testing.T) {
		c := &elastiCacheClient{secretsManagerSvc: &secretsManagerApiFake{}}
		assert.NoError(t, c.DeleteAuthTokenSecret(context.Background(), "name"))
	})

	t.Run("swallows scheduled for deletion", func(t *testing.T) {
		c := &elastiCacheClient{secretsManagerSvc: &secretsManagerApiFake{deleteErr: scheduledForDeletionErr()}}
		assert.NoError(t, c.DeleteAuthTokenSecret(context.Background(), "name"))
	})

	t.Run("swallows not found", func(t *testing.T) {
		c := &elastiCacheClient{secretsManagerSvc: &secretsManagerApiFake{
			deleteErr: &smithy.GenericAPIError{Code: (&secretsmanagertypes.ResourceNotFoundException{}).ErrorCode()},
		}}
		assert.NoError(t, c.DeleteAuthTokenSecret(context.Background(), "name"))
	})

	t.Run("propagates other errors", func(t *testing.T) {
		deleteErr := errors.New("some other error")
		c := &elastiCacheClient{secretsManagerSvc: &secretsManagerApiFake{deleteErr: deleteErr}}
		assert.ErrorIs(t, c.DeleteAuthTokenSecret(context.Background(), "name"), deleteErr)
	})
}
