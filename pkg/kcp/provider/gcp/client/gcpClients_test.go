package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GcpClientsCanCloseOnNilClients(t *testing.T) {
	gcpClients :=  &GcpClients{}
	err := gcpClients.Close()
	assert.NoError(t, err)
}
