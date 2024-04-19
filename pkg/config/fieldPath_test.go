package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAppendFieldPath(t *testing.T) {
	fp := NewFiledPath("aaa", "bbb")
	assert.Len(t, fp, 2)
	assert.Equal(t, 2, fp.Len())
	fp = AppendFieldPath(fp, "ccc")
	assert.Len(t, fp, 3)
	assert.Equal(t, 3, fp.Len())
	assert.Equal(t, "aaa.bbb.ccc", fp.String())
}

func TestPrependFieldPath(t *testing.T) {
	fp := NewFiledPath("aaa", "bbb")
	assert.Len(t, fp, 2)
	fp = PrependFieldPath(fp, "ccc")
	assert.Len(t, fp, 3)
	assert.Equal(t, "ccc.aaa.bbb", fp.String())
}

func TestFieldPathStringEscapesFieldsContainingDot(t *testing.T) {
	fp := NewFiledPath("metadata", "annotations", "cloud-manager.kyma-project.io", "foo")
	assert.Equal(t, "metadata.annotations.\"cloud-manager.kyma-project.io\".foo", fp.String())
}
