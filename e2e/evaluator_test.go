package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEvaluator(t *testing.T) {

	t.Run("simple one resource no deps", func(t *testing.T) {
		handleOne := newClusterEvaluationHandleFake("one")
		handleOne.declare("cmOne", "cmOne", "")
		handleOne.setObj("cmOne", &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: "cmOne",
			},
			Data: map[string]string{
				"alias": "cmOne",
			},
		})

		evaluator, err := NewEvaluatorBuilder().
			Add(handleOne).
			Build(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, evaluator)

		assert.True(t, handleOne.AllResources()[0].Evaluated)

		var s string
		var v interface{}

		s, err = evaluator.EvalTemplate("${cmOne.metadata.name}")
		assert.NoError(t, err)
		assert.Equal(t, "cmOne", s)

		v, err = evaluator.Eval("cmOne.metadata.name")
		assert.NoError(t, err)
		assert.Equal(t, "cmOne", v)
	})

	t.Run("complex with deps", func(t *testing.T) {
		handleOne := newClusterEvaluationHandleFake("one")
		handleOne.declare("cmOne", "cmOne", "")
		handleOne.declare("cmTwo", "${cmOne.data.cmTwoName}", "")

		var evaluator Evaluator
		var err error

		evaluator, err = NewEvaluatorBuilder().
			Add(handleOne).
			Build(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, evaluator)

		assert.True(t, handleOne.GetResource("cmOne").Evaluated)
		assert.False(t, handleOne.GetResource("cmTwo").Evaluated)

		var s string
		var v interface{}

		s, err = evaluator.EvalTemplate("${cmOne.metadata.name}")
		assert.NoError(t, err)
		assert.Equal(t, "", s)

		v, err = evaluator.Eval("cmOne.metadata.name")
		assert.NoError(t, err)
		assert.Nil(t, v)

		// when cmOne is set
		handleOne.setObj("cmOne", &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: "cmOne",
			},
			Data: map[string]string{
				"cmTwoName": "cmTwo",
			},
		})

		// then both items are evaluated

		evaluator, err = NewEvaluatorBuilder().
			Add(handleOne).
			Build(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, evaluator)

		assert.True(t, handleOne.GetResource("cmOne").Evaluated)
		assert.True(t, handleOne.GetResource("cmTwo").Evaluated)

		s, err = evaluator.EvalTemplate("${cmOne.metadata.name}")
		assert.NoError(t, err)
		assert.Equal(t, "cmOne", s)

		v, err = evaluator.Eval("cmOne.metadata.name")
		assert.NoError(t, err)
		assert.Equal(t, "cmOne", v)

	})
}
