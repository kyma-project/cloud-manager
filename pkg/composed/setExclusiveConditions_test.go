package composed

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

type SetStatusConditionsSuite struct {
	suite.Suite
}

const (
	conditionTypeReady   = "Ready"
	conditionTypeError   = "Error"
	conditionTypeWarning = "Warning"
)

func (me *SetStatusConditionsSuite) TestConditionRemoved() {
	conditions := &[]metav1.Condition{
		{
			Type:   conditionTypeReady,
			Status: metav1.ConditionTrue,
		},
		{
			Type:   conditionTypeError,
			Status: metav1.ConditionTrue,
		},
	}

	assert.True(me.T(), SetExclusiveConditions(conditions, metav1.Condition{
		Type:   conditionTypeReady,
		Status: metav1.ConditionTrue,
	}), "Conditions changes should be detected, but didn't")

	assert.Equal(me.T(), 1, len(*conditions))
}

func (me *SetStatusConditionsSuite) TestConditionAdded() {
	conditions := &[]metav1.Condition{
		{
			Type:   conditionTypeReady,
			Status: metav1.ConditionTrue,
		},
		{
			Type:   conditionTypeError,
			Status: metav1.ConditionTrue,
		},
	}

	assert.True(me.T(), SetExclusiveConditions(conditions,
		[]metav1.Condition{
			{
				Type:   conditionTypeReady,
				Status: metav1.ConditionTrue,
			},
			{
				Type:   conditionTypeError,
				Status: metav1.ConditionTrue,
			},
			{
				Type:   conditionTypeWarning,
				Status: metav1.ConditionTrue,
			},
		}...,
	), "Conditions changes should be detected, but didn't")

	assert.Equal(me.T(), 3, len(*conditions))

}

func (me *SetStatusConditionsSuite) TestConditionChanged() {

	conditions := &[]metav1.Condition{
		{
			Type:   conditionTypeReady,
			Status: metav1.ConditionTrue,
		},
		{
			Type:   conditionTypeError,
			Status: metav1.ConditionTrue,
		},
	}

	assert.True(me.T(), SetExclusiveConditions(conditions,
		[]metav1.Condition{
			{
				Type:   conditionTypeReady,
				Status: metav1.ConditionFalse,
			},
			{
				Type:   conditionTypeError,
				Status: metav1.ConditionTrue,
			},
		}...,
	), "Conditions changes should be detected, but didn't")

	assert.Equal(me.T(), 2, len(*conditions))

}

func TestSetStatusConditionsSuite(t *testing.T) {
	suite.Run(t, new(SetStatusConditionsSuite))
}
