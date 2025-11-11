package integration

import (
	"testing"

	"github.com/gookit/goutil/testutil/assert"
)

func TestCmd_init(t *testing.T) {
	err := app.RunWithArgs([]string{"init"})
	assert.Nil(t, err)
}

