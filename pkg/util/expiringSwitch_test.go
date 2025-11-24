package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ExpiringSwitch(t *testing.T) {

	callCountA := 0
	callCountB := 0

	ExpiringSwitch().
		Key("shoot", "a123").
		IfNotRecently(func() {
			callCountA++
		})
	ExpiringSwitch().
		Key("shoot", "a123").
		IfNotRecently(func() {
			callCountA++
		})

	ExpiringSwitch().
		Key("something", "c22").
		IfNotRecently(func() {
			callCountB++
		})
	ExpiringSwitch().
		Key("something", "c22").
		IfNotRecently(func() {
			callCountB++
		})
	ExpiringSwitch().
		Key("something", "c22").
		IfNotRecently(func() {
			callCountB++
		})

	assert.Equal(t, 1, callCountA)
	assert.Equal(t, 1, callCountB)
}
