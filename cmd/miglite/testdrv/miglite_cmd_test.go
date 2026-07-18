package testdrv

import (
	"os"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
)

func TestCmd_init(t *testing.T) {
	t.Cleanup(func() { _ = os.Remove("sqlite_test.db") })
	err := app.RunWithArgs([]string{"init"})
	assert.Nil(t, err)
}
