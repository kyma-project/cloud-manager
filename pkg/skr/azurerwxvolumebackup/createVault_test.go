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

		})

		t.Run("unhappy path - vault doesn't exist and http error", func(t *testing.T) {

			// Arrange
			backup.Spec.Location = "useast"
			kvp := map[string]string{"CreateVault": "fail"}

			newCtx := addValuesToContext(ctx, kvp)

			// Act
			err, res := createVault(newCtx, state)

			// Assert
			assert.Equal(t, newCtx, res, "should return same context")
			assert.Equal(t, composed.StopWithRequeue, err, "should expect stop with requeue")

		})

	})

}
