package iprange

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/actions/focal"
)

type State struct {
	focal.State

	ec2Client *ec2.Client
}

func NewState(baseState focal.State) *State {
	return &State{
		State: baseState,
	}
}
