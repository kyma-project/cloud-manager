package azurerwxvolumebackup

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateVault(t *testing.T) {

	t.Run("createBackup", func(t *testing.T) {

		ctx := context.Background()
		backup := setupDefaultBackup()
		state := setupDefaultState(ctx, backup)

		t.Run("happy path vault already exists", func(t *testing.T) {

			// Act
			err, res := createVault(ctx, state)

			// Assert
			assert.Equal(t, nil, err)
			assert.Equal(t, ctx, res)

		})

		t.Run("happy path - vault doesn't exist", func(t *testing.T) {

			backup.Spec.Location = "useast"

			err, res := createVault(ctx, state)

			// Assert
			assert.Equal(t, ctx, res, "should return same context")
			assert.Equal(t, composed.StopAndForget, err, "should expect stop and forget")

			//assert.Equal(t, nil, err)
			//assert.Equal(t, nil, res)

		})

		t.Run("unhappy path", func(t *testing.T) {

			// Arrange
			state.resourceGroupName = "http-error"

			// Act
			err, res := createVault(ctx, state)

			// Assert
			assert.Equal(t, ctx, res, "should return same context")
			assert.Equal(t, composed.StopWithRequeue, err, "should expect stop with requeue")

		})

	})

}
