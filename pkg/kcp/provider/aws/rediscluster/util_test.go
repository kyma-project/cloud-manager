package rediscluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUtil(t *testing.T) {
	t.Run("GetDesiredParameters", func(t *testing.T) {
		tests := []struct {
			name           string
			input          map[string]string
			expectedOutput map[string]string
		}{
			{
				name:  "should always have cluster-enabled set to yes",
				input: map[string]string{},
				expectedOutput: map[string]string{
					"active-defrag-cycle-max": "75",
					"acl-pubsub-default":      "allchannels",
					"cluster-enabled":         "yes",
				},
			},
			{
				name: "should not allow overwrite of cluster-enabled to no",
				input: map[string]string{
					"cluster-enabled": "no",
				},
				expectedOutput: map[string]string{
					"active-defrag-cycle-max": "75",
					"acl-pubsub-default":      "allchannels",
					"cluster-enabled":         "yes",
				},
			},
			{
				name: "should overwrite default param with desired param",
				input: map[string]string{
					"active-defrag-cycle-max": "85",
				},
				expectedOutput: map[string]string{
					"active-defrag-cycle-max": "85",
					"acl-pubsub-default":      "allchannels",
					"cluster-enabled":         "yes",
				},
			},
		}

		defaultParams := map[string]string{
			"active-defrag-cycle-max": "75",
			"acl-pubsub-default":      "allchannels",
			"cluster-enabled":         "no",
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {

				actual := GetDesiredParameters(defaultParams, test.input)

				assert.Equal(t, test.expectedOutput, actual, "actual params do not equal expected params")
			})
		}

	})
}
