package unit_test

import (
	"fmt"
	"testing"

	"github.com/gookit/goutil/sysutil"
	"github.com/gookit/goutil/testutil/assert"
	"github.com/gookit/miglite/pkg/command"
)

var app = command.NewApp("miglite", "0.0.1", "Go minimal database migration tool")

func TestMain(m *testing.M) {
	fmt.Println("Workdir:", sysutil.Workdir())
	m.Run()
}

func TestCmd_init(t *testing.T) {
	err := app.RunWithArgs([]string{"init"})
	assert.Nil(t, err)
}

