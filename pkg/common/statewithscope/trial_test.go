package statewithscope

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsTrialPredicate(t *testing.T) {
	testData := []struct {
		title    string
		state    composed.State
		expected bool
	}{
		{
			"Focal state with aws plan",
			sb(s().WithBrokerPlan("aws").Build()).BuildFocal(),
			false,
		},
		{
			"Focal state with trial plan",
			sb(s().WithBrokerPlan("trial").Build()).BuildFocal(),
			true,
		},
		{
			"ObjAsScope with azure plan",
			sb(s().WithBrokerPlan("azure").Build()).BuildObjAsScope(),
			false,
		},
		{
			"ObjAsScope with trial plan",
			sb(s().WithBrokerPlan("trial").Build()).BuildObjAsScope(),
			true,
		},
		{
			"Scope with gcp plan",
			sb(s().WithBrokerPlan("gcp").Build()).BuildScope(),
			false,
		},
		{
			"Scope with trial plan",
			sb(s().WithBrokerPlan("trial").Build()).BuildScope(),
			true,
		},
	}

	for _, tt := range testData {
		t.Run(tt.title, func(t *testing.T) {
			ctx := context.Background()
			actual := IsTrialPredicate(ctx, tt.state)
			assert.Equal(t, tt.expected, actual, tt.title)
		})
	}
}
