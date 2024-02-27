package gcpnfsvolume

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type pvEventHandlerSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *pvEventHandlerSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *pvEventHandlerSuite) TestIsMatchingPV() {
	//Update the capacity in spec.
	pv := v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pv",
			Namespace: "test",
			Labels: map[string]string{
				v1beta1.LabelNfsVolName: "test-pv",
				v1beta1.LabelNfsVolNS:   "test",
			},
		},
	}

	matching, key := isMatchingPV(&pv)
	assert.True(suite.T(), matching)
	assert.Equal(suite.T(), pv.Namespace, key.Namespace)
	assert.Equal(suite.T(), pv.Name, key.Name)
}

func (suite *pvEventHandlerSuite) TestNotMatchingPV() {
	//Update the capacity in spec.
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test",
		},
	}

	matching, key := isMatchingPV(&pod)
	assert.False(suite.T(), matching)
	assert.Nil(suite.T(), key)
}

func TestPvEventHandler(t *testing.T) {
	suite.Run(t, new(pvEventHandlerSuite))
}
